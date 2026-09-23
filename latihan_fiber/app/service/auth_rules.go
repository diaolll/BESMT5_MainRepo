package service

import (
	"latihan_fiber/app/model"
	"latihan_fiber/helper"
)

// Wrapper untuk kompatibilitas tes lama — delegasi ke validasi deklaratif.
func ValidateRegister(req model.RegisterRequest) map[string]string {
	if m := helper.ValidateStruct(req); m != nil {
		return m
	}
	return map[string]string{}
}

func ValidateLogin(req model.LoginRequest) map[string]string {
	if m := helper.ValidateStruct(req); m != nil {
		return m
	}
	return map[string]string{}
}
