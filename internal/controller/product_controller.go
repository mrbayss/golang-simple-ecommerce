package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
	"github.com/sirupsen/logrus"
)

type ProductController struct {
	Log             *logrus.Logger
	ProductService  service.ProductService
}

func NewProductController(log *logrus.Logger, productService service.ProductService) *ProductController {
	return &ProductController{
		Log:            log,
		ProductService: productService,
	}
}

func (pc *ProductController) GetByID(ctx fiber.Ctx) error {
	res, err := pc.ProductService.GetByID(ctx, ctx.Params("id"))
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "produk"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "Produk ditemukan"))
}

func (pc *ProductController) GetBySlug(ctx fiber.Ctx) error {
	res, err := pc.ProductService.GetBySlug(ctx, ctx.Params("slug"))
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "produk"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "Produk ditemukan"))
}

func (pc *ProductController) GetAll(ctx fiber.Ctx) error {
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	res, err := pc.ProductService.GetAll(ctx, page, limit)
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "produk"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "Daftar produk"))
}
