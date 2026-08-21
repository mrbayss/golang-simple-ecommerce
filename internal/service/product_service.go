package service

import (
	"context"

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

type ProductService interface {
	GetByID(c context.Context, id string) (*model.ProductRes, error)
	GetBySlug(c context.Context, slug string) (*model.ProductRes, error)
	GetAll(c context.Context, page, limit int) (*model.ProductListRes, error)
}

type productService struct {
	DB              *gorm.DB
	ProductRepository repository.ProductRepository
	Log             *logrus.Logger
	Validate        *validator.Validate
}

func NewProductService(db *gorm.DB, productRepository repository.ProductRepository, log *logrus.Logger, validate *validator.Validate) ProductService {
	return &productService{
		DB:                db,
		ProductRepository: productRepository,
		Log:               log,
		Validate:          validate,
	}
}

func (ps *productService) GetByID(c context.Context, id string) (*model.ProductRes, error) {
	ctx := ps.DB.WithContext(c)

	if _, err := uuid.Parse(id); err != nil {
		return nil, apperror.NewAppError(fiber.StatusBadRequest, "Format id tidak valid")
	}

	product, err := ps.ProductRepository.GetByID(ctx, id)
	if err != nil {
		ps.Log.Errorf("failed to get product by id: %v", err)
		return nil, err
	}

	return toProductRes(product), nil
}

func (ps *productService) GetBySlug(c context.Context, slug string) (*model.ProductRes, error) {
	ctx := ps.DB.WithContext(c)

	product, err := ps.ProductRepository.FindBySlug(ctx, slug)
	if err != nil {
		ps.Log.Errorf("failed to get product by slug: %v", err)
		return nil, err
	}

	return toProductRes(product), nil
}

func (ps *productService) GetAll(c context.Context, page, limit int) (*model.ProductListRes, error) {
	ctx := ps.DB.WithContext(c)

	products, total, err := ps.ProductRepository.FindAllPaginated(ctx, page, limit)
	if err != nil {
		ps.Log.Errorf("failed to get all products: %v", err)
		return nil, err
	}

	items := make([]model.ProductRes, 0, len(products))
	for _, product := range products {
		items = append(items, *toProductRes(&product))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &model.PaginatedRes[model.ProductRes]{
		Data: items,
		Meta: model.PageMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func toProductRes(product *entity.Product) *model.ProductRes {
	res := &model.ProductRes{
		ID:          product.ID.String(),
		Name:        product.Name,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       int64(product.Price),
		Stock:       product.Stock,
		Weight:      product.Weight,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
		Images:      make([]model.ProductImageRes, 0, len(product.Images)),
	}

	if product.CategoryID != nil && product.Category != nil {
		categoryID := product.CategoryID.String()
		res.CategoryID = &categoryID
		res.Category = toCategoryRes(product.Category)
	}

	for _, img := range product.Images {
		res.Images = append(res.Images, model.ProductImageRes{
			ID:        img.ID.String(),
			ImageURL:  img.ImageURL,
			IsPrimary: img.IsPrimary,
		})
	}

	return res
}
