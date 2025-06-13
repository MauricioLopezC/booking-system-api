package schedule

import "time"

type CreateScheduleDto struct {
	WeekDay   int       `json:"weekday" validate:"gt=0"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// CreateScheduleDto for admin with EmployeeID, to allow admin create schedles
// type CreateScheduleDtoAdmin struct {
// 	EmployeeID int       `json:"employeeId"`
// 	WeekDay    int       `json:"weekday"`
// 	StartTime  time.Time `json:"startTime"`
// 	EndTime    time.Time `json:"endTime"`
// }
