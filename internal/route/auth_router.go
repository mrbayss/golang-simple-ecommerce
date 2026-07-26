package route

import "github.com/gofiber/fiber/v3"

func (c *RouteConfig) SetupAuthRoute(api fiber.Router) {
	api.Post("/register", c.AuthController.Register)
}
