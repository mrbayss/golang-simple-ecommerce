package route

import "github.com/gofiber/fiber/v3"

func (c *RouteConfig) SetupCategoryRoute(api fiber.Router) {
	// Public: anyone can browse categories (warung concept)
	api.Get("/categories", c.CategoryController.GetAll)
	api.Get("/categories/:id", c.CategoryController.GetByID)

	// Admin only: manage categories
	admin := api.Group("/admin", c.AuthMiddleware, c.AdminMiddleware)
	admin.Post("/categories", c.CategoryController.Create)
	admin.Put("/categories/:id", c.CategoryController.Update)
	admin.Delete("/categories/:id", c.CategoryController.Delete)
}
