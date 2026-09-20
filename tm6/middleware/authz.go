package middleware

import (
	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm6/helper"
)

// RequirePermission menolak request yang role-nya tidak memiliki
// permission tertentu. Dipasang pada route yang haknya dapat diputuskan
// TANPA melihat isi data — misalnya "boleh melihat daftar seluruh student".
//
// Untuk keputusan yang bergantung pada isi data (misalnya "boleh mengubah
// data ini karena miliknya sendiri"), pemeriksaan tidak bisa di sini:
// middleware belum tahu data siapa yang akan disentuh. Pemeriksaan
// semacam itu berada di layer service.

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "belum terautentikasi",
			})
		}
		if !perms.Can(user.Role, permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "role " + user.Role + " tidak memiliki hak " + permission,
			})
		}
		return c.Next()
	}
}