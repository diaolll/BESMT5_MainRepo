package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakaduts/handlers"
	"siakaduts/middleware"
	"siakaduts/utils"
)

type Deps struct {
	DB      *pgxpool.Pool
	JWT     *utils.JWTManager
	Limiter *middleware.LoginFailLimiter
}

func Register(app *fiber.App, d Deps) {
	auth := &handlers.AuthHandler{DB: d.DB, JWT: d.JWT, Limiter: d.Limiter}
	students := &handlers.StudentHandler{DB: d.DB}
	courses := &handlers.CourseHandler{DB: d.DB}
	enrollments := &handlers.EnrollmentHandler{DB: d.DB}

	api := app.Group("/api/v1")

	// Endpoint 1 — publik (tanpa token).
	api.Post("/auth/login", auth.Login)

	// Semua endpoint lain wajib token.
	protected := api.Group("", middleware.RequireAuth(d.JWT))
	protected.Get("/auth/me", auth.Me) // endpoint 2

	// Endpoint 3,4,6,7 — khusus admin; endpoint 5 campuran (dicek di handler).
	protected.Get("/students", middleware.RequireRole("admin"), students.List)
	protected.Post("/students", middleware.RequireRole("admin"), students.Create)
	protected.Get("/students/:id", students.Get)
	protected.Put("/students/:id", middleware.RequireRole("admin"), students.Update)
	protected.Delete("/students/:id", middleware.RequireRole("admin"), students.Delete)

	// Endpoint 8 — semua role yang login.
	protected.Get("/courses", courses.List)

	// Endpoint 9,10 — khusus mahasiswa (kepemilikan dicek di handler).
	protected.Post("/enrollments", middleware.RequireRole("mahasiswa"), enrollments.Create)
	protected.Delete("/enrollments/:id", middleware.RequireRole("mahasiswa"), enrollments.Delete)
}
