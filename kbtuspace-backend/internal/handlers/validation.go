package handlers

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidationMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return "Invalid input"
	}

	for _, fieldErr := range validationErrors {
		field := strings.ToLower(fieldErr.Field())
		switch fieldErr.Tag() {
		case "required":
			return field + " is required"
		case "email":
			return "Email must be a valid address"
		case "min":
			return field + " must be at least " + fieldErr.Param() + " characters"
		case "max":
			return field + " must be at most " + fieldErr.Param() + " characters"
		case "oneof":
			return field + " has an invalid value"
		}
	}

	return "Invalid input"
}
