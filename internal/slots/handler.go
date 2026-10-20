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

// FindAvailableSlots godoc
// @Summary      Get available time slots
// @Tags         slots
// @Produce      json
// @Param        serviceId   query     int     true   "Service ID"
// @Param        employeeId  query     int     true   "Employee ID"
// @Param        endDate     query     string  true   "End date (YYYY-MM-DD)"
// @Success      200  {object}  object  "Map of date (YYYY-MM-DD) to available time slots"
// @Failure      400  {object}  common.ErrorResponse
// @Failure      409  {object}  common.ErrorResponse
// @Router       /available-slots [get]
func (h *SlotsHandler) FindAvailableSlots(c fiber.Ctx) error {
	serviceID := fiber.Query[int](c, "serviceId", 0)
	if serviceID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "ServiceId must be an integer gt 0",
		})
	}

	employeeID := fiber.Query[int](c, "employeeId", 0)
	if employeeID <= 0 {
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
	endDate, err := time.ParseInLocation(time.DateOnly, endDateQuery, h.slotsService.Location())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}
	startDate := time.Now()

	slots, err := h.slotsService.FindAvailableSlots(uint(employeeID), uint(serviceID), startDate, endDate)
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
