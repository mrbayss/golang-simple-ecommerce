package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Address struct {
	ID             uuid.UUID      `gorm:"column:id;primaryKey;default:uuid_generate_v4();not null"`
	MemberID       uuid.UUID      `gorm:"column:member_id;not null;index"`
	Title          string         `gorm:"column:title;not null"`
	RecipientName  string         `gorm:"column:recipient_name;not null"`
	RecipientPhone string         `gorm:"column:recipient_phone;not null"`
	FullAddress    string         `gorm:"type:text;column:full_address;not null"`
	City           string         `gorm:"column:city;not null"`
	Province       string         `gorm:"column:province;not null"`
	PostalCode     string         `gorm:"column:postal_code;not null"`
	IsDefault      bool           `gorm:"column:is_default;default:false;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
