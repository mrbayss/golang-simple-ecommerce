package entity

import (
	"time"

	"github.com/google/uuid"
)

// OrderItem is append-only history, following the Order lifecycle.
type OrderItem struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey;default:uuid_generate_v4();uniqueIndex;not null"`
	OrderID      uuid.UUID `gorm:"column:order_id;type:uuid;index;not null"`
	ProductID    uuid.UUID `gorm:"column:product_id;type:uuid;index;not null"`
	ProductName  string    `gorm:"column:product_name;type:varchar(255);not null"` // snapshot at order time
	ProductPrice Money     `gorm:"column:product_price;not null"`                  // snapshot at order time
	Quantity     int       `gorm:"column:quantity;not null;check:quantity > 0"`
	Subtotal     Money     `gorm:"column:subtotal;not null"` // ProductPrice * Quantity, computed server-side
	CreatedAt    time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}
