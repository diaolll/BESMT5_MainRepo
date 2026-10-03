package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"siakaduts/config"
	"siakaduts/db"
	"siakaduts/middleware"
	"siakaduts/routes"
	"siakaduts/utils"
)

//go:embed migrations/001_schema.sql
var schemaSQL string

func main() {
	doMigrate := flag.Bool("migrate", false, "jalankan migration skema lalu keluar")
	doSeed := flag.Bool("seed", false, "jalankan seeder lalu keluar")
	flag.Parse()

	cfg := config.Load()
	if len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET wajib minimal 32 karakter")
	}

	ctx := context.Background()
	pool, err := config.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if *doMigrate || *doSeed {
		if *doMigrate {
			if err := db.Migrate(ctx, pool, schemaSQL); err != nil {
				log.Fatalf("migrate gagal: %v", err)
			}
			fmt.Println("migration OK: tabel users, students, courses, enrollments siap")
		}
		if *doSeed {
			if err := db.Seed(ctx, pool); err != nil {
				log.Fatalf("seed gagal: %v", err)
			}
			fmt.Println("seeder OK: 1 admin, 20 mahasiswa, 10 mata kuliah")
		}
		return
	}

	jwtm := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.AccessTTLMin)
	limiter := middleware.NewLoginFailLimiter()

	app := fiber.New(fiber.Config{
		AppName:   "SIAKAD Mini (UTS PBE)",
		BodyLimit: 1 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			if code == fiber.StatusNotFound {
				return utils.Err(c, code, "Endpoint tidak ditemukan")
			}
			return utils.Err(c, code, "Terjadi kesalahan pada server")
		},
	})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return utils.OK(c, fiber.StatusOK, "server berjalan", nil)
	})

	routes.Register(app, routes.Deps{DB: pool, JWT: jwtm, Limiter: limiter})

	fmt.Println("SIAKAD Mini berjalan di port " + cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		fmt.Fprintln(os.Stderr, "server berhenti:", err)
		os.Exit(1)
	}
}
