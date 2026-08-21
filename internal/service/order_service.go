package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/phonenumber"
	"github.com/mrbayss/golang-simple-ecommerce/internal/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OrderService interface {
	Create(c context.Context, request *model.CreateOrderReq) (*model.OrderRes, error)
	GetByCode(c context.Context, code string) (*model.OrderRes, error)
	GetAll(c context.Context, page, limit int, status *entity.OrderStatus) (*model.OrderListRes, error)
	UpdateStatus(c context.Context, id string, request *model.UpdateOrderStatusReq) (*model.OrderRes, error)
}

type orderService struct {
	DB                *gorm.DB
	OrderRepository   repository.OrderRepository
	ProductRepository repository.ProductRepository
	Log               *logrus.Logger
	Validate          *validator.Validate
}

func NewOrderService(db *gorm.DB, orderRepository repository.OrderRepository, productRepository repository.ProductRepository, log *logrus.Logger, validate *validator.Validate) OrderService {
	return &orderService{
		DB:                db,
		OrderRepository:   orderRepository,
		ProductRepository: productRepository,
		Log:               log,
		Validate:          validate,
	}
}

// Create handles guest checkout. All database access runs inside a single
// transaction: products are locked with SELECT ... FOR UPDATE to prevent
// oversell under concurrent orders.
func (os *orderService) Create(c context.Context, request *model.CreateOrderReq) (*model.OrderRes, error) {
	if err := os.Validate.Struct(request); err != nil {
		os.Log.Warnf("validation failed: %v", err)
		return nil, err
	}

	phone, err := phonenumber.Normalize(request.CustomerPhone)
	if err != nil {
		return nil, apperror.NewAppError(fiber.StatusBadRequest, err.Error())
	}

	order := &entity.Order{
		CustomerName:    request.CustomerName,
		CustomerPhone:   phone,
		CustomerAddress: request.CustomerAddress,
		Notes:           request.Notes,
		PaymentMethod:   entity.PaymentMethod(request.PaymentMethod),
		Status:          entity.StatusPending,
	}

	err = os.DB.WithContext(c).Transaction(func(tx *gorm.DB) error {
		total := entity.Money(0)
		items := make([]entity.OrderItem, 0, len(request.Items))
		seenProducts := make(map[uuid.UUID]bool, len(request.Items))

		for _, itemReq := range request.Items {
			productID, err := uuid.Parse(itemReq.ProductID)
			if err != nil {
				os.Log.Warnf("create order rejected, invalid product_id %s: %v", itemReq.ProductID, err)
				return apperror.NewAppError(fiber.StatusBadRequest, "invalid product_id format")
			}

			if seenProducts[productID] {
				os.Log.Warnf("create order rejected, duplicate product %s", productID)
				return apperror.NewAppError(fiber.StatusBadRequest, "duplicate product in order")
			}
			seenProducts[productID] = true

			product, err := os.ProductRepository.FindByIDForUpdate(tx, productID.String())
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					os.Log.Warnf("create order rejected, product not found: %s", itemReq.ProductID)
					return apperror.NewAppError(fiber.StatusBadRequest, fmt.Sprintf("product not found: %s", itemReq.ProductID))
				}
				return fmt.Errorf("lock product %s: %w", productID, err)
			}

			if product.Stock < itemReq.Quantity {
				os.Log.Warnf("create order rejected, insufficient stock for %s (requested %d, remaining %d)", product.Name, itemReq.Quantity, product.Stock)
				return apperror.NewAppError(fiber.StatusConflict, fmt.Sprintf("insufficient stock for %s (remaining %d)", product.Name, product.Stock))
			}

			subtotal := product.Price * entity.Money(itemReq.Quantity)
			total += subtotal

			items = append(items, entity.OrderItem{
				ProductID:    product.ID,
				ProductName:  product.Name,
				ProductPrice: product.Price,
				Quantity:     itemReq.Quantity,
				Subtotal:     subtotal,
			})

			if err := os.ProductRepository.UpdateStock(tx, product.ID.String(), -itemReq.Quantity); err != nil {
				return fmt.Errorf("decrement stock for %s: %w", product.ID, err)
			}
		}

		order.Items = items
		order.TotalPrice = total

		orderCode, err := os.generateOrderCode(tx)
		if err != nil {
			return fmt.Errorf("generate order code: %w", err)
		}
		order.OrderCode = orderCode

		if err := os.OrderRepository.CreateWithItems(tx, order); err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		return nil
	})
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		os.Log.Errorf("failed to create order: %v", err)
		return nil, apperror.NewAppError(fiber.StatusInternalServerError, "internal server error")
	}

	os.Log.Infof("order created: %s total=%s", order.OrderCode, order.TotalPrice.IDR())

	return toOrderRes(order), nil
}

// generateOrderCode produces WB-YYYYMMDD-NNNN based on today's order count.
// Must be called inside the same transaction as the insert so that
// concurrent orders see a consistent count.
func (os *orderService) generateOrderCode(ctx *gorm.DB) (string, error) {
	count, err := os.OrderRepository.CountToday(ctx, time.Now())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("WB-%s-%04d", time.Now().Format("20060102"), count+1), nil
}

func (os *orderService) GetByCode(c context.Context, code string) (*model.OrderRes, error) {
	order, err := os.OrderRepository.FindByCode(os.DB.WithContext(c), code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			os.Log.Warnf("order not found by code: %s", code)
		} else {
			os.Log.Errorf("failed to find order by code %s: %v", code, err)
		}
		return nil, err
	}

	return toOrderRes(order), nil
}

func (os *orderService) GetAll(c context.Context, page, limit int, status *entity.OrderStatus) (*model.OrderListRes, error) {
	orders, total, err := os.OrderRepository.FindAllPaginated(os.DB.WithContext(c), page, limit, status)
	if err != nil {
		os.Log.Errorf("failed to get orders: %v", err)
		return nil, err
	}

	items := make([]model.OrderRes, 0, len(orders))
	for _, order := range orders {
		items = append(items, *toOrderRes(&order))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &model.PaginatedRes[model.OrderRes]{
		Data: items,
		Meta: model.PageMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// allowedTransitions defines the order status state machine.
var allowedTransitions = map[entity.OrderStatus][]entity.OrderStatus{
	entity.StatusPending:   {entity.StatusConfirmed, entity.StatusCancelled},
	entity.StatusConfirmed: {entity.StatusReady, entity.StatusCancelled},
	entity.StatusReady:     {entity.StatusCompleted, entity.StatusCancelled},
	entity.StatusCompleted: {}, // final state
	entity.StatusCancelled: {}, // final state
}

func (os *orderService) UpdateStatus(c context.Context, id string, request *model.UpdateOrderStatusReq) (*model.OrderRes, error) {
	if _, err := uuid.Parse(id); err != nil {
		os.Log.Warnf("update status rejected, invalid id format: %s", id)
		return nil, apperror.NewAppError(fiber.StatusBadRequest, "invalid id format")
	}

	newStatus := request.Status

	order, err := os.OrderRepository.FindByID(os.DB.WithContext(c), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			os.Log.Warnf("update status rejected, order not found: %s", id)
		} else {
			os.Log.Errorf("failed to find order %s: %v", id, err)
		}
		return nil, err
	}

	valid := false
	for _, s := range allowedTransitions[order.Status] {
		if s == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		os.Log.Warnf("update status rejected, invalid transition for %s: %s -> %s", order.OrderCode, order.Status, newStatus)
		return nil, apperror.NewAppError(fiber.StatusConflict,
			fmt.Sprintf("invalid status transition from %s to %s", order.Status, newStatus))
	}

	oldStatus := order.Status

	err = os.DB.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if err := os.OrderRepository.UpdateStatus(tx, order.ID.String(), newStatus); err != nil {
			return fmt.Errorf("update order status: %w", err)
		}

		// Cancellation returns the reserved stock.
		if newStatus == entity.StatusCancelled && oldStatus != entity.StatusCancelled {
			for _, item := range order.Items {
				if err := os.ProductRepository.UpdateStock(tx, item.ProductID.String(), item.Quantity); err != nil {
					return fmt.Errorf("restock product %s: %w", item.ProductID, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		os.Log.Errorf("failed to update order status: %v", err)
		return nil, apperror.NewAppError(fiber.StatusInternalServerError, "internal server error")
	}

	os.Log.Infof("order %s status: %s -> %s", order.OrderCode, oldStatus, newStatus)

	order.Status = newStatus
	return toOrderRes(order), nil
}

func toOrderRes(order *entity.Order) *model.OrderRes {
	res := &model.OrderRes{
		ID:              order.ID.String(),
		OrderCode:       order.OrderCode,
		CustomerName:    order.CustomerName,
		CustomerPhone:   order.CustomerPhone,
		CustomerAddress: order.CustomerAddress,
		Notes:           order.Notes,
		PaymentMethod:   string(order.PaymentMethod),
		Status:          string(order.Status),
		TotalPrice:      order.TotalPrice,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		Items:           make([]model.OrderItemRes, 0, len(order.Items)),
	}

	for _, item := range order.Items {
		res.Items = append(res.Items, model.OrderItemRes{
			ID:           item.ID.String(),
			ProductID:    item.ProductID.String(),
			ProductName:  item.ProductName,
			ProductPrice: item.ProductPrice,
			Quantity:     item.Quantity,
			Subtotal:     item.Subtotal,
		})
	}

	return res
}
