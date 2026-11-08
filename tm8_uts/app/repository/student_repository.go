package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/uts/app/model"
)

type StudentRepository interface {
	FindAll(ctx context.Context, f model.StudentFilter) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindEnrollments(ctx context.Context, studentID int) ([]model.EnrolledCourse, error)
	Create(ctx context.Context, req model.CreateStudentRequest, passwordHash string) (model.Student, error)
	Update(ctx context.Context, id int, req model.UpdateStudentRequest) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

// FindAll mengambil daftar mahasiswa aktif (yang belum di-soft delete).
func (r *studentPostgresRepository) FindAll(ctx context.Context, f model.StudentFilter) ([]model.Student, int, error) {
	where := "WHERE deleted_at IS NULL"
	args := []any{}

	if f.Prodi != "" {
		where += fmt.Sprintf(" AND LOWER(prodi) = LOWER($%d)", len(args)+1)
		args = append(args, f.Prodi)
	}
	if f.Angkatan > 0 {
		where += fmt.Sprintf(" AND angkatan = $%d", len(args)+1)
		args = append(args, f.Angkatan)
	}
	if f.Search != "" {
		where += fmt.Sprintf(" AND (nim ILIKE $%d OR nama ILIKE $%d)", len(args)+1, len(args)+1)
		args = append(args, "%"+f.Search+"%")
	}

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total data: %w", err)
	}

	// whitelist ORDER BY: nilai sort dari user tidak pernah dimasukkan langsung ke SQL
	orderBy := "id ASC"
	switch f.Sort {
	case "nama":
		orderBy = "nama ASC, id ASC"
	case "-ipk_terakhir":
		orderBy = "ipk_terakhir DESC, id ASC"
	}

	offset := (f.Page - 1) * f.PerPage
	sqlText := fmt.Sprintf(`
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir
		FROM students %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, where, orderBy, len(args)+1, len(args)+2)
	args = append(args, f.PerPage, offset)

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPK); err != nil {
			return nil, 0, fmt.Errorf("gagal membaca baris: %w", err)
		}
		hasil = append(hasil, s)
	}
	return hasil, total, nil
}

// FindByID mengambil satu mahasiswa aktif. Yang sudah di-soft delete dianggap tidak ada.
func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir
		 FROM students WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPK)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil mahasiswa: %w", err)
	}
	return s, nil
}

// FindEnrollments mengambil daftar mata kuliah yang diambil seorang mahasiswa.
func (r *studentPostgresRepository) FindEnrollments(ctx context.Context, studentID int) ([]model.EnrolledCourse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, c.kode_mk, c.nama_mk, c.sks, e.tahun_akademik
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1
		 ORDER BY e.id ASC`, studentID)
	if err != nil {
		return nil, fmt.Errorf("mengambil KRS: %w", err)
	}
	defer rows.Close()

	hasil := []model.EnrolledCourse{}
	for rows.Next() {
		var ec model.EnrolledCourse
		if err := rows.Scan(&ec.EnrollmentID, &ec.KodeMK, &ec.NamaMK, &ec.SKS, &ec.TahunAkademik); err != nil {
			return nil, fmt.Errorf("gagal membaca baris: %w", err)
		}
		hasil = append(hasil, ec)
	}
	return hasil, nil
}

// Create membuat akun users dan data students dalam satu transaction.
func (r *studentPostgresRepository) Create(ctx context.Context, req model.CreateStudentRequest, passwordHash string) (model.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Student{}, fmt.Errorf("memulai transaction: %w", err)
	}
	// kalau ada langkah yang gagal, semua perubahan dibatalkan
	defer tx.Rollback(ctx)

	var userID int
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa') RETURNING id`,
		req.Email, passwordHash,
	).Scan(&userID)
	if err != nil {
		return model.Student{}, translateStudentError(err, "membuat akun")
	}

	// IPK opsional, kalau tidak diisi dianggap 0
	ipk := 0.0
	if req.IPK != nil {
		ipk = *req.IPK
	}

	s := model.Student{
		UserID:   userID,
		NIM:      req.NIM,
		Nama:     req.Nama,
		Prodi:    req.Prodi,
		Angkatan: req.Angkatan,
		IPK:      ipk,
	}
	err = tx.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		s.UserID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPK,
	).Scan(&s.ID)
	if err != nil {
		return model.Student{}, translateStudentError(err, "membuat mahasiswa")
	}

	// semua berhasil, simpan permanen
	if err := tx.Commit(ctx); err != nil {
		return model.Student{}, fmt.Errorf("commit transaction: %w", err)
	}
	return s, nil
}

// Update memperbarui data mahasiswa. NIM tidak ikut diubah.
// Kalau IPK tidak dikirim, nilai IPK yang lama dipertahankan.
func (r *studentPostgresRepository) Update(ctx context.Context, id int, req model.UpdateStudentRequest) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`UPDATE students
		 SET nama = $1, prodi = $2, angkatan = $3,
		     ipk_terakhir = COALESCE($4::numeric, ipk_terakhir)
		 WHERE id = $5 AND deleted_at IS NULL
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir`,
		req.Nama, req.Prodi, req.Angkatan, req.IPK, id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPK)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("memperbarui mahasiswa: %w", err)
	}
	return s, nil
}

// SoftDelete mengisi deleted_at, barisnya tidak benar-benar dihapus.
func (r *studentPostgresRepository) SoftDelete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("menghapus mahasiswa: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// translateStudentError mengubah error "data kembar" dari PostgreSQL
// menjadi error yang kita kenal. Error lain dibungkus apa adanya.
func translateStudentError(err error, aksi string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "students_nim_key":
			return ErrDuplicateNIM
		case "users_email_key":
			return ErrDuplicateEmail
		}
	}
	return fmt.Errorf("%s: %w", aksi, err)
}