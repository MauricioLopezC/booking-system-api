package user

import (
	"errors"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v3"
)

type UsersHandler struct {
	usersService *UsersService
}

func NewUsersHandler(s *UsersService) *UsersHandler {
	return &UsersHandler{usersService: s}
}

func (h *UsersHandler) FindAll(c fiber.Ctx) error {
	limit := fiber.Query[int](c, "limit", 0)
	if limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := fiber.Query[int](c, "page", 0)
	if page <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "page must be an integer number and gt 0",
		})
	}

	orderBy := c.Query("orderBy")
	allowedFields := map[string]bool{
		"booking_date": true,
		"created_at":   true,
	}
	orderBy, _ = utils.ParseSortParam(orderBy, allowedFields)

	usersResponse, err := h.usersService.FindAll(limit, page, orderBy)
	if err != nil {
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(usersResponse)
}

func (h *UsersHandler) FindOne(c fiber.Ctx) error {
	id, err := fiber.Params[int](c, "id"), error(nil)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	user, err := h.usersService.FindOne(uint(id))
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	userResponse := UserResponse{
		ID:       user.ID,
		Username: user.Name,
		Email:    user.Email,
		Role:     user.Role,
	}

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

func (h *UsersHandler) Update(c fiber.Ctx) error {
	id, err := fiber.Params[int](c, "id"), error(nil)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	userDto := new(UpdateUserDto)
	if err := c.Bind().Body(userDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	user, err := h.usersService.Update(uint(id), *userDto)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
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

	userResponse := UserResponse{
		ID:       user.ID,
		Username: user.Name,
		Email:    user.Email,
		Role:     user.Role,
	}

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

func (h *UsersHandler) Delete(c fiber.Ctx) error {
	id, err := fiber.Params[int](c, "id"), error(nil)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	user, err := h.usersService.Delete(uint(id))
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	userResponse := UserResponse{
		ID:       user.ID,
		Username: user.Name,
		Email:    user.Email,
		Role:     user.Role,
	}

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

func (h *UsersHandler) AssignServiceToMe(c fiber.Ctx) error {
	var body ConnectServicesToUserDto
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
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

	user, err := h.usersService.AssignServicesToUser(authUser.ID, body.ServiceIDs)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
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

	return c.Status(fiber.StatusCreated).JSON(user)
}
