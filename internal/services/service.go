package services

import (
	"errors"
	"fmt"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"gorm.io/gorm"
)

type ServicesService struct {
	db *gorm.DB
}

func NewServicesService(db *gorm.DB) *ServicesService {
	return &ServicesService{db}
}

func (ss *ServicesService) Create(serviceDto CreateServiceDto) (*model.Service, error) {

	service := model.Service{
		Name:            serviceDto.Name,
		DurationMinutes: serviceDto.DurationMinutes,
		Price:           serviceDto.Price,
	}

	if result := ss.db.Create(&service); result.Error != nil {
		return nil, result.Error
	}

	return &service, nil
}

func (ss *ServicesService) FindAll(limit, page int, orderBy string) (*AllServicesResponse, error) {
	offset := (page - 1) * limit
	query := ss.db.Model(&model.Service{})

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	if orderBy != "" {
		query = query.Order(orderBy)
	}

	services := []model.Service{}
	if err := query.Limit(limit).Offset(offset).Find(&services).Error; err != nil {
		return nil, err
	}

	return &AllServicesResponse{
		Data:     services,
		Metadata: metadata{TotalItems: count},
	}, nil
}

func (ss *ServicesService) FindOne(id uint) (*model.Service, error) {
	service := new(model.Service)
	if restult := ss.db.First(service, id); restult.Error != nil {
		if errors.Is(restult.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "booking")
		}
		return nil, restult.Error
	}

	return service, nil
}

func (ss *ServicesService) Update(id uint, serviceDto UpdateServiceDto) (*model.Service, error) {
	var service model.Service
	if restult := ss.db.First(&service, id); restult.Error != nil {
		if errors.Is(restult.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "service")
		}
		return nil, restult.Error
	}

	if serviceDto.Name != nil {
		service.Name = *serviceDto.Name
	}
	if serviceDto.DurationMinutes != nil {
		service.DurationMinutes = *serviceDto.DurationMinutes
	}
	if serviceDto.Price != nil {
		service.Price = *serviceDto.Price
	}

	if result := ss.db.Save(&service); result.Error != nil {
		return nil, result.Error
	}

	return &service, nil
}

func (ss *ServicesService) Delete(id uint) (*model.Service, error) {
	var service model.Service

	if result := ss.db.First(&service, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "service")
		}
		return nil, result.Error
	}

	if result := ss.db.Delete(&service); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "service")
		}
		return nil, result.Error
	}
	return &service, nil
}
