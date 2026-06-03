package auth

import (
	"slices"

	"github.com/MauricioLopezC/ReservationSystem/config"
	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Protected middleware
func Protected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(config.Config("SECRET"))},
		ErrorHandler: jwtError,
	})
}

func jwtError(c fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Missing or malformed JWT",
		})
	}
	return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
		StatusCode: fiber.StatusUnauthorized,
		Message:    "Invalid or expired JWT",
	})
}

// RoleMiddleware Use this middleware after the Protected middleware to extract the JWT claims.
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// get user info
		user := jwtware.FromContext(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusUnauthorized,
				Message:    "Invalid auth token",
			})
		}
		claims, ok2 := user.Claims.(jwt.MapClaims)
		if !ok2 || claims == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusUnauthorized,
				Message:    "Invalid auth token",
			})
		}

		role, ok := claims["role"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusUnauthorized,
				Message:    "Invalid auth token",
			})
		}

		if slices.Contains(allowedRoles, role) {
			return c.Next()
		}

		return c.Status(fiber.StatusForbidden).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusForbidden,
			Message:    "Access denied, insufficient permissions!",
		})
	}
}
