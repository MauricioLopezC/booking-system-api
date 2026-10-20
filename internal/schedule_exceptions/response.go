package scheduleexceptions

import (
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
)

type metadata struct {
	TotalItems int64
}

type AllSchedules struct {
	Data     []model.ScheduleOverride
	Metadata metadata
}

// BookingConflict an active booking that blocks the creation of an exception.
type BookingConflict struct {
	BookingID     uint      `json:"bookingId"`
	BookingDate   time.Time `json:"bookingDate"`
	EndDate       time.Time `json:"endDate"`
	CustomerName  string    `json:"customerName"`
	CustomerPhone string    `json:"customerPhone"`
}

// ConflictResponse 409 body listing the bookings to cancel or reschedule first.
type ConflictResponse struct {
	common.ErrorResponse
	Conflicts []BookingConflict `json:"conflicts"`
}

func newConflictResponse(e *ConflictingBookingsError) ConflictResponse {
	conflicts := make([]BookingConflict, len(e.Bookings))
	for i, b := range e.Bookings {
		conflicts[i] = BookingConflict{
			BookingID:     b.ID,
			BookingDate:   b.BookingDate,
			EndDate:       b.EndDate,
			CustomerName:  b.Customer.Name,
			CustomerPhone: b.Customer.Phone,
		}
	}
	return ConflictResponse{
		ErrorResponse: common.ErrorResponse{StatusCode: 409, Message: e.Error()},
		Conflicts:     conflicts,
	}
}
