package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/jwt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetEnv returns the environment variable value or fallback if not set.
func GetEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// SetupTestDB initializes connection to the test PostgreSQL database and migrates schemas.
// If the database is not accessible, it skips the test rather than failing the build.
func SetupTestDB(t *testing.T) *gorm.DB {
	host := GetEnv("TEST_DB_HOST", "localhost")
	port := GetEnv("TEST_DB_PORT", "5432")
	user := GetEnv("TEST_DB_USER", "postgres")
	pass := GetEnv("TEST_DB_PASSWORD", "12345")
	dbname := GetEnv("TEST_DB_NAME", "simple_ecommerce_test")
	sslmode := GetEnv("TEST_DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		host, user, pass, dbname, port, sslmode,
	)

	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger: logger.New(
			log,
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		t.Skipf("Skipping test: unable to connect to PostgreSQL test database (%s:%s): %v", host, port, err)
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Skipf("Skipping test: unable to get sql.DB: %v", err)
		return nil
	}

	// Optimize connection pool for concurrency testing
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		t.Skipf("Skipping test: test database ping failed: %v", err)
		return nil
	}

	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	err = db.AutoMigrate(
		&entity.User{},
		&entity.Member{},
		&entity.Address{},
		&entity.Category{},
		&entity.Product{},
		&entity.ProductImage{},
		&entity.Order{},
		&entity.OrderItem{},
	)
	if err != nil {
		t.Fatalf("Failed to auto migrate test database: %v", err)
	}

	return db
}

// CleanDB cleans up all test tables in foreign-key safe order.
func CleanDB(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}
	tables := []string{
		"order_items",
		"orders",
		"product_images",
		"products",
		"categories",
		"addresses",
		"members",
		"users",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
			t.Logf("warning: failed to truncate %s: %v", table, err)
		}
	}
}

// NewTestLogger creates a logrus logger suitable for testing (silent or error only).
func NewTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	return log
}

// NewTestValidator creates a validator instance for testing.
func NewTestValidator() *validator.Validate {
	return validator.New()
}

// NewTestJWT creates a JWT Key instance for testing.
func NewTestJWT() *jwt.Key {
	v := viper.New()
	v.Set("jwt.secret_key", "test-secret-key-1234567890-secure")
	return jwt.NewJWTToken(v)
}
