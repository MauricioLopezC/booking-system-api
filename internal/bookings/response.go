package bookings

import "github.com/MauricioLopezC/ReservationSystem/internal/model"

type metadata struct {
	TotalItems int64
}
type AllBookingsResponse struct {
	Data     []model.Booking
	Metadata metadata
}
