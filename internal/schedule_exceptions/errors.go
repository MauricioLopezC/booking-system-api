package scheduleexceptions

import (
	"fmt"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
)

// ConflictingBookingsError is returned when a blocking exception overlaps active bookings.
// They must be cancelled or rescheduled first: nobody would notify the customers otherwise.
type ConflictingBookingsError struct {
	Bookings []model.Booking // with Customer preloaded
}

func (e *ConflictingBookingsError) Error() string {
	return fmt.Sprintf("%s: the exception overlaps %d active booking(s), cancel or reschedule them first", common.ErrConflict, len(e.Bookings))
}

func (e *ConflictingBookingsError) Unwrap() error {
	return common.ErrConflict
}
