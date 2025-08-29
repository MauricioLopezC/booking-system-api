package common

import "github.com/gofiber/fiber/v2"

type ErrorResponse struct {
	StatusCode uint
	Message    string
}

func NewInternalServerErrorResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		StatusCode: fiber.StatusInternalServerError,
		Message:    "Internal server error",
	})
}
