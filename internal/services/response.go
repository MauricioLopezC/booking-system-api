package services

import "github.com/MauricioLopezC/ReservationSystem/internal/model"

type metadata struct {
	TotalItems int64
}

type AllServicesResponse struct {
	Data     []model.Service
	Metadata metadata
}
