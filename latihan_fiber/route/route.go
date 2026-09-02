package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan_fiber/app/service"
	"latihan_fiber/helper"
	"latihan_fiber/middleware"
)

// Register memetakan URL ke method pada service.
//
// Tidak ada logika bisnis, tidak ada query, tidak ada validasi di sini —
// hanya daftar alamat dan siapa yang melayaninya.
func Register(
	app *fiber.App, pool *pgxpool.Pool,
	userService *service.UserService, studentService *service.StudentService,
) {
	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(pool))

	users := api.Group("/users", middleware.RequireJSON)
	users.Get("/", userService.List)
	users.Get("/:id", userService.Get)
	users.Post("/", userService.Create)
	users.Put("/:id", userService.Replace)
	users.Patch("/:id", userService.Patch)
	users.Delete("/:id", userService.Delete)

	students := api.Group("/students", middleware.RequireJSON)
	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)
}

// healthCheck melaporkan kondisi layanan beserta databasenya.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}