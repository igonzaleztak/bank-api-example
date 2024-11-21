package binding

import (
	errors "bank_test/internal/api_errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func handleBindingErrors(err error) error {
	validationErrors := err.(validator.ValidationErrors)
	validationErr := validationErrors[0]

	fieldName := validationErr.Field()

	apiError := errors.ErrInvalidBody

	switch validationErr.Tag() {
	case "required":
		apiError.Message = fieldName + " is required and must be a " + validationErr.Type().String()
	case "oneof":
		apiError.Message = fieldName + " must be one of: " + strings.Join(strings.Split(validationErr.Param(), " "), ", ")
	default:
		apiError.Message = err.Error()
	}

	return apiError
}
