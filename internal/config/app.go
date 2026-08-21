package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/controller"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/jwt"
	"github.com/mrbayss/golang-simple-ecommerce/internal/repository"
	"github.com/mrbayss/golang-simple-ecommerce/internal/route"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
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
	Jwt       *jwt.Key
	Redis     *redis.Client
}

func Bootstrap(config *BootstrapConfig) {
	userRepository := repository.NewUserRepository()
	categoryRepository := repository.NewCategoryRepository()
	productRepository := repository.NewProductRepository()
	orderRepository := repository.NewOrderRepository()

	authService := service.NewAuthService(config.DB, userRepository, config.Log, config.Validator, config.Jwt)
	categoryService := service.NewCategoryService(config.DB, categoryRepository, config.Log, config.Validator)
	productService := service.NewProductService(config.DB, productRepository, config.Log, config.Validator)
	orderService := service.NewOrderService(config.DB, orderRepository, productRepository, config.Log, config.Validator)

	authController := controller.NewAuthController(config.Log, authService)
	categoryController := controller.NewCategoryController(config.Log, categoryService)
	productController := controller.NewProductController(config.Log, productService)
	orderController := controller.NewOrderController(config.Log, orderService)

	healthController := controller.NewHealthController(config.Log, config.DB, config.Redis)

	authMiddleware := jwt.NewJWTMiddleware(&jwt.MiddlewareConfig{
		Log: config.Log,
		Jwt: config.Jwt,
	})
	adminMiddleware := authMiddleware.RequireAdmin()

	routeConfig := &route.RouteConfig{
		App:                config.App,
		AuthController:     authController,
		CategoryController: categoryController,
		ProductController:  productController,
		OrderController:    orderController,
		HealthController:   healthController,
		AuthMiddleware:     authMiddleware.Handle(),
		AdminMiddleware:    adminMiddleware,
	}

	routeConfig.Setup()
}
