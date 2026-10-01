package models

import "time"

// ---------- users ----------

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Password  string `json:"-"` // tidak pernah diserialkan ke JSON
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// ---------- students ----------

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id,omitempty"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// ---------- courses ----------

type Course struct {
	ID        int    `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`
	SisaKuota int    `json:"sisa_kuota"`
}

// ---------- enrollments ----------

type Enrollment struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id"`
	CourseID      int       `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

// ---------- request DTO ----------

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type CreateStudentRequest struct {
	NIM         string  `json:"nim" validate:"required,nim12"`
	Nama        string  `json:"nama" validate:"required,min=3,max=100"`
	Email       string  `json:"email" validate:"required,email,max=120"`
	Prodi       string  `json:"prodi" validate:"required,min=2,max=100"`
	Angkatan    int     `json:"angkatan" validate:"required,angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitnil,gte=0,lte=4"`
}

type UpdateStudentRequest struct {
	Nama        *string  `json:"nama,omitempty" validate:"omitnil,min=3,max=100"`
	Prodi       *string  `json:"prodi,omitempty" validate:"omitnil,min=2,max=100"`
	Angkatan    *int     `json:"angkatan,omitempty" validate:"omitnil,angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitnil,gte=0,lte=4"`
}

type CreateEnrollmentRequest struct {
	CourseID      int    `json:"course_id" validate:"required,gt=0"`
	TahunAkademik string `json:"tahun_akademik" validate:"required,tahunakademik"`
}

// ---------- klaim auth di context ----------

type AuthUser struct {
	UserID int
	Email  string
	Role   string
}

// BatasSKS mengimplementasikan business rule 1 soal UTS.
func BatasSKS(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}
