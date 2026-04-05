package model

import "time"

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
}

// ServiceResponse represents a service in API responses.
type ServiceResponse struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Name            string    `json:"name"`
	DurationMinutes int       `json:"durationMinutes"`
	Price           float64   `json:"price"`
}

// BookingResponse represents a booking in API responses.
type BookingResponse struct {
	ID          uint            `json:"id"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	CustomerID  uint            `json:"customerId"`
	EmployeeID  uint            `json:"employeeId"`
	ServiceID   uint            `json:"serviceId"`
	BookingDate time.Time       `json:"bookingDate"`
	Status      string          `json:"status"`
	Notes       string          `json:"notes"`
	Customer    UserResponse    `json:"customer"`
	Employee    UserResponse    `json:"employee"`
	Service     ServiceResponse `json:"service"`
}

// ScheduleResponse represents a schedule entry in API responses.
type ScheduleResponse struct {
	ID         uint         `json:"id"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	EmployeeID uint         `json:"employeeId"`
	WeekDay    int          `json:"weekDay"`
	StartTime  time.Time    `json:"startTime"`
	EndTime    time.Time    `json:"endTime"`
	Employee   UserResponse `json:"employee"`
}

// ScheduleOverrideResponse represents a schedule exception in API responses.
type ScheduleOverrideResponse struct {
	ID          uint         `json:"id"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	EmployeeID  uint         `json:"employeeId"`
	Date        time.Time    `json:"date"`
	Reason      string       `json:"reason"`
	AllDay      bool         `json:"allDay"`
	StartTime   *time.Time   `json:"startTime,omitempty"`
	EndTime     *time.Time   `json:"endTime,omitempty"`
	IsAvailable bool         `json:"isAvailable"`
	Employee    UserResponse `json:"employee"`
}

// TimeSlotResponse represents a time slot in API responses.
type TimeSlotResponse struct {
	ID          uint         `json:"id"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	EmployeeID  uint         `json:"employeeId"`
	StartTime   time.Time    `json:"startTime"`
	EndTime     time.Time    `json:"endTime"`
	IsAvailable bool         `json:"isAvailable"`
	Employee    UserResponse `json:"employee"`
}

// PaginationMetadata holds pagination info for paginated responses.
type PaginationMetadata struct {
	TotalItems int64 `json:"totalItems"`
}

// AllBookingsSwaggerResponse is the paginated bookings response.
type AllBookingsSwaggerResponse struct {
	Data     []BookingResponse  `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

// AllSchedulesSwaggerResponse is the paginated schedules response.
type AllSchedulesSwaggerResponse struct {
	Data     []ScheduleResponse `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

// AllServicesSwaggerResponse is the paginated services response.
type AllServicesSwaggerResponse struct {
	Data     []ServiceResponse  `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

// AllScheduleOverridesSwaggerResponse is the paginated schedule exceptions response.
type AllScheduleOverridesSwaggerResponse struct {
	Data     []ScheduleOverrideResponse `json:"data"`
	Metadata PaginationMetadata         `json:"metadata"`
}

// AllUsersSwaggerResponse is the paginated users response.
type AllUsersSwaggerResponse struct {
	Data     []UserResponse     `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}
