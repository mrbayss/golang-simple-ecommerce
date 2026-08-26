package repository

import (
	"time"

	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateWithItems(db *gorm.DB, order *entity.Order) error
	FindByCode(db *gorm.DB, code string) (*entity.Order, error)
	FindByID(db *gorm.DB, id string) (*entity.Order, error)
	FindAllPaginated(db *gorm.DB, page, limit int, status *entity.OrderStatus) ([]entity.Order, int64, error)
	CountToday(db *gorm.DB, day time.Time) (int64, error)
	UpdateStatus(db *gorm.DB, id string, status entity.OrderStatus) error
}

type orderRepository struct {
}

func NewOrderRepository() OrderRepository {
	return &orderRepository{}
}

// CreateWithItems inserts the order together with its items.
// GORM creates associated rows automatically when order.Items is set.
func (or *orderRepository) CreateWithItems(db *gorm.DB, order *entity.Order) error {
	return db.Create(order).Error
}

func (or *orderRepository) FindByCode(db *gorm.DB, code string) (*entity.Order, error) {
	var order entity.Order

	if err := db.Preload("Items").Where("order_code = ?", code).First(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

func (or *orderRepository) FindByID(db *gorm.DB, id string) (*entity.Order, error) {
	var order entity.Order

	if err := db.Preload("Items").Where("id = ?", id).First(&order).Error; err != nil {
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

func (or *orderRepository) UpdateStatus(db *gorm.DB, id string, status entity.OrderStatus) error {
	return db.Model(new(entity.Order)).Where("id = ?", id).Update("status", status).Error
}
