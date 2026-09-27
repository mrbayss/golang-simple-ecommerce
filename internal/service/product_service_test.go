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

func TestProductService_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	t.Run("Success create product without images", func(t *testing.T) {
		mockProdRepo := new(testutil.MockProductRepository)
		mockCatRepo := new(testutil.MockCategoryRepository)
		mockStorage := new(testutil.MockFileStorage)
		productSvc := service.NewProductService(db, mockProdRepo, mockCatRepo, mockStorage, logger, val)

		req := &model.CreateProductReq{
			Name:        "Sepatu Running",
			Slug:        "sepatu-running",
			Description: "Sepatu lari ringan",
			Price:       250000,
			Stock:       15,
			Weight:      500,
		}

		mockProdRepo.On("FindBySlugExcludingID", mock.Anything, "sepatu-running", "").
			Return(nil, gorm.ErrRecordNotFound).Once()

		mockProdRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p.Name == req.Name && p.Stock == 15
		})).Return(&entity.Product{
			ID:    uuid.New(),
			Name:  req.Name,
			Slug:  req.Slug,
			Price: entity.Money(req.Price),
			Stock: req.Stock,
		}, nil).Once()

		savedProduct := &entity.Product{
			ID:        uuid.New(),
			Name:      req.Name,
			Slug:      req.Slug,
			Price:     entity.Money(req.Price),
			Stock:     req.Stock,
			CreatedAt: time.Now(),
		}
		mockProdRepo.On("GetByIDWithRelations", mock.Anything, mock.Anything).
			Return(savedProduct, nil).Once()

		res, err := productSvc.Create(context.Background(), req, nil)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, req.Name, res.Name)
		assert.Equal(t, req.Stock, res.Stock)
		mockProdRepo.AssertExpectations(t)
	})

	t.Run("Validation error on invalid price or name", func(t *testing.T) {
		mockProdRepo := new(testutil.MockProductRepository)
		mockCatRepo := new(testutil.MockCategoryRepository)
		mockStorage := new(testutil.MockFileStorage)
		productSvc := service.NewProductService(db, mockProdRepo, mockCatRepo, mockStorage, logger, val)

		req := &model.CreateProductReq{
			Name:  "AB", // min=3
			Price: 0,    // gt=0
			Stock: -1,   // gte=0
		}

		res, err := productSvc.Create(context.Background(), req, nil)
		require.Error(t, err)
		assert.Nil(t, res)
		mockProdRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Error when category does not exist", func(t *testing.T) {
		mockProdRepo := new(testutil.MockProductRepository)
		mockCatRepo := new(testutil.MockCategoryRepository)
		mockStorage := new(testutil.MockFileStorage)
		productSvc := service.NewProductService(db, mockProdRepo, mockCatRepo, mockStorage, logger, val)

		nonExistentCatID := uuid.New().String()
		req := &model.CreateProductReq{
			Name:       "Baju Batik",
			Price:      100000,
			Stock:      5,
			CategoryID: nonExistentCatID,
		}

		mockProdRepo.On("FindBySlugExcludingID", mock.Anything, "baju-batik", "").
			Return(nil, gorm.ErrRecordNotFound).Once()

		mockCatRepo.On("GetByID", mock.Anything, nonExistentCatID).
			Return(nil, gorm.ErrRecordNotFound).Once()

		res, err := productSvc.Create(context.Background(), req, nil)
		require.Error(t, err)
		assert.Nil(t, res)
		mockProdRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Conflict when slug is already in use", func(t *testing.T) {
		mockProdRepo := new(testutil.MockProductRepository)
		mockCatRepo := new(testutil.MockCategoryRepository)
		mockStorage := new(testutil.MockFileStorage)
		productSvc := service.NewProductService(db, mockProdRepo, mockCatRepo, mockStorage, logger, val)

		req := &model.CreateProductReq{
			Name:  "Baju Koko",
			Slug:  "baju-koko",
			Price: 85000,
			Stock: 10,
		}

		mockProdRepo.On("FindBySlugExcludingID", mock.Anything, "baju-koko", "").
			Return(&entity.Product{
				ID:   uuid.New(),
				Name: "Existing Product",
				Slug: "baju-koko",
			}, nil).Once()

		res, err := productSvc.Create(context.Background(), req, nil)
		require.Error(t, err)
		assert.Nil(t, res)
		mockProdRepo.AssertNotCalled(t, "Create")
	})
}

func TestProductService_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	productID := uuid.New()

	t.Run("Success get product by ID", func(t *testing.T) {
		mockProdRepo := new(testutil.MockProductRepository)
		mockCatRepo := new(testutil.MockCategoryRepository)
		mockStorage := new(testutil.MockFileStorage)
		productSvc := service.NewProductService(db, mockProdRepo, mockCatRepo, mockStorage, logger, val)

		mockProdRepo.On("GetByIDWithRelations", mock.Anything, productID.String()).
			Return(&entity.Product{
				ID:        productID,
				Name:      "Jam Tangan",
				Slug:      "jam-tangan",
				Price:     150000,
				Stock:     10,
				CreatedAt: time.Now(),
			}, nil).Once()

		res, err := productSvc.GetByID(context.Background(), productID.String())
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, productID.String(), res.ID)
		assert.Equal(t, "Jam Tangan", res.Name)
		mockProdRepo.AssertExpectations(t)
	})

	t.Run("Error on invalid UUID", func(t *testing.T) {
		mockProdRepo := new(testutil.MockProductRepository)
		productSvc := service.NewProductService(db, mockProdRepo, nil, nil, logger, val)

		res, err := productSvc.GetByID(context.Background(), "invalid-uuid")
		require.Error(t, err)
		assert.Nil(t, res)
		mockProdRepo.AssertNotCalled(t, "GetByIDWithRelations")
	})
}

func TestProductService_GetBySlug(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	mockProdRepo := new(testutil.MockProductRepository)
	productSvc := service.NewProductService(db, mockProdRepo, nil, nil, logger, val)

	mockProdRepo.On("FindBySlug", mock.Anything, "kamera-dslr").
		Return(&entity.Product{
			ID:        uuid.New(),
			Name:      "Kamera DSLR",
			Slug:      "kamera-dslr",
			Price:     4500000,
			Stock:     3,
			CreatedAt: time.Now(),
		}, nil).Once()

	res, err := productSvc.GetBySlug(context.Background(), "kamera-dslr")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "Kamera DSLR", res.Name)
	mockProdRepo.AssertExpectations(t)
}

func TestProductService_GetAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	mockProdRepo := new(testutil.MockProductRepository)
	productSvc := service.NewProductService(db, mockProdRepo, nil, nil, logger, val)

	products := []entity.Product{
		{ID: uuid.New(), Name: "Produk 1", Slug: "produk-1", Price: 10000, Stock: 5},
		{ID: uuid.New(), Name: "Produk 2", Slug: "produk-2", Price: 20000, Stock: 10},
	}

	mockProdRepo.On("FindAllPaginated", mock.Anything, 1, 10).
		Return(products, int64(2), nil).Once()

	res, err := productSvc.GetAll(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Data, 2)
	assert.Equal(t, int64(2), res.Meta.Total)
	mockProdRepo.AssertExpectations(t)
}

func TestProductService_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	productID := uuid.New()

	t.Run("Success delete product", func(t *testing.T) {
		mockProdRepo := new(testutil.MockProductRepository)
		mockStorage := new(testutil.MockFileStorage)
		productSvc := service.NewProductService(db, mockProdRepo, nil, mockStorage, logger, val)

		mockProdRepo.On("GetByID", mock.Anything, productID.String()).
			Return(&entity.Product{
				ID:   productID,
				Name: "Tas Ransel",
			}, nil).Once()

		mockProdRepo.On("Delete", mock.Anything, productID.String()).
			Return(nil).Once()

		err := productSvc.Delete(context.Background(), productID.String())
		require.NoError(t, err)
		mockProdRepo.AssertExpectations(t)
	})

	t.Run("Error when product to delete is not found", func(t *testing.T) {
		mockProdRepo := new(testutil.MockProductRepository)
		productSvc := service.NewProductService(db, mockProdRepo, nil, nil, logger, val)

		mockProdRepo.On("GetByID", mock.Anything, productID.String()).
			Return(nil, gorm.ErrRecordNotFound).Once()

		err := productSvc.Delete(context.Background(), productID.String())
		require.Error(t, err)
		mockProdRepo.AssertNotCalled(t, "Delete")
	})
}
