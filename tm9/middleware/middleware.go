package middleware

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm7/helper"
)

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		requestID, _ := c.Locals("requestid").(string)

		status := c.Response().StatusCode()
		if err != nil {
			var appErr *helper.AppError
			if errors.As(err, &appErr) {
				status = appErr.Status
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		}

		logger.Info("http_request", attrs...)
		return err
	}
}
