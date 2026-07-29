package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID        uuid.UUID      `gorm:"column:id;primaryKey;default:uuid_generate_v4();uniqueIndex;not null"`
	Name      string         `gorm:"column:name;type:varchar(100);not null"`
	Slug      string         `gorm:"column:slug;type:varchar(100);uniqueIndex:idx_category_slug_deleted_at,where:deleted_at IS NULL"`
	Products  []Product      `gorm:"foreignKey:CategoryID;references:ID"`
	CreatedAt time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
