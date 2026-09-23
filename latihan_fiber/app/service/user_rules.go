package service

import (
	"strings"

	"latihan_fiber/app/model"
)

// ApplyPatch menggabungkan — validasi sudah via tag, jadi tidak mengembalikan error.
// Namun untuk kompatibilitas tes lama, versi 2-return tetap disediakan: map selalu nil.
func ApplyPatch(current model.User, req model.PatchUserRequest) (model.User, map[string]string) {
	if req.Username != nil {
		current.Username = strings.TrimSpace(*req.Username)
	}
	if req.Email != nil {
		current.Email = strings.TrimSpace(*req.Email)
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current, nil
}

// applyPatchInternal dipakai service yang sudah tidak butuh map.
func applyPatch(current model.User, req model.PatchUserRequest) model.User {
	u, _ := ApplyPatch(current, req)
	return u
}

func IsEmptyPatch(req model.PatchUserRequest) bool {
	return req.Username == nil && req.Email == nil && req.IsActive == nil
}

func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
