package scheduleexceptions

import (
	"fmt"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
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

	dtoDate, err := time.Parse(time.DateOnly, scheduleExDto.Date)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "date must be in YYYY-MM-DD format")
	}

	// Date > today date (both YYYY-MM-DD, so string comparison is chronological)
	if scheduleExDto.Date <= ss.today() {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "exception date must be greater than current date")
	}

	scheduleEx := model.ScheduleOverride{
		EmployeeID:  userID,
		Date:        dtoDate,
		Reason:      scheduleExDto.Reason,
		AllDay:      scheduleExDto.AllDay,
		IsAvailable: scheduleExDto.IsAvailable,
	}

	if !scheduleExDto.AllDay {
		if scheduleExDto.StartTime == nil || scheduleExDto.EndTime == nil {
			return nil, fmt.Errorf("%w: %s", common.ErrValidation, "startTime and endTime are required")
		}

		dtoStartTime, errStart := time.ParseInLocation(time.TimeOnly, *scheduleExDto.StartTime, time.UTC)
		dtoEndTime, errEnd := time.ParseInLocation(time.TimeOnly, *scheduleExDto.EndTime, time.UTC)
		if errStart != nil || errEnd != nil {
			return nil, fmt.Errorf("%w: %s", common.ErrValidation, "startTime and endTime must be in HH:MM:SS format")
		}

		// Start Time < End Time
		if !dtoStartTime.Before(dtoEndTime) {
			return nil, fmt.Errorf("%w: %s", common.ErrValidation, "end time must be greater than the start time")
		}

		scheduleEx.StartTime = &dtoStartTime
		scheduleEx.EndTime = &dtoEndTime
	}

	err = ss.db.Transaction(func(tx *gorm.DB) error {
		// Same lock as booking creation, so no booking can be created between the
		// conflict check below and the insert.
		if err := common.LockEmployeeAgenda(tx, userID); err != nil {
			return err
		}

		schedulesFound := []model.ScheduleOverride{}
		if result := tx.Where("date = ? AND employee_id = ?", scheduleExDto.Date, userID).Find(&schedulesFound); result.Error != nil {
			return result.Error
		}

		if scheduleEx.AllDay && len(schedulesFound) != 0 {
			return fmt.Errorf("%w: %s", common.ErrConflict, "there is already an exception for this date")
		}

		if !scheduleEx.AllDay {
			if hasOverlap, matchSchedule := hasScheduleExceptionOverlap(*scheduleEx.StartTime, *scheduleEx.EndTime, schedulesFound); hasOverlap {
				return exceptionOverlapError(scheduleExDto, matchSchedule)
			}
		}

		// Extra availability can't collide with bookings, only blocking exceptions can
		if !scheduleEx.IsAvailable {
			conflicts, err := ss.findConflictingBookings(tx, scheduleEx)
			if err != nil {
				return err
			}
			if len(conflicts) > 0 {
				return &ConflictingBookingsError{Bookings: conflicts}
			}
		}

		return tx.Create(&scheduleEx).Error
	})
	if err != nil {
		return nil, err
	}

	return &scheduleEx, nil
}

func exceptionOverlapError(scheduleExDto CreateScheduleExDto, matchSchedule *model.ScheduleOverride) error {
	if matchSchedule.AllDay {
		return fmt.Errorf("%w: %s", common.ErrConflict, "An all-day exception already exists for this date")
	}
	layout := "15:04"
	return fmt.Errorf("%w: %s", common.ErrConflict, fmt.Sprintf("time block [%s , %s] overlaps with another exception [%s , %s]",
		*scheduleExDto.StartTime,
		*scheduleExDto.EndTime,
		matchSchedule.StartTime.UTC().Format(layout),
		matchSchedule.EndTime.UTC().Format(layout),
	))
}

// findConflictingBookings returns the active bookings of the employee that overlap the
// time blocked by ex (the whole local day for all-day exceptions).
func (ss *ScheduleExService) findConflictingBookings(tx *gorm.DB, ex model.ScheduleOverride) ([]model.Booking, error) {
	day := time.Date(ex.Date.Year(), ex.Date.Month(), ex.Date.Day(), 0, 0, 0, 0, ss.loc)
	start, end := day, day.AddDate(0, 0, 1)
	if !ex.AllDay {
		start = utils.NewTimeOfDay(*ex.StartTime).ToTimeOnDate(day)
		end = utils.NewTimeOfDay(*ex.EndTime).ToTimeOnDate(day)
	}

	var bookings []model.Booking
	err := tx.Preload("Customer").
		Where("employee_id = ? AND status = ? AND booking_date < ? AND end_date > ?",
			ex.EmployeeID, model.StatusBooked.String(), end, start).
		Order("booking_date").
		Find(&bookings).Error
	return bookings, err
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
	if err := query.Find(&schedules).Error; err != nil {
		return nil, err
	}

	var count int64
	if err := ss.db.Model(&model.ScheduleOverride{}).Count(&count).Error; err != nil {
		return nil, err
	}

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
	if err := query.Where("date >= ? AND employee_id = ?", ss.today(), userID).Find(&schedules).Error; err != nil {
		return nil, err
	}

	var count int64
	if err := ss.db.Model(&model.ScheduleOverride{}).Where("date >= ? AND employee_id = ?", ss.today(), userID).Count(&count).Error; err != nil {
		return nil, err
	}

	return &AllSchedules{
		Data:     schedules,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (ss *ScheduleExService) DeleteByUserID(id, employeeID uint) (*model.ScheduleOverride, error) {
	var deletedScheduleEx model.ScheduleOverride

	result := ss.db.Clauses(clause.Returning{}).Where("id = ? AND employee_id = ?", id, employeeID).Delete(&deletedScheduleEx)
	if result.Error != nil {
		return nil, result.Error
	}
	// Delete doesn't return gorm.ErrRecordNotFound, nothing deleted means not found (or not yours)
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "schedule exception")
	}
	return &deletedScheduleEx, nil
}
