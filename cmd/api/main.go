package main

import (
	"fmt"

	"github.com/mrbayss/golang-simple-ecommerce/internal/config"
	"github.com/mrbayss/golang-simple-ecommerce/internal/pkg/jwt"
)

func main() {
	cfg := config.NewViper()
	log := config.NewLogger(cfg)
	db := config.NewDatabase(cfg, log)
	validate := config.NewValidator(cfg)
	app := config.NewFiber(cfg, log)
	jwtKey := jwt.NewJWTToken(cfg)
	if jwtKey.SecretKey == "" {
		log.Fatal("jwt.secret_key is required in config")
	}
	redisClient := config.NewRedis(cfg)

	config.Bootstrap(&config.BootstrapConfig{
		DB:        db,
		App:       app,
		Log:       log,
		Validator: validate,
		Config:    cfg,
		Jwt:       jwtKey,
		Redis:     redisClient,
	})

	webPort := cfg.GetInt("app.port")
	host := cfg.GetString("app.host")
	log.Infof("starting server on %s:%d", host, webPort)
	err := app.Listen(fmt.Sprintf("%s:%d", host, webPort))
	if err != nil {
		log.Fatalf("failed to start server : %v", err)
	}
}
