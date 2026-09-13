package config

import "github.com/gofiber/fiber/v2"

func NewFiberConfig() fiber.Config {
	return fiber.Config{
		AppName: "Student Management API v1.0",
	}
}