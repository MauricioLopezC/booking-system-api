// Package slots
package slots

import (
	"errors"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/gofiber/fiber/v3"
)

type SlotsHandler struct {
	slotsService *SlotsService
}

func NewSlotsHandler(slotsService *SlotsService) *SlotsHandler {
	return &SlotsHandler{slotsService: slotsService}
}

func (h *SlotsHandler) FindAvailableSlots(c fiber.Ctx) error {
	serviceId := fiber.Query[int](c, "serviceId", 0)
	if serviceId <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "ServiceId must be an integer gt 0",
		})
	}

	employeeId := fiber.Query[int](c, "employeeId", 0)
	if employeeId <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "employeeId must be an integer gt 0",
		})
	}

	endDateQuery := c.Query("endDate")
	if endDateQuery == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Required query parameter: endDate",
		})
	}
	endDate, err := time.Parse(time.DateOnly, endDateQuery)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}
	startDate := time.Now()

	slots, err := h.slotsService.FindAvailableSlots(uint(employeeId), uint(serviceId), startDate, endDate)
	if err != nil {
		if errors.Is(err, common.ErrValidation) {
			return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusBadRequest,
				Message:    err.Error(),
			})
		}
		if errors.Is(err, common.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusConflict,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}
	return c.Status(fiber.StatusOK).JSON(slots)
}
