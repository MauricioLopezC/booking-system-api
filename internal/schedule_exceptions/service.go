package scheduleexceptions

import (
	"errors"
	"fmt"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ScheduleExService struct {
	db *gorm.DB
}

func NewScheduleExService(db *gorm.DB) *ScheduleExService {
	return &ScheduleExService{db: db}
}

func (ss *ScheduleExService) Create(userId uint, scheduleExDto CreateScheduleExDto) (*model.ScheduleOverride, error) {
	dtoDate, _ := time.Parse(time.DateOnly, scheduleExDto.Date)
	schedulesFound := []model.ScheduleOverride{}
	ss.db.Where("DATE(date) = ? AND employee_id = ?", scheduleExDto.Date, userId).Find(&schedulesFound)

	if scheduleExDto.AllDay && len(schedulesFound) != 0 {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "there is already an exception for this date")
	}

	if scheduleExDto.AllDay {
		scheduleEx := model.ScheduleOverride{
			EmployeeID: userId,
			Date:       dtoDate,
			Reason:     scheduleExDto.Reason,
			AllDay:     scheduleExDto.AllDay,
		}

		if result := ss.db.Create(&scheduleEx); result.Error != nil {
			return nil, result.Error
		}

		return &scheduleEx, nil
	}

	if scheduleExDto.StartTime == nil || scheduleExDto.EndTime == nil {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "startTime and endTime are required")
	}

	layout := "15:04"
	dtoStartTime, _ := time.Parse(time.TimeOnly, *scheduleExDto.StartTime)
	dtoEndTime, _ := time.Parse(time.TimeOnly, *scheduleExDto.EndTime)

	fmt.Println(dtoStartTime, dtoEndTime)
	//validations
	// Start Time < End Time
	if !dtoStartTime.Before(dtoEndTime) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "end time must be greater than the start time")
	}

	// Date > today date
	if !time.Now().Before(dtoDate) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "exception date must be greater than current date")
	}

	hasOverlap := false
	matchIndex := -1

	for i, s := range schedulesFound {
		existingStartTime := time.Date(0, 0, 0, s.StartTime.Hour(), s.StartTime.Minute(), 0, 0, time.Local)
		existingEndTime := time.Date(0, 0, 0, s.EndTime.Hour(), s.EndTime.Minute(), 0, 0, time.Local)
		//Note that we already verified that dtoStartTime < dtoEndTime
		if !(dtoEndTime.Before(existingStartTime) ||
			dtoStartTime.After(existingEndTime)) {
			hasOverlap = true
			matchIndex = i
			break
		}
	}

	if hasOverlap {
		matchSchedule := schedulesFound[matchIndex]
		errMessage := fmt.Sprintf("time block [%s , %s] must be in an interval other than [%s , %s]",
			*scheduleExDto.StartTime,
			*scheduleExDto.EndTime,
			matchSchedule.StartTime.Local().Format(layout),
			matchSchedule.EndTime.Local().Format(layout),
		)

		return nil, fmt.Errorf("%w: %s", common.ErrConflict, errMessage)
	}

	scheduleEx := model.ScheduleOverride{
		EmployeeID: userId,
		Date:       dtoDate,
		Reason:     scheduleExDto.Reason,
		AllDay:     scheduleExDto.AllDay,
		StartTime:  &dtoStartTime,
		EndTime:    &dtoEndTime,
	}

	if result := ss.db.Create(&scheduleEx); result.Error != nil {
		return nil, result.Error
	}

	return &scheduleEx, nil
}

func (ss *ScheduleExService) FindAll(limit, page int, orderBy string) (*AllSchedulesResponse, error) {
	offset := (page - 1) * limit
	query := ss.db.Limit(limit).Offset(offset)

	if orderBy != "" { //if orderBy is not allowed ignore
		query = query.Order(orderBy)
	}

	schedules := []model.ScheduleOverride{}
	query.Find(&schedules)

	var count int64
	ss.db.Model(&model.Schedule{}).Count(&count)

	return &AllSchedulesResponse{
		Data:     schedules,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (ss *ScheduleExService) FindAllByUSer(limit, page int, orderBy string, userId uint) (*AllSchedulesResponse, error) {
	offset := (page - 1) * limit
	query := ss.db.Limit(limit).Offset(offset)

	if orderBy != "" { //if orderBy is not allowed ignore
		query = query.Order(orderBy)
	}

	schedules := []model.ScheduleOverride{}
	query.Where("date > ? AND employee_id = ?", time.Now().UTC(), userId).Find(schedules)

	var count int64
	ss.db.Model(&model.Schedule{}).Count(&count)

	return &AllSchedulesResponse{
		Data:     schedules,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (ss *ScheduleExService) DeleteByUser(id, employeeId uint) (*model.ScheduleOverride, error) {
	var deletedScheduleEx model.ScheduleOverride

	if result := ss.db.Clauses(clause.Returning{}).Where("id = ? AND employee_id = ?", id, employeeId).Delete(&deletedScheduleEx); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "schedule")
		}
		return nil, result.Error
	}
	return &deletedScheduleEx, nil
}
