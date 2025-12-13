package auth

import (
	"errors"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(as *AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var loginDto LoginDto
	if err := c.BodyParser(&loginDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	if err := utils.Validate(&loginDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Message,
		})
	}

	loginResponse, err := h.authService.Login(loginDto)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		if errors.Is(err, common.ErrUnauthorized) {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusUnauthorized,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(loginResponse)
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	userDto := new(CreateUserDto)

	if err := c.BodyParser(userDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	if err := utils.Validate(userDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Message,
		})
	}

	newUser, err := h.authService.Register(*userDto)
	if err != nil {
		if errors.Is(err, common.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusConflict,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusCreated).JSON(UserResponse{
		ID:       newUser.ID,
		Username: newUser.Name,
		Email:    newUser.Email,
		Role:     newUser.Role,
	})
}
