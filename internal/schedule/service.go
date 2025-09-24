package schedule

import (
	"errors"
	"fmt"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"gorm.io/gorm"
)

type ScheduleService struct {
	db *gorm.DB
}

func NewScheduleService(db *gorm.DB) *ScheduleService {
	return &ScheduleService{db: db}
}

func (ss *ScheduleService) CreateSchedule(userId uint, scheduleDto CreateScheduleDto) (*model.Schedule, error) {
	layout := "15:04"

	dtoStartTime, err := time.Parse(layout, scheduleDto.StartTime)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "start time must be in HH:MM format")
	}

	dtoEndTime, err := time.Parse(layout, scheduleDto.EndTime)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "start time must be in HH:MM format")
	}

	if !dtoStartTime.Before(dtoEndTime) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "end time must be greater than the start time")
	}

	schedulesFound := []model.Schedule{}
	ss.db.Where("employee_id = ? AND week_day = ?", userId, scheduleDto.WeekDay).Find(&schedulesFound)

	hasOverlap := false
	matchIndex := -1
	for i, schedule := range schedulesFound {
		existingStartTime := time.Date(0, 0, 0, schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, time.Local)
		existingEndTime := time.Date(0, 0, 0, schedule.EndTime.Hour(), schedule.EndTime.Minute(), 0, 0, time.Local)
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
		errMessage := fmt.Sprintf("time block [%s , %s] must be in an interval other than [%s , %s]",
			scheduleDto.StartTime,
			scheduleDto.EndTime,
			matchSchedule.StartTime.Local().Format(layout),
			matchSchedule.EndTime.Local().Format(layout),
		)

		return nil, fmt.Errorf("%w: %s", common.ErrConflict, errMessage)
	}

	newSchedule := model.Schedule{
		EmployeeID: userId,
		WeekDay:    scheduleDto.WeekDay,
		StartTime:  dtoStartTime,
		EndTime:    dtoEndTime,
	}

	ss.db.Create(&newSchedule)
	return &newSchedule, nil
}

func (ss *ScheduleService) FindAllSchedules(limit, page int, orderBy string) (*AllSchedulesResponse, error) {
	offset := (page - 1) * limit
	query := ss.db.Limit(limit).Offset(offset)

	if orderBy != "" { //if orderBy is not allowed ignore
		query = query.Order(orderBy)
	}

	schedules := []model.Schedule{}
	query.Find(&schedules)

	var count int64
	ss.db.Model(&model.Schedule{}).Count(&count)

	return &AllSchedulesResponse{
		Data:     schedules,
		Metadata: metadata{TotalItems: count},
	}, nil

}

func (ss *ScheduleService) FindOneSchedule(id uint) (*model.Schedule, error) {
	schedule := new(model.Schedule)
	result := ss.db.First(&schedule, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "schedule")
	}

	return schedule, nil
}

// there is no need for paginate because employee needs to see all his schedules
func (ss *ScheduleService) FindSchedulesByEmployee(employeeId uint) ([]model.Schedule, error) {
	schedules := []model.Schedule{}
	if result := ss.db.Where("employee_id = ?", employeeId).Order("week_day").Find(&schedules); result.Error != nil {
		return nil, result.Error
	}
	return schedules, nil
}

func (ss *ScheduleService) DeleteSchedule(id uint) (*model.Schedule, error) {
	var schedule model.Schedule
	if result := ss.db.Where("id = ?", id).Delete(&schedule); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "schedule")
		}
		return nil, result.Error
	}
	return &schedule, nil
}
