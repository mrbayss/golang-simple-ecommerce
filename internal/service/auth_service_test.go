package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
	"github.com/mrbayss/golang-simple-ecommerce/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestAuthService_Register(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()
	jwtKey := testutil.NewTestJWT()

	t.Run("Success register new user", func(t *testing.T) {
		mockRepo := new(testutil.MockUserRepository)
		authSvc := service.NewAuthService(db, mockRepo, logger, val, jwtKey)

		req := &model.RegisterReq{
			Email:    "test@example.com",
			Password: "securepassword123",
		}

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Email == "test@example.com" && u.Password != ""
		})).Return(&entity.User{
			ID:    uuid.New(),
			Email: req.Email,
		}, nil).Once()

		err := authSvc.Register(context.Background(), req)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Validation error on invalid email or short password", func(t *testing.T) {
		mockRepo := new(testutil.MockUserRepository)
		authSvc := service.NewAuthService(db, mockRepo, logger, val, jwtKey)

		req := &model.RegisterReq{
			Email:    "invalid-email",
			Password: "short",
		}

		err := authSvc.Register(context.Background(), req)
		require.Error(t, err)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Database error when user already exists", func(t *testing.T) {
		mockRepo := new(testutil.MockUserRepository)
		authSvc := service.NewAuthService(db, mockRepo, logger, val, jwtKey)

		req := &model.RegisterReq{
			Email:    "existing@example.com",
			Password: "securepassword123",
		}

		mockRepo.On("Create", mock.Anything, mock.Anything).
			Return(nil, gorm.ErrDuplicatedKey).Once()

		err := authSvc.Register(context.Background(), req)
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrDuplicatedKey))
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()
	jwtKey := testutil.NewTestJWT()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	existingUser := &entity.User{
		ID:       uuid.New(),
		Email:    "buyer@example.com",
		Password: string(hashedPassword),
		Role:     entity.MemberRole,
	}

	t.Run("Success login with correct credentials", func(t *testing.T) {
		mockRepo := new(testutil.MockUserRepository)
		authSvc := service.NewAuthService(db, mockRepo, logger, val, jwtKey)

		req := &model.LoginReq{
			Email:    "buyer@example.com",
			Password: "password123",
		}

		mockRepo.On("FindByEmail", mock.Anything, req.Email).
			Return(existingUser, nil).Once()

		res, err := authSvc.Login(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.SessionID)
		assert.NotEmpty(t, res.ExpiredAccessToken)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Validation error on empty fields", func(t *testing.T) {
		mockRepo := new(testutil.MockUserRepository)
		authSvc := service.NewAuthService(db, mockRepo, logger, val, jwtKey)

		req := &model.LoginReq{
			Email:    "",
			Password: "",
		}

		res, err := authSvc.Login(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertNotCalled(t, "FindByEmail")
	})

	t.Run("Login fails when email is not found", func(t *testing.T) {
		mockRepo := new(testutil.MockUserRepository)
		authSvc := service.NewAuthService(db, mockRepo, logger, val, jwtKey)

		req := &model.LoginReq{
			Email:    "unknown@example.com",
			Password: "password123",
		}

		mockRepo.On("FindByEmail", mock.Anything, req.Email).
			Return(nil, gorm.ErrRecordNotFound).Once()

		res, err := authSvc.Login(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Login fails with incorrect password", func(t *testing.T) {
		mockRepo := new(testutil.MockUserRepository)
		authSvc := service.NewAuthService(db, mockRepo, logger, val, jwtKey)

		req := &model.LoginReq{
			Email:    "buyer@example.com",
			Password: "wrongpassword",
		}

		mockRepo.On("FindByEmail", mock.Anything, req.Email).
			Return(existingUser, nil).Once()

		res, err := authSvc.Login(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
	})
}
