package service

import (
	"latihan_fiber/app/model"
	"latihan_fiber/helper"
)

// CanAccessStudent memutuskan apakah seseorang boleh menyentuh data student.
//
// Fungsi murni: tidak mengimpor fiber maupun repository, hanya menerima
// nilai. Pemilik data (ownerID sama dengan pemanggil) selalu boleh;
// selain itu diperlukan permission :any yang sesuai.
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
