# latihan_fiber

REST API menggunakan [Fiber](https://docs.gofiber.io/) (Go) dengan basis data PostgreSQL, disusun mengikuti pola **repository pattern**. Proyek ini berisi dua entitas:

- **users** — latihan mengikuti Modul Pertemuan 3 (Langkah Panduan)
- **students** — Tugas Mandiri, menerapkan pola yang sama pada entitas mahasiswa

Kedua entitas berbagi satu basis data yang sama.

---

## Skema Tabel

### `users`

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | SERIAL PRIMARY KEY | Id unik, auto-increment |
| username | VARCHAR(50) NOT NULL | Unik (case-insensitive) via `UNIQUE INDEX` pada `LOWER(username)` |
| email | VARCHAR(255) NOT NULL | Diberi indeks tambahan pada `LOWER(email)` |
| password | VARCHAR(255) NOT NULL | Disimpan sebagai teks pada tugas ini (bukan hash) |
| is_active | BOOLEAN NOT NULL DEFAULT TRUE | Status aktif pengguna |
| created_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | Diisi otomatis oleh basis data |

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_key
    ON users (LOWER(username));

CREATE INDEX IF NOT EXISTS users_email_lower_idx
    ON users (LOWER(email));
```

### `students`

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | SERIAL PRIMARY KEY | Id unik, auto-increment |
| nim | VARCHAR(20) NOT NULL | Wajib unik via `UNIQUE INDEX` (`students_nim_key`) |
| name | VARCHAR(100) NOT NULL | Diberi indeks tambahan pada `LOWER(name)` untuk pencarian `ILIKE` |
| grade | NUMERIC(4,2) NOT NULL | Nilai/IPK mahasiswa |
| is_active | BOOLEAN NOT NULL DEFAULT TRUE | Status aktif mahasiswa |
| created_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | Diisi otomatis oleh basis data |

```sql
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(4,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
    ON students (nim);

CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
```

NIM dijaga unik langsung oleh basis data (bukan oleh pemeriksaan manual di kode Go) agar tidak ada celah *race condition* ketika dua permintaan pembuatan data datang hampir bersamaan.

---

## Menyiapkan Basis Data dari Nol

1. **Buat database kosong:**
   ```bash
   psql -U postgres -c "CREATE DATABASE nama_database_anda;"
   ```

2. **Jalankan migrasi secara berurutan** (users lebih dulu, karena bernomor lebih kecil):
   ```bash
   psql -U postgres -d nama_database_anda -f migrations/001_create_users.sql
   psql -U postgres -d nama_database_anda -f migrations/002_create_students.sql
   ```

3. **Verifikasi kedua tabel sudah terbentuk:**
   ```bash
   psql -U postgres -d nama_database_anda -c "\d users"
   psql -U postgres -d nama_database_anda -c "\d students"
   ```

4. **Salin `.env.example` menjadi `.env`**, lalu isi sesuai kredensial PostgreSQL lokal Anda:
   ```bash
   cp .env.example .env
   ```

5. **Unduh dependency Go dan jalankan server:**
   ```bash
   go mod tidy
   go run .
   ```

6. **Cek server berjalan** (endpoint ini turut memeriksa koneksi basis data):
   ```bash
   curl http://localhost:3000/api/v1/health
   ```

---

## Variabel Environment

Seluruh variabel berikut wajib diisi pada berkas `.env` (lihat `.env.example`):

| Variabel | Keterangan |
|---|---|
| `APP_PORT` | Port tempat server Fiber berjalan (contoh: `3000`) |
| `DB_HOST` | Host PostgreSQL (contoh: `localhost`) |
| `DB_PORT` | Port PostgreSQL (default: `5432`) |
| `DB_USER` | Username PostgreSQL |
| `DB_PASSWORD` | Password PostgreSQL |
| `DB_NAME` | Nama database yang sudah dibuat pada langkah di atas |
| `DB_SSLMODE` | Mode SSL koneksi (`disable` untuk lingkungan lokal) |
| `DB_MAX_CONNS` | Jumlah koneksi maksimum pada connection pool (contoh: `10`) |

> Berkas `.env` **tidak** ikut ter-commit ke repositori (lihat `.gitignore`). Gunakan `.env.example` sebagai acuan nama variabel yang diperlukan.

---

## Endpoint

### Users — `/api/v1/users`
| Method | Path | Keterangan |
|---|---|---|
| GET | `/` | Daftar user (mendukung `search`, `sort`, `order`, `page`, `limit`, `is_active`) |
| GET | `/:id` | Detail satu user |
| POST | `/` | Membuat user baru |
| PUT | `/:id` | Mengganti seluruh data user |
| PATCH | `/:id` | Mengubah sebagian data user |
| DELETE | `/:id` | Menghapus user |

### Students — `/api/v1/students`
| Method | Path | Keterangan |
|---|---|---|
| GET | `/` | Daftar student (mendukung `search`, `sort`, `order`, `page`, `limit`, `is_active`, `min_grade`, `max_grade`) |
| GET | `/:id` | Detail satu student |
| POST | `/` | Membuat student baru |
| PUT | `/:id` | Mengganti seluruh data student |
| PATCH | `/:id` | Mengubah sebagian data student |
| DELETE | `/:id` | Menghapus student |

### Lainnya
| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/v1/health` | Memeriksa status server dan koneksi basis data |

Seluruh endpoint `POST`, `PUT`, `PATCH` mewajibkan header `Content-Type: application/json`.

---

## Struktur Proyek

```
latihan_fiber/
├── app/
│   ├── model/
│   │   └── user.go          # struct User, Student, request/response, WebResponse, dst.
│   └── repository/
│       ├── user_repository.go
│       └── student_repository.go
├── config/
│   └── env.go                # memuat dan membaca variabel environment
├── database/
│   └── postgres.go           # connection pool ke PostgreSQL (pgx)
├── migrations/
│   ├── 001_create_users.sql
│   └── 002_create_students.sql
├── .env.example
├── .gitignore
├── go.mod / go.sum
├── main.go
├── handler.go                 # handler untuk users
├── student_handler.go         # handler untuk students
└── helper.go                  # fungsi respons JSON, parsing query, dsb.
```