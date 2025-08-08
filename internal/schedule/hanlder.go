package schedule

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

// Employees can create, read, update and delete his schedules

// TODO: create a test for this function
func CreateSchedule(c *fiber.Ctx) error {
	scheduleDto := new(CreateScheduleDto)

	if err := c.BodyParser(scheduleDto); err != nil {
		return c.Status(400).JSON(common.ErrorResponse{StatusCode: 400, Message: err.Error()})
	}

	if err := utils.Validate(scheduleDto); err != nil {
		return c.Status(err.Code).JSON(common.ErrorResponse{
			StatusCode: uint(err.Code),
			Message:    err.Message,
		})
	}

	dtoStartTime, err := utils.ParseTimeOfDay(scheduleDto.StartTime)
	if err != nil {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Error StartTime must be in HH:MM:SS format",
		})
	}
	dtoEndTime, err := utils.ParseTimeOfDay(scheduleDto.EndTime)
	if err != nil {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Error EndTime must be in HH:MM:SS format",
		})
	}

	//validate if StartTime < EndTime
	if !dtoStartTime.Before(dtoEndTime) {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "EndTime must be greater than StartTime",
		})
	}

	//get user info
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	employeeId := uint(claims["user_id"].(float64))
	fmt.Println("claims", claims)

	// search employee existing schedules and check for time collisions
	schedulesFound := []model.Schedule{}
	database.DB.Db.Where("employee_id = ? AND week_day = ?", employeeId, scheduleDto.WeekDay).Find(&schedulesFound)
	fmt.Println(len(schedulesFound))

	hasOverlap := false
	matchIndex := -1
	for i, schedule := range schedulesFound {
		existingStartTime := utils.NewTimeOfDay(schedule.StartTime)
		existingEndTime := utils.NewTimeOfDay(schedule.EndTime)
		//Note that we already verified that dtoStartTime < dtoEndTime
		if !(dtoEndTime.Before(existingStartTime) ||
			dtoStartTime.After(existingEndTime) ||
			dtoEndTime.Equal(existingStartTime) ||
			dtoStartTime.Equal(existingEndTime)) {
			hasOverlap = true
			matchIndex = i
			break
		}
	}

	if hasOverlap {
		matchSchedule := schedulesFound[matchIndex]
		layout := "15:04"
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message: fmt.Sprintf("time block [%s , %s] must be in an interval other than [%s , %s]",
				scheduleDto.StartTime,
				scheduleDto.EndTime,
				matchSchedule.StartTime.Local().Format(layout),
				matchSchedule.EndTime.Local().Format(layout),
			),
		})
	}

	//TODO: - create a schedule response dto for select fields to response,
	//instead of fiber.Map
	newSchedule := model.Schedule{
		EmployeeID: employeeId,
		WeekDay:    scheduleDto.WeekDay,
		StartTime:  dtoStartTime.ToTime(),
		EndTime:    dtoEndTime.ToTime(),
	}

	database.DB.Db.Create(&newSchedule)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"ID":         newSchedule.ID,
		"CreatedAt":  newSchedule.CreatedAt,
		"UpdatedAt":  newSchedule.UpdatedAt,
		"DeletedAt":  newSchedule.DeletedAt,
		"EmployeeID": newSchedule.EmployeeID,
		"WeekDay":    newSchedule.WeekDay,
		"StartTime":  newSchedule.StartTime.Format("15:04"),
		"EndTime":    newSchedule.EndTime.Format("15:04"),
	})
}

// NOTE: we can use c.queryParser, is like body parser but for query params,
// using a struct with query and validate tags
func FindAllSchedules(c *fiber.Ctx) error {
	//validate query params
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "page must be an integer number and gt 0",
		})
	}

	offset := (page - 1) * limit
	query := database.DB.Db.Limit(limit).Offset(offset)

	orderBy := c.Query("orderBy")
	allowedFields := map[string]bool{
		"updated_at": true,
		"created_at": true,
	}
	orderBy, err = common.ParseSortParam(orderBy, allowedFields)
	fmt.Println(orderBy, err)
	if err == nil { //if orderBy is not allowed ignore
		query = query.Order(orderBy)
	}

	schedules := []model.Schedule{}
	query.Find(&schedules)

	var count int64
	database.DB.Db.Model(&model.Schedule{}).Count(&count)

	return c.Status(fiber.StatusOK).JSON(AllSchedulesResponse{
		Data:     schedules,
		Metadata: metadata{TotalItems: count},
	})
}

func FindOneSchedule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	schedule := new(model.Schedule)
	result := database.DB.Db.First(&schedule, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusNotFound,
			Message:    "Schedule not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(schedule)
}

// Employee gets his weekly schedule.
// Use in endpoint: GET /me/schedules
// there is no need for paginate because employee needs to see all his schedules
func FindSchedulesByEmployee(c *fiber.Ctx) error {
	//get user data from context
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	if role == model.RoleCustomer.String() {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Access denied! only employees can se his schedules",
		})
	}

	schedules := []model.Schedule{}
	if result := database.DB.Db.Where("employee_id = ?", userId).Order("week_day").Find(&schedules); result.Error != nil {
		return common.NewInternalServerError(c)
	}
	return c.Status(fiber.StatusOK).JSON(schedules)
}

// Use in endpoint /DELETE /schedule/:id
// with role employee required and check
// if schedule belongs to the employee
func DeleteSchedule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))

	hasSchedule, err := common.EmployeeHasSchedule(database.DB.Db, userId, uint(id))
	if err != nil {
		return common.NewInternalServerError(c)
	}
	if !hasSchedule {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Access denied! you can only delete your schedules!",
		})
	}

	var schedule model.Schedule
	if result := database.DB.Db.Where("id = ?", id).Delete(&schedule); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    "Schedule not found",
			})
		}
		return common.NewInternalServerError(c)
	}
	return c.Status(fiber.StatusOK).JSON(schedule)
}

type SlotMap map[string][]time.Time // fecha en formato YYYY-MM-DD

func FindAvailableSlots(c *fiber.Ctx) error {
	serviceId := c.QueryInt("serviceId", 0)
	if serviceId <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "ServiceId must be an integer gt 0",
		})
	}

	employeeId := c.QueryInt("employeeId", 0)
	if employeeId <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "employeeId must be an integer gt 0",
		})
	}

	endDateQuery := c.Query("endDate")
	if endDateQuery == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Required query parameter: endDate",
		})
	}
	endDate, err := time.Parse(time.DateOnly, endDateQuery)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}
	startDate := time.Now()

	// step 1. get service and its duration and check employee-service relation
	var service model.Service
	if result := database.DB.Db.First(&service, serviceId); result.Error != nil {
		return common.NewInternalServerError(c)
	}

	hasService, err := common.EmployeeHasService(database.DB.Db, uint(employeeId), uint(serviceId))
	if err != nil {
		return common.NewInternalServerError(c)
	}

	if !hasService {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Service not assigned to employee",
		})
	}

	// step 2. get employee schedules
	schedules := []model.Schedule{}
	if result := database.DB.Db.Where("employee_id = ?", employeeId).Find(&schedules); result.Error != nil {
		return common.NewInternalServerError(c)
	}

	//index schedules by week day for find later
	weeklySchedule := make(map[time.Weekday][]model.Schedule)
	for _, schedule := range schedules {
		fmt.Println(schedule.StartTime.Zone())
		weekday := time.Weekday(schedule.WeekDay)
		weeklySchedule[weekday] = append(weeklySchedule[weekday], schedule)
	}

	//step 3. get schedules exceptions/overrides in date range
	exceptions := []model.ScheduleOverride{}
	query := database.DB.Db.Where("employee_id = ? AND date BETWEEN ? AND ? AND is_available = ?", employeeId, startDate, endDate, false)
	if result := query.Find(&exceptions); result.Error != nil {
		return common.NewInternalServerError(c)
	}
	// Indexing exceptions by date
	exceptionMap := make(map[string][]model.ScheduleOverride)
	for _, ex := range exceptions {
		key := ex.Date.Format("2006-01-02")
		exceptionMap[key] = append(exceptionMap[key], ex)
	}

	//step 4. get existing bookings
	bookings := []model.Booking{}
	query = database.DB.Db.Where("employee_id = ? AND booking_date BETWEEN ? AND ?", employeeId, startDate, endDate)
	if result := query.Find(&bookings); result.Error != nil {
		return common.NewInternalServerError(c)
	}
	//Indexing existing bookings
	bookingsMap := make(map[string][]model.Booking)
	for _, booking := range bookings {
		key := booking.BookingDate.Format("2006-01-02")
		bookingsMap[key] = append(bookingsMap[key], booking)
	}

	//step 5 loop over each range day
	slotMap := make(SlotMap)
	duration := time.Duration(service.DurationMinutes) * time.Minute
	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		weekday := day.Weekday()
		dateKey := day.Format("2006-01-02")
		dailySchedule := weeklySchedule[weekday]
		dayExceptions := exceptionMap[dateKey]
		dayBookings := bookingsMap[dateKey]

		var dailySlots []time.Time
		for _, schedule := range dailySchedule {
			//schedule time interval, for example [start, end] = [08:00, 14:00]
			start := time.Date(day.Year(), day.Month(), day.Day(), schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, time.Local)
			end := time.Date(day.Year(), day.Month(), day.Day(), schedule.EndTime.Hour(), schedule.EndTime.Minute(), 0, 0, time.Local)
			for current := start; current.Add(duration).Before(end) || current.Add(duration).Equal(end); current = current.Add(duration) {
				//current is the start of current slot for example 08:00, 08:30, 09:00, ...
				blocked := false

				//check exceptions
				for _, ex := range dayExceptions {
					if ex.AllDay {
						blocked = true
						break
					}

					if current.Before(*ex.EndTime) && current.Add(duration).After(*ex.StartTime) && !ex.IsAvailable {
						blocked = true
						break
					}
				}

				//check bookings
				for _, booking := range dayBookings {
					bookingStartTime := booking.BookingDate
					bookingEndTime := booking.BookingDate.Add(duration)
					if bookingStartTime.Before(current.Add(duration)) && bookingEndTime.After(current) {
						blocked = true
						break
					}
				}

				if !blocked {
					dailySlots = append(dailySlots, current.Local())
				}

			}
		}
		if len(dailySlots) > 0 {
			slotMap[dateKey] = dailySlots
		}
	}

	fmt.Println(slotMap)
	return c.Status(200).JSON(slotMap)
}
