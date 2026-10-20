package slots

import (
	"testing"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- helpers ---

// mustTimeOnly construye un valor time-only usando TimeOfDay, dejando claro
// que la fecha es irrelevante y solo se usan hora y minuto.
func mustTimeOnly(h, m int) time.Time {
	return utils.TimeOfDay{Hour: h, Minute: m}.ToTimeOnDate(time.Time{})
}

func makeSchedule(startH, startM, endH, endM int) model.Schedule {
	return model.Schedule{
		StartTime: mustTimeOnly(startH, startM),
		EndTime:   mustTimeOnly(endH, endM),
	}
}

func makeAllDayException(date time.Time) model.ScheduleOverride {
	return model.ScheduleOverride{
		Date:        date,
		AllDay:      true,
		IsAvailable: false,
	}
}

func makePartialException(date time.Time, startH, startM, endH, endM int) model.ScheduleOverride {
	start := mustTimeOnly(startH, startM)
	end := mustTimeOnly(endH, endM)
	return model.ScheduleOverride{
		Date:        date,
		AllDay:      false,
		StartTime:   &start,
		EndTime:     &end,
		IsAvailable: false,
	}
}

func makeBooking(bookingDate time.Time, durationMinutes int) model.Booking {
	return model.Booking{
		BookingDate: bookingDate,
		EndDate:     bookingDate.Add(time.Duration(durationMinutes) * time.Minute),
	}
}

func formatSlots(slots []time.Time) []string {
	var out []string
	for _, s := range slots {
		out = append(out, s.Format("2006-01-02 15:04"))
	}
	return out
}

// --- tests ---

// TestGetSlots cubre los casos del algoritmo de generación de turnos.
//
// Fechas base (futuro garantizado, nunca filtradas por "ahora"):
//   tue = 2030-01-08 (martes)
//   wed = 2030-01-09 (miércoles)
//   thu = 2030-01-10 (jueves)
//
// El horario base es martes 08:00-10:00 con servicio de 30 min → slots 08:00, 08:30, 09:00, 09:30.
func TestGetSlots(t *testing.T) {
	ss := NewSlotsService(nil)

	// 2030-01-01 es martes.
	tue := time.Date(2030, 1, 8, 0, 0, 0, 0, time.UTC)
	wed := time.Date(2030, 1, 9, 0, 0, 0, 0, time.UTC)
	thu := time.Date(2030, 1, 10, 0, 0, 0, 0, time.UTC)

	svc30 := model.Service{Model: gorm.Model{ID: 1}, DurationMinutes: 30}
	svc60 := model.Service{Model: gorm.Model{ID: 2}, DurationMinutes: 60}

	tueSchedule := map[time.Weekday][]model.Schedule{
		time.Tuesday: {makeSchedule(8, 0, 10, 0)},
	}
	noExceptions := map[string][]model.ScheduleOverride{}
	noBookings := map[string][]model.Booking{}

	tests := []struct {
		name           string
		startDate      time.Time
		endDate        time.Time
		service        model.Service
		weeklySchedule map[time.Weekday][]model.Schedule
		exceptionMap   map[string][]model.ScheduleOverride
		bookingsMap    map[string][]model.Booking
		checkFn        func(t *testing.T, result SlotMap)
	}{
		// ── Grupo 1: Generación básica ──────────────────────────────────────────

		{
			name:           "T01: un día, un bloque, sin excepciones ni reservas",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap:   noExceptions,
			bookingsMap:    noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				key := "2030-01-08"
				require.Contains(t, result, key)
				assert.Equal(t, []string{
					"2030-01-08 08:00",
					"2030-01-08 08:30",
					"2030-01-08 09:00",
					"2030-01-08 09:30",
				}, formatSlots(result[key]))
			},
		},
		{
			name:      "T02: un día, dos bloques de horario",
			startDate: tue,
			endDate:   tue,
			service:   svc30,
			weeklySchedule: map[time.Weekday][]model.Schedule{
				time.Tuesday: {
					makeSchedule(8, 0, 10, 0),
					makeSchedule(14, 0, 16, 0),
				},
			},
			exceptionMap: noExceptions,
			bookingsMap:  noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				key := "2030-01-08"
				require.Contains(t, result, key)
				assert.Equal(t, []string{
					"2030-01-08 08:00",
					"2030-01-08 08:30",
					"2030-01-08 09:00",
					"2030-01-08 09:30",
					"2030-01-08 14:00",
					"2030-01-08 14:30",
					"2030-01-08 15:00",
					"2030-01-08 15:30",
				}, formatSlots(result[key]))
			},
		},
		{
			name:      "T03: rango de 3 días, solo los días con horario aparecen en el resultado",
			startDate: tue,
			endDate:   thu,
			service:   svc30,
			weeklySchedule: map[time.Weekday][]model.Schedule{
				time.Tuesday:  {makeSchedule(8, 0, 10, 0)},
				time.Thursday: {makeSchedule(9, 0, 11, 0)},
				// miércoles sin horario
			},
			exceptionMap: noExceptions,
			bookingsMap:  noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				assert.Contains(t, result, "2030-01-08")
				assert.NotContains(t, result, "2030-01-09")
				assert.Contains(t, result, "2030-01-10")
				assert.Len(t, result, 2)
			},
		},
		{
			name:           "T04: ningún horario para ese weekday → día no aparece",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: map[time.Weekday][]model.Schedule{},
			exceptionMap:   noExceptions,
			bookingsMap:    noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				assert.Empty(t, result)
			},
		},

		// ── Grupo 2: Filtrado de slots pasados ──────────────────────────────────

		{
			name:      "T05: fecha completamente pasada → ningún slot devuelto",
			startDate: time.Date(2020, 9, 29, 0, 0, 0, 0, time.UTC), // martes pasado
			endDate:   time.Date(2020, 9, 29, 0, 0, 0, 0, time.UTC),
			service:   svc30,
			weeklySchedule: map[time.Weekday][]model.Schedule{
				time.Tuesday: {makeSchedule(8, 0, 10, 0)},
			},
			exceptionMap: noExceptions,
			bookingsMap:  noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				assert.Empty(t, result)
			},
		},
		{
			name:      "T06: hoy como fecha, horario todo el día → ningún slot devuelto es anterior a ahora",
			startDate: time.Now().UTC().Truncate(24 * time.Hour),
			endDate:   time.Now().UTC().Truncate(24 * time.Hour),
			service:   svc30,
			weeklySchedule: func() map[time.Weekday][]model.Schedule {
				today := time.Now().UTC()
				return map[time.Weekday][]model.Schedule{
					today.Weekday(): {makeSchedule(0, 0, 23, 30)},
				}
			}(),
			exceptionMap: noExceptions,
			bookingsMap:  noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				key := time.Now().UTC().Format("2006-01-02")
				now := time.Now().UTC()
				for _, slot := range result[key] {
					assert.False(t, slot.Before(now),
						"slot %s está en el pasado (now=%s)", slot.Format("15:04"), now.Format("15:04"))
				}
			},
		},

		// ── Grupo 3: Excepción AllDay ────────────────────────────────────────────

		{
			name:           "T07: AllDay exception → ningún slot para ese día",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap: map[string][]model.ScheduleOverride{
				"2030-01-08": {makeAllDayException(tue)},
			},
			bookingsMap: noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				assert.NotContains(t, result, "2030-01-08")
			},
		},
		{
			name:      "T08: AllDay en día 1 de rango de 2 días, día 2 no afectado",
			startDate: tue,
			endDate:   wed,
			service:   svc30,
			weeklySchedule: map[time.Weekday][]model.Schedule{
				time.Tuesday:   {makeSchedule(8, 0, 10, 0)},
				time.Wednesday: {makeSchedule(9, 0, 11, 0)},
			},
			exceptionMap: map[string][]model.ScheduleOverride{
				"2030-01-08": {makeAllDayException(tue)},
			},
			bookingsMap: noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				assert.NotContains(t, result, "2030-01-08")
				assert.Contains(t, result, "2030-01-09")
			},
		},

		// ── Grupo 4: Excepción parcial ───────────────────────────────────────────

		{
			name:           "T09: excepción parcial bloquea slot del medio (09:00-09:30)",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap: map[string][]model.ScheduleOverride{
				"2030-01-08": {makePartialException(tue, 9, 0, 9, 30)},
			},
			bookingsMap: noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Equal(t, []string{
					"2030-01-08 08:00",
					"2030-01-08 08:30",
					"2030-01-08 09:30",
				}, formatSlots(result["2030-01-08"]))
			},
		},
		{
			name:           "T10: excepción parcial bloquea el primer slot (08:00-08:30)",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap: map[string][]model.ScheduleOverride{
				"2030-01-08": {makePartialException(tue, 8, 0, 8, 30)},
			},
			bookingsMap: noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Equal(t, []string{
					"2030-01-08 08:30",
					"2030-01-08 09:00",
					"2030-01-08 09:30",
				}, formatSlots(result["2030-01-08"]))
			},
		},
		{
			name:           "T11: excepción parcial bloquea el último slot (09:30-10:00)",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap: map[string][]model.ScheduleOverride{
				"2030-01-08": {makePartialException(tue, 9, 30, 10, 0)},
			},
			bookingsMap: noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Equal(t, []string{
					"2030-01-08 08:00",
					"2030-01-08 08:30",
					"2030-01-08 09:00",
				}, formatSlots(result["2030-01-08"]))
			},
		},
		{
			name:           "T12: excepción parcial cubre todo el bloque → ningún slot",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap: map[string][]model.ScheduleOverride{
				"2030-01-08": {makePartialException(tue, 8, 0, 10, 0)},
			},
			bookingsMap: noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				assert.NotContains(t, result, "2030-01-08")
			},
		},
		{
			name:           "T13: excepción con StartTime/EndTime nil → no panic, slot no bloqueado",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap: map[string][]model.ScheduleOverride{
				"2030-01-08": {
					{
						Date:        tue,
						AllDay:      false,
						StartTime:   nil,
						EndTime:     nil,
						IsAvailable: false,
					},
				},
			},
			bookingsMap: noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Len(t, result["2030-01-08"], 4)
			},
		},

		// ── Grupo 5: Conflicto de reservas ──────────────────────────────────────

		{
			name:           "T14: reserva exactamente en 08:30 bloquea solo ese slot",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap:   noExceptions,
			bookingsMap: map[string][]model.Booking{
				"2030-01-08": {makeBooking(time.Date(2030, 1, 8, 8, 30, 0, 0, time.UTC), 30)},
			},
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Equal(t, []string{
					"2030-01-08 08:00",
					"2030-01-08 09:00",
					"2030-01-08 09:30",
				}, formatSlots(result["2030-01-08"]))
			},
		},
		{
			name:           "T15: reserva a las 08:15 solapa con slots 08:00 y 08:30, ambos bloqueados",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap:   noExceptions,
			bookingsMap: map[string][]model.Booking{
				"2030-01-08": {makeBooking(time.Date(2030, 1, 8, 8, 15, 0, 0, time.UTC), 30)},
			},
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Equal(t, []string{
					"2030-01-08 09:00",
					"2030-01-08 09:30",
				}, formatSlots(result["2030-01-08"]))
			},
		},
		{
			name:           "T16: reserva termina exactamente cuando empieza el slot (adyacente) → slot libre",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap:   noExceptions,
			bookingsMap: map[string][]model.Booking{
				// reserva 07:30-08:00, slot empieza en 08:00 → libre
				"2030-01-08": {makeBooking(time.Date(2030, 1, 8, 7, 30, 0, 0, time.UTC), 30)},
			},
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Len(t, result["2030-01-08"], 4)
			},
		},
		{
			name:           "T17: reserva empieza exactamente cuando termina el slot (adyacente) → slot libre",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap:   noExceptions,
			bookingsMap: map[string][]model.Booking{
				// reserva 09:00-09:30; slot 08:30-09:00 termina justo cuando empieza → libre
				"2030-01-08": {makeBooking(time.Date(2030, 1, 8, 9, 0, 0, 0, time.UTC), 30)},
			},
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Equal(t, []string{
					"2030-01-08 08:00",
					"2030-01-08 08:30",
					"2030-01-08 09:30",
				}, formatSlots(result["2030-01-08"]))
			},
		},
		{
			name:           "T18: múltiples reservas bloquean slots no contiguos",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap:   noExceptions,
			bookingsMap: map[string][]model.Booking{
				"2030-01-08": {
					makeBooking(time.Date(2030, 1, 8, 8, 0, 0, 0, time.UTC), 30),  // bloquea 08:00
					makeBooking(time.Date(2030, 1, 8, 9, 0, 0, 0, time.UTC), 30),  // bloquea 09:00
				},
			},
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Equal(t, []string{
					"2030-01-08 08:30",
					"2030-01-08 09:30",
				}, formatSlots(result["2030-01-08"]))
			},
		},

		// ── Grupo 6: Combinaciones y bordes ─────────────────────────────────────

		{
			name:           "T19: excepción parcial + reserva en el mismo día, ambas aplican",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap: map[string][]model.ScheduleOverride{
				"2030-01-08": {makePartialException(tue, 9, 0, 9, 30)}, // bloquea 09:00
			},
			bookingsMap: map[string][]model.Booking{
				"2030-01-08": {makeBooking(time.Date(2030, 1, 8, 8, 0, 0, 0, time.UTC), 30)}, // bloquea 08:00
			},
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Equal(t, []string{
					"2030-01-08 08:30",
					"2030-01-08 09:30",
				}, formatSlots(result["2030-01-08"]))
			},
		},
		{
			name:           "T20: startDate == endDate, un solo día funciona normalmente",
			startDate:      tue,
			endDate:        tue,
			service:        svc30,
			weeklySchedule: tueSchedule,
			exceptionMap:   noExceptions,
			bookingsMap:    noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				assert.Contains(t, result, "2030-01-08")
				assert.Len(t, result["2030-01-08"], 4)
			},
		},
		{
			name:           "T21: servicio de 60 min en bloque 08:00-10:00 → 2 slots, no 4",
			startDate:      tue,
			endDate:        tue,
			service:        svc60,
			weeklySchedule: tueSchedule,
			exceptionMap:   noExceptions,
			bookingsMap:    noBookings,
			checkFn: func(t *testing.T, result SlotMap) {
				require.Contains(t, result, "2030-01-08")
				assert.Equal(t, []string{
					"2030-01-08 08:00",
					"2030-01-08 09:00",
				}, formatSlots(result["2030-01-08"]))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ss.getSlots(
				tc.startDate, tc.endDate,
				tc.service,
				tc.weeklySchedule,
				tc.exceptionMap,
				tc.bookingsMap,
			)
			tc.checkFn(t, result)
		})
	}
}
