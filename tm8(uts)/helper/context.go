package helper

import (
	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/app/model"
)

// LocalsAuthUser adalah nama "laci" tempat menyimpan identitas user di request.
const LocalsAuthUser = "authUser"

// CurrentUser mengambil identitas user yang sudah disimpan oleh middleware.
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}