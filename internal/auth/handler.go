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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func checkPasswordHash(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func Login(c *fiber.Ctx) error {
	loginDto := new(LoginDto)

	if err := c.BodyParser(loginDto); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	if err := utils.Validate(loginDto); err != nil {
		return c.Status(err.Code).JSON(common.ErrorResponse{
			StatusCode: uint(err.Code),
			Message:    err.Message,
		})
	}

	//get user by email
	user := new(model.User)
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

	return c.Status(fiber.StatusOK).JSON(AuthResponse{
		Token: t,
		UserResponse: UserResponse{
			ID:       user.ID,
			Username: user.Name,
			Email:    user.Email,
			Role:     user.Role},
	})
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func Register(c *fiber.Ctx) error {
	userDto := new(CreateUserDto)

	if err := c.BodyParser(userDto); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	if err := utils.Validate(userDto); err != nil {
		return c.Status(err.Code).JSON(common.ErrorResponse{
			StatusCode: uint(err.Code),
			Message:    err.Message,
		})
	}
	//check if user already exists
	//TODO: check if phone number already exists
	existingUser := new(model.User)
	userResult := database.DB.Db.Where("email = ?", userDto.Email).First(&existingUser)
	if !errors.Is(userResult.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusConflict).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusConflict,
			Message:    "Username already exists!",
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
