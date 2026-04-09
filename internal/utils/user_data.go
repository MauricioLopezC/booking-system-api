package utils

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type AuthUser struct {
	ID   uint
	Role string
}

func GetAuthUser(c fiber.Ctx) (*AuthUser, error) {
	// get user data from context
	userToken := jwtware.FromContext(c)
	if userToken == nil {
		return nil, common.ErrUnauthorized
	}

	claims, ok2 := userToken.Claims.(jwt.MapClaims)
	idF, ok3 := claims["user_id"].(float64)
	role, ok4 := claims["role"].(string)

	if !ok2 || !ok3 || !ok4 {
		return nil, common.ErrUnauthorized
	}

	return &AuthUser{
		ID:   uint(idF),
		Role: role,
	}, nil
}
