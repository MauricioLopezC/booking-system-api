package slots

import (
	"testing"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetSlots_WithModels(t *testing.T) {
	// ==== Datos base ====
	employee := model.User{
		Model:    gorm.Model{ID: 1},
		Name:     "Empleado 1",
		Email:    "emp1@example.com",
		Phone:    "123456",
		Password: "hashedpass",
		Role:     "employee",
	}
	customer := model.User{
		Model:    gorm.Model{ID: 2},
		Name:     "Cliente 1",
		Email:    "cust1@example.com",
		Phone:    "654321",
		Password: "hashedpass",
		Role:     "customer",
	}
	serviceDef := model.Service{
		Model:           gorm.Model{ID: 1},
		Name:            "Haircut",
		DurationMinutes: 30,
		Price:           10,
		Users:           []*model.User{&employee},
	}

	// Fecha de prueba: martes 30 sep 2025
	day := time.Date(2025, 9, 30, 0, 0, 0, 0, time.Local)
	dateKey := day.Format("2006-01-02")

	// ==== Horario semanal ====
	weeklySchedule := map[time.Weekday][]model.Schedule{
		time.Tuesday: {
			{
				Model:      gorm.Model{ID: 1},
				EmployeeID: employee.ID,
				WeekDay:    int(time.Tuesday),
				StartTime:  time.Date(0, 1, 1, 8, 0, 0, 0, time.Local),
				EndTime:    time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
				Employee:   employee,
			},
		},
	}

	// ==== Caso 1: sin excepciones ni reservas ====
	exceptions := map[string][]model.ScheduleOverride{}
	bookings := map[string][]model.Booking{}

	ss := NewSlotsService(nil)
	result := ss.getSlots(day, day, serviceDef, weeklySchedule, exceptions, bookings)

	assert.Contains(t, result, dateKey)
	assert.Equal(t, 4, len(result[dateKey])) // 08:00, 08:30, 09:00, 09:30

	// ==== Caso 2: excepción que bloquea todo el día ====
	exceptions = map[string][]model.ScheduleOverride{
		dateKey: {
			{
				Model:       gorm.Model{ID: 1},
				EmployeeID:  employee.ID,
				Date:        day,
				Reason:      "Vacaciones",
				AllDay:      true,
				IsAvailable: false,
				Employee:    employee,
			},
		},
	}
	result = ss.getSlots(day, day, serviceDef, weeklySchedule, exceptions, bookings)
	_, exists := result[dateKey]
	assert.False(t, exists)

	// ==== Caso 3: reserva que ocupa un slot ====
	bookings = map[string][]model.Booking{
		dateKey: {
			{
				Model:       gorm.Model{ID: 1},
				CustomerID:  customer.ID,
				EmployeeID:  employee.ID,
				ServiceID:   serviceDef.ID,
				BookingDate: time.Date(2025, 9, 30, 8, 30, 0, 0, time.Local), // 08:30
				Status:      "booked",
				Customer:    customer,
				Employee:    employee,
				Service:     serviceDef,
			},
		},
	}
	exceptions = map[string][]model.ScheduleOverride{}
	result = ss.getSlots(day, day, serviceDef, weeklySchedule, exceptions, bookings)

	slots := result[dateKey]
	assert.Equal(t, []string{
		"2025-09-30 08:00",
		"2025-09-30 09:00",
		"2025-09-30 09:30",
	}, formatSlots(slots))

	// ==== Caso 4: excepción parcial (bloquea 09:00–10:00) ====
	startBlock := time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, time.Local)
	endBlock := time.Date(day.Year(), day.Month(), day.Day(), 10, 0, 0, 0, time.Local)
	exceptions = map[string][]model.ScheduleOverride{
		dateKey: {
			{
				Model:       gorm.Model{ID: 2},
				EmployeeID:  employee.ID,
				Date:        day,
				Reason:      "Reunión",
				AllDay:      false,
				StartTime:   &startBlock,
				EndTime:     &endBlock,
				IsAvailable: false,
				Employee:    employee,
			},
		},
	}
	bookings = map[string][]model.Booking{}
	result = ss.getSlots(day, day, serviceDef, weeklySchedule, exceptions, bookings)

	slots = result[dateKey]
	assert.Equal(t, []string{
		"2025-09-30 08:00",
		"2025-09-30 08:30",
	}, formatSlots(slots))
}

func formatSlots(slots []time.Time) []string {
	var out []string
	for _, s := range slots {
		out = append(out, s.Format("2006-01-02 15:04"))
	}
	return out
}
