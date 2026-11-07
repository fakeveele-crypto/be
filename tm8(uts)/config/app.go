package config

import "github.com/gofiber/fiber/v2"

func NewFiberConfig() fiber.Config {
	return fiber.Config{
		AppName: "SIAKAD Mini API v1.0",
	}
}