-- ---------------------------------------------------------------
-- Modul 6 — Tugas Mandiri C.1
-- Permission students + column owner_id pada students
-- ---------------------------------------------------------------

-- Tambah permission untuk students
INSERT INTO permissions (name, description) VALUES
  ('student:list', 'Melihat daftar seluruh student'),
  ('student:read:any', 'Melihat data student mana pun'),
  ('student:create', 'Membuat data student'),
  ('student:update:any', 'Mengubah data student mana pun'),
  ('student:delete', 'Menghapus student')
ON CONFLICT (name) DO NOTHING;

-- Mapping permission students ke role
INSERT INTO role_permissions (role_name, permission_name) VALUES
  ('admin', 'student:list'),
  ('admin', 'student:read:any'),
  ('admin', 'student:create'),
  ('admin', 'student:update:any'),
  ('admin', 'student:delete'),
  ('staff', 'student:list'),
  ('staff', 'student:read:any'),
  ('staff', 'student:create')
ON CONFLICT DO NOTHING;

-- Tambah column owner_id ke students
-- Baris lama belum punya owner -> biarkan NULL dulu, lalu isi fallback.
ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id INTEGER REFERENCES users(id) ON DELETE SET NULL;

-- Isi owner_id untuk data lama: pakai user pertama (admin) jika ada, atau biarkan NULL.
-- Strategi fail-safe: baris tanpa owner hanya bisa diakses oleh pemegang permission :any.
DO $$
DECLARE
  fallback_owner INTEGER;
BEGIN
  SELECT id INTO fallback_owner FROM users ORDER BY id LIMIT 1;
  IF fallback_owner IS NOT NULL THEN
    UPDATE students SET owner_id = fallback_owner WHERE owner_id IS NULL;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students (owner_id);
