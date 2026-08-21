package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
	"github.com/sirupsen/logrus"
)

type OrderController struct {
	Log           *logrus.Logger
	OrderService  service.OrderService
}

func NewOrderController(log *logrus.Logger, orderService service.OrderService) *OrderController {
	return &OrderController{
		Log:          log,
		OrderService: orderService,
	}
}

// Create handles guest checkout — no auth required.
func (oc *OrderController) Create(ctx fiber.Ctx) error {
	var request model.CreateOrderReq

	if err := ctx.Bind().Body(&request); err != nil {
		oc.Log.Warnf("failed to parse body: %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("Format request tidak valid", nil))
	}

	res, err := oc.OrderService.Create(ctx, &request)
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err)
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.SuccessResponse(res, "Pesanan berhasil dibuat"))
}

// GetByCode lets a customer check their order status — no auth required.
func (oc *OrderController) GetByCode(ctx fiber.Ctx) error {
	res, err := oc.OrderService.GetByCode(ctx, ctx.Params("code"))
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "pesanan"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "Pesanan ditemukan"))
}

// GetAll lists orders for admin, optional ?status= filter.
func (oc *OrderController) GetAll(ctx fiber.Ctx) error {
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	var status *entity.OrderStatus
	if s := ctx.Query("status", ""); s != "" {
		st := entity.OrderStatus(s)
		status = &st
	}

	res, err := oc.OrderService.GetAll(ctx, page, limit, status)
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "pesanan"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "Daftar pesanan"))
}

// UpdateStatus changes order status — admin only.
func (oc *OrderController) UpdateStatus(ctx fiber.Ctx) error {
	var request model.UpdateOrderStatusReq

	if err := ctx.Bind().Body(&request); err != nil {
		oc.Log.Warnf("failed to parse body: %v", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("Format request tidak valid", nil))
	}

	res, err := oc.OrderService.UpdateStatus(ctx, ctx.Params("id"), entity.OrderStatus(request.Status))
	if err != nil {
		message, code, validationErrors := apperror.HandleError(err, apperror.ErrorParams{Object: "pesanan"})
		return ctx.Status(code).JSON(model.ErrorResponse(message, validationErrors))
	}

	return ctx.Status(fiber.StatusOK).JSON(model.SuccessResponse(res, "Status pesanan diperbarui"))
}
