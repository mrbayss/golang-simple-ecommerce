package repository

import (
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Repository[entity.Category]
	FindBySlug(db *gorm.DB, slug string) (*entity.Category, error)
	FindBySlugExcludingID(db *gorm.DB, slug, id string) (*entity.Category, error)
	FindAllPaginated(db *gorm.DB, page, limit int) ([]entity.Category, int64, error)
}

type categoryRepository struct {
	RepositoryImpl[entity.Category]
}

func NewCategoryRepository() CategoryRepository {
	return &categoryRepository{}
}

func (cr *categoryRepository) FindBySlug(db *gorm.DB, slug string) (*entity.Category, error) {
	var category entity.Category

	if err := db.Where("slug = ?", slug).First(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

func (cr *categoryRepository) FindBySlugExcludingID(db *gorm.DB, slug, id string) (*entity.Category, error) {
	var category entity.Category

	if err := db.Where("slug = ? AND id <> ?", slug, id).First(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

func (cr *categoryRepository) FindAllPaginated(db *gorm.DB, page, limit int) ([]entity.Category, int64, error) {
	var categories []entity.Category
	var total int64

	if err := db.Model(new(entity.Category)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := db.Offset(offset).Limit(limit).Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}
