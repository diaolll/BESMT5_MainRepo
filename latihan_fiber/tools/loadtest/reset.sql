-- Kembalikan data uji beban ke keadaan konsisten.
-- Jalankan sebelum SETIAP uji beban:
--   psql -U postgres -d pbepraktm3 -f tools/loadtest/reset.sql

DELETE FROM enrollments
WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%');

UPDATE courses c
SET enrolled_count = (
    SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id
);
