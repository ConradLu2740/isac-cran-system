package model

import (
	"errors"
	"fmt"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

func NewValidationErrorf(format string, args ...interface{}) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
