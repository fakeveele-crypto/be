package service

import (
	"latihan-fiber/tm9/app/model"
	"latihan-fiber/tm9/helper"
)

// CanAccessUser memutuskan apakah current boleh mengakses satu baris
// user tertentu. Pemilik akun SELALU boleh; selain itu, pemanggil
// butuh permission ":any" yang sesuai.
func CanAccessUser(
	current model.AuthUser,
	targetID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == targetID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}
