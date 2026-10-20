package utils

import (
	"fmt"
	"time"
)

// TimeOfDay representa una hora del día sin fecha.
type TimeOfDay struct {
	Hour   int // 0-23
	Minute int // 0-59
	Second int // 0-59
}

// NewTimeOfDay crea una instancia desde un time.Time extrayendo solo hora/minuto/segundo.
func NewTimeOfDay(t time.Time) TimeOfDay {
	h, m, s := t.Clock()
	return TimeOfDay{Hour: h, Minute: m, Second: s}
}

// ParseTimeOfDay crea una instancia desde un string "HH:MM:SS".
func ParseTimeOfDay(s string) (TimeOfDay, error) {
	var h, m, sec int
	_, err := fmt.Sscanf(s, "%d:%d:%d", &h, &m, &sec)
	if err != nil {
		return TimeOfDay{}, err
	}
	if h < 0 || h > 23 || m < 0 || m > 59 || sec < 0 || sec > 59 {
		return TimeOfDay{}, fmt.Errorf("invalid time: %s", s)
	}
	return TimeOfDay{Hour: h, Minute: m, Second: sec}, nil
}

// String devuelve la representación "HH:MM:SS".
func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}

// ToTimeOnDate construye un time.Time combinando esta hora con la fecha de day,
// en la zona horaria de day (la hora se interpreta como hora de pared en esa zona).
func (t TimeOfDay) ToTimeOnDate(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(),
		t.Hour, t.Minute, t.Second, 0, day.Location())
}

// Before informa si esta hora es anterior a other.
func (t TimeOfDay) Before(other TimeOfDay) bool {
	if t.Hour != other.Hour {
		return t.Hour < other.Hour
	}
	if t.Minute != other.Minute {
		return t.Minute < other.Minute
	}
	return t.Second < other.Second
}

// After informa si esta hora es posterior a other.
func (t TimeOfDay) After(other TimeOfDay) bool {
	return other.Before(t)
}

// Equal informa si ambas horas son iguales.
func (t TimeOfDay) Equal(other TimeOfDay) bool {
	return t.Hour == other.Hour &&
		t.Minute == other.Minute &&
		t.Second == other.Second
}

// Between informa si esta hora está en el intervalo [start, end] (ambos inclusive).
func (t TimeOfDay) Between(start, end TimeOfDay) bool {
	return (t.After(start) || t.Equal(start)) &&
		(t.Before(end) || t.Equal(end))
}
