package route

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/controller"
)

type RouteConfig struct {
	App                *fiber.App
	AuthController     *controller.AuthController
	CategoryController *controller.CategoryController
	ProductController  *controller.ProductController
	HealthController   *controller.HealthController
	AuthMiddleware     fiber.Handler
	AdminMiddleware    fiber.Handler
}
