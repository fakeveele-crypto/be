package middleware

import (
	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/helper"
)

// RequireRole hanya meloloskan user dengan salah satu role yang diizinkan.
// Harus dipasang SETELAH RequireAuth.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authUser, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak valid", nil)
		}

		for _, role := range roles {
			if authUser.Role == role {
				return c.Next()
			}
		}
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak punya akses ke endpoint ini", nil)
	}
}