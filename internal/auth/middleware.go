package auth

import (
	"slices"

	"github.com/MauricioLopezC/ReservationSystem/config"
	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
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

// This middleware should be used after Protected middleware
// in order to excract jwt claims
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		//get user info
		user := c.Locals("user").(*jwt.Token)
		claims, ok := user.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusUnauthorized,
				Message:    "Invalid token claims",
			})
		}

		role, ok := claims["role"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusUnauthorized,
				Message:    "Role not found in token",
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

// This middleware should be used in endpoints with `/users/:id/resource`
// so that the user can only access their resources. The admin has full access
func UserOwnResourceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		//Get logged-in user data
		userToken, ok1 := c.Locals("user").(*jwt.Token)
		claims, ok2 := userToken.Claims.(jwt.MapClaims)
		userId, ok3 := claims["user_id"].(float64)
		userRole, ok4 := claims["role"].(string)

		if !ok1 || !ok2 || !ok3 || !ok4 {
			return common.NewInternalServerErrorResponse(c)
		}

		//Get param id
		paramUserId, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusBadRequest,
				Message:    "Numeric string is expected for user id",
			})
		}

		if userRole == model.RoleAdmin.String() {
			return c.Next()
		}

		if int(userId) == paramUserId {
			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Access denied, insufficient permissions!",
		})

	}
}
