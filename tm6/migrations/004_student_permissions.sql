-- ---------------------------------------------------------------
-- Permission baru untuk entity students
-- ---------------------------------------------------------------
INSERT INTO permissions (name, description) VALUES
    ('student:list',       'Melihat daftar seluruh student'),
    ('student:read:any',   'Melihat detail student milik siapa pun'),
    ('student:create',     'Menambahkan student baru'),
    ('student:update:any', 'Mengubah data student milik siapa pun'),
    ('student:delete',     'Menghapus data student')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT (role_name, permission_name) DO NOTHING;

-- ---------------------------------------------------------------
-- Kolom owner_id — dasar pemeriksaan kepemilikan
-- ---------------------------------------------------------------
ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id INTEGER REFERENCES users(id);

-- Data lama belum punya owner_id. Kita tandai baris lama itu sebagai
-- milik user paling pertama yang terdaftar, sebagai penanda "data
-- peninggalan sebelum sistem kepemilikan diterapkan", bukan klaim
-- bahwa user itu beneran yang mendaftarkannya.
UPDATE students
SET owner_id = (SELECT id FROM users ORDER BY id LIMIT 1)
WHERE owner_id IS NULL;

ALTER TABLE students ALTER COLUMN owner_id SET NOT NULL;