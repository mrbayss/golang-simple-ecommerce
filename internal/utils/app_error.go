package utils

type AppError struct {
	Code    int
	Message string
	Errors  map[string]string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}
