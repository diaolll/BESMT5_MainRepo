package service

import (
	"strings"

	"latihan_fiber/app/model"
)

// File ini business rules MURNI untuk entitas Student.
// Grade bertipe float64 (kolom NUMERIC(4,2) di database), sehingga
// divalidasi lewat perbandingan angka, BUKAN strings.TrimSpace.

func ValidateCreateStudent(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 4 {
		errs["grade"] = "harus di antara 0.00 dan 4.00"
	}
	return errs
}

func ValidateReplaceStudent(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 4 {
		errs["grade"] = "harus di antara 0.00 dan 4.00"
	}
	return errs
}

func ApplyPatchStudent(
	current model.Student, req model.PatchStudentRequest,
) (model.Student, map[string]string) {
	errs := map[string]string{}
	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			errs["nim"] = "tidak boleh kosong"
		} else {
			current.NIM = *req.NIM
		}
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = *req.Name
		}
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 4 {
			errs["grade"] = "harus di antara 0.00 dan 4.00"
		} else {
			current.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current, errs
}

func IsEmptyPatchStudent(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}