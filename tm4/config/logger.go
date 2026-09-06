package config

import (
	"io"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupLogger(app *fiber.App) {
	_ = os.MkdirAll("./logs", 0755)

	file, err := os.OpenFile("./logs/app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		file = os.Stdout
	}

	app.Use(logger.New(logger.Config{
		Output: io.MultiWriter(os.Stdout, file),
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
}