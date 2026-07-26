package service

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/repository"
	"github.com/mrbayss/golang-simple-ecommerce/internal/utils"
	"github.com/mrbayss/golang-simple-ecommerce/internal/utils/token"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(c context.Context, request *model.RegisterReq) error
}

type authService struct {
	DB             *gorm.DB
	UserRepository repository.UserRepository
	Log            *logrus.Logger
	Validate       *validator.Validate
	Jwt            *token.Key
}

func NewAuthService(db *gorm.DB, userRepository repository.UserRepository, log *logrus.Logger, validate *validator.Validate, jwt *token.Key) AuthService {
	return &authService{
		DB:             db,
		UserRepository: userRepository,
		Log:            log,
		Validate:       validate,
		Jwt:            jwt,
	}
}

func (as *authService) Register(c context.Context, request *model.RegisterReq) error {
	ctx := as.DB.WithContext(c)

	if err := as.Validate.Struct(request); err != nil {
		as.Log.Warnf("validation failed: %v", err)
		return err
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		as.Log.Errorf("failed to bcrypt password: %v", err)
		return utils.NewAppError(http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	create := &entity.User{
		Email:    request.Email,
		Password: string(hashPassword),
	}

	_, err = as.UserRepository.Create(ctx, create)
	if err != nil {
		as.Log.Errorf("failed to create user: %v", err)
		return err
	}

	as.Log.Info("user created successfully")

	return nil
}
