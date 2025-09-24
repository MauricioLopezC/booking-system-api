package model

import (
	"time"

	"gorm.io/gorm"
)

// NOTE: Roles: employee, customer, admin
type User struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex;not null" json:"name"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Phone    string `gorm:"uniqueIndex;not null" json:"phone"`
	Password string `gorm:"not null" json:"password"`
	Role     string `gorm:"not null;default:customer" json:"role"`

	Services []*Service `gorm:"many2many:user_services"`
}

type Service struct {
	gorm.Model
	Name            string
	DurationMinutes int     `gorm:"not null"`
	Price           float64 `gorm:"not null"`

	Users []*User `gorm:"many2many:user_services"`
}

// Status: 'booked', 'cancelled', 'completed'
// Creating unique index for EmployeeID and BookingDate
type Booking struct {
	gorm.Model
	CustomerID  uint      `gorm:"not null"`
	EmployeeID  uint      `gorm:"not null;uniqueIndex:idx_employee_bookingdate"`
	ServiceID   uint      `gorm:"not null"`
	BookingDate time.Time `gorm:"not null;uniqueIndex:idx_employee_bookingdate"`
	Status      string    `gorm:"not null"`
	Notes       string
	Customer    User    `gorm:"foreignKey:CustomerID"`
	Employee    User    `gorm:"foreignKey:EmployeeID"`
	Service     Service `gorm:"foreignKey:ServiceID"`
}

// Base schedule by week day
// ignore date for StartTime and EndTime, use only the time
type Schedule struct {
	gorm.Model
	EmployeeID uint      `gorm:"not null"`
	WeekDay    int       `gorm:"not null"`
	StartTime  time.Time `gorm:"not null;type:time"`
	EndTime    time.Time `gorm:"not null;type:time"`
	Employee   User      `gorm:"foreignKey:EmployeeID"`
}

// Used for schedule exceptions
type ScheduleOverride struct {
	gorm.Model
	EmployeeID  uint
	Date        time.Time `gorm:"not null"`
	Reason      string    `gorm:"not null"`
	AllDay      bool      `gorm:"not null;default:false"`
	StartTime   *time.Time
	EndTime     *time.Time
	IsAvailable bool `gorm:"default:false"`
	Employee    User `gorm:"foreignKey:EmployeeID"`
}

// Used to check daily availability, whem registering a shift,
// you must check if it is available with this table
type TimeSlot struct {
	gorm.Model
	EmployeeID  uint
	StartTime   time.Time // Date and time
	EndTime     time.Time // Date and time
	IsAvailable bool      `gorm:"default:true"`
	Employee    User      `gorm:"foreignKey:EmployeeID"`
}
