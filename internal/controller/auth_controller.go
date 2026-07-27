package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	Log         *logrus.Logger
	AuthService service.AuthService
}

func NewAuthController(log *logrus.Logger, authService service.AuthService) *AuthController {
	return &AuthController{
		Log:         log,
		AuthService: authService,
	}
}

func (ac *AuthController) Register(ctx fiber.Ctx) error {
	var request model.RegisterReq

	if err := ctx.Bind().Body(&request); err != nil {
		ac.Log.Warnf("failed to parse body: %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("Format request tidak valid", nil))
	}

	if err := ac.AuthService.Register(ctx, &request); err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: request.Email})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		model.SuccessResponse(nil, "Registrasi berhasil"),
	)
}

func (ac *AuthController) Login(ctx fiber.Ctx) error {
	var request model.LoginReq

	if err := ctx.Bind().Body(&request); err != nil {
		ac.Log.Warnf("failed to parse body: %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("Format request tidak valid", nil))
	}

	res, err := ac.AuthService.Login(ctx, &request)
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err)
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		model.SuccessResponse(res, "login success"),
	)

}
