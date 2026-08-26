package model

type WebResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    any                `json:"data"`
	Errors  []*ValidationError `json:"errors,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func SuccessResponse(data any, message ...string) *WebResponse {
	resMessage := "Sukses"
	if len(message) > 0 {
		resMessage = message[0]
	}
	return &WebResponse{
		Success: true,
		Message: resMessage,
		Data:    data,
	}
}

func ErrorResponse(message string, fieldErrors ...[]*ValidationError) *WebResponse {
	var errs []*ValidationError
	if len(fieldErrors) > 0 {
		errs = fieldErrors[0]
	}
	return &WebResponse{
		Success: false,
		Message: message,
		Data:    nil,
		Errors:  errs,
	}
}
