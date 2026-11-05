package helper

import "github.com/gofiber/fiber/v2"

// RequestID mengambil id request yang dibuat oleh middleware requestid.
func RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals("requestid").(string)
	return id
}
