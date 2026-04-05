// Package scheduleexceptions
package scheduleexceptions

import (
	"errors"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v3"
)

type ScheduleExHandler struct {
	scheduleExService *ScheduleExService
}

func NewScheduleExHandler(ss *ScheduleExService) *ScheduleExHandler {
	return &ScheduleExHandler{scheduleExService: ss}
}

// CreateMyScheduleEx godoc
// @Summary      Create a schedule exception
// @Tags         schedule-exceptions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateScheduleExDto   true  "Schedule exception data"
// @Success      201   {object}  model.ScheduleOverrideResponse
// @Failure      400   {object}  common.ErrorResponse
// @Failure      401   {object}  common.ErrorResponse
// @Failure      409   {object}  common.ErrorResponse
// @Router       /me/schedule-exceptions [post]
func (h *ScheduleExHandler) CreateMyScheduleEx(c fiber.Ctx) error {
	scheduleDto := new(CreateScheduleExDto)

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

	schedule, err := h.scheduleExService.Create(authUser.ID, *scheduleDto)
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

func (h *ScheduleExHandler) FindAll(c fiber.Ctx) error {
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
		"date":       true,
		"created_at": true,
	}
	orderBy, _ = utils.ParseSortParam(orderBy, allowedFields)

	allSchedueles, _ := h.scheduleExService.FindAll(limit, page, orderBy)

	return c.Status(fiber.StatusOK).JSON(model.AllScheduleOverridesResponse{
		Data:     model.ScheduleOverridesToResponse(allSchedueles.Data),
		Metadata: model.PaginationMetadata{TotalItems: allSchedueles.Metadata.TotalItems},
	})
}

func (h *ScheduleExHandler) FindMyScheduleExs(c fiber.Ctx) error {
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
		"date":       true,
		"created_at": true,
	}
	orderBy, _ = utils.ParseSortParam(orderBy, allowedFields)

	authUser, err := utils.GetAuthUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Invalid auth token",
		})
	}

	schedules, _ := h.scheduleExService.FindByUserID(limit, page, orderBy, authUser.ID)
	return c.Status(fiber.StatusOK).JSON(model.AllScheduleOverridesResponse{
		Data:     model.ScheduleOverridesToResponse(schedules.Data),
		Metadata: model.PaginationMetadata{TotalItems: schedules.Metadata.TotalItems},
	})
}

// DeleteMyScheduleEx godoc
// @Summary      Delete a schedule exception
// @Tags         schedule-exceptions
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Schedule exception ID"
// @Success      200  {object}  model.ScheduleOverrideResponse
// @Failure      400  {object}  common.ErrorResponse
// @Failure      401  {object}  common.ErrorResponse
// @Failure      404  {object}  common.ErrorResponse
// @Router       /me/schedule-exceptions/{id} [delete]
func (h *ScheduleExHandler) DeleteMyScheduleEx(c fiber.Ctx) error {
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

	deletedScheduleEx, err := h.scheduleExService.DeleteByUserID(uint(id), authUser.ID)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(deletedScheduleEx.ToResponse())
}
