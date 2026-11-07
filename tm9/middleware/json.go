package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm9/helper"
)

// RequireJSON memastikan request yang membawa body berisikan Content-Type: application/json
func RequireJSON(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut || c.Method() == fiber.MethodPatch {
		if !strings.HasPrefix(c.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
			return helper.UnsupportedMediaType("Content-Type harus application/json")
		}
	}
	return c.Next()
}
