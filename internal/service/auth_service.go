package service

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
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
	Login(c context.Context, request *model.LoginReq) (*model.TokenRes, error)
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
		return utils.NewAppError(http.StatusInternalServerError, "failed to generate hasing password")
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

func (as *authService) Login(c context.Context, request *model.LoginReq) (*model.TokenRes, error) {
	ctx := as.DB.WithContext(c)

	if err := as.Validate.Struct(request); err != nil {
		as.Log.Warnf("validation failed: %v", err)
		return nil, err
	}

	user, err := as.UserRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		as.Log.Errorf("failed to find by email: %v", err)
		return nil, utils.NewAppError(fiber.StatusInternalServerError, "email atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		as.Log.Errorf("failed to compare password: %v", err)
		return nil, utils.NewAppError(fiber.StatusInternalServerError, "email atau password salah")
	}

	accessToken, accessClaims, err := as.Jwt.GenerateToken(user.ID.String(), user.Email, string(entity.AdminRole), time.Hour*24)
	if err != nil {
		as.Log.Errorf("failed to generate token: %v", err)
		return nil, utils.NewAppError(fiber.StatusInternalServerError, "failed to generate token")
	}

	response := &model.TokenRes{
		SessionID:          accessClaims.RegisteredClaims.ID,
		AccessToken:        accessToken,
		ExpiredAccessToken: accessClaims.ExpiresAt.String(),
	}

	as.Log.Infof("user %s success login", request.Email)

	return response, nil

}
