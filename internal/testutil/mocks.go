package testutil

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(db *gorm.DB, u *entity.User) (*entity.User, error) {
	args := m.Called(db, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(db *gorm.DB, id string) (*entity.User, error) {
	args := m.Called(db, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(db *gorm.DB, u *entity.User, id string) error {
	args := m.Called(db, u, id)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(db *gorm.DB, id string) error {
	args := m.Called(db, id)
	return args.Error(0)
}

func (m *MockUserRepository) SoftDelete(db *gorm.DB, id string) error {
	args := m.Called(db, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetAll(db *gorm.DB) ([]entity.User, error) {
	args := m.Called(db)
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByColumn(db *gorm.DB, col string, val any) (*entity.User, error) {
	args := m.Called(db, col, val)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx *gorm.DB, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// MockCategoryRepository is a mock implementation of repository.CategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(db *gorm.DB, c *entity.Category) (*entity.Category, error) {
	args := m.Called(db, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByID(db *gorm.DB, id string) (*entity.Category, error) {
	args := m.Called(db, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(db *gorm.DB, c *entity.Category, id string) error {
	args := m.Called(db, c, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(db *gorm.DB, id string) error {
	args := m.Called(db, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) SoftDelete(db *gorm.DB, id string) error {
	args := m.Called(db, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetAll(db *gorm.DB) ([]entity.Category, error) {
	args := m.Called(db)
	return args.Get(0).([]entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByColumn(db *gorm.DB, col string, val any) (*entity.Category, error) {
	args := m.Called(db, col, val)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) FindBySlug(db *gorm.DB, slug string) (*entity.Category, error) {
	args := m.Called(db, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) FindBySlugExcludingID(db *gorm.DB, slug, id string) (*entity.Category, error) {
	args := m.Called(db, slug, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) FindAllPaginated(db *gorm.DB, page, limit int) ([]entity.Category, int64, error) {
	args := m.Called(db, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.Category), args.Get(1).(int64), args.Error(2)
}

// MockProductRepository is a mock implementation of repository.ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(db *gorm.DB, p *entity.Product) (*entity.Product, error) {
	args := m.Called(db, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) GetByID(db *gorm.DB, id string) (*entity.Product, error) {
	args := m.Called(db, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) Update(db *gorm.DB, p *entity.Product, id string) error {
	args := m.Called(db, p, id)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(db *gorm.DB, id string) error {
	args := m.Called(db, id)
	return args.Error(0)
}

func (m *MockProductRepository) SoftDelete(db *gorm.DB, id string) error {
	args := m.Called(db, id)
	return args.Error(0)
}

func (m *MockProductRepository) GetAll(db *gorm.DB) ([]entity.Product, error) {
	args := m.Called(db)
	return args.Get(0).([]entity.Product), args.Error(1)
}

func (m *MockProductRepository) GetByColumn(db *gorm.DB, col string, val any) (*entity.Product, error) {
	args := m.Called(db, col, val)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) FindBySlug(db *gorm.DB, slug string) (*entity.Product, error) {
	args := m.Called(db, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) FindBySlugExcludingID(db *gorm.DB, slug, id string) (*entity.Product, error) {
	args := m.Called(db, slug, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) FindAllPaginated(db *gorm.DB, page, limit int) ([]entity.Product, int64, error) {
	args := m.Called(db, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) FindByIDForUpdate(db *gorm.DB, id string) (*entity.Product, error) {
	args := m.Called(db, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateStock(db *gorm.DB, id string, delta int) error {
	args := m.Called(db, id, delta)
	return args.Error(0)
}

func (m *MockProductRepository) FindImageByID(db *gorm.DB, id string) (*entity.ProductImage, error) {
	args := m.Called(db, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductImage), args.Error(1)
}

func (m *MockProductRepository) GetByIDWithRelations(db *gorm.DB, id string) (*entity.Product, error) {
	args := m.Called(db, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

// MockOrderRepository is a mock implementation of repository.OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) CreateWithItems(db *gorm.DB, order *entity.Order) error {
	args := m.Called(db, order)
	return args.Error(0)
}

func (m *MockOrderRepository) FindByCode(db *gorm.DB, code string) (*entity.Order, error) {
	args := m.Called(db, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByID(db *gorm.DB, id string) (*entity.Order, error) {
	args := m.Called(db, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) FindAllPaginated(db *gorm.DB, page, limit int, status *entity.OrderStatus) ([]entity.Order, int64, error) {
	args := m.Called(db, page, limit, status)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) CountToday(db *gorm.DB, day time.Time) (int64, error) {
	args := m.Called(db, day)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(db *gorm.DB, id string, status entity.OrderStatus) error {
	args := m.Called(db, id, status)
	return args.Error(0)
}

// MockFileStorage is a mock implementation of storage.FileStorage
type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) Save(ctx context.Context, folder string, file *multipart.FileHeader) (string, error) {
	args := m.Called(ctx, folder, file)
	return args.String(0), args.Error(1)
}

func (m *MockFileStorage) Delete(ctx context.Context, filePath string) error {
	args := m.Called(ctx, filePath)
	return args.Error(0)
}
