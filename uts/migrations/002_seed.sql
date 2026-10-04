-- Data awal: 1 admin, 20 mahasiswa, 10 mata kuliah.
-- Jalankan SEKALI setelah 001_create_tables.sql.

-- pgcrypto dipakai untuk membuat hash bcrypt langsung di PostgreSQL
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1. admin (password: admin12345)
INSERT INTO users (email, password, role)
VALUES ('admin@siakad.ac.id', crypt('admin12345', gen_salt('bf', 10)), 'admin');

-- 2. data mahasiswa disimpan dulu di tabel sementara
CREATE TEMP TABLE seed_mahasiswa (
    email    VARCHAR(100),
    nim      VARCHAR(12),
    nama     VARCHAR(100),
    prodi    VARCHAR(100),
    angkatan INT,
    ipk      NUMERIC(3,2)
);

INSERT INTO seed_mahasiswa (email, nim, nama, prodi, angkatan, ipk) VALUES
('mhs01@siakad.ac.id', '187221000001', 'Rina Putri',     'Sistem Informasi',   2022, 3.45),
('mhs02@siakad.ac.id', '187221000002', 'Budi Santoso',   'Teknik Informatika', 2022, 3.10),
('mhs03@siakad.ac.id', '187221000003', 'Citra Lestari',  'Teknik Informatika', 2023, 3.80),
('mhs04@siakad.ac.id', '187221000004', 'Dimas Pratama',  'Sistem Informasi',   2023, 2.75),
('mhs05@siakad.ac.id', '187221000005', 'Eka Wulandari',  'Teknik Informatika', 2022, 3.25),
('mhs06@siakad.ac.id', '187221000006', 'Fajar Nugroho',  'Teknik Komputer',    2024, 2.40),
('mhs07@siakad.ac.id', '187221000007', 'Gita Anggraini', 'Sistem Informasi',   2024, 3.60),
('mhs08@siakad.ac.id', '187221000008', 'Hendra Wijaya',  'Teknik Informatika', 2023, 2.90),
('mhs09@siakad.ac.id', '187221000009', 'Indah Permata',  'Sistem Informasi',   2022, 3.35),
('mhs10@siakad.ac.id', '187221000010', 'Joko Susilo',    'Teknik Komputer',    2025, 2.55),
('mhs11@siakad.ac.id', '187221000011', 'Kartika Dewi',   'Teknik Informatika', 2024, 3.05),
('mhs12@siakad.ac.id', '187221000012', 'Lukman Hakim',   'Sistem Informasi',   2023, 2.60),
('mhs13@siakad.ac.id', '187221000013', 'Maya Sari',      'Teknik Informatika', 2022, 3.70),
('mhs14@siakad.ac.id', '187221000014', 'Naufal Rizky',   'Teknik Komputer',    2025, 2.30),
('mhs15@siakad.ac.id', '187221000015', 'Olivia Putri',   'Sistem Informasi',   2024, 3.15),
('mhs16@siakad.ac.id', '187221000016', 'Putra Mahendra', 'Teknik Informatika', 2023, 2.85),
('mhs17@siakad.ac.id', '187221000017', 'Qori Amalia',    'Sistem Informasi',   2022, 3.90),
('mhs18@siakad.ac.id', '187221000018', 'Rudi Hartono',   'Teknik Informatika', 2025, 2.45),
('mhs19@siakad.ac.id', '187221000019', 'Siti Nurhaliza', 'Teknik Komputer',    2023, 3.00),
('mhs20@siakad.ac.id', '187221000020', 'Tegar Ramadhan', 'Sistem Informasi',   2024, 2.70);

-- akun login mahasiswa (password awal = NIM)
INSERT INTO users (email, password, role)
SELECT email, crypt(nim, gen_salt('bf', 10)), 'mahasiswa'
FROM seed_mahasiswa;

-- data mahasiswa, dihubungkan ke akunnya lewat email
INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT u.id, s.nim, s.nama, s.prodi, s.angkatan, s.ipk
FROM seed_mahasiswa s
JOIN users u ON u.email = s.email;

DROP TABLE seed_mahasiswa;

-- 3. mata kuliah (IF110 kuotanya sengaja 2, untuk uji kuota penuh)
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES
('IF101', 'Algoritma dan Pemrograman', 4, 1, 40),
('IF102', 'Basis Data',                3, 3, 35),
('IF103', 'Pemrograman Web',           3, 3, 30),
('IF104', 'Struktur Data',             3, 2, 30),
('IF105', 'Jaringan Komputer',         3, 3, 25),
('IF106', 'Rekayasa Perangkat Lunak',  3, 4, 30),
('IF107', 'Pemrograman Backend',       4, 5, 20),
('IF108', 'Kecerdasan Buatan',         3, 5, 20),
('IF109', 'Pemrograman Mobile',        3, 5, 25),
('IF110', 'Keamanan Informasi',        2, 5, 2);