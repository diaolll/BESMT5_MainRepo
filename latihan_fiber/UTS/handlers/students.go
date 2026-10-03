package handlers

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakaduts/middleware"
	"siakaduts/models"
	"siakaduts/utils"
)

type StudentHandler struct {
	DB *pgxpool.Pool
}

// myStudentID mengembalikan id students milik user login (khusus mahasiswa).
func (h *StudentHandler) myStudentID(ctx context.Context, userID int) (int, error) {
	var id int
	err := h.DB.QueryRow(ctx, `SELECT id FROM students WHERE user_id=$1 AND deleted_at IS NULL`, userID).Scan(&id)
	return id, err
}

// GET /api/v1/students (admin)
func (h *StudentHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}
	prodi := strings.TrimSpace(c.Query("prodi"))
	angkatanStr := strings.TrimSpace(c.Query("angkatan"))
	search := strings.TrimSpace(c.Query("search"))
	sort := strings.TrimSpace(c.Query("sort", ""))

	orderBy := "s.id ASC"
	switch sort {
	case "nama":
		orderBy = "s.nama ASC"
	case "-ipk_terakhir":
		orderBy = "s.ipk_terakhir DESC"
	}

	where := "WHERE s.deleted_at IS NULL"
	args := []any{}
	n := 0
	push := func(v any) string {
		n++
		args = append(args, v)
		return "$" + strconv.Itoa(n)
	}
	if prodi != "" {
		where += " AND s.prodi ILIKE " + push(prodi)
	}
	if angkatanStr != "" {
		if a, err := strconv.Atoi(angkatanStr); err == nil {
			where += " AND s.angkatan = " + push(a)
		}
	}
	if search != "" {
		where += " AND (s.nim ILIKE " + push("%"+search+"%") + " OR s.nama ILIKE " + push("%"+search+"%") + ")"
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var total int
	if err := h.DB.QueryRow(ctx, `SELECT COUNT(*) FROM students s `+where, args...).Scan(&total); err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	lastPage := (total + perPage - 1) / perPage
	if lastPage == 0 {
		lastPage = 1
	}
	limitPh, offsetPh := push(perPage), push((page-1)*perPage)
	rows, err := h.DB.Query(ctx, `
		SELECT s.id, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir
		FROM students s `+where+` ORDER BY `+orderBy+` LIMIT `+limitPh+` OFFSET `+offsetPh, args...)
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	defer rows.Close()

	list := []fiber.Map{}
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir); err != nil {
			return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
		}
		list = append(list, fiber.Map{
			"id": s.ID, "nim": s.NIM, "nama": s.Nama,
			"prodi": s.Prodi, "angkatan": s.Angkatan, "ipk_terakhir": s.IPKTerakhir,
		})
	}
	return utils.OKList(c, "Data mahasiswa berhasil diambil", list, fiber.Map{
		"current_page": page, "per_page": perPage, "total": total, "last_page": lastPage,
	})
}

// POST /api/v1/students (admin) — buat users + students dalam 1 transaksi.
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req models.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrValidation(c, map[string][]string{"_": {"Body harus berupa JSON yang valid"}})
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Nama = strings.TrimSpace(req.Nama)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Prodi = strings.TrimSpace(req.Prodi)
	if errs := utils.ValidateStruct(req); errs != nil {
		return utils.ErrValidation(c, errs)
	}
	ipk := 0.0
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 8*time.Second)
	defer cancel()

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	defer tx.Rollback(ctx)

	hashed, err := utils.HashPassword(req.NIM) // password awal = NIM
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	var userID int
	err = tx.QueryRow(ctx, `INSERT INTO users (email, password, role) VALUES ($1,$2,'mahasiswa') RETURNING id`,
		req.Email, hashed).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return utils.ErrValidation(c, map[string][]string{"email": {"Email sudah terdaftar"}})
		}
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	var s models.Student
	err = tx.QueryRow(ctx, `
		INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, nim, nama, prodi, angkatan, ipk_terakhir`,
		userID, req.NIM, req.Nama, req.Prodi, req.Angkatan, ipk).
		Scan(&s.ID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err != nil {
		if isUniqueViolation(err) {
			return utils.ErrValidation(c, map[string][]string{"nim": {"NIM sudah terdaftar"}})
		}
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if err := tx.Commit(ctx); err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return utils.Created(c, "Mahasiswa berhasil ditambahkan", fiber.Map{
		"id": s.ID, "nim": s.NIM, "nama": s.Nama,
		"prodi": s.Prodi, "angkatan": s.Angkatan, "ipk_terakhir": s.IPKTerakhir,
	})
}

// GET /api/v1/students/{id} — admin bebas, mahasiswa hanya milik sendiri.
func (h *StudentHandler) Get(c *fiber.Ctx) error {
	au, _ := middleware.CurrentUser(c)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return utils.Err(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var s models.Student
	err = h.DB.QueryRow(ctx, `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir
		FROM students WHERE id=$1 AND deleted_at IS NULL`, id).
		Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err == pgx.ErrNoRows {
		return utils.Err(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if au.Role == "mahasiswa" && s.UserID != au.UserID {
		return utils.Err(c, fiber.StatusForbidden, "Anda hanya dapat mengakses data milik sendiri")
	}

	rows, err := h.DB.Query(ctx, `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, e.tahun_akademik
		FROM enrollments e JOIN courses c ON c.id = e.course_id
		WHERE e.student_id=$1 ORDER BY e.id`, id)
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	defer rows.Close()
	courses := []fiber.Map{}
	totalSKS := 0
	for rows.Next() {
		var cid, sks int
		var kode, nama, ta string
		if err := rows.Scan(&cid, &kode, &nama, &sks, &ta); err != nil {
			return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
		}
		totalSKS += sks
		courses = append(courses, fiber.Map{
			"id": cid, "kode_mk": kode, "nama_mk": nama, "sks": sks, "tahun_akademik": ta,
		})
	}
	return utils.OK(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", fiber.Map{
		"id": s.ID, "nim": s.NIM, "nama": s.Nama, "prodi": s.Prodi,
		"angkatan": s.Angkatan, "ipk_terakhir": s.IPKTerakhir,
		"mata_kuliah": courses,
		"total_sks":   totalSKS,
		"batas_sks":   models.BatasSKS(s.IPKTerakhir),
	})
}

// PUT /api/v1/students/{id} (admin) — NIM tidak boleh diubah.
func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return utils.Err(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}
	// Tolak eksplisit bila body mencoba mengubah NIM.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &raw); err != nil {
		return utils.ErrValidation(c, map[string][]string{"_": {"Body harus berupa JSON yang valid"}})
	}
	if _, ok := raw["nim"]; ok {
		return utils.ErrValidation(c, map[string][]string{"nim": {"NIM tidak dapat diubah"}})
	}
	var req models.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrValidation(c, map[string][]string{"_": {"Body harus berupa JSON yang valid"}})
	}
	if req.Nama != nil {
		*req.Nama = strings.TrimSpace(*req.Nama)
	}
	if req.Prodi != nil {
		*req.Prodi = strings.TrimSpace(*req.Prodi)
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return utils.ErrValidation(c, errs)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	var exists bool
	if err := h.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM students WHERE id=$1 AND deleted_at IS NULL)`, id).Scan(&exists); err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if !exists {
		return utils.Err(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}
	// Bangun UPDATE dinamis dari field yang dikirim.
	sets, args := []string{}, []any{}
	if req.Nama != nil {
		args = append(args, *req.Nama)
		sets = append(sets, "nama=$"+strconv.Itoa(len(args)))
	}
	if req.Prodi != nil {
		args = append(args, *req.Prodi)
		sets = append(sets, "prodi=$"+strconv.Itoa(len(args)))
	}
	if req.Angkatan != nil {
		args = append(args, *req.Angkatan)
		sets = append(sets, "angkatan=$"+strconv.Itoa(len(args)))
	}
	if req.IPKTerakhir != nil {
		args = append(args, *req.IPKTerakhir)
		sets = append(sets, "ipk_terakhir=$"+strconv.Itoa(len(args)))
	}
	if len(sets) == 0 {
		return utils.ErrValidation(c, map[string][]string{"_": {"Tidak ada field yang diubah"}})
	}
	args = append(args, id)
	var s models.Student
	err = h.DB.QueryRow(ctx, `UPDATE students SET `+strings.Join(sets, ", ")+
		` WHERE id=$`+strconv.Itoa(len(args))+` AND deleted_at IS NULL
		 RETURNING id, nim, nama, prodi, angkatan, ipk_terakhir`,
		args...).Scan(&s.ID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return utils.OK(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", fiber.Map{
		"id": s.ID, "nim": s.NIM, "nama": s.Nama,
		"prodi": s.Prodi, "angkatan": s.Angkatan, "ipk_terakhir": s.IPKTerakhir,
	})
}

// DELETE /api/v1/students/{id} (admin) — soft delete.
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return utils.Err(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()
	tag, err := h.DB.Exec(ctx, `UPDATE students SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	if tag.RowsAffected() == 0 {
		return utils.Err(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}
	return utils.NoContent(c)
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
