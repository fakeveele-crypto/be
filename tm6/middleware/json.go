package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// RequireJSON memastikan request header berisikan Content-Type: application/json
func RequireJSON(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut || c.Method() == fiber.MethodPatch {
		if c.Get("Content-Type") != "application/json" {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
				"error": "Content-Type harus application/json",
			})
		}
	}
	return c.Next()
}