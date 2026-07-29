package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductImage struct {
	ID        uuid.UUID      `gorm:"column:id;primaryKey;default:uuid_generate_v4();uniqueIndex;not null"`
	ProductID uuid.UUID      `gorm:"column:product_id;not null;index"`
	ImageURL  string         `gorm:"column:image_url;type:varchar(255);not null"`
	IsPrimary bool           `gorm:"column:is_primary;default:false"`
	CreatedAt time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
