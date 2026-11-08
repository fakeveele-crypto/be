-- akun login (admin dan mahasiswa)
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    email      VARCHAR(100) NOT NULL UNIQUE,
    password   VARCHAR(100) NOT NULL,
    role       VARCHAR(20)  NOT NULL CHECK (role IN ('admin', 'mahasiswa')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- data mahasiswa, 1 baris untuk 1 akun users
CREATE TABLE IF NOT EXISTS students (
    id           SERIAL PRIMARY KEY,
    user_id      INT           NOT NULL UNIQUE REFERENCES users(id),
    nim          VARCHAR(12)   NOT NULL UNIQUE,
    nama         VARCHAR(100)  NOT NULL,
    prodi        VARCHAR(100)  NOT NULL,
    angkatan     INT           NOT NULL,
    ipk_terakhir NUMERIC(3,2)  NOT NULL DEFAULT 0 CHECK (ipk_terakhir BETWEEN 0 AND 4),
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- daftar mata kuliah
CREATE TABLE IF NOT EXISTS courses (
    id         SERIAL PRIMARY KEY,
    kode_mk    VARCHAR(10)  NOT NULL UNIQUE,
    nama_mk    VARCHAR(100) NOT NULL,
    sks        INT          NOT NULL CHECK (sks > 0),
    semester   INT          NOT NULL,
    kuota      INT          NOT NULL CHECK (kuota >= 0),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- KRS: mahasiswa mengambil mata kuliah pada tahun akademik tertentu
CREATE TABLE IF NOT EXISTS enrollments (
    id             SERIAL PRIMARY KEY,
    student_id     INT         NOT NULL REFERENCES students(id),
    course_id      INT         NOT NULL REFERENCES courses(id),
    tahun_akademik VARCHAR(20) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, course_id, tahun_akademik)
);