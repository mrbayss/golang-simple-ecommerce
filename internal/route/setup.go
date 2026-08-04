package route

import "github.com/gofiber/fiber/v3"

func (c *RouteConfig) Setup() {

	c.App.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":     "WarBay",
			"version": "1.0.0.0",
		})
	})

	c.App.Get("/health/live", c.HealthController.Healthz)
	c.App.Get("/health/ready", c.HealthController.Readyz)

	api := c.App.Group("/api/v1")

	api.Get("/health", c.HealthController.Check)

	c.SetupAuthRoute(api)
	c.SetupCategoryRoute(api)
}
