package schedule

import (
	"errors"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type ScheduleHandler struct {
	scheduleService *ScheduleService
}

func NewScheduleHandler(scheduleService *ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}

}

func (h *ScheduleHandler) CreateCurrentUserSchedule(c *fiber.Ctx) error {
	scheduleDto := new(CreateScheduleDto)

	if err := c.BodyParser(scheduleDto); err != nil {
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

	schedule, err := h.scheduleService.CreateSchedule(authUser.ID, *scheduleDto)
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
		//other errors
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusCreated).JSON(schedule)
}

// admin only
func (h *ScheduleHandler) FindAll(c *fiber.Ctx) error {
	limit := c.QueryInt("limit")
	if limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := c.QueryInt("page")
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

func (h *ScheduleHandler) FindOne(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	schedule, err := h.scheduleService.FindOneSchedule(uint(id))
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

// Use in endpoint: GET /me/schedules
// there is no need for paginate because employee needs to see all his schedules
func (h *ScheduleHandler) CurrentUserFindAll(c *fiber.Ctx) error {
	authUser, err := utils.GetAuthUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Invalid auth token",
		})
	}

	schedules, err := h.scheduleService.FindSchedulesByEmployee(uint(authUser.ID))
	if err != nil {
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(schedules)
}

func (h *ScheduleHandler) CurrentUserDelete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
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

	schedule, err := h.scheduleService.DeleteScheduleByUser(uint(id), authUser.ID)
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
