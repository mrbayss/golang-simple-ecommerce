package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
	"github.com/sirupsen/logrus"
)

type CategoryController struct {
	Log             *logrus.Logger
	CategoryService service.CategoryService
}

func NewCategoryController(log *logrus.Logger, categoryService service.CategoryService) *CategoryController {
	return &CategoryController{
		Log:             log,
		CategoryService: categoryService,
	}
}

func (cc *CategoryController) Create(ctx fiber.Ctx) error {
	var request model.CreateCategoryReq

	if err := ctx.Bind().Body(&request); err != nil {
		cc.Log.Warnf("failed to parse body: %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("invalid request format", nil))
	}

	res, err := cc.CategoryService.Create(ctx, &request)
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "kategori"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.SuccessResponse(res, "category created successfully"))
}

func (cc *CategoryController) GetByID(ctx fiber.Ctx) error {
	res, err := cc.CategoryService.GetByID(ctx, ctx.Params("id"))
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "kategori"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "category found"))
}

func (cc *CategoryController) GetAll(ctx fiber.Ctx) error {
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	res, err := cc.CategoryService.GetAll(ctx, page, limit)
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "kategori"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "category list"))
}

func (cc *CategoryController) Update(ctx fiber.Ctx) error {
	var request model.UpdateCategoryReq

	if err := ctx.Bind().Body(&request); err != nil {
		cc.Log.Warnf("failed to parse body: %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("invalid request format", nil))
	}

	res, err := cc.CategoryService.Update(ctx, ctx.Params("id"), &request)
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "kategori"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "category updated successfully"))
}

func (cc *CategoryController) Delete(ctx fiber.Ctx) error {
	if err := cc.CategoryService.Delete(ctx, ctx.Params("id")); err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "kategori"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(nil, "category deleted successfully"))
}
