package bookings

import (
	"errors"
	"fmt"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"gorm.io/gorm"
)

type BookingsService struct {
	db *gorm.DB
}

func (bs *BookingsService) CreateBooking(booking model.Booking) (*model.Booking, error) {
	var service model.Service
	if result := bs.db.First(&service, booking.ServiceID); result.Error != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "Service not assigned to employee")
	}

	hasService, err := common.EmployeeHasService(bs.db, booking.EmployeeID, booking.ServiceID)
	if err != nil {
		return nil, err
	}

	if !hasService {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "Service not assigned to employee")
	}

	//1. check if BookingDate match with a available schedule
	//find schedules by EmployeeID
	daySchedule := []model.Schedule{}
	bs.db.Where("employee_id = ? AND week_day = ?",
		booking.EmployeeID,
		int(booking.BookingDate.Local().Weekday()),
	).Find(&daySchedule)

	if len(daySchedule) == 0 {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "Employee or Employee schedule not found")
	}

	//Interval [BookingStart, BookingEnd] outside interval [scheduleStartTime, scheduleEndTime]
	hasMatch := false
	bookingStart := booking.BookingDate.Local()
	bookingEnd := bookingStart.Add(time.Duration(service.DurationMinutes) * time.Minute)
	for _, schedule := range daySchedule {
		scheduleStart := time.Date(bookingStart.Year(), bookingStart.Month(), bookingStart.Day(), schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, time.Local)
		scheduleEnd := time.Date(bookingStart.Year(), bookingStart.Month(), bookingStart.Day(), schedule.EndTime.Hour(), schedule.EndTime.Minute(), 0, 0, time.Local)
		if (bookingStart.After(scheduleStart) || bookingStart.Equal(scheduleStart)) &&
			(bookingEnd.Before(scheduleEnd) || bookingEnd.Equal(scheduleEnd)) {
			hasMatch = true
			break
		}
	}

	if !hasMatch {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "bookingTime not match with an available schedule")
	}

	// Check for schedule exeptions
	bookingDate := booking.BookingDate
	exceptions := []model.ScheduleOverride{}
	if result := bs.db.Where("DATE(exception_date) = DATE(?) ", bookingDate).Find(&exceptions); result.Error != nil {
		return nil, result.Error
	}

	// Check if [bookingStart, bookingEnd] overlaps interval [exStartTime, exEndtime]
	// excluding the extremes, for each exception of the day
	isBlocked := false
	for _, ex := range exceptions {
		if ex.AllDay {
			isBlocked = true
			break
		}
		if bookingStart.Before(*ex.EndTime) && bookingEnd.After(*ex.StartTime) {
			isBlocked = true
			break
		}
	}
	if isBlocked {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "booking not available.")
	}

	// Check the existing bookings for employee in bookingDate
	bookings := []model.Booking{}
	if result := bs.db.Where("employee_id = ? AND DATE(booking_date) = DATE(?) AND status = ?", booking.EmployeeID, bookingDate, model.StatusBooked.String()).Find(&bookings); result.Error != nil {
		return nil, result.Error
	}

	// Check if [bookingStart, bookingEnd] overlaps interval [existingStart, existingEnd]
	for _, booking := range bookings {
		existingStart := booking.BookingDate
		existingEnd := booking.BookingDate.Add(time.Duration(service.DurationMinutes) * time.Minute)
		if bookingStart.Before(existingEnd) && bookingStart.After(existingStart) {
			isBlocked = true
			break
		}
	}
	if isBlocked {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "booking not available.")
	}

	bs.db.Create(&booking)
	return &booking, nil
}

func (bs *BookingsService) FindAllBookings(limit, page int, orderBy string) (*AllBookingsResponse, error) {
	offset := (page - 1) * limit
	query := bs.db.Limit(limit).Offset(offset)
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	bookings := []model.Booking{} //declaring bookings slice
	query.Find(&bookings)

	var count int64
	bs.db.Model(&model.Booking{}).Count(&count)

	return &AllBookingsResponse{
		Data:     bookings,
		Metadata: metadata{TotalItems: count},
	}, nil

}

func (bs *BookingsService) FindOneBooking(id int) (*model.Booking, error) {
	booking := new(model.Booking)
	if restult := bs.db.First(booking, id); restult.Error != nil {
		if errors.Is(restult.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "booking")
		}
		return nil, restult.Error
	}

	return booking, nil
}

func (bs *BookingsService) FindBookingsByUser(limit, page int, orderBy string, userId uint, userRole string) (*AllBookingsResponse, error) {
	offset := (page - 1) * limit

	query := bs.db.Limit(limit).Offset(offset)
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	if userRole == model.RoleCustomer.String() {
		query = query.Where("customer_id = ?", userId)
	} else {
		query = query.Where("employee_id = ?", userId)
	}

	bookings := []model.Booking{} //declaring bookings slice
	query.Find(&bookings)

	var count int64
	bs.db.Model(&model.Booking{}).Count(&count)

	return &AllBookingsResponse{
		Data:     bookings,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (bs *BookingsService) CancelBooking(id int, userId uint, userRole string) (*model.Booking, error) {

	hasBooking, err := common.UserHasBooking(bs.db, userId, uint(id), userRole)
	if err != nil {
		return nil, err
	}
	if !hasBooking {
		return nil, fmt.Errorf("%w: %s", common.ErrUnauthorized, "You do not have permission to cancel this reservation")
	}

	var booking model.Booking
	if result := bs.db.First(&booking, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "booking")
		}
		return nil, err
	}

	//Check if previous booking status is booked
	if booking.Status != model.StatusBooked.String() {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "you can't cancel booking with status completed or cancelled")
	}

	booking.Status = model.StatusCancelled.String()
	if result := bs.db.Save(booking); result.Error != nil {
		return nil, err
	}

	return &booking, nil
}

func (bs *BookingsService) UpdateBooking(booking model.Booking) (*model.Booking, error) {
	//check if booking date is greater than now, and is Before new two weeks
	now := time.Now()
	if !booking.BookingDate.After(now) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "Booking datetime must be greater than the current one")
	}

	if booking.BookingDate.Before(now.AddDate(0, 0, 14)) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "Booking date must no exceed 2 weks")
	}

	var service model.Service
	if result := bs.db.First(&service, booking.ServiceID); result.Error != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "Service not assigned to employee")
	}

	//check if employee has service
	hasService, err := common.EmployeeHasService(bs.db, booking.EmployeeID, booking.ServiceID)
	if err != nil {
		return nil, err
	}

	if !hasService {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "Service not assigned to employee")
	}

	//1. check if BookingDate match with a available schedule
	//find schedules by EmployeeID
	daySchedule := []model.Schedule{}
	bs.db.Where("employee_id = ? AND week_day = ?",
		booking.EmployeeID,
		int(booking.BookingDate.Local().Weekday()),
	).Find(&daySchedule)

	if len(daySchedule) == 0 {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "Employee or Employee schedule not found")
	}

	//exeptions
	exceptions := []model.ScheduleOverride{}
	if result := bs.db.Where("DATE(exception_date) = DATE(?) ", booking.BookingDate).Find(&exceptions); result.Error != nil {
		return nil, result.Error
	}

	//existing existingBookings
	existingBookings := []model.Booking{}
	if result := bs.db.Where("employee_id = ? AND DATE(booking_date) = DATE(?) AND status = ?", booking.EmployeeID, booking.BookingDate, model.StatusBooked.String()).Find(&existingBookings); result.Error != nil {
		return nil, result.Error
	}

	//check if is an available slot then update or return err
	isAvailable := bs.checkSlot(booking, service, daySchedule, exceptions, existingBookings)
	if !isAvailable {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "Booking cannot be scheduled on the selected datetime")
	}

	if result := bs.db.Save(booking); result.Error != nil {
		return nil, result.Error
	}

	return &booking, nil
}

func (bs *BookingsService) checkSlot(
	booking model.Booking,
	service model.Service,
	daySchedule []model.Schedule,
	exceptions []model.ScheduleOverride,
	existingBookings []model.Booking,
) bool {
	//Interval [BookingStart, BookingEnd] outside interval [scheduleStartTime, scheduleEndTime]
	hasMatch := false
	bookingStart := booking.BookingDate.Local()
	bookingEnd := bookingStart.Add(time.Duration(service.DurationMinutes) * time.Minute)
	for _, schedule := range daySchedule {
		scheduleStart := time.Date(bookingStart.Year(), bookingStart.Month(), bookingStart.Day(), schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, time.Local)
		scheduleEnd := time.Date(bookingStart.Year(), bookingStart.Month(), bookingStart.Day(), schedule.EndTime.Hour(), schedule.EndTime.Minute(), 0, 0, time.Local)
		if (bookingStart.After(scheduleStart) || bookingStart.Equal(scheduleStart)) &&
			(bookingEnd.Before(scheduleEnd) || bookingEnd.Equal(scheduleEnd)) {
			hasMatch = true
			break
		}
	}

	if !hasMatch {
		return false
	}

	// Check for schedule exeptions
	// check if [bookingStart, bookingEnd] overlaps interval [exStartTime, exEndtime]
	// excluding the extremes, for each exception of the day
	isBlocked := false
	for _, ex := range exceptions {
		if ex.AllDay {
			isBlocked = true
			break
		}
		if bookingStart.Before(*ex.EndTime) && bookingEnd.After(*ex.StartTime) {
			isBlocked = true
			break
		}
	}
	if isBlocked {
		return false
	}

	// Check the existing existingBookings for employee in bookingDate
	// Check if [bookingStart, bookingEnd] overlaps interval [existingStart, existingEnd]
	for _, booking := range existingBookings {
		existingStart := booking.BookingDate
		existingEnd := booking.BookingDate.Add(time.Duration(service.DurationMinutes) * time.Minute)
		if bookingStart.Before(existingEnd) && bookingStart.After(existingStart) {
			isBlocked = true
			break
		}
	}
	if isBlocked {
		return false
	}
	return true
}
