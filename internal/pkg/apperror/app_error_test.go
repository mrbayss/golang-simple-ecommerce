package apperror_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type sampleRequest struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gt=18"`
	Role  string `validate:"oneof=ADMIN CUSTOMER"`
}

func TestAppError(t *testing.T) {
	err := apperror.NewAppError(http.StatusBadRequest, "bad request happened")
	require.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Equal(t, "bad request happened", err.Error())
}

func TestValidationErrorsAndFields(t *testing.T) {
	val := validator.New()

	req := sampleRequest{
		Name:  "",
		Email: "invalid-email",
		Age:   15,
		Role:  "SUPERUSER",
	}

	err := val.Struct(req)
	require.Error(t, err)

	t.Run("GetValidationErrorMessage", func(t *testing.T) {
		errMsgs := apperror.GetValidationErrorMessage(err)
		require.NotEmpty(t, errMsgs)

		fieldErrors := make(map[string]string)
		for _, e := range errMsgs {
			fieldErrors[e.Field] = e.Message
		}

		assert.Contains(t, fieldErrors, "Name")
		assert.Equal(t, "is required", fieldErrors["Name"])

		assert.Contains(t, fieldErrors, "Email")
		assert.Equal(t, "invalid email format", fieldErrors["Email"])

		assert.Contains(t, fieldErrors, "Age")
		assert.Equal(t, "value must be greater than 18", fieldErrors["Age"])

		assert.Contains(t, fieldErrors, "Role")
		assert.Equal(t, "must be one of ADMIN CUSTOMER", fieldErrors["Role"])
	})

	t.Run("GetValidationErrorMessage on non-validation error", func(t *testing.T) {
		nonValErr := errors.New("database disconnected")
		errMsgs := apperror.GetValidationErrorMessage(nonValErr)
		require.Len(t, errMsgs, 1)
		assert.Equal(t, "invalid input", errMsgs[0].Message)
	})

	t.Run("ValidationFields", func(t *testing.T) {
		fieldsStr := apperror.ValidationFields(err)
		assert.Contains(t, fieldsStr, "Name(required)")
		assert.Contains(t, fieldsStr, "Email(email)")
		assert.Contains(t, fieldsStr, "Age(gt)")
		assert.Contains(t, fieldsStr, "Role(oneof)")

		// Non-validation error returns empty
		assert.Empty(t, apperror.ValidationFields(errors.New("other")))
	})
}

func TestHandleError(t *testing.T) {
	val := validator.New()

	t.Run("Nil error returns 200 OK", func(t *testing.T) {
		msg, code, errs := apperror.HandleError(nil)
		assert.Empty(t, msg)
		assert.Equal(t, http.StatusOK, code)
		assert.Nil(t, errs)
	})

	t.Run("Validator error returns 400 Bad Request with field errors", func(t *testing.T) {
		req := sampleRequest{}
		valErr := val.Struct(req)
		msg, code, errs := apperror.HandleError(valErr)
		assert.Equal(t, "invalid input", msg)
		assert.Equal(t, http.StatusBadRequest, code)
		assert.NotEmpty(t, errs)
	})

	t.Run("AppError returns custom status code and message", func(t *testing.T) {
		appErr := apperror.NewAppError(http.StatusConflict, "slug already exists")
		msg, code, errs := apperror.HandleError(appErr)
		assert.Equal(t, "slug already exists", msg)
		assert.Equal(t, http.StatusConflict, code)
		assert.Nil(t, errs)
	})

	t.Run("gorm.ErrRecordNotFound returns 404 Not Found", func(t *testing.T) {
		msg, code, _ := apperror.HandleError(gorm.ErrRecordNotFound, apperror.ErrorParams{Object: "product"})
		assert.Equal(t, "product not found", msg)
		assert.Equal(t, http.StatusNotFound, code)
	})

	t.Run("gorm.ErrDuplicatedKey returns 409 Conflict", func(t *testing.T) {
		msg, code, _ := apperror.HandleError(gorm.ErrDuplicatedKey, apperror.ErrorParams{Object: "user"})
		assert.Equal(t, "user already exists", msg)
		assert.Equal(t, http.StatusConflict, code)
	})

	t.Run("Invalid UUID error returns 400 Bad Request", func(t *testing.T) {
		uuidErr := errors.New("invalid UUID length 5")
		msg, code, _ := apperror.HandleError(uuidErr, apperror.ErrorParams{Object: "category"})
		assert.Equal(t, "invalid UUID for category", msg)
		assert.Equal(t, http.StatusBadRequest, code)
	})

	t.Run("Fallback error message", func(t *testing.T) {
		genericErr := errors.New("some unexpected error")
		msg, code, _ := apperror.HandleError(genericErr, apperror.ErrorParams{
			Fallback: "custom fallback message",
		})
		assert.Equal(t, "custom fallback message", msg)
		assert.Equal(t, http.StatusInternalServerError, code)
	})

	t.Run("Default generic error returns 500", func(t *testing.T) {
		genericErr := errors.New("unknown error")
		msg, code, _ := apperror.HandleError(genericErr)
		assert.Equal(t, "internal server error", msg)
		assert.Equal(t, http.StatusInternalServerError, code)
	})
}
