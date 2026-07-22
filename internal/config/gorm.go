package config

import (
	"fmt"

	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(config *viper.Viper, log *logrus.Logger) *gorm.DB {
	host := config.GetString("database.host")
	user := config.GetString("database.user")
	pass := config.GetString("database.password")
	port := config.GetInt("database.port")
	dbname := config.GetString("database.dbname")
	sslmode := config.GetString("database.sslmode")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Jakarta",
		host,
		user,
		pass,
		dbname,
		port,
		sslmode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("failed to connect database : %v", err)
	}

	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	if err := db.AutoMigrate(
		&entity.User{},
		&entity.Member{},
		&entity.Address{},
	); err != nil {
		log.Fatalf("Failed to migrate database : %v", err)
	}

	log.Info("success connect to database")

	return db

}
