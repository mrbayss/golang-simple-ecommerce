package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Member struct {
	ID        uuid.UUID      `gorm:"column:id;primaryKey;default:uuid_generate_v4();not null"`
	UserID    uuid.UUID      `gorm:"column:user_id;not null;uniqueIndex"`
	FullName  string         `gorm:"column:full_name;not null"`
	Phone     string         `gorm:"column:phone;not null"`
	Address   []Address      `gorm:"foreignKey:MemberID;references:ID"`
	CreatedAt time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
