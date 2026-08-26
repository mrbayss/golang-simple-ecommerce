package entity

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string
type OrderStatus string

const (
	PaymentCOD      PaymentMethod = "COD"
	PaymentTransfer PaymentMethod = "TRANSFER"
	PaymentQRIS     PaymentMethod = "QRIS"
)

const (
	StatusPending   OrderStatus = "PENDING"   // awaiting admin confirmation
	StatusConfirmed OrderStatus = "CONFIRMED" // being prepared
	StatusReady     OrderStatus = "READY"     // ready for pickup/delivery
	StatusCompleted OrderStatus = "COMPLETED" // finished and paid (final)
	StatusCancelled OrderStatus = "CANCELLED" // cancelled (final)
)

// Order is append-only history: deletion is not supported,
// cancellation is expressed through Status.
type Order struct {
	ID              uuid.UUID     `gorm:"column:id;primaryKey;default:uuid_generate_v4();uniqueIndex;not null"`
	OrderCode       string        `gorm:"column:order_code;type:varchar(20);uniqueIndex;not null"` // human-readable code, e.g. WB-20260821-0001
	CustomerName    string        `gorm:"column:customer_name;type:varchar(100);not null"`
	CustomerPhone   string        `gorm:"column:customer_phone;type:varchar(15);not null"` // normalized E.164 without '+', e.g. 628123456789
	CustomerAddress *string       `gorm:"column:customer_address;type:text"`               // null means pickup at the store
	Notes           *string       `gorm:"column:notes;type:text"`
	PaymentMethod   PaymentMethod `gorm:"column:payment_method;type:varchar(20);not null"`
	Status          OrderStatus   `gorm:"column:status;type:varchar(20);default:'PENDING';not null"`
	TotalPrice      Money         `gorm:"column:total_price;not null"` // computed server-side
	CreatedAt       time.Time     `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt       *time.Time    `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`

	Items []OrderItem `gorm:"foreignKey:OrderID;references:ID"`
}
