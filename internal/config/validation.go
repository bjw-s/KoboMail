// Package config implements all configuration aspects of KoboMail
package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func errorMessageForValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email"
	case "oneof":
		return fmt.Sprintf("Invalid value. Expected one of %v", strings.Split(fe.Param(), " "))
	}
	return fe.Error() // default error
}

// Validate returns if the given configuration is valid and any validation errors
func Validate(config *KoboMailConfig) (bool, validator.ValidationErrors) {
	validate := validator.New()
	validationErrors := validate.Struct(config)
	if validationErrors != nil {
		return false, validationErrors.(validator.ValidationErrors)
	}
	return true, nil
}
