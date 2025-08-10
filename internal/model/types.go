package model

// USER ROLE
type UserRole int

const (
	RoleAdmin UserRole = iota
	RoleCustomer
	RoleEmployee
)

var roleName = map[UserRole]string{
	RoleAdmin:    "admin",
	RoleCustomer: "customer",
	RoleEmployee: "employee",
}

func (ur UserRole) String() string {
	return roleName[ur]
}

// BOOKING STATUS
type BookingStatus int

const (
	StatusBooked BookingStatus = iota
	StatusCancelled
	StatusCompleted
)

var statusName = map[BookingStatus]string{
	StatusBooked:    "booked",
	StatusCancelled: "cancelled",
	StatusCompleted: "completed",
}

func (bs BookingStatus) String() string {
	return statusName[bs]
}
