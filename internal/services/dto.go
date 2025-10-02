package services

type CreateServiceDto struct {
	Name            string  `json:"name" validate:"required,min=1,max=100"`
	DurationMinutes int     `json:"durationMinutes" validate:"required,gt=0,lte=360"`
	Price           float64 `json:"price" validate:"required,gt=0"`
}

type UpdateServiceDto struct {
	Name            *string  `json:"name" validate:"omitempty,min=1,max=100"`
	DurationMinutes *int     `json:"durationMinutes" validate:"omitempty,gt=0,lte=360"`
	Price           *float64 `json:"price" validate:"omitempty,gt=0"`
}
