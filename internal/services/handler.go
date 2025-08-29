package services

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// Use in POST /service endpoint with required roles: admin
func CreateService(c *fiber.Ctx) error {
	serviceDto := new(CreateServiceDto)

	if err := utils.ParseBodyAndValidate(c, serviceDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Message,
		})
	}

	service := model.Service{
		Name:            serviceDto.Name,
		DurationMinutes: serviceDto.DurationMinutes,
		Price:           serviceDto.Price,
	}

	if result := database.DB.Db.Create(service); result.Error != nil {
		return common.NewInternalServerErrorResponse(c)
	}
	return c.Status(fiber.StatusOK).JSON(serviceDto)
}

// Use in GET /services endpoint, public route, use employeeId optional
// param to get the employee services
func FindAllServices(c *fiber.Ctx) error {
	limit := c.QueryInt("limit")
	if limit <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := c.QueryInt("page")
	if page <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	offset := (page - 1) * limit
	query := database.DB.Db.Limit(limit).Offset(offset)

	orderByQuery := c.Query("orderBy")
	allowedFields := map[string]bool{
		"price":            true,
		"duration_minutes": true,
		"created_at":       true,
	}
	orderBy, err := utils.ParseSortParam(orderByQuery, allowedFields)
	if err == nil {
		query.Order(orderBy)
	}

	services := []model.Service{}
	query.Find(&services)

	var count int64
	database.DB.Db.Model(&model.Service{}).Count(&count)
	return c.Status(fiber.StatusOK).JSON(AllServicesResponse{
		Data:     services,
		Metadata: metadata{TotalItems: count},
	})
}

func UpdateService() {

}

func FindServicesByEmployee() {

}
