package handlers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakaduts/models"
	"siakaduts/utils"
)

type CourseHandler struct {
	DB *pgxpool.Pool
}

// GET /api/v1/courses — semua role yang login.
// Query: semester, search (kode_mk/nama_mk), available=true.
func (h *CourseHandler) List(c *fiber.Ctx) error {
	semStr := strings.TrimSpace(c.Query("semester"))
	search := strings.TrimSpace(c.Query("search"))
	onlyAvailable := strings.ToLower(strings.TrimSpace(c.Query("available"))) == "true"

	where := "WHERE 1=1"
	args := []any{}
	push := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}
	if semStr != "" {
		if s, err := strconv.Atoi(semStr); err == nil {
			where += " AND c.semester = " + push(s)
		}
	}
	if search != "" {
		where += " AND (c.kode_mk ILIKE " + push("%"+search+"%") + " OR c.nama_mk ILIKE " + push("%"+search+"%") + ")"
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	rows, err := h.DB.Query(ctx, `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COUNT(e.id)::int AS terisi
		FROM courses c LEFT JOIN enrollments e ON e.course_id = c.id
		`+where+`
		GROUP BY c.id ORDER BY c.semester, c.kode_mk`, args...)
	if err != nil {
		return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	defer rows.Close()

	list := []fiber.Map{}
	for rows.Next() {
		var m models.Course
		if err := rows.Scan(&m.ID, &m.KodeMK, &m.NamaMK, &m.SKS, &m.Semester, &m.Kuota, &m.Terisi); err != nil {
			return utils.Err(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
		}
		m.SisaKuota = m.Kuota - m.Terisi
		if onlyAvailable && m.SisaKuota <= 0 {
			continue
		}
		list = append(list, fiber.Map{
			"id": m.ID, "kode_mk": m.KodeMK, "nama_mk": m.NamaMK,
			"sks": m.SKS, "semester": m.Semester, "kuota": m.Kuota,
			"terisi": m.Terisi, "sisa_kuota": m.SisaKuota,
		})
	}
	return utils.OK(c, fiber.StatusOK, "Data mata kuliah berhasil diambil", list)
}
