package route

import "github.com/gofiber/fiber/v3"

func (c *RouteConfig) SetupOrderRoute(api fiber.Router) {
	// Public: guest checkout + order status check (warung concept)
	api.Post("/orders", c.OrderController.Create)
	api.Get("/orders/:code", c.OrderController.GetByCode)

	// Admin only: manage orders
	admin := api.Group("/admin", c.AuthMiddleware, c.AdminMiddleware)
	admin.Get("/orders", c.OrderController.GetAll)
	admin.Put("/orders/:id/status", c.OrderController.UpdateStatus)
}
