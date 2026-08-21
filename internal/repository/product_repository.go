package repository

import (
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository interface {
	Repository[entity.Product]
	FindBySlug(db *gorm.DB, slug string) (*entity.Product, error)
	FindAllPaginated(db *gorm.DB, page, limit int) ([]entity.Product, int64, error)
	FindByIDForUpdate(db *gorm.DB, id string) (*entity.Product, error)
	UpdateStock(db *gorm.DB, id string, delta int) error
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

// FindByIDForUpdate locks the row (SELECT ... FOR UPDATE) until the
// surrounding transaction commits or rolls back.
func (pr *productRepository) FindByIDForUpdate(db *gorm.DB, id string) (*entity.Product, error) {
	var product entity.Product

	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// UpdateStock applies a signed delta to the stock column.
func (pr *productRepository) UpdateStock(db *gorm.DB, id string, delta int) error {
	return db.Model(new(entity.Product)).Where("id = ?", id).
		Update("stock", gorm.Expr("stock + ?", delta)).Error
}
