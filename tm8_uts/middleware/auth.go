package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/helper"
)

// RequireAuth memeriksa access token di header Authorization.
// Kalau token sah, identitas user disimpan di Locals supaya bisa dibaca handler.
func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak ada atau format salah", nil)
		}

		authUser, err := jwtManager.Parse(token)
		if err != nil {
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Fail(c, fiber.StatusUnauthorized, "token sudah kedaluwarsa", nil)
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak valid", nil)
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

// bearerToken mengambil token dari header "Authorization: Bearer <token>".
func bearerToken(c *fiber.Ctx) (string, error) {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", errors.New("header kosong")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("format bukan Bearer")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token kosong")
	}
	return token, nil
}