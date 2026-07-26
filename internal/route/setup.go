package route

import "github.com/gofiber/fiber/v3"

func (c *RouteConfig) Setup() {

	c.App.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":     "WarBay",
			"version": "1.0.0.0",
		})
	})

	api := c.App.Group("/api/v1")

	c.SetupAuthRoute(api)
}
