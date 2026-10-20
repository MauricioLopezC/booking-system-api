package slots

import (
	"fmt"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"gorm.io/gorm"
)

type SlotsService struct {
	db   *gorm.DB
	step time.Duration  // grid between slot start times, see slotStarts
	loc  *time.Location // business timezone: schedules and exceptions are wall-clock times in it
}

const DefaultSlotStep = 15 * time.Minute

func NewSlotsService(db *gorm.DB, step time.Duration, loc *time.Location) *SlotsService {
	if step <= 0 {
		step = DefaultSlotStep
	}
	if loc == nil {
		loc = time.UTC
	}
	return &SlotsService{db: db, step: step, loc: loc}
}

// Location returns the business timezone.
func (ss *SlotsService) Location() *time.Location {
	return ss.loc
}

// DateKey returns the business-timezone calendar date of t, the key used in SlotMap.
func (ss *SlotsService) DateKey(t time.Time) string {
	return t.In(ss.loc).Format(time.DateOnly)
}

// localMidnight returns the start of t's calendar day in the business timezone.
func (ss *SlotsService) localMidnight(t time.Time) time.Time {
	t = t.In(ss.loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, ss.loc)
}

type SlotMap map[string][]time.Time // fecha local del negocio en formato YYYY-MM-DD -> instantes UTC

const MaxDaysRange = 14

// FindAvailableSlots returns the free slots between the business-timezone calendar
// days of startDate and endDate (both inclusive).
func (ss *SlotsService) FindAvailableSlots(employeeID, serviceID uint, startDate, endDate time.Time) (SlotMap, error) {
	startDate = ss.localMidnight(startDate)
	endDate = ss.localMidnight(endDate)
	endExclusive := endDate.AddDate(0, 0, 1)

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

	// step 3. get schedules exceptions/overrides in date range (both blocking and extra availability)
	exceptions := []model.ScheduleOverride{}
	if result := ss.db.Where(
		"employee_id = ? AND date BETWEEN ? AND ?",
		employeeID, startDate.Format(time.DateOnly), endDate.Format(time.DateOnly),
	).Find(&exceptions); result.Error != nil {
		return nil, result.Error
	}

	// Indexing exceptions by date
	exceptionMap := make(map[string][]model.ScheduleOverride)
	for _, ex := range exceptions {
		key := ex.Date.Format(time.DateOnly) // date column: no timezone conversion
		exceptionMap[key] = append(exceptionMap[key], ex)
	}

	// step 4. get existing active bookings (cancelled/completed ones don't block slots)
	bookings := []model.Booking{}
	if result := ss.db.Where(
		"employee_id = ? AND status = ? AND booking_date < ? AND end_date > ?",
		employeeID, model.StatusBooked.String(), endExclusive, startDate,
	).Find(&bookings); result.Error != nil {
		return nil, result.Error
	}

	// Indexing existing bookings
	bookingsMap := make(map[string][]model.Booking)
	for _, booking := range bookings {
		key := ss.DateKey(booking.BookingDate)
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
	now := time.Now()

	// startDate and endDate are midnights in ss.loc; AddDate keeps the wall clock across DST changes
	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		dateKey := day.Format(time.DateOnly)
		free := freeIntervals(day, weeklySchedule[day.Weekday()], exceptionMap[dateKey], bookingsMap[dateKey])

		dailySlots := slotStarts(free, duration, ss.step, day, now)
		for i := range dailySlots {
			dailySlots[i] = dailySlots[i].UTC()
		}
		if len(dailySlots) > 0 {
			slotMap[dateKey] = dailySlots
		}
	}

	return slotMap
}

// freeIntervals computes the employee's free time on day:
// (weekly schedule + available exceptions) - (blocking exceptions + bookings).
func freeIntervals(
	day time.Time,
	schedules []model.Schedule,
	exceptions []model.ScheduleOverride,
	bookings []model.Booking,
) []Interval {
	var open, busy []Interval

	for _, schedule := range schedules {
		open = append(open, timeRangeOnDate(day, schedule.StartTime, schedule.EndTime))
	}

	for _, ex := range exceptions {
		if ex.AllDay && !ex.IsAvailable {
			return nil // day off
		}
		if ex.StartTime == nil || ex.EndTime == nil {
			continue
		}
		if ex.IsAvailable {
			open = append(open, timeRangeOnDate(day, *ex.StartTime, *ex.EndTime))
		} else {
			busy = append(busy, timeRangeOnDate(day, *ex.StartTime, *ex.EndTime))
		}
	}

	for _, booking := range bookings {
		busy = append(busy, Interval{Start: booking.BookingDate, End: booking.EndDate})
	}

	return subtractIntervals(mergeIntervals(open), mergeIntervals(busy))
}

// timeRangeOnDate places a time-of-day range (dates ignored) on day, in day's timezone.
func timeRangeOnDate(day, start, end time.Time) Interval {
	return Interval{
		Start: utils.NewTimeOfDay(start).ToTimeOnDate(day),
		End:   utils.NewTimeOfDay(end).ToTimeOnDate(day),
	}
}
