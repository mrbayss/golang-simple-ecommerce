package entity

import (
	"time"

	"github.com/google/uuid"
)

type paymentMethod string
type orderStatus string

const (
	PaymentCOD      paymentMethod = "COD"
	PaymentTransfer paymentMethod = "TRANSFER"
	PaymentQRIS     paymentMethod = "QRIS"
)

const (
	StatusPending   orderStatus = "PENDING"   // baru masuk, belum dikonfirmasi admin
	StatusConfirmed orderStatus = "CONFIRMED" // admin konfirmasi, sedang disiapkan
	StatusReady     orderStatus = "READY"     // siap diambil/diantar
	StatusCompleted orderStatus = "COMPLETED" // selesai + dibayar (final)
	StatusCancelled orderStatus = "CANCELLED" // dibatalkan (final)
)

type Order struct {
	ID              uuid.UUID     `gorm:"column:id;primaryKey;default:uuid_generate_v4();uniqueIndex;not null"`
	OrderCode       string        `gorm:"column:order_code;type:varchar(20);uniqueIndex;not null"` // WB-20260821-0001
	CustomerName    string        `gorm:"column:customer_name;type:varchar(100);not null"`
	CustomerPhone   string        `gorm:"column:customer_phone;type:varchar(15);not null"` // normalized E.164 tanpa '+': 628123456789
	CustomerAddress *string       `gorm:"column:customer_address;type:text"`               // null = ambil sendiri di warung
	Notes           *string       `gorm:"column:notes;type:text"`                          // contoh: "pedas, tanpa timun"
	PaymentMethod   paymentMethod `gorm:"column:payment_method;type:varchar(20);not null"` // COD / TRANSFER / QRIS
	Status          orderStatus   `gorm:"column:status;type:varchar(20);default:'PENDING';not null"`
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
