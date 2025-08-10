package bookings

import "time"

type CreateBookingDto struct {
	EmployeeID  uint      `json:"employeeId" validate:"gt=0"`
	ServiceID   uint      `json:"serviceId" validate:"gt=0"`
	BookingDate time.Time `json:"bookingDate" validate:"required"`
	Notes       string    `json:"notes" validate:"max=250"`
}

type UpdateBookingDto struct {
}
