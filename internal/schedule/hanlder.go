package schedule

import (
	"errors"
	"fmt"
	"strconv"

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
	employeeId := int(claims["user_id"].(float64))
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

func FindEmployeSchedules() {

}
