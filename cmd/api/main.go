package main

import (
	"fmt"

	"github.com/mrbayss/golang-simple-ecommerce/internal/config"
	"github.com/mrbayss/golang-simple-ecommerce/internal/utils/token"
)

func main() {
	cfg := config.NewViper()
	log := config.NewLogger(cfg)
	db := config.NewDatabase(cfg, log)
	validate := config.NewValidator(cfg)
	app := config.NewFiber(cfg, log)
	jwt := token.NewJWTToken(cfg)
	redis := config.NewRedis(cfg)

	config.Bootstrap(&config.BootstrapConfig{
		DB:        db,
		App:       app,
		Log:       log,
		Validator: validate,
		Config:    cfg,
		Jwt:       jwt,
		Redis:     redis,
	})

	webPort := cfg.GetInt("app.port")
	host := cfg.GetString("app.host")
	err := app.Listen(fmt.Sprintf("%s:%d", host, webPort))
	if err != nil {
		log.Fatalf("failed to start server : %v", err)
	}

	log.Info("success start server")
}
