package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex;not null" json:"name"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Phone    string `gorm:"uniqueIndex;not null" json:"phone"`
	Password string `gorm:"not null" json:"password"`
	Role     string `gorm:"not null;default:customer" json:"role"`
}

// NOTE: Roles: employee, customer, admin
type Service struct {
	gorm.Model
	Name            string
	DurationMinutes int     `gorm:"not null"`
	Price           float64 `gorm:"not null"`
}

type Booking struct {
	gorm.Model
	CustomerID  int
	EmployeeID  int
	ServiceID   int
	BookingDate time.Time
	Status      string `gorm:"not null"`
	Notes       string
	Customer    User    `gorm:"foreignKey:CustomerID"`
	Employee    User    `gorm:"foreignKey:EmployeeID"`
	Service     Service `gorm:"foreignKey:ServiceID"`
}

// Ignore date for StartTime and EndTime, use only the time
type Schedule struct {
	gorm.Model
	EmployeeID int       `gorm:"not nill"`
	WeekDay    int       `gorm:"not null"`
	StartTime  time.Time `gorm:"not null"`
	EndTime    time.Time `gorm:"not null"`
	Employee   User      `gorm:"foreignKey:EmployeeID"`
}

type ScheduleException struct {
	gorm.Model
	EmployeeID    int
	ExceptionDate time.Time `gorm:"not null"`
	Reason        string
	AllDay        bool `gorm:"not null;default:false"`
	StartTime     time.Time
	Endtime       time.Time
	Employee      User `gorm:"foreignKey:EmployeeID"`
}
