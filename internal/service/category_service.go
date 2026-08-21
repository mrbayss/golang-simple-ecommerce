package service

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/mrbayss/golang-simple-ecommerce/internal/repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CategoryService interface {
	Create(c context.Context, request *model.CreateCategoryReq) (*model.CategoryRes, error)
	GetByID(c context.Context, id string) (*model.CategoryRes, error)
	GetAll(c context.Context, page, limit int) (*model.PaginatedRes[model.CategoryRes], error)
	Update(c context.Context, id string, request *model.UpdateCategoryReq) (*model.CategoryRes, error)
	Delete(c context.Context, id string) error
}

type categoryService struct {
	DB                 *gorm.DB
	CategoryRepository repository.CategoryRepository
	Log                *logrus.Logger
	Validate           *validator.Validate
}

func NewCategoryService(db *gorm.DB, categoryRepository repository.CategoryRepository, log *logrus.Logger, validate *validator.Validate) CategoryService {
	return &categoryService{
		DB:                 db,
		CategoryRepository: categoryRepository,
		Log:                log,
		Validate:           validate,
	}
}

func (cs *categoryService) Create(c context.Context, request *model.CreateCategoryReq) (*model.CategoryRes, error) {
	ctx := cs.DB.WithContext(c)

	if err := cs.Validate.Struct(request); err != nil {
		cs.Log.Warnf("category rejected, invalid fields: %s", apperror.ValidationFields(err))
		return nil, err
	}

	if _, err := cs.CategoryRepository.FindBySlug(ctx, request.Slug); err == nil {
		cs.Log.Warnf("slug already used: %s", request.Slug)
		return nil, apperror.NewAppError(fiber.StatusConflict, "category slug already in use")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		cs.Log.Errorf("failed to find category by slug: %v", err)
		return nil, err
	}

	category := &entity.Category{
		Name: request.Name,
		Slug: request.Slug,
	}

	if _, err := cs.CategoryRepository.Create(ctx, category); err != nil {
		cs.Log.Errorf("failed to create category: %v", err)
		return nil, err
	}

	cs.Log.Infof("category created: %s", category.Name)

	return toCategoryRes(category), nil
}

func (cs *categoryService) GetByID(c context.Context, id string) (*model.CategoryRes, error) {
	ctx := cs.DB.WithContext(c)

	if _, err := uuid.Parse(id); err != nil {
		return nil, apperror.NewAppError(fiber.StatusBadRequest, "invalid id format")
	}

	category, err := cs.CategoryRepository.GetByID(ctx, id)
	if err != nil {
		cs.Log.Errorf("failed to get category by id: %v", err)
		return nil, err
	}

	return toCategoryRes(category), nil
}

func (cs *categoryService) GetAll(c context.Context, page, limit int) (*model.PaginatedRes[model.CategoryRes], error) {
	ctx := cs.DB.WithContext(c)

	categories, total, err := cs.CategoryRepository.FindAllPaginated(ctx, page, limit)
	if err != nil {
		cs.Log.Errorf("failed to get all categories: %v", err)
		return nil, err
	}

	items := make([]model.CategoryRes, 0, len(categories))
	for _, category := range categories {
		items = append(items, *toCategoryRes(&category))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &model.PaginatedRes[model.CategoryRes]{
		Data: items,
		Meta: model.PageMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (cs *categoryService) Update(c context.Context, id string, request *model.UpdateCategoryReq) (*model.CategoryRes, error) {
	ctx := cs.DB.WithContext(c)

	if _, err := uuid.Parse(id); err != nil {
		return nil, apperror.NewAppError(fiber.StatusBadRequest, "invalid id format")
	}

	if err := cs.Validate.Struct(request); err != nil {
		cs.Log.Warnf("category rejected, invalid fields: %s", apperror.ValidationFields(err))
		return nil, err
	}

	category, err := cs.CategoryRepository.GetByID(ctx, id)
	if err != nil {
		cs.Log.Errorf("failed to get category by id: %v", err)
		return nil, err
	}

	if _, err := cs.CategoryRepository.FindBySlugExcludingID(ctx, request.Slug, id); err == nil {
		cs.Log.Warnf("slug already used by other category: %s", request.Slug)
		return nil, apperror.NewAppError(fiber.StatusConflict, "category slug already in use")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		cs.Log.Errorf("failed to find category by slug: %v", err)
		return nil, err
	}

	category.Name = request.Name
	category.Slug = request.Slug

	if err := cs.CategoryRepository.Update(ctx, category, id); err != nil {
		cs.Log.Errorf("failed to update category: %v", err)
		return nil, err
	}

	cs.Log.Infof("category updated: %s", category.Name)

	return toCategoryRes(category), nil
}

func (cs *categoryService) Delete(c context.Context, id string) error {
	ctx := cs.DB.WithContext(c)

	if _, err := uuid.Parse(id); err != nil {
		return apperror.NewAppError(fiber.StatusBadRequest, "invalid id format")
	}

	if _, err := cs.CategoryRepository.GetByID(ctx, id); err != nil {
		cs.Log.Errorf("failed to get category by id: %v", err)
		return err
	}

	if err := cs.CategoryRepository.Delete(ctx, id); err != nil {
		cs.Log.Errorf("failed to delete category: %v", err)
		return err
	}

	cs.Log.Infof("category deleted: %s", id)

	return nil
}

func toCategoryRes(category *entity.Category) *model.CategoryRes {
	return &model.CategoryRes{
		ID:        category.ID.String(),
		Name:      category.Name,
		Slug:      category.Slug,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}
