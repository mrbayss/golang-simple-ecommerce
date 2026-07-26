package utils

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
	"required": "Tidak boleh kosong",
	"email":    "Format email tidak valid",
	"datetime": "Format tanggal tidak valid",
	"min":      "minimal %s karakter",
	"max":      "maksimal %s karakter",
	"gt":       "nominal harus lebih besar dari %s",
	"uuid":     "Format id tidak valid",
	"url":      "Format url tidak valid",
	"oneof":    "harus salah satu dari %s",
	"numeric":  "harus berupa angka",
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
				Message: "Input tidak valid",
			},
		}
	}

	result := make([]*model.ValidationError, 0, len(ve))

	for _, e := range ve {

		msg := ValidatorValidationMessages[e.Tag()]

		if msg == "" {
			msg = "input tidak valid"
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
		return "Input tidak valid", http.StatusBadRequest, GetValidationErrorMessage(err)
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message, appErr.Code, nil
	}

	var p ErrorParams
	if len(params) > 0 {
		p = params[0]
	}

	object := "tersebut"
	if p.Object != "" {
		object = p.Object
	}

	if strings.Contains(err.Error(), "invalid UUID") {
		return fmt.Sprintf("UUID %s tidak valid", object), http.StatusBadRequest, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Sprintf("Data %s tidak ditemukan", object), http.StatusNotFound, nil
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return fmt.Sprintf("Data %s sudah ada", object), http.StatusConflict, nil
	}

	if p.Fallback != "" {
		return p.Fallback, http.StatusInternalServerError, nil
	}

	if p.Object != "" {
		return fmt.Sprintf("Kesalahan tidak terduga untuk data %s : %s", object, err.Error()), http.StatusInternalServerError, nil
	}

	return fmt.Sprintf("Kesalahan tidak terduga : %s", err.Error()), http.StatusInternalServerError, nil
}
