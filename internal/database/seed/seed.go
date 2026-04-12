// Package seed provides functions to populate the database with initial data.
package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Reset drops and recreates the public schema, then runs AutoMigrate.
func Reset(db *gorm.DB) error {
	sqls := []string{
		"DROP SCHEMA public CASCADE",
		"CREATE SCHEMA public",
	}
	for _, s := range sqls {
		if err := db.Exec(s).Error; err != nil {
			return fmt.Errorf("seed reset: %w", err)
		}
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Service{},
		&model.Schedule{},
		&model.Booking{},
		&model.ScheduleOverride{},
	); err != nil {
		return fmt.Errorf("seed reset migrate: %w", err)
	}
	log.Println("seed: schema dropped and recreated")
	return nil
}

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatalf("seed: failed to hash password: %v", err)
	}
	return string(bytes)
}

// SeedServices creates initial services if they don't exist.
func SeedServices(db *gorm.DB) ([]*model.Service, error) {
	seeds := []model.Service{
		{Name: "Corte de pelo", DurationMinutes: 30, Price: 15},
		{Name: "Barba", DurationMinutes: 20, Price: 10},
		{Name: "Corte + Barba", DurationMinutes: 45, Price: 22},
	}

	var result []*model.Service
	for _, s := range seeds {
		var svc model.Service
		tx := db.Where(model.Service{Name: s.Name}).FirstOrCreate(&svc, s)
		if tx.Error != nil {
			return nil, fmt.Errorf("seed services: %w", tx.Error)
		}
		result = append(result, &svc)
	}
	log.Printf("seed: %d services ready\n", len(result))
	return result, nil
}

// SeedUsers creates initial users (admin, employees, customers) if they don't exist.
func SeedUsers(db *gorm.DB, services []*model.Service) ([]*model.User, error) {
	seeds := []struct {
		user     model.User
		services []*model.Service
	}{
		{
			user: model.User{
				Name:     "Admin",
				Email:    "admin@test.com",
				Phone:    "1111111111",
				Password: hashPassword("admin123"),
				Role:     model.RoleAdmin.String(),
			},
		},
		{
			user: model.User{
				Name:     "Empleado Uno",
				Email:    "emp1@test.com",
				Phone:    "2222222222",
				Password: hashPassword("emp123"),
				Role:     model.RoleEmployee.String(),
			},
			services: services, // all services
		},
		{
			user: model.User{
				Name:     "Empleado Dos",
				Email:    "emp2@test.com",
				Phone:    "3333333333",
				Password: hashPassword("emp123"),
				Role:     model.RoleEmployee.String(),
			},
			services: services[:1], // only first service
		},
		{
			user: model.User{
				Name:     "Cliente Uno",
				Email:    "client1@test.com",
				Phone:    "4444444444",
				Password: hashPassword("client123"),
				Role:     model.RoleCustomer.String(),
			},
		},
		{
			user: model.User{
				Name:     "Cliente Dos",
				Email:    "client2@test.com",
				Phone:    "5555555555",
				Password: hashPassword("client123"),
				Role:     model.RoleCustomer.String(),
			},
		},
	}

	var result []*model.User
	for _, s := range seeds {
		var u model.User
		tx := db.Where(model.User{Email: s.user.Email}).FirstOrCreate(&u, s.user)
		if tx.Error != nil {
			return nil, fmt.Errorf("seed users: %w", tx.Error)
		}
		if len(s.services) > 0 {
			if err := db.Model(&u).Association("Services").Replace(s.services); err != nil {
				return nil, fmt.Errorf("seed users: assign services: %w", err)
			}
		}
		result = append(result, &u)
	}
	log.Printf("seed: %d users ready\n", len(result))
	return result, nil
}

// SeedSchedules creates weekly schedules for employees (Mon–Fri, 09:00–18:00).
func SeedSchedules(db *gorm.DB, employees []*model.User) error {
	// Weekdays: 1=Monday … 5=Friday (matches time.Weekday)
	start := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)

	count := 0
	for _, emp := range employees {
		for weekday := 1; weekday <= 5; weekday++ {
			var s model.Schedule
			tx := db.Where(model.Schedule{EmployeeID: emp.ID, WeekDay: weekday}).
				FirstOrCreate(&s, model.Schedule{
					EmployeeID: emp.ID,
					WeekDay:    weekday,
					StartTime:  start,
					EndTime:    end,
				})
			if tx.Error != nil {
				return fmt.Errorf("seed schedules: %w", tx.Error)
			}
			count++
		}
	}
	log.Printf("seed: %d schedule entries ready\n", count)
	return nil
}

// SeedBookings creates a few sample bookings.
func SeedBookings(db *gorm.DB, customers []*model.User, employees []*model.User, services []*model.Service) error {
	now := time.Now().UTC()
	// Next Monday at 10:00
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	monday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 10, 0, 0, 0, time.UTC)

	seeds := []model.Booking{
		{
			CustomerID:  customers[0].ID,
			EmployeeID:  employees[0].ID,
			ServiceID:   services[0].ID,
			BookingDate: monday,
			Status:      model.StatusBooked.String(),
			Notes:       "Primera reserva de prueba",
		},
		{
			CustomerID:  customers[1].ID,
			EmployeeID:  employees[1].ID,
			ServiceID:   services[0].ID,
			BookingDate: monday.Add(1 * time.Hour),
			Status:      model.StatusCompleted.String(),
		},
		{
			CustomerID:  customers[0].ID,
			EmployeeID:  employees[0].ID,
			ServiceID:   services[1].ID,
			BookingDate: monday.Add(2 * time.Hour),
			Status:      model.StatusBooked.String(),
		},
	}

	count := 0
	for _, b := range seeds {
		var existing model.Booking
		tx := db.Where(model.Booking{EmployeeID: b.EmployeeID, BookingDate: b.BookingDate}).
			FirstOrCreate(&existing, b)
		if tx.Error != nil {
			return fmt.Errorf("seed bookings: %w", tx.Error)
		}
		count++
	}
	log.Printf("seed: %d bookings ready\n", count)
	return nil
}
