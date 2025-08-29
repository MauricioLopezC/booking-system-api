package common

import (
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrValidation     = errors.New("validation error")
	ErrConflict       = errors.New("conflict")
	ErrUnauthorized   = errors.New("unauthorized")
)
