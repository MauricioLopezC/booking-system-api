package schedule

// TODO: change StartTime and EndTime for string type and validate with datetime tag
// then check createSchedule function
type CreateScheduleDto struct {
	WeekDay   int    `json:"weekday" validate:"gte=0,lte=6"`
	StartTime string `json:"startTime" validate:"required,datetime=15:04:05"`
	EndTime   string `json:"endTime" validate:"required,datetime=15:04:05"`
}
