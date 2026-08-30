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