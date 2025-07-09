package schedule

import "github.com/MauricioLopezC/ReservationSystem/internal/model"

type CreateScheduleResponse struct {
}

type metadata struct {
	TotalItems int64
}
type AllSchedulesResponse struct {
	Data     []model.Schedule
	Metadata metadata
}
