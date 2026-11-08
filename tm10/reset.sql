-- Kembalikan keadaan sebelum SETIAP uji beban.
-- psql -U postgres -d praktikum_backend -f reset.sql
DELETE FROM enrollments
WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%');

UPDATE courses c
SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id);
