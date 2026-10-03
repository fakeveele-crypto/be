package helper

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm7/app/model"
)

// Success membalas response berhasil dalam bentuk WebResponse yang seragam.
func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ParamID membaca parameter ":id" dari URL dan memastikan ia angka positif.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := c.ParamsInt("id")
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// RequestContext memberi batas waktu pada setiap operasi basis data,
// agar query yang menggantung tidak menahan koneksi selamanya.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}