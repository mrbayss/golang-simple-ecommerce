package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
	"github.com/mrbayss/golang-simple-ecommerce/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOrderService_Validation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	mockOrderRepo := new(testutil.MockOrderRepository)
	mockProdRepo := new(testutil.MockProductRepository)
	orderSvc := service.NewOrderService(db, mockOrderRepo, mockProdRepo, logger, val)

	t.Run("Validation failure on missing customer name and invalid payment method", func(t *testing.T) {
		req := &model.CreateOrderReq{
			CustomerName:  "",
			CustomerPhone: "08123456789",
			PaymentMethod: "BITCOIN", // not allowed
			Items: []model.CreateOrderItemReq{
				{ProductID: uuid.New().String(), Quantity: 1},
			},
		}

		res, err := orderSvc.Create(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Validation failure on empty items", func(t *testing.T) {
		req := &model.CreateOrderReq{
			CustomerName:  "Budi Santoso",
			CustomerPhone: "08123456789",
			PaymentMethod: "COD",
			Items:         []model.CreateOrderItemReq{},
		}

		res, err := orderSvc.Create(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Failure on invalid phone number", func(t *testing.T) {
		req := &model.CreateOrderReq{
			CustomerName:  "Budi Santoso",
			CustomerPhone: "1234567890", // invalid prefix
			PaymentMethod: "COD",
			Items: []model.CreateOrderItemReq{
				{ProductID: uuid.New().String(), Quantity: 1},
			},
		}

		res, err := orderSvc.Create(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("Failure on duplicate products in single order", func(t *testing.T) {
		prodUUID := uuid.New()
		prodID := prodUUID.String()

		mockProdRepo.On("FindByIDForUpdate", mock.Anything, prodID).
			Return(&entity.Product{
				ID:    prodUUID,
				Name:  "Test Item",
				Price: 10000,
				Stock: 5,
			}, nil).Once()
		mockProdRepo.On("UpdateStock", mock.Anything, prodID, -1).
			Return(nil).Once()

		req := &model.CreateOrderReq{
			CustomerName:  "Budi Santoso",
			CustomerPhone: "08123456789",
			PaymentMethod: "COD",
			Items: []model.CreateOrderItemReq{
				{ProductID: prodID, Quantity: 1},
				{ProductID: prodID, Quantity: 2}, // duplicate
			},
		}

		res, err := orderSvc.Create(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate product")
		assert.Nil(t, res)
	})
}

func TestOrderService_GetByCode(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	mockOrderRepo := new(testutil.MockOrderRepository)
	mockProdRepo := new(testutil.MockProductRepository)
	orderSvc := service.NewOrderService(db, mockOrderRepo, mockProdRepo, logger, val)

	t.Run("Success find by order code", func(t *testing.T) {
		orderCode := "WB-20260927-0001"
		mockOrderRepo.On("FindByCode", mock.Anything, orderCode).
			Return(&entity.Order{
				ID:            uuid.New(),
				OrderCode:     orderCode,
				CustomerName:  "Budi",
				CustomerPhone: "628123456789",
				PaymentMethod: entity.PaymentCOD,
				Status:        entity.StatusPending,
				TotalPrice:    100000,
				CreatedAt:     time.Now(),
			}, nil).Once()

		res, err := orderSvc.GetByCode(context.Background(), orderCode)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, orderCode, res.OrderCode)
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("Error when order code not found", func(t *testing.T) {
		mockOrderRepo.On("FindByCode", mock.Anything, "NOT-EXIST").
			Return(nil, gorm.ErrRecordNotFound).Once()

		res, err := orderSvc.GetByCode(context.Background(), "NOT-EXIST")
		require.Error(t, err)
		assert.Nil(t, res)
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderService_UpdateStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	orderID := uuid.New()
	prodID := uuid.New()

	t.Run("Invalid status transition rejected (e.g. COMPLETED -> PENDING)", func(t *testing.T) {
		mockOrderRepo := new(testutil.MockOrderRepository)
		mockProdRepo := new(testutil.MockProductRepository)
		orderSvc := service.NewOrderService(db, mockOrderRepo, mockProdRepo, logger, val)

		mockOrderRepo.On("FindByID", mock.Anything, orderID.String()).
			Return(&entity.Order{
				ID:        orderID,
				OrderCode: "WB-20260927-0001",
				Status:    entity.StatusCompleted, // already completed
			}, nil).Once()

		req := &model.UpdateOrderStatusReq{
			Status: entity.StatusPending,
		}

		res, err := orderSvc.UpdateStatus(context.Background(), orderID.String(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status transition")
		assert.Nil(t, res)
		mockOrderRepo.AssertNotCalled(t, "UpdateStatus")
	})

	t.Run("Valid status transition (PENDING -> CONFIRMED)", func(t *testing.T) {
		mockOrderRepo := new(testutil.MockOrderRepository)
		mockProdRepo := new(testutil.MockProductRepository)
		orderSvc := service.NewOrderService(db, mockOrderRepo, mockProdRepo, logger, val)

		mockOrderRepo.On("FindByID", mock.Anything, orderID.String()).
			Return(&entity.Order{
				ID:        orderID,
				OrderCode: "WB-20260927-0001",
				Status:    entity.StatusPending,
			}, nil).Once()

		mockOrderRepo.On("UpdateStatus", mock.Anything, orderID.String(), entity.StatusConfirmed).
			Return(nil).Once()

		req := &model.UpdateOrderStatusReq{
			Status: entity.StatusConfirmed,
		}

		res, err := orderSvc.UpdateStatus(context.Background(), orderID.String(), req)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, string(entity.StatusConfirmed), res.Status)
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("Cancellation restores reserved product stock", func(t *testing.T) {
		mockOrderRepo := new(testutil.MockOrderRepository)
		mockProdRepo := new(testutil.MockProductRepository)
		orderSvc := service.NewOrderService(db, mockOrderRepo, mockProdRepo, logger, val)

		mockOrderRepo.On("FindByID", mock.Anything, orderID.String()).
			Return(&entity.Order{
				ID:        orderID,
				OrderCode: "WB-20260927-0001",
				Status:    entity.StatusPending,
				Items: []entity.OrderItem{
					{
						ProductID: prodID,
						Quantity:  3,
					},
				},
			}, nil).Once()

		mockOrderRepo.On("UpdateStatus", mock.Anything, orderID.String(), entity.StatusCancelled).
			Return(nil).Once()

		// Product stock must be updated with positive delta (+3) to restock
		mockProdRepo.On("UpdateStock", mock.Anything, prodID.String(), 3).
			Return(nil).Once()

		req := &model.UpdateOrderStatusReq{
			Status: entity.StatusCancelled,
		}

		res, err := orderSvc.UpdateStatus(context.Background(), orderID.String(), req)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, string(entity.StatusCancelled), res.Status)
		mockProdRepo.AssertExpectations(t)
		mockOrderRepo.AssertExpectations(t)
	})
}
