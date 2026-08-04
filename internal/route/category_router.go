package route

import "github.com/gofiber/fiber/v3"

func (c *RouteConfig) SetupCategoryRoute(api fiber.Router) {
	api.Get("/categories", c.CategoryController.GetAll)
	api.Get("/categories/:id", c.CategoryController.GetByID)

	api.Post("/categories", c.AuthMiddleware, c.CategoryController.Create)
	api.Put("/categories/:id", c.AuthMiddleware, c.CategoryController.Update)
	api.Delete("/categories/:id", c.AuthMiddleware, c.CategoryController.Delete)
}
