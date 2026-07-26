package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
	"github.com/mrbayss/golang-simple-ecommerce/internal/utils"
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

func (c *AuthController) Register(ctx fiber.Ctx) error {
	var request model.RegisterReq

	if err := ctx.Bind().Body(&request); err != nil {
		c.Log.Warnf("failed to parse body: %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("Format request tidak valid", nil))
	}

	if err := c.AuthService.Register(ctx, &request); err != nil {
		message, code, validationErrors := utils.HandleError(err, utils.ErrorParams{Object: request.Email})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusCreated).JSON(
		model.SuccessResponse(nil, "Registrasi berhasil"),
	)
}
