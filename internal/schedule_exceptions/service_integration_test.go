package scheduleexceptions_test

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/bookings"
	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	scheduleexceptions "github.com/MauricioLopezC/ReservationSystem/internal/schedule_exceptions"
	"github.com/MauricioLopezC/ReservationSystem/internal/slots"
	"github.com/MauricioLopezC/ReservationSystem/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Same as cmd/api/main.go: postgres time columns are read in time.Local.
func TestMain(m *testing.M) {
	time.Local = time.UTC
	os.Exit(m.Run())
}

type fixture struct {
	db         *gorm.DB
	exceptions *scheduleexceptions.ScheduleExService
	bookings   *bookings.BookingsService
	employee   model.User
	customer   model.User
	service    model.Service
	day        time.Time // tomorrow 00:00 in the business timezone
}

func (f fixture) at(h, m int) time.Time {
	return f.day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute)
}

func (f fixture) date() string {
	return f.day.Format(time.DateOnly)
}

func (f fixture) book(t *testing.T, start time.Time) *model.Booking {
	t.Helper()
	b, err := f.bookings.CreateBooking(bookings.BookingDto{EmployeeID: f.employee.ID, ServiceID: f.service.ID, BookingDate: start}, f.customer.ID)
	require.NoError(t, err)
	return b
}

func blockDto(date, start, end string) scheduleexceptions.CreateScheduleExDto {
	return scheduleexceptions.CreateScheduleExDto{Date: date, Reason: "turno médico", StartTime: &start, EndTime: &end}
}

func allDayDto(date string) scheduleexceptions.CreateScheduleExDto {
	return scheduleexceptions.CreateScheduleExDto{Date: date, Reason: "feriado", AllDay: true}
}

// newFixture: employee working tomorrow 08:00-22:00 (business time in loc), 60 min service.
func newFixture(t *testing.T, loc *time.Location) fixture {
	db := testutil.OpenTestDB(t)

	service := model.Service{Name: "Corte", DurationMinutes: 60, Price: 10}
	employee := model.User{Name: "emp", Email: "emp@test.com", Phone: "1", Password: "x", Role: model.RoleEmployee.String()}
	customer := model.User{Name: "Cliente Uno", Email: "cli@test.com", Phone: "+5491100000000", Password: "x", Role: model.RoleCustomer.String()}
	require.NoError(t, db.Create(&service).Error)
	require.NoError(t, db.Create(&employee).Error)
	require.NoError(t, db.Create(&customer).Error)
	require.NoError(t, db.Model(&employee).Association("Services").Append(&service))

	now := time.Now().In(loc)
	day := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	require.NoError(t, db.Create(&model.Schedule{
		EmployeeID: employee.ID,
		WeekDay:    int(day.Weekday()),
		StartTime:  time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC),
	}).Error)

	return fixture{
		db:         db,
		exceptions: scheduleexceptions.NewScheduleExService(db, loc),
		bookings:   bookings.NewBookingsService(db, slots.NewSlotsService(db, 30*time.Minute, loc)),
		employee:   employee,
		customer:   customer,
		service:    service,
		day:        day,
	}
}

func countExceptions(t *testing.T, db *gorm.DB) int64 {
	var n int64
	require.NoError(t, db.Model(&model.ScheduleOverride{}).Count(&n).Error)
	return n
}

func TestBlockingExceptionOverBookings(t *testing.T) {
	f := newFixture(t, time.UTC)
	booking := f.book(t, f.at(10, 0)) // 10:00-11:00

	t.Run("bloqueo parcial que se superpone devuelve las reservas en conflicto", func(t *testing.T) {
		_, err := f.exceptions.Create(f.employee.ID, blockDto(f.date(), "10:30:00", "12:00:00"))

		var conflictErr *scheduleexceptions.ConflictingBookingsError
		require.True(t, errors.As(err, &conflictErr), "got %v", err)
		assert.ErrorIs(t, err, common.ErrConflict)
		require.Len(t, conflictErr.Bookings, 1)
		assert.Equal(t, booking.ID, conflictErr.Bookings[0].ID)
		assert.Equal(t, "Cliente Uno", conflictErr.Bookings[0].Customer.Name, "customer preloaded so the employee can contact them")
		assert.Zero(t, countExceptions(t, f.db))
	})

	t.Run("bloqueo de todo el día también", func(t *testing.T) {
		_, err := f.exceptions.Create(f.employee.ID, allDayDto(f.date()))
		var conflictErr *scheduleexceptions.ConflictingBookingsError
		assert.True(t, errors.As(err, &conflictErr), "got %v", err)
		assert.Zero(t, countExceptions(t, f.db))
	})

	t.Run("bloqueo adyacente no choca", func(t *testing.T) {
		_, err := f.exceptions.Create(f.employee.ID, blockDto(f.date(), "11:00:00", "12:00:00"))
		assert.NoError(t, err)
	})

	t.Run("horario extra disponible nunca choca con reservas", func(t *testing.T) {
		dto := blockDto(f.date(), "06:00:00", "07:00:00")
		dto.IsAvailable = true
		_, err := f.exceptions.Create(f.employee.ID, dto)
		assert.NoError(t, err)
	})
}

func TestCancelledBookingsDontBlockExceptions(t *testing.T) {
	f := newFixture(t, time.UTC)
	booking := f.book(t, f.at(10, 0))
	_, err := f.bookings.CancelBooking(booking.ID, f.customer.ID, model.RoleCustomer.String())
	require.NoError(t, err)

	_, err = f.exceptions.Create(f.employee.ID, allDayDto(f.date()))
	assert.NoError(t, err)
}

func TestAllDayExceptionUsesBusinessTimezone(t *testing.T) {
	art, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	require.NoError(t, err)
	f := newFixture(t, art)

	// 21:00 in Argentina is already the next day in UTC, but it's part of the local day.
	f.book(t, f.at(21, 0))

	_, err = f.exceptions.Create(f.employee.ID, allDayDto(f.date()))
	var conflictErr *scheduleexceptions.ConflictingBookingsError
	assert.True(t, errors.As(err, &conflictErr), "got %v", err)

	// The next local day is free even though the booking falls on it in UTC.
	_, err = f.exceptions.Create(f.employee.ID, allDayDto(f.day.AddDate(0, 0, 1).Format(time.DateOnly)))
	assert.NoError(t, err)
}

// A booking and an all-day exception for the same day, created concurrently,
// must never both succeed.
func TestBookingAndExceptionRace(t *testing.T) {
	f := newFixture(t, time.UTC)

	for i := range 30 {
		require.NoError(t, f.db.Exec("DELETE FROM bookings").Error)
		require.NoError(t, f.db.Exec("DELETE FROM schedule_overrides").Error)

		var wg sync.WaitGroup
		var bookingErr, exceptionErr error
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, bookingErr = f.bookings.CreateBooking(bookings.BookingDto{EmployeeID: f.employee.ID, ServiceID: f.service.ID, BookingDate: f.at(10, 0)}, f.customer.ID)
		}()
		go func() {
			defer wg.Done()
			_, exceptionErr = f.exceptions.Create(f.employee.ID, allDayDto(f.date()))
		}()
		wg.Wait()

		require.False(t, bookingErr == nil && exceptionErr == nil, "iteration %d: booking and blocking exception both created", i)
		require.True(t, bookingErr == nil || exceptionErr == nil, "iteration %d: one of them must win: %v / %v", i, bookingErr, exceptionErr)
	}
}
