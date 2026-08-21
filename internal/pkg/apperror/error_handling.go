package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"gorm.io/gorm"
)

var ValidatorValidationMessages = map[string]string{
	"required": "is required",
	"email":    "invalid email format",
	"datetime": "invalid date format",
	"min":      "minimum %s characters",
	"max":      "maximum %s characters",
	"gt":       "value must be greater than %s",
	"uuid":     "invalid id format",
	"url":      "invalid url format",
	"oneof":    "must be one of %s",
	"numeric":  "must be a number",
}

type ErrorParams struct {
	Object   string
	Fallback string
}

func GetValidationErrorMessage(err error) []*model.ValidationError {
	var ve validator.ValidationErrors

	if !errors.As(err, &ve) {
		return []*model.ValidationError{
			{
				Field:   "",
				Message: "invalid input",
			},
		}
	}

	result := make([]*model.ValidationError, 0, len(ve))

	for _, e := range ve {

		msg := ValidatorValidationMessages[e.Tag()]

		if msg == "" {
			msg = "invalid input"
		}

		if e.Param() != "" {
			msg = fmt.Sprintf(msg, e.Param())
		}

		result = append(result, &model.ValidationError{
			Field:   e.Field(),
			Message: msg,
		})
	}

	return result
}

func HandleError(err error, params ...ErrorParams) (string, int, []*model.ValidationError) {
	if err == nil {
		return "", http.StatusOK, nil
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return "invalid input", http.StatusBadRequest, GetValidationErrorMessage(err)
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message, appErr.Code, nil
	}

	var p ErrorParams
	if len(params) > 0 {
		p = params[0]
	}

	object := "item"
	if p.Object != "" {
		object = p.Object
	}

	if strings.Contains(err.Error(), "invalid UUID") {
		return fmt.Sprintf("invalid UUID for %s", object), http.StatusBadRequest, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Sprintf("%s not found", object), http.StatusNotFound, nil
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return fmt.Sprintf("%s already exists", object), http.StatusConflict, nil
	}

	if p.Fallback != "" {
		return p.Fallback, http.StatusInternalServerError, nil
	}

	if p.Object != "" {
		return fmt.Sprintf("unexpected error for %s", object), http.StatusInternalServerError, nil
	}

	return "internal server error", http.StatusInternalServerError, nil
}
