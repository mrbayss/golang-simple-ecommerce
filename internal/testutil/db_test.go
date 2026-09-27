package testutil_test

import (
	"testing"

	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTestDB(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if db == nil {
		return
	}

	testutil.CleanDB(t, db)

	// Verify we can insert and query an entity
	cat := &entity.Category{
		Name: "Test Electronics",
		Slug: "test-electronics",
	}

	err := db.Create(cat).Error
	require.NoError(t, err)
	assert.NotEmpty(t, cat.ID)

	var retrieved entity.Category
	err = db.First(&retrieved, "id = ?", cat.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Test Electronics", retrieved.Name)

	testutil.CleanDB(t, db)
}
