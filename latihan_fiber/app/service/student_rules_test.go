package service

import (
	"testing"

	"latihan_fiber/app/model"
)

func TestValidateCreateStudent_Kosong(t *testing.T) {
	errs := ValidateCreateStudent(model.CreateStudentRequest{})
	// NIM dan Name kosong menghasilkan error; Grade bernilai zero-value 0.0
	// masih dianggap valid karena berada dalam rentang 0.00–4.00.
	if len(errs) != 2 {
		t.Fatalf("harap 2 error (nim, name), dapat %d: %v", len(errs), errs)
	}
}

func TestValidateCreateStudent_GradeDiLuarRentang(t *testing.T) {
	errs := ValidateCreateStudent(model.CreateStudentRequest{
		NIM: "434241074", Name: "Diaul Haq", Grade: 5.0,
	})
	if _, ok := errs["grade"]; !ok {
		t.Fatal("grade di luar rentang 0-4 seharusnya menghasilkan error")
	}
}

func TestApplyPatchStudent_UbahNama(t *testing.T) {
	current := model.Student{ID: 1, NIM: "434241074", Name: "Diaul", Grade: 3.75, IsActive: true}
	namaBaru := "Diaul Haq"
	result, errs := ApplyPatchStudent(current, model.PatchStudentRequest{Name: &namaBaru})
	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya error: %v", errs)
	}
	if result.Name != "Diaul Haq" {
		t.Error("name seharusnya berubah")
	}
	if result.NIM != "434241074" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestIsEmptyPatchStudent(t *testing.T) {
	if !IsEmptyPatchStudent(model.PatchStudentRequest{}) {
		t.Error("request kosong harusnya dianggap empty patch")
	}
}