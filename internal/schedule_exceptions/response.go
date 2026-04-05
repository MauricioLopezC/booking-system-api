package scheduleexceptions

import "github.com/MauricioLopezC/ReservationSystem/internal/model"

type metadata struct {
	TotalItems int64
}

type AllSchedules struct {
	Data     []model.ScheduleOverride
	Metadata metadata
}
