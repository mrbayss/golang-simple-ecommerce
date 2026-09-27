package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/mrbayss/golang-simple-ecommerce/internal/repository"
	"github.com/mrbayss/golang-simple-ecommerce/internal/service"
	"github.com/mrbayss/golang-simple-ecommerce/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrderService_ConcurrentCheckout_PreventOversell simulates a flash-sale scenario
// where multiple concurrent goroutines compete to purchase limited stock simultaneously.
// Verifies that:
// 1. SELECT ... FOR UPDATE serializes stock checks.
// 2. Exactly InitialStock orders succeed.
// 3. Excess requests fail with 409 Conflict (insufficient stock).
// 4. Product stock in database never drops below 0 (no overselling).
func TestOrderService_ConcurrentCheckout_PreventOversell(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	testutil.CleanDB(t, db)
	defer testutil.CleanDB(t, db)

	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	prodRepo := repository.NewProductRepository()
	orderRepo := repository.NewOrderRepository()
	orderSvc := service.NewOrderService(db, orderRepo, prodRepo, logger, val)

	// 1. Seed Category and Product with stock = 10
	cat := &entity.Category{
		Name: "Flash Sale Category",
		Slug: "flash-sale-category",
	}
	require.NoError(t, db.Create(cat).Error)

	initialStock := 10
	product := &entity.Product{
		CategoryID:  &cat.ID,
		Name:        "Limited Edition Sneakers",
		Slug:        "limited-edition-sneakers",
		Description: "Only 10 available",
		Price:       500000,
		Stock:       initialStock,
	}
	require.NoError(t, db.Create(product).Error)

	// 2. Simulate 30 concurrent buyers attempting to buy 1 item each
	totalRequests := 30
	var successCount int64
	var conflictCount int64
	var otherErrorCount int64

	startBarrier := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(totalRequests)

	for i := 0; i < totalRequests; i++ {
		go func(buyerID int) {
			defer wg.Done()

			req := &model.CreateOrderReq{
				CustomerName:  fmt.Sprintf("Buyer %02d", buyerID),
				CustomerPhone: fmt.Sprintf("08123456%04d", buyerID),
				PaymentMethod: "COD",
				Items: []model.CreateOrderItemReq{
					{
						ProductID: product.ID.String(),
						Quantity:  1,
					},
				},
			}

			// Wait for barrier signal so all goroutines trigger simultaneously
			<-startBarrier

			_, err := orderSvc.Create(context.Background(), req)
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			} else {
				var appErr *apperror.AppError
				if errors.As(err, &appErr) && appErr.Code == fiber.StatusConflict {
					atomic.AddInt64(&conflictCount, 1)
				} else {
					atomic.AddInt64(&otherErrorCount, 1)
					t.Logf("unexpected error for buyer %d: %v", buyerID, err)
				}
			}
		}(i)
	}

	// Release the barrier
	close(startBarrier)
	wg.Wait()

	// 3. Verify results
	assert.Equal(t, int64(initialStock), successCount, "successful orders must equal initial stock")
	assert.Equal(t, int64(totalRequests-initialStock), conflictCount, "rejected orders must equal excess requests")
	assert.Equal(t, int64(0), otherErrorCount, "no unexpected errors should occur")

	// 4. Verify product stock in DB is exactly 0 (no negative stock)
	var reloadedProduct entity.Product
	err := db.First(&reloadedProduct, "id = ?", product.ID).Error
	require.NoError(t, err)
	assert.Equal(t, 0, reloadedProduct.Stock, "final stock in DB must be exactly 0")

	// 5. Verify total orders created in DB is exactly 10
	var orderCount int64
	err = db.Model(new(entity.Order)).Count(&orderCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(initialStock), orderCount, "total orders in DB must match initial stock")
}

// TestOrderService_ConcurrentCheckout_LastItemRace tests the extreme edge case
// where only 1 item remains in stock and 10 goroutines try to buy it simultaneously.
func TestOrderService_ConcurrentCheckout_LastItemRace(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	testutil.CleanDB(t, db)
	defer testutil.CleanDB(t, db)

	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	prodRepo := repository.NewProductRepository()
	orderRepo := repository.NewOrderRepository()
	orderSvc := service.NewOrderService(db, orderRepo, prodRepo, logger, val)

	product := &entity.Product{
		Name:        "Last Concert Ticket",
		Slug:        "last-concert-ticket",
		Description: "Last single item",
		Price:       1500000,
		Stock:       1, // only 1 in stock!
	}
	require.NoError(t, db.Create(product).Error)

	totalRequests := 10
	var successCount int64
	var conflictCount int64

	startBarrier := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(totalRequests)

	for i := 0; i < totalRequests; i++ {
		go func(buyerID int) {
			defer wg.Done()

			req := &model.CreateOrderReq{
				CustomerName:  fmt.Sprintf("Fan %d", buyerID),
				CustomerPhone: fmt.Sprintf("08123456%04d", buyerID),
				PaymentMethod: "TRANSFER",
				Items: []model.CreateOrderItemReq{
					{
						ProductID: product.ID.String(),
						Quantity:  1,
					},
				},
			}

			<-startBarrier

			_, err := orderSvc.Create(context.Background(), req)
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			} else {
				var appErr *apperror.AppError
				if errors.As(err, &appErr) && appErr.Code == fiber.StatusConflict {
					atomic.AddInt64(&conflictCount, 1)
				}
			}
		}(i)
	}

	close(startBarrier)
	wg.Wait()

	assert.Equal(t, int64(1), successCount, "only 1 order should succeed for the last item")
	assert.Equal(t, int64(totalRequests-1), conflictCount, "all other orders must be rejected")

	var reloadedProduct entity.Product
	err := db.First(&reloadedProduct, "id = ?", product.ID).Error
	require.NoError(t, err)
	assert.Equal(t, 0, reloadedProduct.Stock, "final stock must be 0")
}

// TestOrderService_ConcurrentCancellation_Restock verifies that concurrent cancellations
// correctly restore product stock without race conditions or lost updates.
func TestOrderService_ConcurrentCancellation_Restock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	testutil.CleanDB(t, db)
	defer testutil.CleanDB(t, db)

	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	prodRepo := repository.NewProductRepository()
	orderRepo := repository.NewOrderRepository()
	orderSvc := service.NewOrderService(db, orderRepo, prodRepo, logger, val)

	// Initial product with stock = 10
	product := &entity.Product{
		Name:        "Restock Test Item",
		Slug:        "restock-test-item",
		Description: "Testing restock on cancellation",
		Price:       100000,
		Stock:       10,
	}
	require.NoError(t, db.Create(product).Error)

	// Create 5 orders sequentially, each buying 2 units (stock becomes 0)
	numOrders := 5
	orderIDs := make([]string, numOrders)
	for i := 0; i < numOrders; i++ {
		res, err := orderSvc.Create(context.Background(), &model.CreateOrderReq{
			CustomerName:  fmt.Sprintf("Customer %d", i),
			CustomerPhone: fmt.Sprintf("08123456%04d", i),
			PaymentMethod: "COD",
			Items: []model.CreateOrderItemReq{
				{
					ProductID: product.ID.String(),
					Quantity:  2,
				},
			},
		})
		require.NoError(t, err)
		orderIDs[i] = res.ID
	}

	// Verify stock is now 0
	var midProduct entity.Product
	require.NoError(t, db.First(&midProduct, "id = ?", product.ID).Error)
	assert.Equal(t, 0, midProduct.Stock)

	// Now cancel all 5 orders concurrently
	startBarrier := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(numOrders)

	var cancelErrors int64

	for i := 0; i < numOrders; i++ {
		go func(orderID string) {
			defer wg.Done()
			<-startBarrier

			_, err := orderSvc.UpdateStatus(context.Background(), orderID, &model.UpdateOrderStatusReq{
				Status: entity.StatusCancelled,
			})
			if err != nil {
				atomic.AddInt64(&cancelErrors, 1)
			}
		}(orderIDs[i])
	}

	close(startBarrier)
	wg.Wait()

	assert.Equal(t, int64(0), cancelErrors, "all cancellations should succeed")

	// Verify product stock is accurately restored to 10
	var finalProduct entity.Product
	require.NoError(t, db.First(&finalProduct, "id = ?", product.ID).Error)
	assert.Equal(t, 10, finalProduct.Stock, "stock must be restored to 10 without lost updates")
}
