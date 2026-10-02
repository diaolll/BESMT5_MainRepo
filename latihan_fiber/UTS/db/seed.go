package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"siakaduts/utils"
)

// Seed mengisi data minimal soal: 1 admin, 20 mahasiswa (+akun user),
// 10 mata kuliah. Seluruh password di-hash bcrypt.
// Password awal mahasiswa = NIM-nya (sesuai proses endpoint 4).
// Idempoten: aman dijalankan ulang (ON CONFLICT DO NOTHING)./res
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	adminHash, err := utils.HashPassword("Admin12345")
	if err != nil {
		return err
	}
	var adminID int
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, password, role)
		VALUES ('admin@siakad.ac.id', $1, 'admin')
		ON CONFLICT ((LOWER(email))) DO UPDATE SET email=EXCLUDED.email
		RETURNING id`, adminHash).Scan(&adminID)
	if err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	prodis := []string{"Sistem Informasi", "Informatika", "Teknik Elektro", "Manajemen"}
	names := []string{
		"Rina Putri", "Budi Santoso", "Siti Aminah", "Agus Wijaya", "Dewi Lestari",
		"Joko Prasetyo", "Ayu Wulandari", "Rudi Hartono", "Nina Kurnia", "Dedi Supriadi",
		"Maya Anggraini", "Fajar Nugroho", "Lina Marlina", "Hendra Gunawan", "Tina Susanti",
		"Rudi Hermawan", "Sari Melati", "Eko Purnomo", "Fitri Handayani", "Yoga Saputra",
	}
	ipks := []float64{3.45, 2.75, 2.10, 3.80, 3.10, 2.60, 1.90, 3.25, 2.90, 3.00,
		2.40, 3.65, 2.20, 3.15, 2.85, 3.55, 2.35, 3.90, 2.55, 3.05}

	for i := 0; i < 20; i++ {
		nim := fmt.Sprintf("187221%06d", i+1) // 12 digit, unik
		email := fmt.Sprintf("mhs%02d@siakad.ac.id", i+1)
		hash, err := utils.HashPassword(nim)
		if err != nil {
			return err
		}
		var uid int
		err = tx.QueryRow(ctx, `
			INSERT INTO users (email, password, role)
			VALUES ($1, $2, 'mahasiswa')
			ON CONFLICT ((LOWER(email))) DO UPDATE SET email=EXCLUDED.email
			RETURNING id`, email, hash).Scan(&uid)
		if err != nil {
			return fmt.Errorf("seed user %d: %w", i+1, err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (nim) DO UPDATE SET nama=EXCLUDED.nama, prodi=EXCLUDED.prodi,
				angkatan=EXCLUDED.angkatan, ipk_terakhir=EXCLUDED.ipk_terakhir`,
			uid, nim, names[i], prodis[i%len(prodis)], 2022+(i%3), ipks[i])
		if err != nil {
			return fmt.Errorf("seed student %d: %w", i+1, err)
		}
	}

	courses := []struct {
		kode, nama string
		sks, smt, kuota int
	}{
		{"SI101", "Pengantar Sistem Informasi", 3, 1, 30},
		{"SI102", "Algoritma dan Pemrograman", 4, 1, 30},
		{"SI201", "Basis Data", 3, 2, 25},
		{"SI202", "Pemrograman Web", 3, 2, 25},
		{"SI203", "Jaringan Komputer", 3, 3, 2}, // kuota kecil untuk uji 422 kuota penuh
		{"SI301", "Rekayasa Perangkat Lunak", 3, 3, 30},
		{"SI302", "Pemrograman Backend", 4, 4, 30},
		{"SI303", "Kecerdasan Buatan", 3, 4, 30},
		{"SI401", "Proyek Perangkat Lunak", 4, 5, 30},
		{"SI402", "Keamanan Sistem Informasi", 3, 5, 30},
	}
	for _, mk := range courses {
		_, err := tx.Exec(ctx, `
			INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (kode_mk) DO UPDATE SET nama_mk=EXCLUDED.nama_mk,
				sks=EXCLUDED.sks, semester=EXCLUDED.semester, kuota=EXCLUDED.kuota`,
			mk.kode, mk.nama, mk.sks, mk.smt, mk.kuota)
		if err != nil {
			return fmt.Errorf("seed course %s: %w", mk.kode, err)
		}
	}

	return tx.Commit(ctx)
}

// Migrate menjalankan 001_schema.sql yang disematkan via go:embed di main.
func Migrate(ctx context.Context, pool *pgxpool.Pool, schemaSQL string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, err := pool.Exec(ctx, schemaSQL)
	return err
}
