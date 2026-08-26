
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade VARCHAR(5) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- NIM wajib unik. Dijaga di basis data (bukan di kode Go) supaya
-- tidak ada celah race condition ketika dua permintaan pembuatan
-- data datang hampir bersamaan — basis datalah yang memutuskan
-- siapa lebih dulu, sama seperti kasus username pada modul.
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
    ON students (nim);

-- Indeks tambahan untuk mempercepat pencarian berdasarkan nama,
-- karena endpoint daftar akan memakai ILIKE pada kolom ini.
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));