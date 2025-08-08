package bookings

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// TODO: employee should be able to reschedule bookings
func CreateBooking(c *fiber.Ctx) error {
	bookingDto := new(CreateBookingDto)

	if err := c.BodyParser(bookingDto); err != nil {
		return c.Status(400).JSON(common.ErrorResponse{StatusCode: 400, Message: err.Error()})
	}
	fmt.Println(bookingDto)

	if err := utils.Validate(bookingDto); err != nil {
		return c.Status(err.Code).JSON(common.ErrorResponse{
			StatusCode: uint(err.Code),
			Message:    err.Message,
		})
	}

	//check if BookingDate is gte today date and lte max date, for exmple next month
	if !time.Now().Before(bookingDto.BookingDate.Local()) {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Booking date and time must be greater than the current ones",
		})
	}

	var service model.Service
	if result := database.DB.Db.First(&service, bookingDto.ServiceID); result.Error != nil {
		return common.NewInternalServerError(c)
	}

	hasService, err := common.EmployeeHasService(database.DB.Db, bookingDto.EmployeeID, bookingDto.ServiceID)
	if err != nil {
		return common.NewInternalServerError(c)
	}

	if !hasService {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Service not assigned to employee",
		})
	}

	//1. check if BookingDate match with a available schedule
	//find schedules by EmployeeID
	daySchedule := []model.Schedule{}
	database.DB.Db.Where("employee_id = ? AND week_day = ?",
		bookingDto.EmployeeID,
		int(bookingDto.BookingDate.Local().Weekday()),
	).Find(&daySchedule)

	if len(daySchedule) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Employee or Employee schedule not found",
		})
	}

	//Interval [BookingStart, BookingEnd] outside interval [scheduleStartTime, scheduleEndTime]
	hasMatch := false
	bookingStart := bookingDto.BookingDate.Local()
	bookingEnd := bookingStart.Add(time.Duration(service.DurationMinutes) * time.Minute)
	for _, schedule := range daySchedule {
		scheduleStart := time.Date(bookingStart.Year(), bookingStart.Month(), bookingStart.Day(), schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, time.Local)
		scheduleEnd := time.Date(bookingStart.Year(), bookingStart.Month(), bookingStart.Day(), schedule.EndTime.Hour(), schedule.EndTime.Minute(), 0, 0, time.Local)
		if (bookingStart.After(scheduleStart) || bookingStart.Equal(scheduleStart)) &&
			(bookingEnd.Before(scheduleEnd) || bookingEnd.Equal(scheduleEnd)) {
			hasMatch = true
			break
		}
	}

	if !hasMatch {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Error bookingTime not match with an available schedule",
		})
	}

	// Check for schedule exeptions
	bookingDate := bookingDto.BookingDate
	exceptions := []model.ScheduleOverride{}
	if result := database.DB.Db.Where("DATE(exception_date) = DATE(?) ", bookingDate).Find(&exceptions); result.Error != nil {
		return common.NewInternalServerError(c)
	}

	// Check if [bookingStart, bookingEnd] overlaps interval [exStartTime, exEndtime]
	// excluding the extremes, for each exception of the day
	isBlocked := false
	for _, ex := range exceptions {
		if ex.AllDay {
			isBlocked = true
			break
		}
		if bookingStart.Before(*ex.EndTime) && bookingEnd.After(*ex.StartTime) {
			isBlocked = true
			break
		}
	}
	if isBlocked {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Error: booking not available.",
		})
	}

	// Check the existing bookings for employee in bookingDate
	bookings := []model.Booking{}
	if result := database.DB.Db.Where("employee_id = ? AND DATE(booking_date) = DATE(?) AND status = ?", bookingDto.EmployeeID, bookingDate, model.SatusBooked.String()).Find(&bookings); result.Error != nil {
		return common.NewInternalServerError(c)
	}

	// Check if [bookingStart, bookingEnd] overlaps interval [existingStart, existingEnd]
	for _, booking := range bookings {
		existingStart := booking.BookingDate
		existingEnd := booking.BookingDate.Add(time.Duration(service.DurationMinutes) * time.Minute)
		if bookingStart.Before(existingEnd) && bookingStart.After(existingStart) {
			isBlocked = true
			break
		}
	}
	if isBlocked {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Error: booking not available.",
		})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))
	fmt.Println("claims", claims)

	booking := model.Booking{
		CustomerID:  userId,
		EmployeeID:  bookingDto.EmployeeID,
		ServiceID:   bookingDto.ServiceID,
		BookingDate: bookingDto.BookingDate,
		Status:      model.SatusBooked.String(),
		Notes:       bookingDto.Notes,
	}

	database.DB.Db.Create(&booking)
	return c.Status(fiber.StatusCreated).JSON(booking)
}

// Use in endpoint: GET /bookins
// Required roles: admin
func FindAllBookings(c *fiber.Ctx) error {
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "page must be an integer number and gt 0",
		})
	}

	offset := (page - 1) * limit
	query := database.DB.Db.Limit(limit).Offset(offset)

	orderBy := c.Query("orderBy")
	allowedFields := map[string]bool{
		"booking_date": true,
		"created_at":   true,
	}
	orderBy, err = common.ParseSortParam(orderBy, allowedFields)
	fmt.Println(orderBy, err)
	if err == nil {
		query = query.Order(orderBy)
	}

	bookings := []model.Booking{} //declaring bookings slice
	query.Find(&bookings)

	var count int64
	database.DB.Db.Model(&model.Booking{}).Count(&count)

	return c.Status(200).JSON(AllBookingsResponse{
		Data:     bookings,
		Metadata: metadata{TotalItems: count},
	})
}

func FindBooking(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Numeric string is expected",
		})
	}

	booking := new(model.Booking)
	restult := database.DB.Db.First(booking, id)
	if restult.Error != nil {
		return c.Status(404).JSON(common.ErrorResponse{
			StatusCode: 404,
			Message:    "Booking not found",
		})
	}

	return c.Status(200).JSON(booking)
}

// Employee or customer gets his bookings.
// Use in endpoint GET /me/bookings
func FindBookingsByUser(c *fiber.Ctx) error {
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "page must be an integer number and gt 0",
		})
	}

	offset := (page - 1) * limit
	orderBy := c.Query("orderBy")

	allowedFields := map[string]bool{
		"booking_date": true,
		"created_at":   true,
	}
	orderBy, err = common.ParseSortParam(orderBy, allowedFields)
	fmt.Println(orderBy, err)

	query := database.DB.Db.Limit(limit).Offset(offset)
	if err == nil {
		query = query.Order(orderBy)
	}

	//get user data from context
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	if role == model.RoleCustomer.String() {
		query = query.Where("customer_id = ?", userId)
	} else {
		query = query.Where("employee_id = ?", userId)
	}

	bookings := []model.Booking{} //declaring bookings slice
	query.Find(&bookings)

	var count int64
	database.DB.Db.Model(&model.Booking{}).Count(&count)

	return c.Status(200).JSON(AllBookingsResponse{
		Data:     bookings,
		Metadata: metadata{TotalItems: count},
	})
}

// Employee or customer can cancel their bookings, use in
// POST /bookings/:id/cancel endpoint
func CancelBooking(c *fiber.Ctx) error {
	//Get booking id
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}
	//Get user id
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	//check if employee or customer has the associated booking
	hasBooking, err := common.UserHasBooking(database.DB.Db, uint(userId), uint(id), role)
	if err != nil {
		return common.NewInternalServerError(c)
	}
	if !hasBooking {
		return c.Status(fiber.StatusForbidden).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusForbidden,
			Message:    "You do not have permission to cancel this reservation",
		})
	}

	var booking model.Booking
	result := database.DB.Db.Model(&booking).Where("id = ?", id).Update("status", model.StatusCancelled.String())
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    "booking not found",
			})
		}
		return common.NewInternalServerError(c)
	}
	return c.Status(fiber.StatusOK).JSON(booking)

}
