package model

import (
	"fmt"
	"time"
)

// Student adalah entitas utama.
// Field dan method di bawah melanjutkan struct Student dari tugas pertemuan 1,
// ditambah NIM sebagai penanda unik dan CreatedAt untuk kebutuhan REST API.
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// GetInfo — dari tugas pertemuan 1, dipakai untuk logging/debug di handler.
func (s Student) GetInfo() string {
	return fmt.Sprintf("[%d] %s (%s) - Nilai: %.1f - Status: %v",
		s.ID, s.Name, s.NIM, s.Grade, s.IsActive)
}

// UpdateGrade — dari tugas pertemuan 1, dipakai di PUT dan PATCH.
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate — dari tugas pertemuan 1, dipakai saat PATCH is_active=true.
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate — dari tugas pertemuan 1, dipakai saat PATCH is_active=false.
func (s *Student) Deactivate() {
	s.IsActive = false
}

// POST — semua field wajib
type CreateStudentRequest struct {
	NIM   string  `json:"nim"`
	Name  string  `json:"name"`
	Grade float64 `json:"grade"`
}

// PUT — ganti seluruh isi, field bertipe biasa dan semuanya wajib
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PATCH — ubah sebagian, field bertipe pointer supaya bisa dibedakan
// antara "tidak dikirim" (nil) dan "dikirim bernilai kosong/nol"
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// Amplop baku untuk semua respons
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListQuery menampung hasil parsing query string endpoint daftar.
type ListQuery struct {
	Page      int
	Limit     int
	Search    string
	Sort      string
	Order     string
	IsActive  *bool
	MinGrade  *float64
	MaxGrade  *float64
}