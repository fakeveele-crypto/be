package helper

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm4/app/model"
)

func ok(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func fail(c *fiber.Ctx, status int, message string, errors any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// metodeBerbody adalah daftar metode HTTP yang wajib mengirim body,
// sehingga perlu dicek Content-Type-nya.
var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// requireJSON menolak request berisi body yang Content-Type-nya bukan JSON.
// Status yang tepat untuk kasus ini adalah 415, bukan 400.
func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return fail(c, fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json", nil)
		}
	}
	return c.Next()
}