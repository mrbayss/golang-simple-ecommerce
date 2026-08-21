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
	UpdateStatus(c context.Context, id string, newStatus entity.OrderStatus) (*model.OrderRes, error)
}

type orderService struct {
	DB              *gorm.DB
	OrderRepository repository.OrderRepository
	Log             *logrus.Logger
	Validate        *validator.Validate
}

func NewOrderService(db *gorm.DB, orderRepository repository.OrderRepository, log *logrus.Logger, validate *validator.Validate) OrderService {
	return &orderService{
		DB:              db,
		OrderRepository: orderRepository,
		Log:             log,
		Validate:        validate,
	}
}

// Create handles guest checkout: validate -> normalize phone -> TX:
// lock products, check stock, compute totals server-side, snapshot items,
// insert order + items, decrement stock.
func (os *orderService) Create(c context.Context, request *model.CreateOrderReq) (*model.OrderRes, error) {
	if err := os.Validate.Struct(request); err != nil {
		os.Log.Warnf("validation failed: %v", err)
		return nil, err
	}

	phone, err := phonenumber.Normalize(request.CustomerPhone)
	if err != nil {
		return nil, apperror.NewAppError(fiber.StatusBadRequest, err.Error())
	}

	ctx := os.DB.WithContext(c).Begin()
	defer func() {
		if r := recover(); r != nil {
			ctx.Rollback()
			panic(r)
		}
	}()

	order := &entity.Order{
		CustomerName:    request.CustomerName,
		CustomerPhone:   phone,
		CustomerAddress: request.CustomerAddress,
		Notes:           request.Notes,
		PaymentMethod:   entity.PaymentMethod(request.PaymentMethod),
		Status:          entity.StatusPending,
	}

	total := entity.Money(0)
	items := make([]entity.OrderItem, 0, len(request.Items))
	seenProducts := make(map[uuid.UUID]int, len(request.Items))

	for _, itemReq := range request.Items {
		productID, err := uuid.Parse(itemReq.ProductID)
		if err != nil {
			ctx.Rollback()
			return nil, apperror.NewAppError(fiber.StatusBadRequest, "Format product_id tidak valid")
		}

		seenProducts[productID]++
		if seenProducts[productID] > 1 {
			ctx.Rollback()
			return nil, apperror.NewAppError(fiber.StatusBadRequest, "Produk duplikat di dalam order")
		}

		// Lock row to prevent oversell on concurrent orders.
		var product entity.Product
		if err := ctx.Set("gorm:query_option", "FOR UPDATE").First(&product, "id = ?", productID).Error; err != nil {
			ctx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperror.NewAppError(fiber.StatusBadRequest, fmt.Sprintf("Produk tidak ditemukan: %s", itemReq.ProductID))
			}
			os.Log.Errorf("failed to lock product %s: %v", productID, err)
			return nil, apperror.NewAppError(fiber.StatusInternalServerError, "kesalahan server internal")
		}

		if product.Stock < itemReq.Quantity {
			ctx.Rollback()
			return nil, apperror.NewAppError(fiber.StatusConflict, fmt.Sprintf("Stok produk %s tidak cukup (sisa %d)", product.Name, product.Stock))
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

		// Decrement stock.
		if err := ctx.Model(&entity.Product{}).Where("id = ?", product.ID).
			Update("stock", gorm.Expr("stock - ?", itemReq.Quantity)).Error; err != nil {
			ctx.Rollback()
			os.Log.Errorf("failed to decrement stock for %s: %v", product.ID, err)
			return nil, apperror.NewAppError(fiber.StatusInternalServerError, "kesalahan server internal")
		}
	}

	order.Items = items
	order.TotalPrice = total

	orderCode, err := os.generateOrderCode(ctx)
	if err != nil {
		ctx.Rollback()
		os.Log.Errorf("failed to generate order code: %v", err)
		return nil, apperror.NewAppError(fiber.StatusInternalServerError, "kesalahan server internal")
	}
	order.OrderCode = orderCode

	if err := ctx.Create(order).Error; err != nil {
		ctx.Rollback()
		os.Log.Errorf("failed to create order: %v", err)
		return nil, apperror.NewAppError(fiber.StatusInternalServerError, "kesalahan server internal")
	}

	if err := ctx.Commit().Error; err != nil {
		os.Log.Errorf("failed to commit order transaction: %v", err)
		return nil, apperror.NewAppError(fiber.StatusInternalServerError, "kesalahan server internal")
	}

	os.Log.Infof("order created: %s total=%s", order.OrderCode, order.TotalPrice.IDR())

	return toOrderRes(order), nil
}

// generateOrderCode produces WB-YYYYMMDD-NNNN based on today's order count.
func (os *orderService) generateOrderCode(ctx *gorm.DB) (string, error) {
	count, err := os.OrderRepository.CountToday(ctx, time.Now())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("WB-%s-%04d", time.Now().Format("20060102"), count+1), nil
}

func (os *orderService) GetByCode(c context.Context, code string) (*model.OrderRes, error) {
	ctx := os.DB.WithContext(c)

	order, err := os.OrderRepository.FindByCode(ctx, code)
	if err != nil {
		os.Log.Errorf("failed to find order by code %s: %v", code, err)
		return nil, err
	}

	return toOrderRes(order), nil
}

func (os *orderService) GetAll(c context.Context, page, limit int, status *entity.OrderStatus) (*model.OrderListRes, error) {
	ctx := os.DB.WithContext(c)

	orders, total, err := os.OrderRepository.FindAllPaginated(ctx, page, limit, status)
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
	entity.StatusCompleted: {}, // final
	entity.StatusCancelled: {}, // final
}

func (os *orderService) UpdateStatus(c context.Context, id string, newStatus entity.OrderStatus) (*model.OrderRes, error) {
	ctx := os.DB.WithContext(c)

	if _, err := uuid.Parse(id); err != nil {
		return nil, apperror.NewAppError(fiber.StatusBadRequest, "Format id tidak valid")
	}

	var order entity.Order
	if err := ctx.Preload("Items").First(&order, "id = ?", id).Error; err != nil {
		os.Log.Errorf("failed to find order %s: %v", id, err)
		return nil, err
	}

	allowed := allowedTransitions[order.Status]
	valid := false
	for _, s := range allowed {
		if s == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return nil, apperror.NewAppError(fiber.StatusConflict,
			fmt.Sprintf("Transisi status tidak valid dari %s ke %s", order.Status, newStatus))
	}

	oldStatus := order.Status
	order.Status = newStatus

	err := ctx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Order{}).Where("id = ?", order.ID).Update("status", newStatus).Error; err != nil {
			return err
		}

		// Cancellation restocks the items.
		if newStatus == entity.StatusCancelled && oldStatus != entity.StatusCancelled {
			for _, item := range order.Items {
				if err := tx.Model(&entity.Product{}).Where("id = ?", item.ProductID).
					Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		os.Log.Errorf("failed to update order status: %v", err)
		return nil, apperror.NewAppError(fiber.StatusInternalServerError, "kesalahan server internal")
	}

	os.Log.Infof("order %s status: %s -> %s", order.OrderCode, oldStatus, newStatus)

	return toOrderRes(&order), nil
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
		TotalPrice:      int64(order.TotalPrice),
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		Items:           make([]model.OrderItemRes, 0, len(order.Items)),
	}

	for _, item := range order.Items {
		res.Items = append(res.Items, model.OrderItemRes{
			ID:           item.ID.String(),
			ProductID:    item.ProductID.String(),
			ProductName:  item.ProductName,
			ProductPrice: int64(item.ProductPrice),
			Quantity:     item.Quantity,
			Subtotal:     int64(item.Subtotal),
		})
	}

	return res
}
