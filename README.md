# Kontrak API Students

| Metode | Endpoint | Parameter | Contoh body permintaan | Status yang mungkin dikembalikan | Contoh respons |
| --- | --- | --- | --- | --- | --- |
| GET | `/api/v1/students` | Query: `page`, `limit`, `search`, `sort`, `order`, `is_active`, `min_grade`, `max_grade` | - | `200 OK` | `{"success":true,"message":"daftar student berhasil diambil","data":[{"id":1,"nim":"434241088","name":"Valerina","grade":3.8,"is_active":true},{"id":2,"nim":"434241002","name":"Alisya","grade":3.5,"is_active":true}],"meta":{"page":1,"limit":10,"total":2,"total_pages":1}}` |
| GET | `/api/v1/students/:id` | Path: `id` (angka) | - | `200 OK`, `400 Bad Request`, `404 Not Found` | `{"success":true,"message":"student berhasil ditemukan","data":{"id":1,"nim":"434241088","name":"Valerina","grade":3.8,"is_active":true}}` |
| POST | `/api/v1/students` | - | `{"nim":"434241099","name":"Budi","grade":3.9}` | `201 Created`, `400 Bad Request`, `415 Unsupported Media Type`, `409 Conflict`, `422 Unprocessable Entity` | `{"success":true,"message":"student berhasil dibuat","data":{"id":4,"nim":"434241099","name":"Budi","grade":3.9,"is_active":true}}` |
| PUT | `/api/v1/students/:id` | Path: `id` (angka) | `{"nim":"434241088","name":"Valerina","grade":3.8,"is_active":true}` | `200 OK`, `400 Bad Request`, `404 Not Found`, `409 Conflict`, `415 Unsupported Media Type`, `422 Unprocessable Entity` | `{"success":true,"message":"student berhasil diperbarui total","data":{"id":1,"nim":"434241088","name":"Valerina","grade":3.8,"is_active":true}}` |
| PATCH | `/api/v1/students/:id` | Path: `id` (angka) | `{"name":"Valerina Sari","grade":3.9}` | `200 OK`, `400 Bad Request`, `404 Not Found`, `415 Unsupported Media Type`, `422 Unprocessable Entity` | `{"success":true,"message":"student berhasil diperbarui sebagian","data":{"id":1,"nim":"434241088","name":"Valerina Sari","grade":3.9,"is_active":true}}` |
| DELETE | `/api/v1/students/:id` | Path: `id` (angka) | - | `204 No Content`, `400 Bad Request`, `404 Not Found` | `` (kosong / tanpa body) |

## Skema Tabel

Tabel `students` menyimpan data mahasiswa.

| Kolom | Tipe | Keterangan |
| --- | --- | --- |
| id | SERIAL PRIMARY KEY | ID unik, dibuat otomatis oleh database |
| nim | VARCHAR(20) NOT NULL | Nomor Induk Mahasiswa, wajib unik |
| name | VARCHAR(100) NOT NULL | Nama mahasiswa |
| grade | DECIMAL(3,2) NOT NULL | Nilai mahasiswa |
| is_active | BOOLEAN NOT NULL DEFAULT TRUE | Status aktif mahasiswa |
| created_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | Waktu data dibuat |

Indeks yang diterapkan:
- `students_nim_key` — UNIQUE INDEX pada kolom `nim`, mencegah data dengan NIM ganda tersimpan
- `students_name_lower_idx` — INDEX pada `LOWER(name)`, mempercepat pencarian nama yang tidak membedakan huruf besar dan kecil

## Cara Menyiapkan Basis Data dari Nol

1. Pastikan PostgreSQL sudah terpasang dan berjalan
2. Buat database baru:
   ```
   psql -U postgres -c "CREATE DATABASE pbl;"
   ```
3. Jalankan berkas migrasi untuk membuat tabel:
   ```
   psql -U postgres -d pbl -f migrations/001_create_students.sql
   ```
4. Salin `.env.example` menjadi `.env`, lalu isi sesuai kredensial PostgreSQL masing-masing
5. Jalankan aplikasi:
   ```
   go run .
   ```

## Variabel Environment

| Variabel | Keterangan | Contoh |
| --- | --- | --- |
| APP_PORT | Port aplikasi berjalan | 3000 |
| DB_HOST | Host PostgreSQL | localhost |
| DB_PORT | Port PostgreSQL | 5432 |
| DB_USER | Username PostgreSQL | postgres |
| DB_PASSWORD | Password PostgreSQL | (isi sesuai konfigurasi masing-masing) |
| DB_NAME | Nama database | pbl |
| DB_SSLMODE | Mode SSL koneksi | disable |
| DB_MAX_CONNS | Jumlah maksimum koneksi pada connection pool | 10 |

# SIAKAD Mini API

RESTful API back end untuk layanan akademik sederhana (SIAKAD Mini) yang mengelola data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS). Dibuat untuk UTS Praktikum Pemrograman Backend Lanjut.

- Bahasa: Go
- Framework: Fiber v2
- Database: PostgreSQL (driver pgx)
- Autentikasi: JWT (access token), password disimpan dalam bentuk hash bcrypt

API dipakai oleh dua peran:

1. **admin**: mengelola data mahasiswa.
2. **mahasiswa**: melihat profil sendiri, serta mengambil dan membatalkan mata kuliah pada KRS.

## Struktur Folder

```
uts/
├── main.go            titik awal aplikasi
├── config/            pembacaan .env dan pengaturan Fiber
├── database/          koneksi PostgreSQL (connection pool)
├── migrations/        001_create_tables.sql, 002_seed.sql
├── app/
│   ├── model/         struct data dan bentuk response
│   ├── repository/    query ke database
│   └── service/       validasi, aturan bisnis, dan handler
├── helper/            response, JWT, bcrypt
├── middleware/        pengecekan token, role, dan pembatas login
└── route/             daftar alamat endpoint
```

## Cara Menjalankan

1. Pasang Go dan PostgreSQL.
2. Buat database:

   ```
   psql -U postgres -c "CREATE DATABASE siakad_mini;"
   ```

3. Jalankan migration dan seeder dari folder `uts`:

   ```
   psql -U postgres -d siakad_mini -f migrations/001_create_tables.sql
   psql -U postgres -d siakad_mini -f migrations/002_seed.sql
   ```

   Seeder hanya dijalankan sekali. Kalau perlu mengulang, kosongkan tabel dulu:

   ```
   psql -U postgres -d siakad_mini -c "TRUNCATE users, students, courses, enrollments RESTART IDENTITY CASCADE;"
   ```

4. Salin `.env.example` menjadi `.env`, lalu isi nilainya (lihat tabel di bawah).
5. Jalankan dari dalam folder `uts` supaya file `.env` terbaca:

   ```
   go run .
   ```

Server berjalan di `http://localhost:3000` (sesuai `APP_PORT`). Alamat dasar semua endpoint adalah `/api/v1`.

### Variabel Environment

| Variabel | Keterangan | Contoh |
|---|---|---|
| APP_PORT | Port server | 3000 |
| DB_HOST | Alamat PostgreSQL | localhost |
| DB_PORT | Port PostgreSQL | 5432 |
| DB_USER | User PostgreSQL | postgres |
| DB_PASSWORD | Password PostgreSQL | (isi sendiri) |
| DB_NAME | Nama database | siakad_mini |
| DB_SSLMODE | Mode SSL | disable |
| DB_MAX_CONNS | Maksimal koneksi dalam pool | 10 |
| JWT_SECRET | Kunci penanda tangan token, wajib diisi, gunakan teks acak yang panjang | (isi sendiri) |
| JWT_ISSUER | Penerbit token | siakad-mini |
| JWT_ACCESS_TTL_MINUTES | Masa berlaku access token (menit) | 60 |

File `.env` sudah dikecualikan lewat `.gitignore` dan tidak boleh di-commit.

### Akun Contoh (dari seeder)

Akun ini hanya untuk pengujian.

| Peran | Email | Password |
|---|---|---|
| admin | admin@siakad.ac.id | admin12345 |
| mahasiswa | mhs01@siakad.ac.id sampai mhs20@siakad.ac.id | NIM masing-masing (187221000001 sampai 187221000020) |

Data seeder: 1 admin, 20 mahasiswa, dan 10 mata kuliah. Mata kuliah `IF110` sengaja berkuota 2 untuk menguji kuota penuh.

## Format Response

Semua response berformat JSON dengan struktur seragam (kecuali status 204 yang tidak punya isi).

Response sukses:

```json
{
  "success": true,
  "message": "Data mahasiswa berhasil diambil",
  "data": [ ... ],
  "meta": { "current_page": 1, "per_page": 10, "total": 20, "last_page": 2 }
}
```

Field `meta` hanya muncul pada endpoint daftar yang memakai pagination.

Response error:

```json
{
  "success": false,
  "message": "Anda tidak punya akses ke endpoint ini"
}
```

Response error validasi (422) memuat field `errors` berisi pesan per field:

```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": {
    "nim": ["NIM harus 12 digit angka"],
    "email": ["Format email tidak valid"]
  }
}
```

## Autentikasi

Semua endpoint kecuali login dan health check memerlukan header:

```
Authorization: Bearer <access_token>
```

Token yang tidak ada, salah format, tidak valid, atau sudah kedaluwarsa dijawab dengan status 401.

## Status Code

| Kode | Arti | Contoh pemakaian |
|---|---|---|
| 200 | OK | Data berhasil diambil atau diperbarui |
| 201 | Created | Mahasiswa atau KRS berhasil ditambahkan |
| 204 | No Content | Hapus berhasil, tanpa isi response |
| 400 | Bad Request | Body bukan JSON yang valid |
| 401 | Unauthorized | Token tidak ada atau tidak valid, kredensial login salah |
| 403 | Forbidden | Role tidak berhak, atau mengakses data milik orang lain |
| 404 | Not Found | Data tidak ditemukan atau sudah di-soft delete |
| 409 | Conflict | Mata kuliah sudah pernah diambil pada tahun akademik yang sama |
| 422 | Unprocessable Entity | Validasi gagal, kuota penuh, atau batas SKS terlampaui |
| 429 | Too Many Requests | Terlalu banyak percobaan login yang gagal |
| 500 | Internal Server Error | Kesalahan server (detail hanya dicatat di log server) |

## Ringkasan Endpoint

| No | Method | Endpoint | Akses | Status sukses |
|---|---|---|---|---|
| 1 | POST | /api/v1/auth/login | Publik | 200 |
| 2 | GET | /api/v1/auth/me | Semua role | 200 |
| 3 | GET | /api/v1/students | Admin | 200 |
| 4 | POST | /api/v1/students | Admin | 201 |
| 5 | GET | /api/v1/students/{id} | Admin, mahasiswa (data sendiri) | 200 |
| 6 | PUT | /api/v1/students/{id} | Admin | 200 |
| 7 | DELETE | /api/v1/students/{id} | Admin | 204 |
| 8 | GET | /api/v1/courses | Semua role | 200 |
| 9 | POST | /api/v1/enrollments | Mahasiswa | 201 |
| 10 | DELETE | /api/v1/enrollments/{id} | Mahasiswa (milik sendiri) | 204 |

Endpoint tambahan: `GET /api/v1/health` (publik, 200) untuk mengecek server dan koneksi database.

## Detail Endpoint

### 1. POST /api/v1/auth/login

Body:

| Field | Aturan |
|---|---|
| email | wajib, format email |
| password | wajib, minimal 8 karakter |

Response 200:

```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "access_token": "eyJ...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": { "id": 1, "email": "admin@siakad.ac.id", "role": "admin" }
  }
}
```

Error: 401 (email atau password salah, atau mahasiswa sudah di-soft delete), 422 (validasi), 429 (lebih dari 5 kali gagal dalam 1 menit dari alamat IP yang sama, login yang berhasil tidak dihitung).

### 2. GET /api/v1/auth/me

Mengembalikan data pengguna yang sedang login. Jika role-nya mahasiswa, ikut disertakan data `student`.

```json
{
  "success": true,
  "message": "Data pengguna berhasil diambil",
  "data": {
    "id": 2,
    "email": "mhs01@siakad.ac.id",
    "role": "mahasiswa",
    "student": {
      "nim": "187221000001",
      "nama": "Rina Putri",
      "prodi": "Sistem Informasi",
      "angkatan": 2022
    }
  }
}
```

Untuk admin, bagian `student` tidak ada. Error: 401.

### 3. GET /api/v1/students

Hanya admin. Mahasiswa yang sudah di-soft delete tidak ikut ditampilkan.

Query parameter:

| Parameter | Keterangan |
|---|---|
| page | nomor halaman, default 1 |
| per_page | jumlah per halaman, default 10, maksimal 50 |
| prodi | filter prodi (tidak membedakan huruf besar kecil) |
| angkatan | filter angkatan |
| search | mencari NIM atau nama |
| sort | `nama` atau `-ipk_terakhir` |

Response 200 berisi array mahasiswa (`id`, `nim`, `nama`, `prodi`, `angkatan`, `ipk_terakhir`) dan `meta` (`current_page`, `per_page`, `total`, `last_page`).

Error: 401, 403 (bukan admin), 422 (parameter tidak valid, misalnya `sort` di luar pilihan).

### 4. POST /api/v1/students

Hanya admin. Membuat akun `users` (role mahasiswa, password awal = NIM yang di-hash) dan data `students` dalam satu transaction. Jika salah satu gagal, keduanya dibatalkan.

Body:

| Field | Aturan |
|---|---|
| nim | wajib, unik, 12 digit angka |
| nama | wajib |
| email | wajib, unik, format email |
| prodi | wajib |
| angkatan | wajib, 4 digit, tidak melebihi tahun berjalan |
| ipk_terakhir | opsional, 0,00 sampai 4,00 (jika tidak diisi dianggap 0) |

Response 201 berisi data mahasiswa yang dibuat.

Error: 401, 403, 422 (validasi gagal, atau NIM atau email sudah terdaftar).

### 5. GET /api/v1/students/{id}

Admin boleh membuka semua data. Mahasiswa hanya boleh membuka data miliknya sendiri.

Response 200:

```json
{
  "success": true,
  "message": "Detail mahasiswa berhasil diambil",
  "data": {
    "id": 6,
    "nim": "187221000006",
    "nama": "Fajar Nugroho",
    "prodi": "Teknik Komputer",
    "angkatan": 2024,
    "ipk_terakhir": 2.4,
    "mata_kuliah": [
      {
        "enrollment_id": 1,
        "kode_mk": "IF101",
        "nama_mk": "Algoritma dan Pemrograman",
        "sks": 4,
        "tahun_akademik": "2026/2027-Ganjil"
      }
    ],
    "total_sks": 4,
    "batas_sks": 18
  }
}
```

Error: 401, 403 (mahasiswa membuka data mahasiswa lain), 404 (tidak ditemukan atau sudah di-soft delete).

### 6. PUT /api/v1/students/{id}

Hanya admin. NIM dan email tidak dapat diubah.

Body:

| Field | Aturan |
|---|---|
| nama | wajib |
| prodi | wajib |
| angkatan | wajib, 4 digit, tidak melebihi tahun berjalan |
| ipk_terakhir | opsional, 0,00 sampai 4,00 (jika tidak dikirim, nilai lama dipertahankan) |

Response 200 berisi data mahasiswa setelah diperbarui.

Error: 401, 403, 404, 422.

### 7. DELETE /api/v1/students/{id}

Hanya admin. Soft delete: kolom `deleted_at` diisi, barisnya tidak dihapus dari database. Mahasiswa yang dihapus tidak muncul di endpoint 3, tidak ditemukan di endpoint 5, dan tidak dapat login.

Response 204 tanpa isi. Error: 401, 403, 404.

### 8. GET /api/v1/courses

Semua role yang sudah login.

Query parameter:

| Parameter | Keterangan |
|---|---|
| semester | filter semester (angka) |
| search | mencari kode atau nama mata kuliah |
| available | `true` untuk hanya menampilkan yang kuotanya belum penuh |

Setiap mata kuliah menyertakan `terisi` dan `sisa_kuota` yang dihitung dari tabel `enrollments`:

```json
{
  "id": 10,
  "kode_mk": "IF110",
  "nama_mk": "Keamanan Informasi",
  "sks": 2,
  "semester": 5,
  "kuota": 2,
  "terisi": 1,
  "sisa_kuota": 1
}
```

Error: 401, 422 (semester bukan angka positif).

### 9. POST /api/v1/enrollments

Hanya mahasiswa. Seluruh proses berjalan dalam satu transaction, dengan baris mata kuliah dikunci (`FOR UPDATE`) saat pengecekan kuota supaya dua permintaan yang bersamaan tidak sama-sama lolos.

Body:

| Field | Aturan |
|---|---|
| course_id | wajib, mata kuliah harus ada |
| tahun_akademik | wajib, format `2026/2027-Ganjil` atau `2026/2027-Genap` |

Urutan pengecekan: data mahasiswa, mata kuliah, duplikat, kuota, lalu batas SKS.

Response 201:

```json
{
  "success": true,
  "message": "Mata kuliah berhasil ditambahkan ke KRS",
  "data": {
    "id": 1,
    "course_id": 1,
    "tahun_akademik": "2026/2027-Ganjil",
    "created_at": "2026-10-05T10:00:00+07:00"
  }
}
```

Error: 401, 403 (bukan mahasiswa), 409 (sudah pernah diambil pada tahun akademik yang sama), 422 (validasi gagal, mata kuliah tidak ditemukan, kuota penuh, atau total SKS melebihi batas dengan pesan yang menyebut sisa SKS).

### 10. DELETE /api/v1/enrollments/{id}

Hanya mahasiswa, dan hanya untuk KRS miliknya sendiri. Kuota mata kuliah otomatis bertambah kembali karena dihitung dari isi tabel `enrollments`.

Response 204 tanpa isi. Error: 401, 403 (KRS milik mahasiswa lain), 404.

## Aturan Bisnis

1. Batas SKS per tahun akademik ditentukan dari IPK terakhir:

   | IPK terakhir | Batas SKS |
   |---|---|
   | 3,00 atau lebih | 24 |
   | 2,50 sampai 2,99 | 21 |
   | kurang dari 2,50 | 18 |

2. Mahasiswa tidak dapat mengambil mata kuliah yang sama dua kali pada tahun akademik yang sama (dijaga oleh pengecekan di aplikasi dan constraint `UNIQUE` di database).
3. Mata kuliah yang kuotanya sudah penuh tidak dapat diambil.
4. Mahasiswa hanya dapat mengakses dan mengubah KRS miliknya sendiri.

## Asumsi dan Catatan

- `terisi` dan pengecekan kuota dihitung dari seluruh data KRS pada mata kuliah tersebut, belum dipisah per tahun akademik. Batas SKS dihitung per tahun akademik.
- `total_sks` pada detail mahasiswa (endpoint 5) dihitung dari seluruh KRS yang pernah diambil.
- NIM mahasiswa yang sudah di-soft delete tetap tidak bisa dipakai ulang, karena kolom `nim` tetap unik.
- Panjang teks (misalnya `nama` lebih dari 100 karakter) belum divalidasi di aplikasi. Data seperti itu ditolak oleh database dan dijawab dengan status 500.
- Token mahasiswa yang di-soft delete tetap berlaku sampai kedaluwarsa pada endpoint yang tidak mengecek status akun.