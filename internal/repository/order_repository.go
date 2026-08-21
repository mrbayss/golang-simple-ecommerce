package repository

import (
	"time"

	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateWithItems(db *gorm.DB, order *entity.Order) error
	FindByCode(db *gorm.DB, code string) (*entity.Order, error)
	FindAllPaginated(db *gorm.DB, page, limit int, status *entity.OrderStatus) ([]entity.Order, int64, error)
	CountToday(db *gorm.DB, day time.Time) (int64, error)
}

type orderRepository struct {
}

func NewOrderRepository() OrderRepository {
	return &orderRepository{}
}

// CreateWithItems inserts the order and its items in one transaction.
func (or *orderRepository) CreateWithItems(db *gorm.DB, order *entity.Order) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		return nil // items dibuat via association di service (order.Items sudah terisi)
	})
}

func (or *orderRepository) FindByCode(db *gorm.DB, code string) (*entity.Order, error) {
	var order entity.Order

	if err := db.Preload("Items").Where("order_code = ?", code).First(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

func (or *orderRepository) FindAllPaginated(db *gorm.DB, page, limit int, status *entity.OrderStatus) ([]entity.Order, int64, error) {
	var orders []entity.Order
	var total int64

	query := db.Model(new(entity.Order))
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("Items").Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (or *orderRepository) CountToday(db *gorm.DB, day time.Time) (int64, error) {
	var count int64
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.AddDate(0, 0, 1)

	if err := db.Model(new(entity.Order)).Where("created_at >= ? AND created_at < ?", start, end).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
