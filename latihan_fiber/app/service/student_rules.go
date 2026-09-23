package service

import (
	"strings"

	"latihan_fiber/app/model"
	"latihan_fiber/helper"
)

func ApplyPatchStudent(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current, nil
}

func ValidateCreateStudent(req model.CreateStudentRequest) map[string]string {
	if m := helper.ValidateStruct(req); m != nil {
		return m
	}
	return map[string]string{}
}
func ValidateReplaceStudent(req model.ReplaceStudentRequest) map[string]string {
	if m := helper.ValidateStruct(req); m != nil {
		return m
	}
	return map[string]string{}
}

func IsEmptyPatchStudent(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}
