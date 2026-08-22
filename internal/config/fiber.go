package config

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewFiber(config *viper.Viper, log *logrus.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: config.GetString("app.name"),
	})

	app.Use(recover.New())

	// Serve static files from ./public at /public
	app.Use("/public", static.New("./public"))

	return app
}
