package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakaduts/middleware"
	"siakaduts/models"
	"siakaduts/utils"
)

type EnrollmentHandler struct {
	DB *pgxpool.Pool
}

// POST /api/v1/enrollments (mahasiswa) — ambil MK dalam 1 transaksi:
// cek duplikasi -> kunci baris MK (FOR UPDATE) -> cek kuota -> cek batas SKS.
func (h *EnrollmentHandler) Create(c *fiber.Ctx) error {
	au, _ := middleware.CurrentUser(c)
	var req models.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrValidation(c, map[string][]string{"_": {"Body harus berupa JSON yang valid"}})
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return utils.ErrValidation(c, errs)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 8*time.Second)
	defer cancel()

	// Identitas mahasiswa dari token (rule 4: hanya KRS sendiri).
	var studentID int
	var ipk float64
	err := h.DB.QueryRow(ctx, `SELECT id, ipk_terakhir FROM students WHERE user_id=$1 AND deleted_at IS NULL`, au.UserID).
		Scan(&studentID, &ipk)
	if err == pgx.ErrNoRows {
		return utils.Err(c, fiber.StatusForbidden, "Akun Anda tidak terhubung ke data mahasiswa aktif")
	}
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	defer tx.Rollback(ctx)

	// 1. Mata kuliah harus ada + kunci baris agar kuota tidak race.
	var sks, kuota int
	var kode, nama string
	err = tx.QueryRow(ctx, `SELECT sks, kuota, kode_mk, nama_mk FROM courses WHERE id=$1 FOR UPDATE`, req.CourseID).
		Scan(&sks, &kuota, &kode, &nama)
	if err == pgx.ErrNoRows {
		return utils.ErrValidation(c, map[string][]string{"course_id": {"Mata kuliah tidak ditemukan"}})
	}
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	// 2. Rule 2: tidak boleh ambil MK yang sama 2x di tahun akademik yang sama.
	var dupe bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM enrollments WHERE student_id=$1 AND course_id=$2 AND tahun_akademik=$3)`,
		studentID, req.CourseID, req.TahunAkademik).Scan(&dupe); err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if dupe {
		return utils.Err(c, fiber.StatusConflict, "Mata kuliah "+kode+" sudah diambil pada tahun akademik "+req.TahunAkademik)
	}

	// 3. Rule 3: kuota penuh tidak dapat diambil.
	var terisi int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM enrollments WHERE course_id=$1`, req.CourseID).Scan(&terisi); err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if terisi >= kuota {
		return utils.ErrValidation(c, map[string][]string{"course_id": {"Kuota mata kuliah " + kode + " sudah penuh"}})
	}

	// 4. Rule 1: batas SKS sesuai IPK terakhir.
	batas := models.BatasSKS(ipk)
	var dipakai int
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(c.sks),0) FROM enrollments e
		JOIN courses c ON c.id=e.course_id
		WHERE e.student_id=$1 AND e.tahun_akademik=$2`, studentID, req.TahunAkademik).Scan(&dipakai); err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if dipakai+sks > batas {
		sisa := batas - dipakai
		if sisa < 0 {
			sisa = 0
		}
		return utils.ErrValidation(c, map[string][]string{
			"course_id": {"Total SKS (" + strconv.Itoa(dipakai+sks) + ") melebihi batas " + strconv.Itoa(batas) +
				" SKS untuk IPK " + strconv.FormatFloat(ipk, 'f', 2, 64) +
				". Sisa SKS Anda: " + strconv.Itoa(sisa)},
		})
	}

	var e models.Enrollment
	err = tx.QueryRow(ctx, `
		INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		VALUES ($1,$2,$3) RETURNING id, student_id, course_id, tahun_akademik, created_at`,
		studentID, req.CourseID, req.TahunAkademik).
		Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return utils.Err(c, fiber.StatusConflict, "Mata kuliah sudah diambil pada tahun akademik tersebut")
		}
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if err := tx.Commit(ctx); err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return utils.Created(c, "Mata kuliah "+nama+" berhasil diambil", fiber.Map{
		"id": e.ID, "student_id": e.StudentID, "course_id": e.CourseID,
		"tahun_akademik": e.TahunAkademik, "created_at": e.CreatedAt,
	})
}

// DELETE /api/v1/enrollments/{id} — hanya milik sendiri.
func (h *EnrollmentHandler) Delete(c *fiber.Ctx) error {
	au, _ := middleware.CurrentUser(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return utils.Err(c, fiber.StatusNotFound, "Data KRS tidak ditemukan")
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var studentID int
	if err := h.DB.QueryRow(ctx, `SELECT id FROM students WHERE user_id=$1 AND deleted_at IS NULL`, au.UserID).Scan(&studentID); err != nil {
		return utils.Err(c, fiber.StatusForbidden, "Akun Anda tidak terhubung ke data mahasiswa aktif")
	}
	var owner int
	err = h.DB.QueryRow(ctx, `SELECT student_id FROM enrollments WHERE id=$1`, id).Scan(&owner)
	if err == pgx.ErrNoRows {
		return utils.Err(c, fiber.StatusNotFound, "Data KRS tidak ditemukan")
	}
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if owner != studentID {
		return utils.Err(c, fiber.StatusForbidden, "Anda hanya dapat membatalkan KRS milik sendiri")
	}
	if _, err := h.DB.Exec(ctx, `DELETE FROM enrollments WHERE id=$1`, id); err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return utils.NoContent(c)
}
