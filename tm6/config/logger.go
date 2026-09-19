package config

import (
	"io"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupLogger(app *fiber.App) {
	app.Use(requestid.New())

	fileLogger := &lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    10, 
		MaxBackups: 3, 
		MaxAge:     28, 
		Compress:   true,
	}

	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	jsonFormat := `{"time":"${time}","request_id":"${locals:requestid}","metode":"${method}","jalur":"${path}","status":${status},"durasi":"${latency}"}` + "\n"

	app.Use(logger.New(logger.Config{
		Output:     multiWriter,
		Format:     jsonFormat,
		TimeFormat: "2006-01-02T15:04:05Z07:00",
	}))
}