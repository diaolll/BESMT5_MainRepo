package utils

import (
	"regexp"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

var Validate = newValidator()

var reTahunAkademik = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

func newValidator() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("nim12", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if len(s) != 12 {
			return false
		}
		for _, r := range s {
			if r < '0' || r > '9' {
				return false
			}
		}
		return true
	})
	_ = v.RegisterValidation("angkatan", func(fl validator.FieldLevel) bool {
		var thn int
		switch fl.Field().Kind() {
		case 2, 3, 4, 5, 6: // int kinds
			thn = int(fl.Field().Int())
		default:
			return false
		}
		if thn < 1000 || thn > 9999 {
			return false
		}
		return thn <= time.Now().Year()
	})
	_ = v.RegisterValidation("tahunakademik", func(fl validator.FieldLevel) bool {
		return reTahunAkademik.MatchString(fl.Field().String())
	})
	return v
}

// ValidateStruct mengembalikan map field -> pesan (nil bila valid).
func ValidateStruct(s any) map[string][]string {
	if err := Validate.Struct(s); err == nil {
		return nil
	} else if ves, ok := err.(validator.ValidationErrors); ok {
		out := make(map[string][]string, len(ves))
		for _, fe := range ves {
			field := jsonName(fe)
			out[field] = append(out[field], messageFor(fe))
		}
		return out
	}
	return map[string][]string{"_": {"validasi gagal"}}
}

func jsonName(fe validator.FieldError) string {
	// validator sudah dikonfigurasi TagNameFunc? kita pakai lowercase field.
	// Sederhananya: pakai nama JSON dari struct via lookup manual.
	// Karena DTO kecil, petakan dari StructField.
	name := fe.StructField()
	switch name {
	case "NIM":
		return "nim"
	case "Nama":
		return "nama"
	case "Email":
		return "email"
	case "Prodi":
		return "prodi"
	case "Angkatan":
		return "angkatan"
	case "IPKTerakhir":
		return "ipk_terakhir"
	case "CourseID":
		return "course_id"
	case "TahunAkademik":
		return "tahun_akademik"
	case "Password":
		return "password"
	}
	return fe.Field()
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		return "minimal " + fe.Param() + " karakter"
	case "max":
		return "maksimal " + fe.Param() + " karakter"
	case "gte":
		return "nilai minimal " + fe.Param()
	case "lte":
		return "nilai maksimal " + fe.Param()
	case "gt":
		return "harus lebih dari " + fe.Param()
	case "nim12":
		return "NIM wajib 12 digit angka"
	case "angkatan":
		return "angkatan wajib 4 digit dan ≤ tahun berjalan (" + strconv.Itoa(time.Now().Year()) + ")"
	case "tahunakademik":
		return "format wajib YYYY/YYYY-Ganjil atau YYYY/YYYY-Genap (contoh 2026/2027-Ganjil)"
	default:
		return "tidak valid (" + fe.Tag() + ")"
	}
}
