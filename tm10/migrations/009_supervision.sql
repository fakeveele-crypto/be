-- Migration 009: dosen pembimbing tugas akhir (Modul 9, tugas D.2)
-- psql -U postgres -d praktikum_backend -f migrations/009_supervision.sql
CREATE TABLE IF NOT EXISTS supervisors (
    id               SERIAL PRIMARY KEY,
    name             VARCHAR(120) NOT NULL,
    quota            INTEGER      NOT NULL CHECK (quota > 0),
    supervised_count INTEGER      NOT NULL DEFAULT 0,
    version          INTEGER      NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- Jaring pengaman terakhir, sama seperti pada courses.
    CONSTRAINT supervisors_count_within_quota
        CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

-- Aturan "seorang mahasiswa hanya boleh memiliki SATU pembimbing" dijaga
-- database lewat PRIMARY KEY pada student_id: satu mahasiswa hanya dapat
-- muncul satu kali di tabel ini, sepintar apa pun urutan request-nya.
CREATE TABLE IF NOT EXISTS supervisions (
    student_id    INTEGER     PRIMARY KEY REFERENCES students(id)    ON DELETE CASCADE,
    supervisor_id INTEGER     NOT NULL    REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS supervisions_supervisor_idx ON supervisions (supervisor_id);

INSERT INTO supervisors (name, quota)
SELECT v.name, v.quota
FROM (VALUES ('Dosen Pembimbing A', 5),
             ('Dosen Pembimbing B', 5),
             ('Dosen Pembimbing C', 3)) AS v(name, quota)
WHERE NOT EXISTS (SELECT 1 FROM supervisors);

INSERT INTO permissions (name, description) VALUES
    ('supervision:manage', 'Menetapkan, memindahkan, dan mengatur kuota dosen pembimbing')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'supervision:manage')
ON CONFLICT DO NOTHING;
