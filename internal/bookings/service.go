package bookings

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/slots"
	"gorm.io/gorm"
)

type BookingsService struct {
	db           *gorm.DB
	slotsService *slots.SlotsService
}

func NewBookingsService(db *gorm.DB, slotsS *slots.SlotsService) *BookingsService {
	return &BookingsService{db: db, slotsService: slotsS}
}

func (bs *BookingsService) CreateBooking(bookingDto BookingDto, customerID uint) (*model.Booking, error) {
	booking := model.Booking{
		CustomerID:  customerID,
		EmployeeID:  bookingDto.EmployeeID,
		ServiceID:   bookingDto.ServiceID,
		BookingDate: bookingDto.BookingDate,
		Status:      model.StatusBooked.String(),
		Notes:       bookingDto.Notes,
	}

	//check if booking date is greater than now, and is Before new two weeks
	now := time.Now()
	if !booking.BookingDate.After(now) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "BookingDate must be greater than the current datetime")
	}

	if !booking.BookingDate.Before(now.AddDate(0, 0, 14)) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "BookingDate must no exceed 2 weks")
	}

	availableSlots, err := bs.slotsService.FindAvailableSlots(booking.EmployeeID, booking.ServiceID, booking.BookingDate, booking.BookingDate)
	if err != nil {
		return nil, err
	}

	bookingDateKey := booking.BookingDate.Format("2006-01-02")
	daySlots, ok := availableSlots[bookingDateKey]
	if !ok {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "There are no available appointments for the booking date")
	}
	if !slices.Contains(daySlots, booking.BookingDate) {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "The booking time does't not match  with any of available slots for the day")
	}

	if result := bs.db.Create(&booking); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, fmt.Errorf("%w: %s", common.ErrConflict, "The booking has already been reserved!")
		}
		return nil, result.Error
	}

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

func (bs *BookingsService) FindOneBooking(id uint) (*model.Booking, error) {
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

// WARNING: check this funtion
func (bs *BookingsService) Update(bookingID uint) (*model.Booking, error) {
	//check if booking date is greater than now, and is Before new two weeks

	booking, err := bs.FindOneBooking(bookingID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "Booking not found")
	}

	now := time.Now()
	if !booking.BookingDate.After(now) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "Booking datetime must be greater than the current one")
	}

	if booking.BookingDate.After(now.AddDate(0, 0, 14)) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "Booking date must no exceed 2 weks")
	}

	//check if employee has service
	hasService, err := common.EmployeeHasService(bs.db, booking.EmployeeID, booking.ServiceID)
	if err != nil {
		return nil, err
	}

	if !hasService {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "Service not assigned to employee")
	}

	availableSlots, err := bs.slotsService.FindAvailableSlots(booking.EmployeeID, booking.ServiceID, booking.BookingDate, booking.BookingDate)
	if err != nil {
		return nil, err
	}

	bookingDateKey := booking.BookingDate.Format("2006-01-02")
	daySlots, ok := availableSlots[bookingDateKey]
	if !ok {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "There are no available appointments for the booking date")
	}
	if !slices.Contains(daySlots, booking.BookingDate) {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "The booking time does't not match  with any of available slots for the day")
	}

	if result := bs.db.Save(&booking); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, fmt.Errorf("%w: %s", common.ErrConflict, "The booking has already been reserved!")
		}
		return nil, result.Error
	}

	return booking, nil //booking is a pointer
}
