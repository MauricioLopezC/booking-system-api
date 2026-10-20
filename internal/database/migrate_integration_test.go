package database_test

import (
	"testing"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A database created before bookings.end_date existed must be upgraded in place.
func TestMigrateUpgradesLegacyBookings(t *testing.T) {
	db := testutil.OpenTestDB(t)

	// Rebuild the legacy shape: no end_date, unique (employee_id, booking_date) index.
	require.NoError(t, db.Exec(`ALTER TABLE bookings DROP CONSTRAINT bookings_no_overlap`).Error)
	require.NoError(t, db.Exec(`ALTER TABLE bookings DROP COLUMN end_date`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX idx_employee_bookingdate ON bookings (employee_id, booking_date)`).Error)

	service := model.Service{Name: "Corte", DurationMinutes: 45, Price: 10}
	require.NoError(t, db.Create(&service).Error)
	user := model.User{Name: "u", Email: "u@test.com", Phone: "1", Password: "x"}
	require.NoError(t, db.Create(&user).Error)
	start := time.Date(2030, 1, 8, 10, 0, 0, 0, time.UTC)
	require.NoError(t, db.Exec(
		`INSERT INTO bookings (created_at, updated_at, customer_id, employee_id, service_id, booking_date, status)
		 VALUES (now(), now(), ?, ?, ?, ?, 'cancelled')`, user.ID, user.ID, service.ID, start).Error)

	require.NoError(t, database.Migrate(db))
	require.NoError(t, database.Migrate(db), "migrations must be idempotent")

	var booking model.Booking
	require.NoError(t, db.First(&booking).Error)
	assert.True(t, booking.EndDate.Equal(start.Add(45*time.Minute)), "end_date backfilled from the service, got %s", booking.EndDate)

	var legacyIndexes int64
	db.Raw(`SELECT count(*) FROM pg_indexes WHERE indexname = 'idx_employee_bookingdate'`).Scan(&legacyIndexes)
	assert.Zero(t, legacyIndexes, "legacy unique index must be dropped")

	// The cancelled legacy booking no longer blocks its time.
	assert.NoError(t, db.Create(&model.Booking{
		CustomerID: user.ID, EmployeeID: user.ID, ServiceID: service.ID,
		BookingDate: start, EndDate: start.Add(45 * time.Minute), Status: model.StatusBooked.String(),
	}).Error)
}
