package bookings

import "github.com/MauricioLopezC/ReservationSystem/internal/model"

type metadata struct {
	TotalItems int64
}
type AllBookings struct {
	Data     []model.Booking
	Metadata metadata
}
