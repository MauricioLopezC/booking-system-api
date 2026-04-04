// Package schedule
package schedule

import (
	"errors"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v3"
)

type ScheduleHandler struct {
	scheduleService *ScheduleService
}

func NewScheduleHandler(scheduleService *ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

func (h *ScheduleHandler) CreateMySchedule(c fiber.Ctx) error {
	scheduleDto := new(CreateScheduleDto)

	if err := c.Bind().Body(scheduleDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{StatusCode: fiber.StatusBadRequest, Message: err.Error()})
	}

	if err := utils.Validate(scheduleDto); err != nil {
		return c.Status(err.Code).JSON(common.ErrorResponse{
			StatusCode: uint(err.Code),
			Message:    err.Message,
		})
	}

	authUser, err := utils.GetAuthUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Invalid auth token",
		})
	}

	schedule, err := h.scheduleService.Create(authUser.ID, *scheduleDto)
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
		// other errors
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusCreated).JSON(schedule)
}

// FindAll admin only
func (h *ScheduleHandler) FindAll(c fiber.Ctx) error {
	limit := fiber.Query[int](c, "limit")
	if limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := fiber.Query[int](c, "page")
	if page <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "page must be an integer number and gt 0",
		})
	}

	orderBy := c.Query("orderBy")
	allowedFields := map[string]bool{
		"updated_at": true,
		"created_at": true,
	}
	orderBy, _ = utils.ParseSortParam(orderBy, allowedFields)

	allSchedueles, _ := h.scheduleService.FindAll(limit, page, orderBy)
	return c.Status(fiber.StatusOK).JSON(allSchedueles)
}

// FindOne Admin only
func (h *ScheduleHandler) FindOne(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	schedule, err := h.scheduleService.FindOne(uint(id))
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}
	return c.Status(fiber.StatusOK).JSON(schedule)
}

// FindMySchedules use in endpoint: GET /me/schedules
// there is no need for paginate because employee needs to see all his schedules
func (h *ScheduleHandler) FindMySchedules(c fiber.Ctx) error {
	authUser, err := utils.GetAuthUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Invalid auth token",
		})
	}

	schedules, err := h.scheduleService.FindByUserId(uint(authUser.ID))
	if err != nil {
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(schedules)
}

func (h *ScheduleHandler) DeleteMySchedules(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	authUser, err := utils.GetAuthUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Invalid auth token",
		})
	}

	schedule, err := h.scheduleService.DeleteByUserId(uint(id), authUser.ID)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(schedule)
}
