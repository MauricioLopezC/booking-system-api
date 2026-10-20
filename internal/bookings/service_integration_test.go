package bookings

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/slots"
	"github.com/MauricioLopezC/ReservationSystem/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fixture struct {
	db       *gorm.DB
	svc      *BookingsService
	service  model.Service
	employee model.User
	customer model.User
	day      time.Time // tomorrow 00:00 UTC
}

// at returns tomorrow at hh:mm UTC.
func (f fixture) at(h, m int) time.Time {
	return f.day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute)
}

func (f fixture) dto(start time.Time) BookingDto {
	return BookingDto{EmployeeID: f.employee.ID, ServiceID: f.service.ID, BookingDate: start}
}

// newFixture: one employee working tomorrow 08:00-20:00 offering a 60 min service.
func newFixture(t *testing.T) fixture {
	db := testutil.OpenTestDB(t)

	service := model.Service{Name: "Corte", DurationMinutes: 60, Price: 10}
	require.NoError(t, db.Create(&service).Error)

	employee := model.User{Name: "emp", Email: "emp@test.com", Phone: "1", Password: "x", Role: model.RoleEmployee.String()}
	customer := model.User{Name: "cli", Email: "cli@test.com", Phone: "2", Password: "x", Role: model.RoleCustomer.String()}
	require.NoError(t, db.Create(&employee).Error)
	require.NoError(t, db.Create(&customer).Error)
	require.NoError(t, db.Model(&employee).Association("Services").Append(&service))

	now := time.Now().UTC()
	day := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&model.Schedule{
		EmployeeID: employee.ID,
		WeekDay:    int(day.Weekday()),
		StartTime:  time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC),
	}).Error)

	return fixture{
		db:       db,
		svc:      NewBookingsService(db, slots.NewSlotsService(db, 30*time.Minute)),
		service:  service,
		employee: employee,
		customer: customer,
		day:      day,
	}
}

func TestOverlapConstraint(t *testing.T) {
	f := newFixture(t)

	insert := func(start, end time.Time, status model.BookingStatus) error {
		return f.db.Create(&model.Booking{
			CustomerID: f.customer.ID, EmployeeID: f.employee.ID, ServiceID: f.service.ID,
			BookingDate: start, EndDate: end, Status: status.String(),
		}).Error
	}

	require.NoError(t, insert(f.at(10, 0), f.at(11, 0), model.StatusBooked))

	err := insert(f.at(10, 30), f.at(11, 30), model.StatusBooked)
	assert.True(t, common.IsExclusionViolation(err), "overlapping booking must be rejected by the DB, got %v", err)

	assert.NoError(t, insert(f.at(11, 0), f.at(12, 0), model.StatusBooked), "adjacent bookings don't overlap")
	assert.NoError(t, insert(f.at(10, 0), f.at(11, 0), model.StatusCancelled), "cancelled bookings don't take part in the constraint")
}

func TestCancelledBookingFreesTheSlot(t *testing.T) {
	f := newFixture(t)

	first, err := f.svc.CreateBooking(f.dto(f.at(10, 0)), f.customer.ID)
	require.NoError(t, err)

	_, err = f.svc.CreateBooking(f.dto(f.at(10, 0)), f.customer.ID)
	require.ErrorIs(t, err, common.ErrConflict)

	cancelled, err := f.svc.CancelBooking(first.ID, f.customer.ID, model.RoleCustomer.String())
	require.NoError(t, err)
	assert.Equal(t, model.StatusCancelled.String(), cancelled.Status)

	_, err = f.svc.CancelBooking(first.ID, f.customer.ID, model.RoleCustomer.String())
	assert.ErrorIs(t, err, common.ErrConflict, "a booking can't be cancelled twice")

	_, err = f.svc.CreateBooking(f.dto(f.at(10, 0)), f.customer.ID)
	assert.NoError(t, err, "the time of a cancelled booking must be bookable again")
}

func TestConcurrentOverlappingBookings(t *testing.T) {
	f := newFixture(t)

	// 60 min service: 10:00 and 10:30 overlap, so only one request may win.
	starts := []time.Time{f.at(10, 0), f.at(10, 30), f.at(10, 0), f.at(10, 30), f.at(10, 0), f.at(10, 30), f.at(10, 0), f.at(10, 30)}

	var wg sync.WaitGroup
	errs := make([]error, len(starts))
	for i, start := range starts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = f.svc.CreateBooking(f.dto(start), f.customer.ID)
		}()
	}
	wg.Wait()

	succeeded := 0
	for _, err := range errs {
		if err == nil {
			succeeded++
			continue
		}
		assert.True(t, errors.Is(err, common.ErrConflict), "losers must get ErrConflict, got %v", err)
	}
	assert.Equal(t, 1, succeeded)

	var count int64
	f.db.Model(&model.Booking{}).Where("status = ?", model.StatusBooked.String()).Count(&count)
	assert.EqualValues(t, 1, count)
}

func TestBookingDateWithOffset(t *testing.T) {
	f := newFixture(t)

	// 07:00-03:00 is 10:00 UTC, a valid slot.
	local := f.at(10, 0).In(time.FixedZone("ART", -3*60*60))
	booking, err := f.svc.CreateBooking(f.dto(local), f.customer.ID)
	require.NoError(t, err)
	assert.True(t, booking.BookingDate.Equal(f.at(10, 0)))
	assert.True(t, booking.EndDate.Equal(f.at(11, 0)))
}

func TestBookingKeepsItsDurationWhenServiceChanges(t *testing.T) {
	f := newFixture(t)

	_, err := f.svc.CreateBooking(f.dto(f.at(10, 0)), f.customer.ID)
	require.NoError(t, err)

	// Shortening and then deleting the service must not shrink the existing booking.
	require.NoError(t, f.db.Model(&f.service).Update("duration_minutes", 15).Error)
	require.NoError(t, f.db.Delete(&f.service).Error)

	other := model.Service{Name: "Barba", DurationMinutes: 15, Price: 5}
	require.NoError(t, f.db.Create(&other).Error)
	require.NoError(t, f.db.Model(&f.employee).Association("Services").Append(&other))

	available, err := f.svc.slotsService.FindAvailableSlots(f.employee.ID, other.ID, f.day, f.day)
	require.NoError(t, err)
	for _, slot := range available[f.day.Format(time.DateOnly)] {
		assert.False(t, !slot.Before(f.at(10, 0)) && slot.Before(f.at(11, 0)), "slot %s overlaps the existing booking", slot)
	}
}

func TestAvailableExceptionExtendsWorkingHours(t *testing.T) {
	f := newFixture(t)

	// Base schedule ends at 20:00; the exception adds 20:00-21:00 tomorrow.
	_, err := f.svc.CreateBooking(f.dto(f.at(20, 0)), f.customer.ID)
	require.ErrorIs(t, err, common.ErrConflict)

	start := time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC)
	end := time.Date(0, 1, 1, 21, 0, 0, 0, time.UTC)
	require.NoError(t, f.db.Create(&model.ScheduleOverride{
		EmployeeID: f.employee.ID, Date: f.day, Reason: "turno extra",
		StartTime: &start, EndTime: &end, IsAvailable: true,
	}).Error)

	booking, err := f.svc.CreateBooking(f.dto(f.at(20, 0)), f.customer.ID)
	require.NoError(t, err)
	assert.True(t, booking.EndDate.Equal(f.at(21, 0)))
}
