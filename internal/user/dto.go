package user

import "github.com/MauricioLopezC/ReservationSystem/internal/model"

type CreateUserDto struct {
	Username string `json:"username" validate:"required,min=1,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Phone    string `json:"phone" validate:"required,e164"`
}

type metadata struct {
	TotalItems int64
}

type AllUsers struct {
	Data     []model.User
	Metadata metadata
}

type UpdateUserDto struct {
	Username *string `json:"username" validate:"omitempty,min=1,max=100"`
	Email    *string `json:"email" validate:"omitempty,email"`
	Password *string `json:"password" validate:"omitempty,min=1,max=100"`
	Phone    *string `json:"phone" validate:"omitempty,e164"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type ConnectServicesToUserDto struct {
	ServiceIDs []uint `json:"serviceIds" validate:"required,min=1,max=20"`
}
