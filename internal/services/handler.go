package services

import (
	"fmt"

	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// Use in POST /service endpoint with required roles: admin
func CreateService(c *fiber.Ctx) error {
	serviceDto := new(CreateServiceDto)

	// if err := c.BodyParser(serviceDto); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
	// 		StatusCode: fiber.StatusBadRequest,
	// 		Message:    err.Error(),
	// 	})
	// }
	//
	// if err := utils.Validate(serviceDto); err != nil {
	// 	return c.Status(err.Code).JSON(common.ErrorResponse{
	// 		StatusCode: uint(err.Code),
	// 		Message:    err.Message,
	// 	})
	// }

	fmt.Println("hELLOOOO")
	utils.ParseBodyAndValidate(c, serviceDto)

	return c.Status(fiber.StatusOK).JSON(serviceDto)

}
