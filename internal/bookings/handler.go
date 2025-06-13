package bookings

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// create a booking
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
	if !time.Now().Before(bookingDto.BookingDate) {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Booking date and time must be greater than the current ones",
		})
	}

	//check if BookingDate match with a available schedule
	//find schedules by EmployeeID
	avSchedules := []model.Schedule{}
	database.DB.Db.Where("employee_id = ?", bookingDto.EmployeeID).Find(&avSchedules)
	fmt.Println(avSchedules)

	if len(avSchedules) == 0 {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Employee or Employee schedule not found",
		})
	}

	//if booking weekday is not in avSchedules.weekday return error
	i := slices.IndexFunc(avSchedules, func(schedule model.Schedule) bool {
		return schedule.WeekDay == int(bookingDto.BookingDate.Weekday())
	})
	if i == -1 {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    fmt.Sprintf("Employee schedule not found for the day %s", bookingDto.BookingDate.Weekday().String()),
		})
	}

	//else check the booking time in range of schedules end and start time
	scheduleFound := avSchedules[i]
	bookingTime := bookingDto.BookingDate
	scheduleStartTime := scheduleFound.StartTime
	scheduleEndTime := scheduleFound.EndTime

	//check booking time > schedule StartTime
	if result := utils.CompareTimeOnly(bookingTime, scheduleStartTime); result < 0 {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message: fmt.Sprintf(
				"Booking time must be greater than %d:%d",
				scheduleFound.StartTime.Hour(),
				scheduleFound.StartTime.Minute(),
			),
		})
	}

	//check booking time < schedule EndTime
	if result := utils.CompareTimeOnly(bookingTime, scheduleEndTime); result > 0 {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message: fmt.Sprintf(
				"Booking time must be less than %d:%d",
				scheduleFound.EndTime.Hour(),
				scheduleFound.EndTime.Minute(),
			),
		})
	}

	//Prevent to user to pass his own CustomerID, get his id with:
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	customerId, err := strconv.Atoi(claims["string"].(string))
	if err != nil {
		return c.Status(500).JSON(common.ErrorResponse{
			StatusCode: 500,
			Message:    "Error getting user info",
		})
	}
	fmt.Println("user_id: ", customerId)

	booking := model.Booking{
		CustomerID:  customerId,
		EmployeeID:  bookingDto.EmployeeID,
		ServiceID:   bookingDto.ServiceID,
		BookingDate: bookingDto.BookingDate,
		Status:      "booked",
		Notes:       bookingDto.Notes,
	}

	database.DB.Db.Create(&booking)
	return c.Status(201).JSON(booking)
}

// Admin can see all Bookings
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
	bookings := []model.Booking{} //declaring bookings slice

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

// func FindBookingByEmployeeId(c *fiber.Ctx) error {
// 	id, err := strconv.Atoi(c.Params("id"))
// 	if err != nil {
// 		return c.Status(400).JSON(common.ErrorResponse{
// 			StatusCode: 400,
// 			Message:    "Numeric string is expected",
// 		})
// 	}
//
// 	booking := new(model.Booking)
// 	restult := database.DB.Db.First(booking, id)
// 	if restult.Error != nil {
// 		return c.Status(404).JSON(common.ErrorResponse{
// 			StatusCode: 404,
// 			Message:    "Booking not found",
// 		})
// 	}
//
// 	return c.Status(200).JSON(booking)
// }
