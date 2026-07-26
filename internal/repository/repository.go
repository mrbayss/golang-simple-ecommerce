package repository

import (
	"time"

	"gorm.io/gorm"
)

type Repository[T any] interface {
	Create(db *gorm.DB, entity *T) (*T, error)
	GetByID(db *gorm.DB, id string) (*T, error)
	Update(db *gorm.DB, entity *T, id string) error
	Delete(db *gorm.DB, id string) error
	SoftDelete(db *gorm.DB, id string) error
	GetAll(db *gorm.DB) ([]T, error)
	GetByColumn(db *gorm.DB, column string, value any) (*T, error)
}

type RepositoryImpl[T any] struct{}

func (r *RepositoryImpl[T]) Create(db *gorm.DB, entity *T) (*T, error) {
	if err := db.Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *RepositoryImpl[T]) GetByID(db *gorm.DB, id string) (*T, error) {
	var t T

	if err := db.Where("id = ?", id).Take(&t).Error; err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *RepositoryImpl[T]) Update(db *gorm.DB, entity *T, id string) error {
	return db.Model(new(T)).Where("id = ?", id).Updates(entity).Error
}

func (r *RepositoryImpl[T]) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(new(T)).Error
}

func (r *RepositoryImpl[T]) GetAll(db *gorm.DB) ([]T, error) {
	var t []T

	if err := db.Find(&t).Error; err != nil {
		return nil, err
	}

	return t, nil
}

func (r *RepositoryImpl[T]) SoftDelete(db *gorm.DB, id string) error {
	return db.Model(new(T)).Where("id = ?", id).Update("deleted_at", time.Now().Unix()).Error

}

func (r *RepositoryImpl[T]) GetByColumn(db *gorm.DB, column string, value any) (*T, error) {
	var t T
	return &t, db.Where(map[string]any{column: value}).First(&t).Error
}
