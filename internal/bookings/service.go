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

const MaxBookingDateDays = 14

func (bs *BookingsService) CreateBooking(bookingDto BookingDto, customerID uint) (*model.Booking, error) {
	booking := model.Booking{
		CustomerID:  customerID,
		EmployeeID:  bookingDto.EmployeeID,
		ServiceID:   bookingDto.ServiceID,
		BookingDate: bookingDto.BookingDate,
		Status:      model.StatusBooked.String(),
		Notes:       bookingDto.Notes,
	}

	// check if booking date is greater than now, and is Before new two weeks
	now := time.Now().UTC()
	if !booking.BookingDate.After(now) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "BookingDate must be greater than the current datetime")
	}

	if !booking.BookingDate.Before(now.AddDate(0, 0, MaxBookingDateDays)) {
		return nil, fmt.Errorf("%w: %s", common.ErrValidation, "BookingDate must not exceed 2 weeks from today")
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
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "The booking time doesn't match any available slot for the day")
	}

	if result := bs.db.Create(&booking); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, fmt.Errorf("%w: %s", common.ErrConflict, "The booking has already been reserved!")
		}
		return nil, result.Error
	}

	return &booking, nil
}

func (bs *BookingsService) FindAll(limit, page int, orderBy string) (*AllBookings, error) {
	offset := (page - 1) * limit
	query := bs.db.Model(&model.Booking{})

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	if orderBy != "" {
		query = query.Order(orderBy)
	}

	bookings := []model.Booking{}
	if err := query.Preload("Customer").Preload("Employee").Preload("Service").Limit(limit).Offset(offset).Find(&bookings).Error; err != nil {
		return nil, err
	}

	return &AllBookings{
		Data:     bookings,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (bs *BookingsService) FindByID(id uint) (*model.Booking, error) {
	booking := new(model.Booking)
	if result := bs.db.Preload("Customer").Preload("Employee").Preload("Service").First(booking, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "booking")
		}
		return nil, result.Error
	}

	return booking, nil
}

func (bs *BookingsService) FindByUserID(limit, page int, orderBy string, userID uint, userRole string) (*AllBookings, error) {
	offset := (page - 1) * limit

	query := bs.db.Model(&model.Booking{})

	if userRole == model.RoleCustomer.String() {
		query = query.Where("customer_id = ?", userID)
	} else {
		query = query.Where("employee_id = ?", userID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	// 3. Apply sorting and pagination for data retrieval
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	bookings := []model.Booking{}
	// Note: query.Limit/Offset returns a new instance, safe to use after Count
	if err := query.Preload("Customer").Preload("Employee").Preload("Service").Limit(limit).Offset(offset).Find(&bookings).Error; err != nil {
		return nil, err
	}

	return &AllBookings{
		Data:     bookings,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (bs *BookingsService) CancelBooking(id uint, userID uint, userRole string) (*model.Booking, error) {
	hasBooking, err := common.UserHasBooking(bs.db, userID, id, userRole)
	if err != nil {
		return nil, err
	}
	if !hasBooking {
		return nil, fmt.Errorf("%w: %s", common.ErrUnauthorized, "You do not have permission to cancel this reservation")
	}

	var booking model.Booking
	if result := bs.db.Preload("Customer").Preload("Employee").Preload("Service").First(&booking, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "booking")
		}
		return nil, result.Error
	}

	// Check if previous booking status is booked
	if booking.Status != model.StatusBooked.String() {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "You can't cancel a booking with status completed or cancelled")
	}

	booking.Status = model.StatusCancelled.String()
	if result := bs.db.Save(booking); result.Error != nil {
		return nil, result.Error
	}

	return &booking, nil
}
