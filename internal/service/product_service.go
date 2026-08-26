package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model/converter"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/apperror"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/slugutil"
	"github.com/mrbayss/golang-simple-ecommerce/internal/repository"
	"github.com/mrbayss/golang-simple-ecommerce/internal/storage"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ProductService interface {
	GetByID(c context.Context, id string) (*model.ProductRes, error)
	GetBySlug(c context.Context, slug string) (*model.ProductRes, error)
	GetAll(c context.Context, page, limit int) (*model.ProductListRes, error)
	Create(c context.Context, request *model.CreateProductReq, images []*multipart.FileHeader) (*model.ProductRes, error)
	Update(c context.Context, id string, request *model.UpdateProductReq, newImages []*multipart.FileHeader) (*model.ProductRes, error)
	Delete(c context.Context, id string) error
	DeleteImage(c context.Context, imageID string) error
}

type productService struct {
	DB                 *gorm.DB
	ProductRepository  repository.ProductRepository
	CategoryRepository repository.CategoryRepository
	FileStorage        storage.FileStorage
	Log                *logrus.Logger
	Validate           *validator.Validate
}

func NewProductService(
	db *gorm.DB,
	productRepository repository.ProductRepository,
	categoryRepository repository.CategoryRepository,
	fileStorage storage.FileStorage,
	log *logrus.Logger,
	validate *validator.Validate,
) ProductService {
	return &productService{
		DB:                 db,
		ProductRepository:  productRepository,
		CategoryRepository: categoryRepository,
		FileStorage:        fileStorage,
		Log:                log,
		Validate:           validate,
	}
}

func (ps *productService) GetByID(c context.Context, id string) (*model.ProductRes, error) {
	ctx := ps.DB.WithContext(c)

	if _, err := uuid.Parse(id); err != nil {
		return nil, apperror.NewAppError(fiber.StatusBadRequest, "invalid id format")
	}

	product, err := ps.ProductRepository.GetByIDWithRelations(ctx, id)
	if err != nil {
		ps.Log.Errorf("failed to get product by id: %v", err)
		return nil, err
	}

	return converter.ToProductRes(product), nil
}

func (ps *productService) GetBySlug(c context.Context, slug string) (*model.ProductRes, error) {
	ctx := ps.DB.WithContext(c)

	product, err := ps.ProductRepository.FindBySlug(ctx, slug)
	if err != nil {
		ps.Log.Errorf("failed to get product by slug: %v", err)
		return nil, err
	}

	return converter.ToProductRes(product), nil
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
		items = append(items, *converter.ToProductRes(&product))
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

// Create stores a new product with its uploaded images in one transaction.
func (ps *productService) Create(c context.Context, request *model.CreateProductReq, images []*multipart.FileHeader) (*model.ProductRes, error) {
	ctx := ps.DB.WithContext(c)

	if err := ps.Validate.Struct(request); err != nil {
		ps.Log.Warnf("create product rejected, invalid fields: %s", apperror.ValidationFields(err))
		return nil, apperror.NewAppError(fiber.StatusBadRequest, "invalid request fields")
	}

	slug := slugutil.Slugify(request.Slug)
	if slug == "" {
		slug = slugutil.Slugify(request.Name)
	}
	if err := ps.ensureSlugFree(c, slug, ""); err != nil {
		return nil, err
	}

	var categoryID *uuid.UUID
	if request.CategoryID != "" {
		catID, err := uuid.Parse(request.CategoryID)
		if err != nil {
			return nil, apperror.NewAppError(fiber.StatusBadRequest, "invalid category_id format")
		}
		if _, err := ps.CategoryRepository.GetByID(ctx, catID.String()); err != nil {
			ps.Log.Warnf("create product rejected, category not found: %s", request.CategoryID)
			return nil, apperror.NewAppError(fiber.StatusBadRequest, "category not found")
		}
		categoryID = &catID
	}

	product := &entity.Product{
		CategoryID:  categoryID,
		Name:        request.Name,
		Slug:        slug,
		Description: request.Description,
		Price:       entity.Money(request.Price),
		Stock:       request.Stock,
		Weight:      request.Weight,
	}

	err := ps.DB.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if _, err := ps.ProductRepository.Create(tx, product); err != nil {
			return fmt.Errorf("create product: %w", err)
		}
		if len(images) > 0 {
			if err := ps.saveImages(tx, product.ID, images); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		// AppError = business rejection; anything else = technical failure.
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		ps.Log.Errorf("failed to create product: %v", err)
		return nil, apperror.NewAppError(fiber.StatusInternalServerError, "internal server error")
	}

	ps.Log.Infof("product created: %s (%s), %d image(s)", product.Name, product.Slug, len(images))

	saved, err := ps.ProductRepository.GetByIDWithRelations(ctx, product.ID.String())
	if err != nil {
		ps.Log.Errorf("failed to reload created product: %v", err)
		return nil, err
	}

	return converter.ToProductRes(saved), nil
}

// Update applies only the sent fields and optionally adds new images.
func (ps *productService) Update(c context.Context, id string, request *model.UpdateProductReq, newImages []*multipart.FileHeader) (*model.ProductRes, error) {
	ctx := ps.DB.WithContext(c)

	if _, err := uuid.Parse(id); err != nil {
		ps.Log.Warnf("update product rejected, invalid id format: %s", id)
		return nil, apperror.NewAppError(fiber.StatusBadRequest, "invalid id format")
	}

	product, err := ps.ProductRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ps.Log.Warnf("update rejected, product not found: %s", id)
		} else {
			ps.Log.Errorf("failed to find product %s: %v", id, err)
		}
		return nil, err
	}

	updates := map[string]any{}

	if request.Name != nil && *request.Name != "" {
		updates["name"] = *request.Name
	}
	if request.Description != nil {
		updates["description"] = *request.Description
	}
	if request.Price != nil {
		updates["price"] = entity.Money(*request.Price)
	}
	if request.Stock != nil {
		updates["stock"] = *request.Stock
	}
	if request.Weight != nil {
		updates["weight"] = *request.Weight
	}

	if request.Slug != nil && *request.Slug != "" {
		slug := slugutil.Slugify(*request.Slug)
		if slug == "" && updates["name"] != nil {
			slug = slugutil.Slugify(updates["name"].(string))
		}
		if slug != "" {
			if err := ps.ensureSlugFree(c, slug, id); err != nil {
				return nil, err
			}
			updates["slug"] = slug
		}
	} else if updates["name"] != nil {
		// Name changed without an explicit slug: keep the slug stable.
		_ = updates["name"]
	}

	if request.CategoryID != nil {
		if *request.CategoryID == "" {
			updates["category_id"] = nil
		} else {
			catID, err := uuid.Parse(*request.CategoryID)
			if err != nil {
				return nil, apperror.NewAppError(fiber.StatusBadRequest, "invalid category_id format")
			}
			if _, err := ps.CategoryRepository.GetByID(ctx, catID.String()); err != nil {
				ps.Log.Warnf("update product rejected, category not found: %s", *request.CategoryID)
				return nil, apperror.NewAppError(fiber.StatusBadRequest, "category not found")
			}
			updates["category_id"] = catID
		}
	}

	err = ps.DB.WithContext(c).Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(product).Updates(updates).Error; err != nil {
				return fmt.Errorf("update product: %w", err)
			}
		}
		if len(newImages) > 0 {
			if err := ps.saveImages(tx, product.ID, newImages); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		ps.Log.Errorf("failed to update product %s: %v", id, err)
		return nil, apperror.NewAppError(fiber.StatusInternalServerError, "internal server error")
	}

	ps.Log.Infof("product updated: %s (%d field(s), %d new image(s))", product.Slug, len(updates), len(newImages))

	saved, err := ps.ProductRepository.GetByIDWithRelations(ctx, id)
	if err != nil {
		ps.Log.Errorf("failed to reload updated product: %v", err)
		return nil, err
	}

	return converter.ToProductRes(saved), nil
}

// Delete soft-deletes the product. Stored image files are removed from disk.
func (ps *productService) Delete(c context.Context, id string) error {
	ctx := ps.DB.WithContext(c)

	if _, err := uuid.Parse(id); err != nil {
		ps.Log.Warnf("delete product rejected, invalid id format: %s", id)
		return apperror.NewAppError(fiber.StatusBadRequest, "invalid id format")
	}

	product, err := ps.ProductRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ps.Log.Warnf("delete rejected, product not found: %s", id)
		} else {
			ps.Log.Errorf("failed to find product %s: %v", id, err)
		}
		return err
	}

	if err := ps.ProductRepository.Delete(ctx, id); err != nil {
		ps.Log.Errorf("failed to delete product %s: %v", id, err)
		return err
	}

	ps.removeFiles(product)

	ps.Log.Infof("product deleted: %s (%s)", product.Name, product.Slug)
	return nil
}

// DeleteImage removes a single product image row and its file.
func (ps *productService) DeleteImage(c context.Context, imageID string) error {
	ctx := ps.DB.WithContext(c)

	if _, err := uuid.Parse(imageID); err != nil {
		ps.Log.Warnf("delete image rejected, invalid id format: %s", imageID)
		return apperror.NewAppError(fiber.StatusBadRequest, "invalid id format")
	}

	image, err := ps.ProductRepository.FindImageByID(ctx, imageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ps.Log.Warnf("delete image rejected, image not found: %s", imageID)
		} else {
			ps.Log.Errorf("failed to find image %s: %v", imageID, err)
		}
		return err
	}

	if err := ps.DB.WithContext(c).Delete(image).Error; err != nil {
		ps.Log.Errorf("failed to delete image %s: %v", imageID, err)
		return err
	}

	if err := ps.FileStorage.Delete(c, image.ImageURL); err != nil {
		// Row is already gone; log the orphaned file for manual cleanup.
		ps.Log.Warnf("image row deleted but file remains: %v", err)
	}

	ps.Log.Infof("product image deleted: %s", image.ImageURL)
	return nil
}

// saveImages uploads every file and inserts one ProductImage row per file.
// On any failure the already-uploaded files are removed again — the caller's
// transaction rolls back the rows.
func (ps *productService) saveImages(tx *gorm.DB, productID uuid.UUID, files []*multipart.FileHeader) error {
	savedURLs := make([]string, 0, len(files))
	defer func() {
		// Runs on error return paths only when savedURLs is captured below.
	}()

	isPrimary := true // first image becomes the primary one
	for _, file := range files {
		url, err := ps.FileStorage.Save(context.Background(), "product", file)
		if err != nil {
			ps.cleanupFiles(savedURLs)
			ps.Log.Warnf("upload rejected: %v", err)
			return apperror.NewAppError(fiber.StatusBadRequest, fmt.Sprintf("invalid image %q: %v", file.Filename, err))
		}
		savedURLs = append(savedURLs, url)

		image := &entity.ProductImage{
			ProductID: productID,
			ImageURL:  url,
			IsPrimary: isPrimary,
		}
		if err := tx.Create(image).Error; err != nil {
			ps.cleanupFiles(append(savedURLs, url))
			return fmt.Errorf("save image record %s: %w", url, err)
		}
		isPrimary = false
	}
	return nil
}

func (ps *productService) ensureSlugFree(ctx context.Context, slug, excludingID string) error {
	_, err := ps.ProductRepository.FindBySlugExcludingID(ps.DB.WithContext(ctx), slug, excludingID)
	if err == nil {
		ps.Log.Warnf("slug already used: %s", slug)
		return apperror.NewAppError(fiber.StatusConflict, "product slug already in use")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ps.Log.Errorf("failed to check slug %s: %v", slug, err)
		return err
	}
	return nil
}

func (ps *productService) removeFiles(product *entity.Product) {
	ps.cleanupFiles(imageURLs(product.Images))
}

func (ps *productService) cleanupFiles(urls []string) {
	for _, url := range urls {
		if err := ps.FileStorage.Delete(context.Background(), url); err != nil {
			ps.Log.Warnf("failed to remove stored file %s: %v", url, err)
		}
	}
}

func imageURLs(images []entity.ProductImage) []string {
	urls := make([]string, 0, len(images))
	for _, img := range images {
		urls = append(urls, img.ImageURL)
	}
	return urls
}
