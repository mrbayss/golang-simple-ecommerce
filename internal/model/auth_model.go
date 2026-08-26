package model

type RegisterReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type TokenRes struct {
	SessionID          string `json:"session_id"`
	AccessToken        string `json:"access_token"`
	ExpiredAccessToken string `json:"expired_access_token"`
}
