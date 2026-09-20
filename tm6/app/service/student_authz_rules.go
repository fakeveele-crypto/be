package service

import (
	"latihan-fiber/tm6/app/model"
	"latihan-fiber/tm6/helper"
)

// CanAccessStudent memutuskan apakah current boleh mengakses satu baris
// student tertentu.
//
// Pemilik data (ownerID sama dengan identitas pemanggil) SELALU boleh.
// Selain itu, pemanggil butuh permission ":any" yang sesuai (misalnya
// "student:read:any" untuk membaca, "student:update:any" untuk mengubah).
//
// Fungsi ini murni: tidak mengimpor fiber maupun repository, sehingga
// dapat diuji tanpa server maupun database.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == ownerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}