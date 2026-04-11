package schedule

import (
	"errors"
	"fmt"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ScheduleService struct {
	db *gorm.DB
}

func NewScheduleService(db *gorm.DB) *ScheduleService {
	return &ScheduleService{db: db}
}

func (ss *ScheduleService) Create(userID uint, scheduleDto CreateScheduleDto) (*model.Schedule, error) {
	layout := "15:04"

	dtoStartTime, err := time.ParseInLocation(layout, scheduleDto.StartTime, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "start time must be in HH:MM format")
	}

	dtoEndTime, err := time.ParseInLocation(layout, scheduleDto.EndTime, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "start time must be in HH:MM format")
	}

	dtoStart := utils.NewTimeOfDay(dtoStartTime)
	dtoEnd := utils.NewTimeOfDay(dtoEndTime)

	if !dtoStart.Before(dtoEnd) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "end time must be greater than the start time")
	}

	schedulesFound := []model.Schedule{}
	ss.db.Where("employee_id = ? AND week_day = ?", userID, scheduleDto.WeekDay).Find(&schedulesFound)

	hasOverlap, matchSchedule := hasScheduleOverlap(dtoStart, dtoEnd, schedulesFound)

	if hasOverlap {
		errMessage := fmt.Sprintf("time block [%s , %s] must be in an interval other than [%s , %s]",
			scheduleDto.StartTime,
			scheduleDto.EndTime,
			matchSchedule.StartTime.UTC().Format(layout),
			matchSchedule.EndTime.UTC().Format(layout),
		)

		return nil, fmt.Errorf("%w: %s", common.ErrConflict, errMessage)
	}

	newSchedule := model.Schedule{
		EmployeeID: userID,
		WeekDay:    scheduleDto.WeekDay,
		StartTime:  dtoStartTime,
		EndTime:    dtoEndTime,
	}

	ss.db.Create(&newSchedule)
	return &newSchedule, nil
}

func hasScheduleOverlap(dtoStart, dtoEnd utils.TimeOfDay, existing []model.Schedule) (bool, *model.Schedule) {
	for i, schedule := range existing {
		existingStart := utils.NewTimeOfDay(schedule.StartTime)
		existingEnd := utils.NewTimeOfDay(schedule.EndTime)

		// OVERLAP
		if dtoStart.Before(existingEnd) && dtoEnd.After(existingStart) {
			return true, &existing[i]
		}
	}
	return false, nil
}

func (ss *ScheduleService) FindAll(limit, page int, orderBy string) (*AllSchedules, error) {
	offset := (page - 1) * limit
	query := ss.db.Model(&model.Schedule{})

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	if orderBy != "" {
		query = query.Order(orderBy)
	}

	schedules := []model.Schedule{}
	if err := query.Limit(limit).Offset(offset).Find(&schedules).Error; err != nil {
		return nil, err
	}

	return &AllSchedules{
		Data:     schedules,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (ss *ScheduleService) FindOne(id uint) (*model.Schedule, error) {
	schedule := new(model.Schedule)
	result := ss.db.First(&schedule, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "schedule")
	}

	return schedule, nil
}

// FindByUserID there is no need for paginate because employee needs to see all his schedules
func (ss *ScheduleService) FindByUserID(userID uint) ([]model.Schedule, error) {
	schedules := []model.Schedule{}
	if result := ss.db.Where("employee_id = ?", userID).Order("week_day").Find(&schedules); result.Error != nil {
		return nil, result.Error
	}
	return schedules, nil
}

func (ss *ScheduleService) Delete(id uint) (*model.Schedule, error) {
	var schedule model.Schedule

	if result := ss.db.Clauses(clause.Returning{}).Delete(&schedule, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "schedule")
		}
		return nil, result.Error
	}
	return &schedule, nil
}

func (ss *ScheduleService) DeleteByUserID(id, userID uint) (*model.Schedule, error) {
	var schedule model.Schedule

	if result := ss.db.Clauses(clause.Returning{}).Where("employee_id = ? AND id = ?", userID, id).Delete(&schedule); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "schedule")
		}
		return nil, result.Error
	}
	return &schedule, nil
}
