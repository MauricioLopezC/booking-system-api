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
	db  *gorm.DB
	loc *time.Location // business timezone, defines what "today" is
}

func NewScheduleExService(db *gorm.DB, loc *time.Location) *ScheduleExService {
	return &ScheduleExService{db: db, loc: loc}
}

func (ss *ScheduleExService) today() string {
	return time.Now().In(ss.loc).Format(time.DateOnly)
}

func (ss *ScheduleExService) Create(userID uint, scheduleExDto CreateScheduleExDto) (*model.ScheduleOverride, error) {
	if scheduleExDto.IsAvailable && scheduleExDto.AllDay {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "an available exception must define startTime and endTime, allDay is only for days off")
	}

	dtoDate, _ := time.Parse(time.DateOnly, scheduleExDto.Date)
	schedulesFound := []model.ScheduleOverride{}
	ss.db.Where("DATE(date) = ? AND employee_id = ?", scheduleExDto.Date, userID).Find(&schedulesFound)

	if scheduleExDto.AllDay && len(schedulesFound) != 0 {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "there is already an exception for this date")
	}

	if scheduleExDto.AllDay && len(schedulesFound) == 0 {
		scheduleEx := model.ScheduleOverride{
			EmployeeID: userID,
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

	dtoStartTime, _ := time.ParseInLocation(time.TimeOnly, *scheduleExDto.StartTime, time.UTC)
	dtoEndTime, _ := time.ParseInLocation(time.TimeOnly, *scheduleExDto.EndTime, time.UTC)

	fmt.Println(dtoStartTime, dtoEndTime)
	// validations
	// Start Time < End Time
	if !dtoStartTime.Before(dtoEndTime) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "end time must be greater than the start time")
	}

	// Date > today date (both YYYY-MM-DD, so string comparison is chronological)
	if scheduleExDto.Date <= ss.today() {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "exception date must be greater than current date")
	}

	hasOverlap, matchSchedule := hasScheduleExceptionOverlap(dtoStartTime, dtoEndTime, schedulesFound)

	if hasOverlap {
		layout := "15:04"
		var errMessage string
		if matchSchedule.AllDay {
			errMessage = "An all-day exception already exists for this date"
		} else {
			errMessage = fmt.Sprintf("time block [%s , %s] overlaps with another exception [%s , %s]",
				*scheduleExDto.StartTime,
				*scheduleExDto.EndTime,
				matchSchedule.StartTime.UTC().Format(layout),
				matchSchedule.EndTime.UTC().Format(layout),
			)
		}
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, errMessage)
	}

	scheduleEx := model.ScheduleOverride{
		EmployeeID:  userID,
		Date:        dtoDate,
		Reason:      scheduleExDto.Reason,
		AllDay:      scheduleExDto.AllDay,
		StartTime:   &dtoStartTime,
		EndTime:     &dtoEndTime,
		IsAvailable: scheduleExDto.IsAvailable,
	}

	if result := ss.db.Create(&scheduleEx); result.Error != nil {
		return nil, result.Error
	}

	return &scheduleEx, nil
}

func hasScheduleExceptionOverlap(dtoStart, dtoEnd time.Time, existing []model.ScheduleOverride) (bool, *model.ScheduleOverride) {
	for i, schedule := range existing {
		if schedule.AllDay {
			return true, &existing[i]
		}

		if schedule.StartTime == nil || schedule.EndTime == nil {
			continue
		}

		existingStart := time.Date(
			0, 1, 1,
			schedule.StartTime.Hour(),
			schedule.StartTime.Minute(),
			0, 0,
			time.UTC,
		)

		existingEnd := time.Date(
			0, 1, 1,
			schedule.EndTime.Hour(),
			schedule.EndTime.Minute(),
			0, 0,
			time.UTC,
		)

		if dtoStart.Before(existingEnd) && dtoEnd.After(existingStart) {
			return true, &existing[i]
		}
	}
	return false, nil
}

func (ss *ScheduleExService) FindAll(limit, page int, orderBy string) (*AllSchedules, error) {
	offset := (page - 1) * limit
	query := ss.db.Limit(limit).Offset(offset)

	if orderBy != "" { // if orderBy is not allowed ignore
		query = query.Order(orderBy)
	}

	schedules := []model.ScheduleOverride{}
	query.Find(&schedules)

	var count int64
	ss.db.Model(&model.ScheduleOverride{}).Count(&count)

	return &AllSchedules{
		Data:     schedules,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (ss *ScheduleExService) FindByUserID(limit, page int, orderBy string, userID uint) (*AllSchedules, error) {
	offset := (page - 1) * limit
	query := ss.db.Limit(limit).Offset(offset)

	if orderBy != "" { // if orderBy is not allowed ignore
		query = query.Order(orderBy)
	}

	schedules := []model.ScheduleOverride{}
	query.Where("date >= ? AND employee_id = ?", ss.today(), userID).Find(&schedules)

	var count int64
	ss.db.Model(&model.ScheduleOverride{}).Where("date >= ? AND employee_id = ?", ss.today(), userID).Count(&count)

	return &AllSchedules{
		Data:     schedules,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (ss *ScheduleExService) DeleteByUserID(id, employeeID uint) (*model.ScheduleOverride, error) {
	var deletedScheduleEx model.ScheduleOverride

	if result := ss.db.Clauses(clause.Returning{}).Where("id = ? AND employee_id = ?", id, employeeID).Delete(&deletedScheduleEx); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "schedule")
		}
		return nil, result.Error
	}
	return &deletedScheduleEx, nil
}
