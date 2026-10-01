package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Config struct {
	AppPort        string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	JWTSecret      string
	JWTIssuer      string
	AccessTTLMin   int
	AllowedOrigins string
}

func Load() Config {
	_ = godotenv.Load() // abaikan bila .env tidak ada (pakai env sistem)
	return Config{
		AppPort:        env("APP_PORT", "3001"),
		DBHost:         env("DB_HOST", "/tmp"),
		DBPort:         env("DB_PORT", "5432"),
		DBUser:         env("DB_USER", "diaul"),
		DBPassword:     env("DB_PASSWORD", ""),
		DBName:         env("DB_NAME", "siakad_uts"),
		DBSSLMode:      env("DB_SSLMODE", "disable"),
		JWTSecret:      env("JWT_SECRET", "ganti-minimal-32-karakter-rahasia-uts"),
		JWTIssuer:      env("JWT_ISSUER", "siakad-mini"),
		AccessTTLMin:   envInt("JWT_ACCESS_TTL_MINUTES", 60),
		AllowedOrigins: env("ALLOWED_ORIGINS", "http://localhost:5173"),
	}
}

func (c Config) DSN() string {
	// Dukungan socket Unix (DB_HOST=/tmp) maupun TCP (DB_HOST=localhost).
	if c.DBHost == "" || c.DBHost[0] == '/' {
		if c.DBPassword == "" {
			return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s&host=%s",
				c.DBUser, "localhost", c.DBPort, c.DBName, c.DBSSLMode, c.DBHost)
		}
		return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=%s&host=%s",
			c.DBUser, c.DBPassword, c.DBPort, c.DBName, c.DBSSLMode, c.DBHost)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("konfigurasi database tidak valid: %w", err)
	}
	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal terhubung ke database: %w", err)
	}
	return pool, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("peringatan: %s bukan angka (%q), memakai bawaan %d", key, v, fallback)
		return fallback
	}
	return n
}
