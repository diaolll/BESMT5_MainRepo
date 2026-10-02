-- SIAKAD Mini — skema database (soal UTS PBE)
-- Urutan dibuat agar FK selalu menunjuk tabel yang sudah ada.

CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    email      VARCHAR(120) NOT NULL,
    password   VARCHAR(255) NOT NULL,          -- hash bcrypt, bukan plaintext
    role       VARCHAR(20)  NOT NULL CHECK (role IN ('admin','mahasiswa')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_key ON users (LOWER(email));

CREATE TABLE IF NOT EXISTS students (
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nim          VARCHAR(12) NOT NULL,
    nama         VARCHAR(100) NOT NULL,
    prodi        VARCHAR(100) NOT NULL,
    angkatan     INTEGER NOT NULL,
    ipk_terakhir NUMERIC(3,2) NOT NULL DEFAULT 0 CHECK (ipk_terakhir >= 0 AND ipk_terakhir <= 4),
    deleted_at   TIMESTAMPTZ NULL,             -- soft delete endpoint 7
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key ON students (nim);
CREATE INDEX IF NOT EXISTS students_prodi_idx ON students (prodi);
CREATE INDEX IF NOT EXISTS students_nama_lower_idx ON students (LOWER(nama));

CREATE TABLE IF NOT EXISTS courses (
    id        SERIAL PRIMARY KEY,
    kode_mk   VARCHAR(20)  NOT NULL,
    nama_mk   VARCHAR(150) NOT NULL,
    sks       INTEGER NOT NULL CHECK (sks > 0),
    semester  INTEGER NOT NULL CHECK (semester > 0),
    kuota     INTEGER NOT NULL CHECK (kuota > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS courses_kode_key ON courses (kode_mk);

CREATE TABLE IF NOT EXISTS enrollments (
    id             SERIAL PRIMARY KEY,
    student_id     INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id      INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    tahun_akademik VARCHAR(20) NOT NULL,       -- format: 2026/2027-Ganjil
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT enrollments_unique UNIQUE (student_id, course_id, tahun_akademik)
);
CREATE INDEX IF NOT EXISTS enrollments_student_idx ON enrollments (student_id);
CREATE INDEX IF NOT EXISTS enrollments_course_idx ON enrollments (course_id);
