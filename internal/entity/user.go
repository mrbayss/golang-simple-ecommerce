package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID      `gorm:"column:id;primaryKey;default:uuid_generate_v4();uniqueIndex;not null"`
	Email     string         `gorm:"column:email;uniqueIndex:email_deleted_at,where:deleted_at IS NULL;not null"`
	Password  string         `gorm:"column:password;not null"`
	Role      userRole       `gorm:"column:role;default:'MEMBER';not null"`
	Member    *Member        `gorm:"foreignKey:UserID;references:ID"`
	CreatedAt time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type userRole string

const (
	AdminRole  userRole = "ADMIN"
	MemberRole userRole = "MEMBER"
)
