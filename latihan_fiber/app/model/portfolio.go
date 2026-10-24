package model

import "time"

// Portfolio adalah bentuk portofolio mahasiswa yang dikenal aplikasi.
// ID proyek disimpan sebagai string agar package model tidak bergantung pada
// tipe ObjectID milik driver MongoDB.
type Portfolio struct {
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	Link         string   `json:"link"`
	Year         int      `json:"year"`
}

type Certificate struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
}

// ReplacePortfolioRequest dipakai PUT. Seluruh isi portofolio diganti, tetapi
// student_id tidak pernah diambil dari body; nilainya berasal dari parameter
// jalur dan diverifikasi ke PostgreSQL.
type ReplacePortfolioRequest struct {
	Headline     string               `json:"headline" validate:"required,min=3,max=120"`
	Bio          string               `json:"bio" validate:"required,min=10,max=1000"`
	Projects     []ProjectRequest     `json:"projects" validate:"max=20"`
	Certificates []CertificateRequest `json:"certificates" validate:"max=20"`
}

type ProjectRequest struct {
	Title        string   `json:"title" validate:"required,min=3,max=120"`
	Technologies []string `json:"technologies" validate:"required,min=1,max=10,dive,min=1,max=30"`
	Link         string   `json:"link" validate:"omitempty,url,max=300"`
	Year         int      `json:"year" validate:"required,min=1990,max=2100"`
}

type CertificateRequest struct {
	Name   string `json:"name" validate:"required,min=3,max=120"`
	Issuer string `json:"issuer" validate:"required,min=3,max=120"`
	Date   string `json:"date" validate:"required,date"`
}

// AddProjectRequest adalah body POST /students/:id/portfolio/projects.
// ID proyek dibuat server, bukan dikirim client.
type AddProjectRequest struct {
	Title        string   `json:"title" validate:"required,min=3,max=120"`
	Technologies []string `json:"technologies" validate:"required,min=1,max=10,dive,min=1,max=30"`
	Link         string   `json:"link" validate:"omitempty,url,max=300"`
	Year         int      `json:"year" validate:"required,min=1990,max=2100"`
}

type TechnologyCount struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}
