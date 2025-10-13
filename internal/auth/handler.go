package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/config"
	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(as *AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var loginDto LoginDto
	if err := c.BodyParser(loginDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	if err := utils.Validate(loginDto); err != nil {
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

func Login(c *fiber.Ctx) error {
	loginDto := new(LoginDto)

	if err := c.BodyParser(loginDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	if err := utils.Validate(loginDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Message,
		})
	}

	//get user by email
	var user model.User
	result := database.DB.Db.Where("email = ?", loginDto.Email).First(&user)
	if result.Error != nil {
		return c.Status(fiber.ErrUnauthorized.Code).JSON(common.ErrorResponse{
			StatusCode: uint(fiber.ErrUnauthorized.Code),
			Message:    "Wrong email",
		})
	}

	//check password
	if !checkPasswordHash(loginDto.Password, user.Password) {
		return c.Status(fiber.ErrUnauthorized.Code).JSON(common.ErrorResponse{
			StatusCode: uint(fiber.ErrUnauthorized.Code),
			Message:    "Wrong password",
		})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	//claims are token payload
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["user_id"] = user.ID
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(config.Config("SECRET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusInternalServerError,
			Message:    "Error: try again later",
		})
	}

	return c.Status(fiber.StatusOK).JSON(LoginResponse{
		Token: t,
		UserResponse: UserResponse{
			ID:       user.ID,
			Username: user.Name,
			Email:    user.Email,
			Role:     user.Role},
	})
}

func Register(c *fiber.Ctx) error {
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

	//check if user already exists
	existingUser := new(model.User)
	userResult := database.DB.Db.Where("email = ?", userDto.Email).First(&existingUser)
	if errors.Is(userResult.Error, gorm.ErrDuplicatedKey) {
		return c.Status(fiber.StatusConflict).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusConflict,
			Message:    "username, email or phone number already exists!",
		})
	}

	hash, err := hashPassword(userDto.Password)
	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(common.ErrorResponse{
			StatusCode: 500,
			Message:    "Couldn't hash password",
		})
	}

	userDto.Password = hash

	newUser := model.User{
		Name:     strings.ToLower(userDto.Username),
		Email:    userDto.Email,
		Password: userDto.Password,
		Phone:    userDto.Phone,
	}

	result := database.DB.Db.Create(&newUser)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusInternalServerError,
			Message:    "Error creating user!",
		})
	}

	return c.Status(201).JSON(UserResponse{
		ID:       newUser.ID,
		Username: newUser.Name,
		Email:    newUser.Email,
		Role:     newUser.Role,
	})

}
