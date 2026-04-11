package slots

import (
	"fmt"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"gorm.io/gorm"
)

type SlotsService struct {
	db *gorm.DB
}

func NewSlotsService(db *gorm.DB) *SlotsService {
	return &SlotsService{db: db}
}

type SlotMap map[string][]time.Time // fecha en formato YYYY-MM-DD

const MaxDaysRange = 14

func (ss *SlotsService) FindAvailableSlots(employeeID, serviceID uint, startDate, endDate time.Time) (SlotMap, error) {
	if endDate.After(startDate.AddDate(0, 0, MaxDaysRange)) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "endDate must no exceed two weeks from today")
	}

	if endDate.Before(startDate) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "endDate must be greater than startDate")
	}

	var service model.Service
	if result := ss.db.First(&service, serviceID); result.Error != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "service not found")
	}

	hasService, err := common.EmployeeHasService(ss.db, employeeID, serviceID)
	if err != nil {
		return nil, err
	}

	if !hasService {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "Employee does not have the selected service")
	}

	// step 2. get employee schedules
	schedules := []model.Schedule{}
	if result := ss.db.Where("employee_id = ?", employeeID).Order("week_day").Find(&schedules); result.Error != nil {
		return nil, result.Error
	}

	// index schedules by week day for find later
	weeklySchedule := make(map[time.Weekday][]model.Schedule)
	for _, schedule := range schedules {
		weekday := time.Weekday(schedule.WeekDay)
		weeklySchedule[weekday] = append(weeklySchedule[weekday], schedule)
	}

	// step 3. get schedules exceptions/overrides in date range
	exceptions := []model.ScheduleOverride{}

	if startDate.Equal(endDate) {
		result := ss.db.Where("employee_id = ? AND date BETWEEN ? AND ? AND is_available = ?", employeeID, startDate, endDate.AddDate(0, 0, 1), false).Find(&exceptions)
		if result.Error != nil {
			return nil, err
		}
	} else {
		result := ss.db.Where("employee_id = ? AND date BETWEEN ? AND ? AND is_available = ?", employeeID, startDate, endDate, false).Find(&exceptions)
		if result.Error != nil {
			return nil, err
		}
	}

	// Indexing exceptions by date
	exceptionMap := make(map[string][]model.ScheduleOverride)
	for _, ex := range exceptions {
		key := ex.Date.Format("2006-01-02")
		exceptionMap[key] = append(exceptionMap[key], ex)
	}

	// step 4. get existing bookings
	bookings := []model.Booking{}
	if startDate.Equal(endDate) {
		result := ss.db.Preload("Service").Where("employee_id = ? AND booking_date BETWEEN ? AND ?", employeeID, startDate, endDate.AddDate(0, 0, 1)).Find(&bookings)
		if result.Error != nil {
			return nil, err
		}
	} else {
		result := ss.db.Preload("Service").Where("employee_id = ? AND booking_date BETWEEN ? AND ?", employeeID, startDate, endDate).Find(&bookings)
		if result.Error != nil {
			return nil, err
		}
	}

	// Indexing existing bookings
	bookingsMap := make(map[string][]model.Booking)
	for _, booking := range bookings {
		key := booking.BookingDate.Format("2006-01-02")
		bookingsMap[key] = append(bookingsMap[key], booking)
	}

	// step 5 loop over each range day
	slotMap := ss.getSlots(startDate, endDate, service, weeklySchedule, exceptionMap, bookingsMap)

	return slotMap, nil
}

func (ss *SlotsService) getSlots(
	startDate,
	endDate time.Time,
	service model.Service,
	weeklySchedule map[time.Weekday][]model.Schedule,
	exceptionMap map[string][]model.ScheduleOverride,
	bookingsMap map[string][]model.Booking,
) SlotMap {
	slotMap := make(SlotMap)
	duration := time.Duration(service.DurationMinutes) * time.Minute

	// Loop over each range day
	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		weekday := day.Weekday()
		dateKey := day.Format("2006-01-02")
		dailySchedule := weeklySchedule[weekday]
		dayExceptions := exceptionMap[dateKey]
		dayBookings := bookingsMap[dateKey]

		var dailySlots []time.Time
		for _, schedule := range dailySchedule {
			// schedule time interval, for example [start, end] = [08:00, 14:00], [start, end] = [16:00, 20:00]
			start := time.Date(day.Year(), day.Month(), day.Day(), schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, time.UTC)
			end := time.Date(day.Year(), day.Month(), day.Day(), schedule.EndTime.Hour(), schedule.EndTime.Minute(), 0, 0, time.UTC)
			for current := start; current.Add(duration).Before(end) || current.Add(duration).Equal(end); current = current.Add(duration) {
				// current is the start of current slot for example (duration 30min);  08:00, 08:30, 09:00, ...
				blocked := false

				// check exceptions
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

				// check bookings
				for _, booking := range dayBookings {
					existingStart := booking.BookingDate
					existingEnd := booking.BookingDate.Add(time.Duration(booking.Service.DurationMinutes) * time.Minute)
					currentEnd := current.Add(duration)
					if existingStart.Before(currentEnd) && existingEnd.After(current) {
						blocked = true
						break
					}
				}

				if !blocked {
					dailySlots = append(dailySlots, current.UTC())
				}

			}
		}
		if len(dailySlots) > 0 {
			slotMap[dateKey] = dailySlots
		}
	}

	return slotMap
}
