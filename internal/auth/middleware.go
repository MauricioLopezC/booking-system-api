package auth

import (
	"slices"

	"github.com/MauricioLopezC/ReservationSystem/config"
	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected middleware
func Protected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(config.Config("SECRET"))},
		ErrorHandler: jwtError,
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Missing or malformed JWT",
		})
	}
	return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
		StatusCode: fiber.StatusUnauthorized,
		Message:    "Invalid or expired JWT",
	})
}

// Use this middleware after the Protected middleware to extract the JWT claims.
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		//get user info
		user, ok1 := c.Locals("user").(*jwt.Token)
		claims, ok2 := user.Claims.(jwt.MapClaims)
		if !ok1 || !ok2 {
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

		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Access denied, insufficient permissions!",
		})
	}
}
