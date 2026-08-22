package route

import "github.com/gofiber/fiber/v3"

func (c *RouteConfig) SetupProductRoute(api fiber.Router) {
	// Public: anyone can browse products (warung concept)
	api.Get("/products", c.ProductController.GetAll)
	api.Get("/products/slug/:slug", c.ProductController.GetBySlug)
	api.Get("/products/:id", c.ProductController.GetByID)

	// Admin only: manage products
	admin := api.Group("/admin", c.AuthMiddleware, c.AdminMiddleware)
	admin.Post("/products", c.ProductController.Create)
	admin.Put("/products/:id", c.ProductController.Update)
	admin.Delete("/products/:id", c.ProductController.Delete)
	admin.Delete("/images/:image_id", c.ProductController.DeleteImage)
}
