// Package database
package database

import (
	"fmt"
	"log"

	"github.com/MauricioLopezC/ReservationSystem/config"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Dbinstance struct {
	Db *gorm.DB
}

var DB Dbinstance

func ConnectDB() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Config("DB_HOST"),
		config.Config("DB_PORT"),
		config.Config("DB_USER"),
		config.Config("DB_PASSWORD"),
		config.Config("DB_NAME"),
	)

	db, err := Open(dsn, logLevel(config.Get("DB_LOG_LEVEL", "warn")))
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
	}
	log.Println("connected")

	DB = Dbinstance{
		Db: db,
	}
}

// logLevel maps DB_LOG_LEVEL (silent | error | warn | info) to a GORM log level.
// "info" logs every SQL statement, useful in development only.
func logLevel(name string) logger.LogLevel {
	switch name {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}

// Open connects to postgres with the GORM settings shared by the API, the seed and the tests.
// TranslateError makes GORM return gorm.ErrDuplicatedKey & co. instead of raw driver errors.
func Open(dsn string, logLevel logger.LogLevel) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logLevel),
		TranslateError: true,
	})
}

// Migrate migrates all the database tables and applies the constraints that
// AutoMigrate can't express. Every step is idempotent, so it runs on every startup.
func Migrate(db *gorm.DB) error {
	if err := preMigrate(db); err != nil {
		return fmt.Errorf("pre-migrate: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Service{},
		&model.Schedule{},
		&model.Booking{},
		&model.ScheduleOverride{},
	); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}

	if err := postMigrate(db); err != nil {
		return fmt.Errorf("post-migrate: %w", err)
	}
	return nil
}

// preMigrate prepares existing databases for changes AutoMigrate can't do safely:
//   - adds bookings.end_date (a NOT NULL column can't be added to a table with rows)
//   - turns schedule_overrides.date from timestamptz into date, casting in UTC
//     (where the old values were stored at midnight) instead of the session timezone
func preMigrate(db *gorm.DB) error {
	return execAll(db,
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'schedule_overrides' AND column_name = 'date' AND data_type <> 'date'
			) THEN
				ALTER TABLE schedule_overrides ALTER COLUMN date TYPE date USING (date AT TIME ZONE 'UTC')::date;
			END IF;
		END $$`,
		`ALTER TABLE IF EXISTS bookings ADD COLUMN IF NOT EXISTS end_date timestamptz`,
		`DO $$ BEGIN
			IF to_regclass('bookings') IS NOT NULL THEN
				UPDATE bookings b
				SET end_date = b.booking_date + s.duration_minutes * interval '1 minute'
				FROM services s
				WHERE s.id = b.service_id AND b.end_date IS NULL;
			END IF;
		END $$`,
		`ALTER TABLE IF EXISTS bookings ALTER COLUMN end_date SET NOT NULL`,
	)
}

// postMigrate replaces the old (employee_id, booking_date) unique index with an
// exclusion constraint: two active bookings of the same employee can't overlap in time.
// Cancelled and soft-deleted bookings don't take part, so their time can be booked again.
func postMigrate(db *gorm.DB) error {
	return execAll(db,
		`CREATE EXTENSION IF NOT EXISTS btree_gist`,
		`DROP INDEX IF EXISTS idx_employee_bookingdate`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'bookings_no_overlap') THEN
				ALTER TABLE bookings ADD CONSTRAINT bookings_no_overlap
				EXCLUDE USING gist (
					employee_id WITH =,
					tstzrange(booking_date, end_date) WITH &&
				) WHERE (status = 'booked' AND deleted_at IS NULL);
			END IF;
		END $$`,
	)
}

func execAll(db *gorm.DB, statements ...string) error {
	for _, s := range statements {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}
