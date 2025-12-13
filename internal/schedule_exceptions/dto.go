package scheduleexceptions

type CreateScheduleExDto struct {
	Date      string  `json:"date" validate:"datetime=2006-01-02"`
	Reason    string  `json:"reason" validate:"required"`
	AllDay    bool    `json:"allDay" validate:"boolean"`
	StartTime *string `json:"startTime" validate:"omitempty,datetime=15:04:05"`
	EndTime   *string `json:"endTime" validate:"omitempty,datetime=15:04:05"`
}
