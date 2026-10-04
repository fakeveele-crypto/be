package helper

import (
	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/app/model"
)

// Ok untuk response sukses 200. meta diisi nil kalau bukan endpoint daftar.
func Ok(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created untuk response sukses 201.
func Created(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent untuk response 204, tanpa isi.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Fail untuk semua response error. errors diisi nil kalau tidak ada detail.
func Fail(c *fiber.Ctx, status int, message string, errors any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}