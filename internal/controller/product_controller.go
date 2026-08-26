package controller

import (
	"mime/multipart"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
)

type ProductController struct {
	ProductService service.ProductService
}

func NewProductController(productService service.ProductService) *ProductController {
	return &ProductController{ProductService: productService}
}

// Public read endpoints
func (pc *ProductController) GetAll(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	res, err := pc.ProductService.GetAll(ctx, page, limit)
	if err != nil {
		message, code, _ := apperror.HandleError(err)
		return ctx.Status(code).JSON(model.ErrorResponse(message, nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "products retrieved successfully"))
}

func (pc *ProductController) GetBySlug(ctx fiber.Ctx) error {
	res, err := pc.ProductService.GetBySlug(ctx, ctx.Params("slug"))
	if err != nil {
		message, code, _ := apperror.HandleError(err)
		return ctx.Status(code).JSON(model.ErrorResponse(message, nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "product retrieved successfully"))
}

func (pc *ProductController) GetByID(ctx fiber.Ctx) error {
	res, err := pc.ProductService.GetByID(ctx, ctx.Params("id"))
	if err != nil {
		message, code, _ := apperror.HandleError(err)
		return ctx.Status(code).JSON(model.ErrorResponse(message, nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "product retrieved successfully"))
}

// Create handles multipart/form-data product creation with image uploads.
func (pc *ProductController) Create(ctx fiber.Ctx) error {
	var request model.CreateProductReq
	if err := ctx.Bind().Form(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("invalid form data", nil))
	}

	images := collectFiles(ctx, "images")

	res, err := pc.ProductService.Create(ctx, &request, images)
	if err != nil {
		message, code, _ := apperror.HandleError(err)
		return ctx.Status(code).JSON(model.ErrorResponse(message, nil))
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.SuccessResponse(res, "product created successfully"))
}

// Update applies sent fields and optionally appends new images.
func (pc *ProductController) Update(ctx fiber.Ctx) error {
	var request model.UpdateProductReq
	if err := ctx.Bind().Form(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("invalid form data", nil))
	}

	newImages := collectFiles(ctx, "images")

	res, err := pc.ProductService.Update(ctx, ctx.Params("id"), &request, newImages)
	if err != nil {
		message, code, _ := apperror.HandleError(err)
		return ctx.Status(code).JSON(model.ErrorResponse(message, nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "product updated successfully"))
}

func (pc *ProductController) Delete(ctx fiber.Ctx) error {
	err := pc.ProductService.Delete(ctx, ctx.Params("id"))
	if err != nil {
		message, code, _ := apperror.HandleError(err)
		return ctx.Status(code).JSON(model.ErrorResponse(message, nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(nil, "product deleted successfully"))
}

func (pc *ProductController) DeleteImage(ctx fiber.Ctx) error {
	err := pc.ProductService.DeleteImage(ctx, ctx.Params("image_id"))
	if err != nil {
		message, code, _ := apperror.HandleError(err)
		return ctx.Status(code).JSON(model.ErrorResponse(message, nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(nil, "image deleted successfully"))
}

// collectFiles gathers all multipart files posted under the given key.
func collectFiles(ctx fiber.Ctx, key string) []*multipart.FileHeader {
	form, err := ctx.MultipartForm()
	if err != nil || form.File == nil {
		return nil
	}
	return form.File[key]
}
