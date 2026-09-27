package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/repository"
	"github.com/mrbayss/golang-simple-ecommerce/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	testutil.CleanDB(t, db)
	defer testutil.CleanDB(t, db)

	repo := repository.NewUserRepository()

	user := &entity.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     entity.MemberRole,
	}

	created, err := repo.Create(db, user)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)

	found, err := repo.FindByEmail(db, "test@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "test@example.com", found.Email)
}

func TestCategoryRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	testutil.CleanDB(t, db)
	defer testutil.CleanDB(t, db)

	repo := repository.NewCategoryRepository()

	cat1 := &entity.Category{Name: "Buku", Slug: "buku"}
	cat2 := &entity.Category{Name: "Elektronik", Slug: "elektronik"}

	_, err := repo.Create(db, cat1)
	require.NoError(t, err)
	_, err = repo.Create(db, cat2)
	require.NoError(t, err)

	// FindBySlug
	found, err := repo.FindBySlug(db, "buku")
	require.NoError(t, err)
	assert.Equal(t, "Buku", found.Name)

	// FindBySlugExcludingID
	foundExcl, err := repo.FindBySlugExcludingID(db, "buku", cat2.ID.String())
	require.NoError(t, err)
	assert.Equal(t, cat1.ID, foundExcl.ID)

	// FindAllPaginated
	cats, total, err := repo.FindAllPaginated(db, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, cats, 2)
}

func TestProductRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	testutil.CleanDB(t, db)
	defer testutil.CleanDB(t, db)

	repo := repository.NewProductRepository()

	prod := &entity.Product{
		Name:        "Kopi Robusta",
		Slug:        "kopi-robusta",
		Description: "Kopi asli Lampung",
		Price:       45000,
		Stock:       20,
	}

	created, err := repo.Create(db, prod)
	require.NoError(t, err)

	// FindByIDForUpdate inside transaction
	tx := db.Begin()
	lockedProd, err := repo.FindByIDForUpdate(tx, created.ID.String())
	require.NoError(t, err)
	assert.Equal(t, 20, lockedProd.Stock)

	// UpdateStock
	err = repo.UpdateStock(tx, created.ID.String(), -5)
	require.NoError(t, err)
	tx.Commit()

	reloaded, err := repo.GetByID(db, created.ID.String())
	require.NoError(t, err)
	assert.Equal(t, 15, reloaded.Stock)
}

func TestOrderRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}
	testutil.CleanDB(t, db)
	defer testutil.CleanDB(t, db)

	orderRepo := repository.NewOrderRepository()

	order := &entity.Order{
		OrderCode:     "WB-20260927-0001",
		CustomerName:  "Andi",
		CustomerPhone: "628123456789",
		PaymentMethod: entity.PaymentCOD,
		Status:        entity.StatusPending,
		TotalPrice:    150000,
		Items: []entity.OrderItem{
			{
				ProductID:    uuid.New(),
				ProductName:  "Kemeja",
				ProductPrice: 75000,
				Quantity:     2,
				Subtotal:     150000,
			},
		},
	}

	err := orderRepo.CreateWithItems(db, order)
	require.NoError(t, err)
	assert.NotEmpty(t, order.ID)

	// FindByCode
	foundByCode, err := orderRepo.FindByCode(db, "WB-20260927-0001")
	require.NoError(t, err)
	require.NotNil(t, foundByCode)
	assert.Len(t, foundByCode.Items, 1)

	// CountToday
	count, err := orderRepo.CountToday(db, time.Now())
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// UpdateStatus
	err = orderRepo.UpdateStatus(db, order.ID.String(), entity.StatusCompleted)
	require.NoError(t, err)

	updated, err := orderRepo.FindByID(db, order.ID.String())
	require.NoError(t, err)
	assert.Equal(t, entity.StatusCompleted, updated.Status)
}
