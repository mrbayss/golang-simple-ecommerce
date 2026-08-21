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
	StatusPending   OrderStatus = "PENDING"   // baru masuk, belum dikonfirmasi admin
	StatusConfirmed OrderStatus = "CONFIRMED" // admin konfirmasi, sedang disiapkan
	StatusReady     OrderStatus = "READY"     // siap diambil/diantar
	StatusCompleted OrderStatus = "COMPLETED" // selesai + dibayar (final)
	StatusCancelled OrderStatus = "CANCELLED" // dibatalkan (final)
)

type Order struct {
	ID              uuid.UUID     `gorm:"column:id;primaryKey;default:uuid_generate_v4();uniqueIndex;not null"`
	OrderCode       string        `gorm:"column:order_code;type:varchar(20);uniqueIndex;not null"` // WB-20260821-0001
	CustomerName    string        `gorm:"column:customer_name;type:varchar(100);not null"`
	CustomerPhone   string        `gorm:"column:customer_phone;type:varchar(15);not null"` // normalized E.164 tanpa '+': 628123456789
	CustomerAddress *string       `gorm:"column:customer_address;type:text"`               // null = ambil sendiri di warung
	Notes           *string       `gorm:"column:notes;type:text"`                          // contoh: "pedas, tanpa timun"
	PaymentMethod   PaymentMethod `gorm:"column:payment_method;type:varchar(20);not null"` // COD / TRANSFER / QRIS
	Status          OrderStatus   `gorm:"column:status;type:varchar(20);default:'PENDING';not null"`
	TotalPrice      Money         `gorm:"column:total_price;not null"` // dihitung server-side
	CreatedAt       time.Time     `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt       *time.Time    `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`

	// Relasi
	Items []OrderItem `gorm:"foreignKey:OrderID;references:ID"`

	// SENGAJA TANPA DeletedAt — orders adalah append-only history.
	// Pembatalan lewat Status = CANCELLED, bukan delete.
}

// WhatsAppLink returns a wa.me link for contacting the customer.
func (o *Order) WhatsAppLink() string {
	return "https://wa.me/" + o.CustomerPhone
}
