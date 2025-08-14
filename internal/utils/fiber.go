package utils

import (
	"github.com/gofiber/fiber/v2"
)

// ParseBodyAndValidate first binds the request body to a struct using fiber c.BodyParser
// then validates struct fields using validator
func ParseBodyAndValidate(c *fiber.Ctx, body any) *fiber.Error {
	if err := c.BodyParser(body); err != nil {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	if err := Validate(body); err != nil {
		return err
	}

	return nil
}
