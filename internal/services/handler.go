// Package services
package services

import (
	"errors"
	"fmt"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v3"
)

type ServicesHandler struct {
	servicesService *ServicesService
}

func NewServicesHandler(servicesService *ServicesService) *ServicesHandler {
	return &ServicesHandler{servicesService: servicesService}
}

// Use in POST /service endpoint with required roles: admin
func (h *ServicesHandler) Create(c fiber.Ctx) error {
	serviceDto := new(CreateServiceDto)

	if err := c.Bind().Body(serviceDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	if err := utils.Validate(serviceDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	service, err := h.servicesService.Create(*serviceDto)
	if err != nil {
		// unexpected error
		fmt.Print(err.Error())
		return common.NewInternalServerErrorResponse(c)
	}
	return c.Status(fiber.StatusOK).JSON(service)
}

// Use in GET /services endpoint, public route, use employeeId optional
// param to get the employee services
func (h *ServicesHandler) FindAll(c fiber.Ctx) error {
	limit := fiber.Query[int](c, "limit")
	if limit <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := fiber.Query[int](c, "page")
	if page <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	orderBy := c.Query("orderBy")
	allowedFields := map[string]bool{
		"booking_date": true,
		"created_at":   true,
	}
	orderBy, _ = utils.ParseSortParam(orderBy, allowedFields)

	allServicesResponse, err := h.servicesService.FindAll(limit, page, orderBy)
	if err != nil {
		return common.NewInternalServerErrorResponse(c)
	}
	return c.Status(fiber.StatusOK).JSON(allServicesResponse)
}

func (h *ServicesHandler) Update(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	var serviceDto UpdateServiceDto
	if err := c.Bind().Body(&serviceDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	updatedService, err := h.servicesService.Update(uint(id), serviceDto)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(updatedService)
}

func (h *ServicesHandler) Delete(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	deletedService, err := h.servicesService.Delete(uint(id))
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}
	return c.Status(fiber.StatusOK).JSON(deletedService)
}

//TODO: find user services, connect service with employee
//
