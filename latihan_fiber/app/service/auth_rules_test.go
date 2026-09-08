package service

import (
	"testing"

	"latihan_fiber/app/model"
)

func TestValidateRegister_PasswordTerlaluPendek(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username: "sari", Email: "sari@example.com", Password: "abc1",
	})
	if _, ok := errs["password"]; !ok {
		t.Fatal("password <8 karakter harus error")
	}
}

func TestValidateRegister_PasswordTanpaAngka(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username: "sari", Email: "sari@example.com", Password: "rahasiaaa",
	})
	if _, ok := errs["password"]; !ok {
		t.Fatal("password tanpa angka harus error")
	}
}

func TestValidateRegister_PasswordUmumDitolak(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username: "budi", Email: "budi@example.com", Password: "password123",
	})
	if _, ok := errs["password"]; !ok {
		t.Fatal("password umum harus ditolak")
	}
}

func TestValidateRegister_Valid(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username: "sari", Email: "sari@example.com", Password: "rahasia123",
	})
	if len(errs) != 0 {
		t.Fatalf("seharusnya valid, dapat errs: %v", errs)
	}
}

func TestValidateLogin_Kosong(t *testing.T) {
	errs := ValidateLogin(model.LoginRequest{})
	if len(errs) != 2 {
		t.Fatalf("harap 2 error (username, password), dapat %v", errs)
	}
}

func TestValidateRegister_UsernameTidakValid(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username: "sa ri!", Email: "sari@example.com", Password: "rahasia123",
	})
	if _, ok := errs["username"]; !ok {
		t.Fatal("username dengan spasi/simbol harus error")
	}
}
