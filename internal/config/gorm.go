package config

import (
	"fmt"

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

	fmt.Println("Host    :", config.GetString("database.host"))
	fmt.Println("Port    :", config.GetInt("database.port"))
	fmt.Println("DB Name :", config.GetString("database.dbname"))
	fmt.Println("User    :", config.GetString("database.user"))

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

	log.Info("success connect to database")

	return db

}
