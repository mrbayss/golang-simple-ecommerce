package route

import "github.com/gofiber/fiber/v3"

func (c *RouteConfig) SetupProductRoute(api fiber.Router) {
	// Public: anyone can browse products (warung concept)
	api.Get("/products", c.ProductController.GetAll)
	api.Get("/products/slug/:slug", c.ProductController.GetBySlug)
	api.Get("/products/:id", c.ProductController.GetByID)
}
