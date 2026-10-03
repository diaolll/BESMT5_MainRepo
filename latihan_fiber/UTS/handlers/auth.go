package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakaduts/middleware"
	"siakaduts/models"
	"siakaduts/utils"
)

type AuthHandler struct {
	DB      *pgxpool.Pool
	JWT     *utils.JWTManager
	Limiter *middleware.LoginFailLimiter
}

// POST /api/v1/auth/login (publik)
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	if !h.Limiter.Check(c) {
		return nil // respons 429 sudah ditulis oleh limiter
	}
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.Limiter.RecordFailed(c)
		return utils.ErrValidation(c, map[string][]string{"_": {"Body harus berupa JSON yang valid"}})
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if errs := utils.ValidateStruct(req); errs != nil {
		h.Limiter.RecordFailed(c)
		return utils.ErrValidation(c, errs)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var u models.User
	var hash string
	var studentDeleted *time.Time
	err := h.DB.QueryRow(ctx, `
		SELECT u.id, u.email, u.password, u.role, s.deleted_at
		FROM users u LEFT JOIN students s ON s.user_id = u.id
		WHERE LOWER(u.email) = $1`, req.Email).
		Scan(&u.ID, &u.Email, &hash, &u.Role, &studentDeleted)
	if err != nil || !utils.VerifyPassword(hash, req.Password) {
		h.Limiter.RecordFailed(c)
		return utils.Err(c, fiber.StatusUnauthorized, "Email atau password salah")
	}
	// Business rule endpoint 7: mahasiswa yang di-soft delete tidak dapat login.
	if u.Role == "mahasiswa" && studentDeleted != nil {
		h.Limiter.RecordFailed(c)
		return utils.Err(c, fiber.StatusUnauthorized, "Akun mahasiswa sudah tidak aktif")
	}

	h.Limiter.Reset(c)
	token, err := h.JWT.Generate(u)
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return utils.OK(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   h.JWT.TTLSeconds(),
		"user": fiber.Map{
			"id":    u.ID,
			"email": u.Email,
			"role":  u.Role,
		},
	})
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	au, ok := middleware.CurrentUser(c)
	if !ok {
		return utils.Err(c, fiber.StatusUnauthorized, "Belum terautentikasi")
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var u models.User
	var hash string
	err := h.DB.QueryRow(ctx, `SELECT id, email, role FROM users WHERE id=$1`, au.UserID).
		Scan(&u.ID, &u.Email, &u.Role)
	if err == pgx.ErrNoRows {
		return utils.Err(c, fiber.StatusUnauthorized, "User tidak ditemukan")
	}
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	_ = hash

	data := fiber.Map{
		"id":    u.ID,
		"email": u.Email,
		"role":  u.Role,
	}
	if u.Role == "mahasiswa" {
		var s models.Student
		err := h.DB.QueryRow(ctx, `
			SELECT id, nim, nama, prodi, angkatan
			FROM students WHERE user_id=$1 AND deleted_at IS NULL`, u.ID).
			Scan(&s.ID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan)
		if err == pgx.ErrNoRows {
			return utils.Err(c, fiber.StatusUnauthorized, "Data mahasiswa tidak ditemukan")
		}
		if err != nil {
			return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
		}
		data["student"] = fiber.Map{
			"nim":      s.NIM,
			"nama":     s.Nama,
			"prodi":    s.Prodi,
			"angkatan": s.Angkatan,
		}
	}
	return utils.OK(c, fiber.StatusOK, "Profil berhasil diambil", data)
}
