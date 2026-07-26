package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/controller"
	"github.com/mrbayss/golang-simple-ecommerce/internal/repository"
	"github.com/mrbayss/golang-simple-ecommerce/internal/route"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
	"github.com/mrbayss/golang-simple-ecommerce/internal/utils/token"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB        *gorm.DB
	App       *fiber.App
	Log       *logrus.Logger
	Validator *validator.Validate
	Config    *viper.Viper
	Jwt       *token.Key
	Redis     *redis.Client
}

func Bootstrap(config *BootstrapConfig) {
	userRepository := repository.NewUserRepository()

	authService := service.NewAuthService(config.DB, userRepository, config.Log, config.Validator, config.Jwt)

	authController := controller.NewAuthController(config.Log, authService)

	routeConfig := &route.RouteConfig{
		App:            config.App,
		AuthController: authController,
	}

	routeConfig.Setup()
}
