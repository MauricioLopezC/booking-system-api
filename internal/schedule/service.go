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

type SlotMap map[string][]time.Time // fecha en formato YYYY-MM-DD

func (ss *ScheduleService) FindAvailableSlots(employeeId, serviceId uint, startDate, endDate time.Time) (SlotMap, error) {

	if endDate.After(startDate.Add(time.Hour * 24 * 14)) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "End date must no exceed two weeks from today")
	}

	var service model.Service
	if result := ss.db.First(&service, serviceId); result.Error != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "Employee does not have the selected service")
	}

	hasService, err := common.EmployeeHasService(ss.db, uint(employeeId), uint(serviceId))
	if err != nil {
		return nil, err
	}

	if !hasService {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "Employee does not have the selected service")
	}

	// step 2. get employee schedules
	schedules, err := ss.FindSchedulesByEmployee(employeeId)
	if err != nil {
		return nil, err
	}

	//index schedules by week day for find later
	weeklySchedule := make(map[time.Weekday][]model.Schedule)
	for _, schedule := range schedules {
		fmt.Println(schedule.StartTime.Zone())
		weekday := time.Weekday(schedule.WeekDay)
		weeklySchedule[weekday] = append(weeklySchedule[weekday], schedule)
	}

	//step 3. get schedules exceptions/overrides in date range
	exceptions := []model.ScheduleOverride{}
	query := ss.db.Where("employee_id = ? AND date BETWEEN ? AND ? AND is_available = ?", employeeId, startDate, endDate, false)
	if result := query.Find(&exceptions); result.Error != nil {
		return nil, err
	}
	// Indexing exceptions by date
	exceptionMap := make(map[string][]model.ScheduleOverride)
	for _, ex := range exceptions {
		key := ex.Date.Format("2006-01-02")
		exceptionMap[key] = append(exceptionMap[key], ex)
	}

	//step 4. get existing bookings
	bookings := []model.Booking{}
	query = ss.db.Where("employee_id = ? AND booking_date BETWEEN ? AND ?", employeeId, startDate, endDate)
	if result := query.Find(&bookings); result.Error != nil {
		return nil, err
	}
	//Indexing existing bookings
	bookingsMap := make(map[string][]model.Booking)
	for _, booking := range bookings {
		key := booking.BookingDate.Format("2006-01-02")
		bookingsMap[key] = append(bookingsMap[key], booking)
	}

	//step 5 loop over each range day
	//TODO: All business logic is in the following loop. extract it to a separate function,
	//this makes testing easier.
	slotMap := make(SlotMap)
	duration := time.Duration(service.DurationMinutes) * time.Minute
	fmt.Println("DurationMinutes: ", duration.Minutes())
	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		weekday := day.Weekday()
		dateKey := day.Format("2006-01-02")
		dailySchedule := weeklySchedule[weekday]
		dayExceptions := exceptionMap[dateKey]
		dayBookings := bookingsMap[dateKey]

		var dailySlots []time.Time
		for _, schedule := range dailySchedule {
			//schedule time interval, for example [start, end] = [08:00, 14:00]
			start := time.Date(day.Year(), day.Month(), day.Day(), schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, time.Local)
			end := time.Date(day.Year(), day.Month(), day.Day(), schedule.EndTime.Hour(), schedule.EndTime.Minute(), 0, 0, time.Local)
			for current := start; current.Add(duration).Before(end) || current.Add(duration).Equal(end); current = current.Add(duration) {
				//current is the start of current slot for example 08:00, 08:30, 09:00, ...
				blocked := false

				//check exceptions
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

				//check bookings
				for _, booking := range dayBookings {
					bookingStartTime := booking.BookingDate
					bookingEndTime := booking.BookingDate.Add(duration)
					if bookingStartTime.Before(current.Add(duration)) && bookingEndTime.After(current) {
						blocked = true
						break
					}
				}

				if !blocked {
					dailySlots = append(dailySlots, current.Local())
				}

			}
		}
		if len(dailySlots) > 0 {
			slotMap[dateKey] = dailySlots
		}
	}

	fmt.Println(slotMap)
	return slotMap, nil
}

func (ss *ScheduleService) getSlots(
	startDate,
	endDate time.Time,
	service model.Service,
	weeklySchedule map[time.Weekday][]model.Schedule,
	exceptionMap map[string][]model.ScheduleOverride,
	bookingsMap map[string][]model.Booking,
) SlotMap {
	slotMap := make(SlotMap)
	duration := time.Duration(service.DurationMinutes) * time.Minute

	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		weekday := day.Weekday()
		dateKey := day.Format("2006-01-02")
		dailySchedule := weeklySchedule[weekday]
		dayExceptions := exceptionMap[dateKey]
		dayBookings := bookingsMap[dateKey]

		var dailySlots []time.Time
		for _, schedule := range dailySchedule {
			//schedule time interval, for example [start, end] = [08:00, 14:00]
			start := time.Date(day.Year(), day.Month(), day.Day(), schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, time.Local)
			end := time.Date(day.Year(), day.Month(), day.Day(), schedule.EndTime.Hour(), schedule.EndTime.Minute(), 0, 0, time.Local)
			for current := start; current.Add(duration).Before(end) || current.Add(duration).Equal(end); current = current.Add(duration) {
				//current is the start of current slot for example 08:00, 08:30, 09:00, ...
				blocked := false

				//check exceptions
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

				//check bookings
				for _, booking := range dayBookings {
					bookingStartTime := booking.BookingDate
					bookingEndTime := booking.BookingDate.Add(duration)
					if bookingStartTime.Before(current.Add(duration)) && bookingEndTime.After(current) {
						blocked = true
						break
					}
				}

				if !blocked {
					dailySlots = append(dailySlots, current.Local())
				}

			}
		}
		if len(dailySlots) > 0 {
			slotMap[dateKey] = dailySlots
		}
	}

	return slotMap
}
