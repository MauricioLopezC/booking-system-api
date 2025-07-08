package schedule

import (
	"fmt"
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
	employeeId := int(claims["user_id"].(float64))
	fmt.Println("===>", claims)

	// search employee existing schedules and check  for time collisions
	schedulesFound := []model.Schedule{}
	database.DB.Db.Where("employee_id = ? AND week_day = ?", employeeId, scheduleDto.WeekDay).Find(&schedulesFound)
	fmt.Println(len(schedulesFound))

	hasOverlap := false
	matchIndex := -1
	for i, schedule := range schedulesFound {
		scheduleStartTime := utils.TimeOfDate(schedule.StartTime)
		scheduleEndTime := utils.TimeOfDate(schedule.EndTime)
		dtoStartTime := utils.TimeOfDate(scheduleDto.StartTime)
		dtoEndTime := utils.TimeOfDate(scheduleDto.EndTime)
		fmt.Println(scheduleStartTime, scheduleEndTime, dtoStartTime, dtoEndTime)
		hasOverlap = checkOverlap(scheduleStartTime, scheduleEndTime, dtoStartTime, dtoEndTime)
		if hasOverlap {
			matchIndex = i
			break
		}
	}

	if hasOverlap {
		matchSchedule := schedulesFound[matchIndex]
		layout := "15:04"
		fmt.Println("*****")
		fmt.Println(matchSchedule.StartTime.Location())
		fmt.Println(matchSchedule.StartTime.In(time.UTC).Location())
		fmt.Println(matchSchedule.StartTime.In(time.UTC))
		fmt.Println(matchSchedule.StartTime)
		return c.Status(400).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message: fmt.Sprintf("time block [%s , %s] must be in an interval other than [%s , %s]",
				scheduleDto.StartTime.Format(layout),
				scheduleDto.EndTime.Format(layout),
				matchSchedule.StartTime.UTC().Format(layout),
				matchSchedule.EndTime.UTC().Format(layout),
			),
		})
	}
	//TODO: create a test for this function

	newSchedule := model.Schedule{
		EmployeeID: employeeId,
		WeekDay:    scheduleDto.WeekDay,
		StartTime:  scheduleDto.StartTime.In(time.UTC),
		EndTime:    scheduleDto.EndTime.In(time.UTC),
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
