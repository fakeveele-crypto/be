package helper

import (
	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm5/app/model"
)

// LocalsAuthUser adalah kunci penyimpanan identitas pemakai di context request.
const LocalsAuthUser = "authUser"

func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}