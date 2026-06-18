package validation

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors,omitempty"`
}

func RegisterJSONTagNameFunc() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

func FormatValidationErrors(err error) []FieldError {
	var validationErrors validator.ValidationErrors

	if !errors.As(err, &validationErrors) {
		return []FieldError{
			{Field: "body", Message: err.Error()},
		}
	}

	errs := make([]FieldError, 0, len(validationErrors))

	for _, fe := range validationErrors {
		field := fe.Field()

		errs = append(errs, FieldError{
			Field:   field,
			Message: messageForTag(field, fe),
		})
	}

	return errs
}

func messageForTag(field string, fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "uuid":
		return field + " must be in valid uuid format"
	default:
		return field + " is invalid"
	}
}
