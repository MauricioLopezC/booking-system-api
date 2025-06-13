package schedule

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

// Employees can create, read, update and delete his schedules

// Create, ingoring date from StartTime and EndTime, only consider time
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

	//validate if StartTime < EndTime
	//wee must compare only the hours and minutes
	if result := utils.CompareTimeOnly(scheduleDto.EndTime, scheduleDto.StartTime); result < 1 {
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "EndTime must be gt StartTime",
		})
	}

	//get user info
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	employeeId, err := strconv.Atoi(claims["string"].(string))
	if err != nil {
		return c.Status(500).JSON(common.ErrorResponse{
			StatusCode: 500,
			Message:    "Error getting user info",
		})
	}
	fmt.Println("user_id: ", employeeId)

	// search employee existing schedles and check  for time collisions
	schedulesFound := []model.Schedule{}
	database.DB.Db.Where("employee_id = ?", employeeId).Find(&schedulesFound)
	i := slices.IndexFunc(schedulesFound, func(s model.Schedule) bool {
		return s.WeekDay == scheduleDto.WeekDay
	})
	if i != -1 {
		scheduleFound := schedulesFound[i]
		scheduleStartTime := utils.TimeOfDate(scheduleFound.StartTime)
		scheduleEndTime := utils.TimeOfDate(scheduleFound.EndTime)
		dtoStartTime := utils.TimeOfDate(scheduleDto.StartTime)
		dtoEndTime := utils.TimeOfDate(scheduleDto.EndTime)

		if checkOverlap(scheduleStartTime, scheduleEndTime, dtoStartTime, dtoEndTime) {
			return c.Status(400).JSON(common.ErrorResponse{
				StatusCode: 400,
				Message:    fmt.Sprintf("time block must be in an interval other than [%s,%s]", scheduleStartTime, scheduleEndTime),
			})
		}
	}

	newSchedule := model.Schedule{
		EmployeeID: employeeId,
		WeekDay:    scheduleDto.WeekDay,
		StartTime:  scheduleDto.StartTime,
		EndTime:    scheduleDto.EndTime,
	}

	database.DB.Db.Create(&newSchedule)
	return c.Status(201).JSON(newSchedule)
}

// ts1: timeStart1, te1: timeEnd1...
func checkOverlap(ts1, te1, ts2, te2 time.Time) bool {
	if te1.Before(ts2) || te2.Before(ts1) || te1.Equal(ts2) || te2.Equal(ts1) {
		return false
	}
	return true
}
