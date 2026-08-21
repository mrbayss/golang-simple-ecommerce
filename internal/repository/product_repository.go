package repository

import (
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Repository[entity.Product]
	FindBySlug(db *gorm.DB, slug string) (*entity.Product, error)
	FindAllPaginated(db *gorm.DB, page, limit int) ([]entity.Product, int64, error)
}

type productRepository struct {
	RepositoryImpl[entity.Product]
}

func NewProductRepository() ProductRepository {
	return &productRepository{}
}

func (pr *productRepository) baseQuery(db *gorm.DB) *gorm.DB {
	return db.Preload("Category").Preload("Images")
}

func (pr *productRepository) FindBySlug(db *gorm.DB, slug string) (*entity.Product, error) {
	var product entity.Product

	if err := pr.baseQuery(db).Where("slug = ?", slug).First(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (pr *productRepository) FindAllPaginated(db *gorm.DB, page, limit int) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64

	if err := db.Model(new(entity.Product)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := pr.baseQuery(db).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
