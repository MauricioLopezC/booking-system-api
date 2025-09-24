package schedule

import (
	"errors"
	"strconv"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type ScheduleHandler struct {
	scheduleService *ScheduleService
}

func NewScheduleHandler(scheduleService *ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}

}

func (h *ScheduleHandler) CreateSchedule(c *fiber.Ctx) error {
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

	//get user info
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))

	schedule, err := h.scheduleService.CreateSchedule(userId, *scheduleDto)
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

func (h *ScheduleHandler) FindAllSchedules(c *fiber.Ctx) error {
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
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
	orderBy, err = utils.ParseSortParam(orderBy, allowedFields)

	allSchedueles, err := h.scheduleService.FindAllSchedules(limit, page, orderBy)
	return c.Status(fiber.StatusOK).JSON(allSchedueles)
}

func (h *ScheduleHandler) FindOneSchedule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
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

// Employee gets his weekly schedule.
// Use in endpoint: GET /me/schedules
// there is no need for paginate because employee needs to see all his schedules
func (h *ScheduleHandler) FindSchedulesByEmployee(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	if role == model.RoleCustomer.String() {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Access denied! only employees can se his schedules",
		})
	}

	schedules, err := h.scheduleService.FindSchedulesByEmployee(userId)
	if err != nil {
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(schedules)
}

func (h *ScheduleHandler) DeleteSchedule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))

	hasSchedule, err := common.EmployeeHasSchedule(database.DB.Db, userId, uint(id))
	if err != nil {
		return common.NewInternalServerErrorResponse(c)
	}
	if !hasSchedule {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Access denied! you can only delete your schedules!",
		})
	}

	schedule, err := h.scheduleService.DeleteSchedule(uint(id))
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
