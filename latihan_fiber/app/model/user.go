package model

import (
	"fmt"
	"time"
)

// Student adalah entitas utama.
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   *int      `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s Student) GetInfo() string {
	return fmt.Sprintf("[%d] %s (%s) - Nilai: %.1f - Status: %v",
		s.ID, s.Name, s.NIM, s.Grade, s.IsActive)
}
func (s *Student) UpdateGrade(grade float64) { s.Grade = grade }
func (s *Student) Activate()                 { s.IsActive = true }
func (s *Student) Deactivate()               { s.IsActive = false }

// POST — semua field wajib (deklaratif)
type CreateStudentRequest struct {
	NIM   string  `json:"nim" validate:"required,nim"`
	Name  string  `json:"name" validate:"required,min=3,max=100,nospace"`
	Grade float64 `json:"grade" validate:"gte=0,lte=4"`
}

// PUT — ganti seluruh isi
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,nim"`
	Name     string  `json:"name" validate:"required,min=3,max=100"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=4"`
	IsActive bool    `json:"is_active"`
}

// PATCH — ubah sebagian, pointer + omitnil
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty" validate:"omitnil,nim"`
	Name     *string  `json:"name,omitempty" validate:"omitnil,min=3,max=100,nospace"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,gte=0,lte=4"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// User entitas
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type AssignRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin staff user"`
}

// POST — deklaratif
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email" validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace"`
}

type ReplaceUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email" validate:"required,email,max=120"`
	IsActive bool   `json:"is_active"`
}

// PATCH — pointer + omitnil (FIX: Username sebelumnya string, harus *string agar bisa bandingkan nil)
type PatchUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitnil,min=3,max=30,alphanum"`
	Email    *string `json:"email,omitempty" validate:"omitnil,email,max=120"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// Amplop baku
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Meta offset pagination (legacy)
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ErrorResponse — bentuk baru kegagalan terpusat
type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// Cursor pagination
type Cursor struct {
	CreatedAt time.Time
	ID        int
}

type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	MinGrade *float64
	MaxGrade *float64
	After    *Cursor
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// ListQuery legacy offset
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
	MinGrade *float64
	MaxGrade *float64
}

func (q ListQuery) Offset() int { return (q.Page - 1) * q.Limit }
