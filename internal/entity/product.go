package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID          uuid.UUID      `gorm:"column:id;primaryKey;default:uuid_generate_v4();uniqueIndex;not null"`
	CategoryID  *uuid.UUID     `gorm:"column:category_id;index"`
	Category    *Category      `gorm:"foreignKey:CategoryID;references:ID"`
	Name        string         `gorm:"column:name;type:varchar(255);not null"`
	Slug        string         `gorm:"column:slug;type:varchar(255);uniqueIndex:idx_product_slug_deleted_at,where:deleted_at IS NULL"`
	Description string         `gorm:"column:description;type:text"`
	Price       int64          `gorm:"column:price;not null"`
	Stock       int            `gorm:"column:stock;default:0;not null"`
	Weight      int            `gorm:"column:weight;default:0"`
	Images      []ProductImage `gorm:"foreignKey:ProductID;references:ID"`
	CreatedAt   time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
