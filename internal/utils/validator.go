// Package utils
package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(payload any) error {
	err := validate.Struct(payload)
	if err != nil {
		var errorsList []string

		for _, err := range err.(validator.ValidationErrors) {
			fmt.Println(err.Param())
			errorsList = append(
				errorsList,
				fmt.Sprintf("`%v` with value `%v` doesn't satisfy the `%v` constraint", err.Field(), err.Value(), err.Tag()),
			)
		}
		return errors.New(strings.Join(errorsList, ","))

	}

	return nil
}
