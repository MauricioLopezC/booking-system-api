package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func Validate(payload any) *fiber.Error {
	err := validate.Struct(payload)
	if err != nil {
		var errors []string

		for _, err := range err.(validator.ValidationErrors) {
			fmt.Println(err.Param())
			errors = append(
				errors,
				fmt.Sprintf("`%v` with value `%v` doesn't satisfy the `%v` constraint", err.Field(), err.Value(), err.Tag()),
			)
		}

		//NOTE: fiber.Error is not the final api response, is a Struct, compatible with
		//Error interface
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: strings.Join(errors, ","),
		}
	}

	return nil
}
