-- Urutan column HARUS sama persis dengan ORDER BY, termasuk DESC
CREATE INDEX IF NOT EXISTS users_created_at_id_desc_idx
  ON users (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
  ON students (created_at DESC, id DESC);
