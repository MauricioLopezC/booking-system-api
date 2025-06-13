package bookings

import "time"

type CreateBookingDto struct {
	EmployeeID  int       `json:"employeeId" validate:"gt=0"`
	ServiceID   int       `json:"serviceId" validate:"gt=0"`
	BookingDate time.Time `json:"bookingDate" validate:"required"`
	Notes       string    `json:"notes" validate:"max=250"`
}
