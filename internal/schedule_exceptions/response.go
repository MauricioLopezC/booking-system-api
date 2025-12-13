package scheduleexceptions

import "github.com/MauricioLopezC/ReservationSystem/internal/model"

type metadata struct {
	TotalItems int64
}

type AllSchedulesResponse struct {
	Data     []model.ScheduleOverride
	Metadata metadata
}
