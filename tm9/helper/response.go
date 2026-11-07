package helper

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm9/app/model"
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

func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func SuccessCursor(c *fiber.Ctx, message string, data any, meta *model.CursorMeta) error {
	return c.Status(fiber.StatusOK).JSON(struct {
		Success bool              `json:"success"`
		Message string            `json:"message"`
		Data    any               `json:"data"`
		Meta    *model.CursorMeta `json:"meta"`
	}{true, message, data, meta})
}

// Created membalas 201 dan menyertakan header Location ke resource baru.
func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set(fiber.HeaderLocation, location)
	return Success(c, fiber.StatusCreated, message, data)
}

// NoContent membalas 204 tanpa body.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
