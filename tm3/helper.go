package main

import "github.com/gofiber/fiber/v2"

func ok(c *fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta, 
	})
}

func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location) 
	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func fail(c *fiber.Ctx, status int, message string, errors any) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}