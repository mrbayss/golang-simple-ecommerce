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

func TestCategoryService_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	t.Run("Success create category", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		req := &model.CreateCategoryReq{
			Name: "Fashion Pria",
			Slug: "fashion-pria",
		}

		mockRepo.On("FindBySlug", mock.Anything, req.Slug).
			Return(nil, gorm.ErrRecordNotFound).Once()

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Category) bool {
			return c.Name == req.Name && c.Slug == req.Slug
		})).Return(&entity.Category{
			ID:        uuid.New(),
			Name:      req.Name,
			Slug:      req.Slug,
			CreatedAt: time.Now(),
		}, nil).Once()

		res, err := categorySvc.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, req.Name, res.Name)
		assert.Equal(t, req.Slug, res.Slug)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Validation error on empty fields", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		req := &model.CreateCategoryReq{
			Name: "",
			Slug: "",
		}

		res, err := categorySvc.Create(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertNotCalled(t, "FindBySlug")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Conflict when slug already exists", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		req := &model.CreateCategoryReq{
			Name: "Fashion Pria",
			Slug: "fashion-pria",
		}

		mockRepo.On("FindBySlug", mock.Anything, req.Slug).
			Return(&entity.Category{
				ID:   uuid.New(),
				Name: "Fashion Pria",
				Slug: "fashion-pria",
			}, nil).Once()

		res, err := categorySvc.Create(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestCategoryService_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	categoryID := uuid.New()

	t.Run("Success get category by valid ID", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		mockRepo.On("GetByID", mock.Anything, categoryID.String()).
			Return(&entity.Category{
				ID:        categoryID,
				Name:      "Elektronik",
				Slug:      "elektronik",
				CreatedAt: time.Now(),
			}, nil).Once()

		res, err := categorySvc.GetByID(context.Background(), categoryID.String())
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, categoryID.String(), res.ID)
		assert.Equal(t, "Elektronik", res.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error on invalid UUID format", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		res, err := categorySvc.GetByID(context.Background(), "invalid-uuid")
		require.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertNotCalled(t, "GetByID")
	})

	t.Run("Error when category not found", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		mockRepo.On("GetByID", mock.Anything, categoryID.String()).
			Return(nil, gorm.ErrRecordNotFound).Once()

		res, err := categorySvc.GetByID(context.Background(), categoryID.String())
		require.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryService_GetAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	mockRepo := new(testutil.MockCategoryRepository)
	categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

	categories := []entity.Category{
		{ID: uuid.New(), Name: "Kategori 1", Slug: "kategori-1"},
		{ID: uuid.New(), Name: "Kategori 2", Slug: "kategori-2"},
	}

	mockRepo.On("FindAllPaginated", mock.Anything, 1, 10).
		Return(categories, int64(2), nil).Once()

	res, err := categorySvc.GetAll(context.Background(), 1, 10)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Data, 2)
	assert.Equal(t, int64(2), res.Meta.Total)
	assert.Equal(t, 1, res.Meta.Page)
	assert.Equal(t, 10, res.Meta.Limit)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	categoryID := uuid.New()

	t.Run("Success update category", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		req := &model.UpdateCategoryReq{
			Name: "Fashion Pria Baru",
			Slug: "fashion-pria-baru",
		}

		mockRepo.On("GetByID", mock.Anything, categoryID.String()).
			Return(&entity.Category{
				ID:   categoryID,
				Name: "Fashion Pria Lama",
				Slug: "fashion-pria-lama",
			}, nil).Once()

		mockRepo.On("FindBySlugExcludingID", mock.Anything, req.Slug, categoryID.String()).
			Return(nil, gorm.ErrRecordNotFound).Once()

		mockRepo.On("Update", mock.Anything, mock.Anything, categoryID.String()).
			Return(nil).Once()

		res, err := categorySvc.Update(context.Background(), categoryID.String(), req)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, req.Name, res.Name)
		assert.Equal(t, req.Slug, res.Slug)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error when slug already used by another category", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		req := &model.UpdateCategoryReq{
			Name: "Fashion Pria Baru",
			Slug: "fashion-pria-conflict",
		}

		mockRepo.On("GetByID", mock.Anything, categoryID.String()).
			Return(&entity.Category{
				ID:   categoryID,
				Name: "Fashion Pria",
				Slug: "fashion-pria",
			}, nil).Once()

		otherID := uuid.New()
		mockRepo.On("FindBySlugExcludingID", mock.Anything, req.Slug, categoryID.String()).
			Return(&entity.Category{
				ID:   otherID,
				Name: "Other Category",
				Slug: req.Slug,
			}, nil).Once()

		res, err := categorySvc.Update(context.Background(), categoryID.String(), req)
		require.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertNotCalled(t, "Update")
	})
}

func TestCategoryService_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	logger := testutil.NewTestLogger()
	val := testutil.NewTestValidator()

	categoryID := uuid.New()

	t.Run("Success delete category", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		mockRepo.On("GetByID", mock.Anything, categoryID.String()).
			Return(&entity.Category{
				ID:   categoryID,
				Name: "Gadgets",
				Slug: "gadgets",
			}, nil).Once()

		mockRepo.On("Delete", mock.Anything, categoryID.String()).
			Return(nil).Once()

		err := categorySvc.Delete(context.Background(), categoryID.String())
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error when deleting non-existent category", func(t *testing.T) {
		mockRepo := new(testutil.MockCategoryRepository)
		categorySvc := service.NewCategoryService(db, mockRepo, logger, val)

		mockRepo.On("GetByID", mock.Anything, categoryID.String()).
			Return(nil, gorm.ErrRecordNotFound).Once()

		err := categorySvc.Delete(context.Background(), categoryID.String())
		require.Error(t, err)
		mockRepo.AssertNotCalled(t, "Delete")
	})
}
