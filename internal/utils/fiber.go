package utils

import (
	"fmt"
	"reflect"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/gofiber/fiber/v2"
)

// REFACTOR
func ParseBodyAndValidate(c *fiber.Ctx, body any) error {
	if err := c.BodyParser(body); err != nil {
		fmt.Println("FALLA")
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	fmt.Printf("kind=%s, isNil=%v, value=%#v\n", reflect.ValueOf(body).Kind(), reflect.ValueOf(body).IsNil(), body)
	if err := Validate(body); err != nil {
		return c.Status(err.Code).JSON(common.ErrorResponse{
			StatusCode: uint(err.Code),
			Message:    err.Message,
		})
	}

	return nil
}
