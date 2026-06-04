// Package schedule
package schedule

import (
	"errors"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v3"
)

type ScheduleHandler struct {
	scheduleService *ScheduleService
}

func NewScheduleHandler(scheduleService *ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

// CreateMySchedule godoc
// @Summary      Create a schedule entry
// @Tags         schedules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateScheduleDto     true  "Schedule data"
// @Success      201   {object}  model.ScheduleResponse
// @Failure      400   {object}  common.ErrorResponse
// @Failure      401   {object}  common.ErrorResponse
// @Failure      409   {object}  common.ErrorResponse
// @Router       /me/schedules [post]
func (h *ScheduleHandler) CreateMySchedule(c fiber.Ctx) error {
	scheduleDto := new(CreateScheduleDto)

	if err := c.Bind().Body(scheduleDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{StatusCode: fiber.StatusBadRequest, Message: err.Error()})
	}

	if err := utils.Validate(scheduleDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: uint(fiber.StatusBadRequest),
			Message:    err.Error(),
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

	return c.Status(fiber.StatusCreated).JSON(schedule.ToResponse())
}

// FindAll godoc
// @Summary      Get all schedules (admin)
// @Tags         schedules
// @Produce      json
// @Security     BearerAuth
// @Param        limit    query     int     true   "Items per page"
// @Param        page     query     int     true   "Page number"
// @Param        orderBy  query     string  false  "Sort field: updated_at | created_at"
// @Success      200  {object}  model.AllSchedulesResponse
// @Failure      400  {object}  common.ErrorResponse
// @Failure      401  {object}  common.ErrorResponse
// @Router       /schedules [get]
// FindAll admin only
func (h *ScheduleHandler) FindAll(c fiber.Ctx) error {
	limit := fiber.Query(c, "limit", 10)
	if limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := fiber.Query(c, "page", 1)
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

	allSchedueles, err := h.scheduleService.FindAll(limit, page, orderBy)
	if err != nil {
		return common.NewInternalServerErrorResponse(c)
	}
	return c.Status(fiber.StatusOK).JSON(model.AllSchedulesResponse{
		Data:     model.SchedulesToResponse(allSchedueles.Data),
		Metadata: model.PaginationMetadata{TotalItems: allSchedueles.Metadata.TotalItems},
	})
}

// FindOne godoc
// @Summary      Get schedule by ID (admin)
// @Tags         schedules
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Schedule ID"
// @Success      200  {object}  model.ScheduleResponse
// @Failure      400  {object}  common.ErrorResponse
// @Failure      401  {object}  common.ErrorResponse
// @Failure      404  {object}  common.ErrorResponse
// @Router       /schedules/{id} [get]
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
	return c.Status(fiber.StatusOK).JSON(schedule.ToResponse())
}

// FindMySchedules godoc
// @Summary      Get current user schedules
// @Tags         schedules
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.ScheduleResponse
// @Failure      401  {object}  common.ErrorResponse
// @Router       /me/schedules [get]
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

	schedules, err := h.scheduleService.FindByUserID(uint(authUser.ID))
	if err != nil {
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(model.SchedulesToResponse(schedules))
}

// DeleteMySchedules godoc
// @Summary      Delete a schedule entry
// @Tags         schedules
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Schedule ID"
// @Success      200  {object}  model.ScheduleResponse
// @Failure      400  {object}  common.ErrorResponse
// @Failure      401  {object}  common.ErrorResponse
// @Failure      404  {object}  common.ErrorResponse
// @Router       /me/schedules/{id} [delete]
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

	schedule, err := h.scheduleService.DeleteByUserID(uint(id), authUser.ID)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(schedule.ToResponse())
}
