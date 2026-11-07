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

type EnrollmentRepository interface {
	Enroll(ctx context.Context, userID, courseID int, tahun string, batasSKS func(ipk float64) int) (model.Enrollment, error)
	FindOwnerUserID(ctx context.Context, enrollmentID int) (int, error)
	Delete(ctx context.Context, enrollmentID int) error
}

type enrollmentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentPostgresRepository{pool: pool}
}

// Enroll mengambil satu mata kuliah untuk mahasiswa yang sedang login.
// Semua pengecekan dan penyimpanan dilakukan dalam satu transaction.
// Aturan hitung batas SKS dititipkan lewat parameter batasSKS (diisi oleh service).
func (r *enrollmentPostgresRepository) Enroll(ctx context.Context, userID, courseID int, tahun string, batasSKS func(ipk float64) int) (model.Enrollment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("memulai transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. cari data mahasiswa milik user yang login
	var studentID int
	var ipk float64
	err = tx.QueryRow(ctx,
		`SELECT id, ipk_terakhir FROM students WHERE user_id = $1 AND deleted_at IS NULL`, userID,
	).Scan(&studentID, &ipk)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("mengambil mahasiswa: %w", err)
	}

	// 2. ambil data mata kuliah sambil mengunci barisnya (FOR UPDATE)
	var sks, kuota int
	err = tx.QueryRow(ctx,
		`SELECT sks, kuota FROM courses WHERE id = $1 FOR UPDATE`, courseID,
	).Scan(&sks, &kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrCourseNotFound
		}
		return model.Enrollment{}, fmt.Errorf("mengambil mata kuliah: %w", err)
	}

	// 3. cek apakah sudah pernah diambil pada tahun akademik yang sama
	var sudah bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM enrollments
			WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3)`,
		studentID, courseID, tahun,
	).Scan(&sudah)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("cek duplikat: %w", err)
	}
	if sudah {
		return model.Enrollment{}, ErrAlreadyEnrolled
	}

	// 4. cek kuota
	var terisi int
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM enrollments WHERE course_id = $1`, courseID,
	).Scan(&terisi)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("cek kuota: %w", err)
	}
	if terisi >= kuota {
		return model.Enrollment{}, ErrCourseFull
	}

	// 5. cek batas SKS pada tahun akademik ini
	var terpakai int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahun,
	).Scan(&terpakai)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("cek SKS: %w", err)
	}
	batas := batasSKS(ipk)
	if terpakai+sks > batas {
		return model.Enrollment{}, &SKSExceededError{Batas: batas, Terpakai: terpakai, SKSMatkul: sks}
	}

	// 6. semua aturan lolos, simpan
	e := model.Enrollment{CourseID: courseID, TahunAkademik: tahun}
	err = tx.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3) RETURNING id, created_at`,
		studentID, courseID, tahun,
	).Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		// jaga-jaga kalau dua permintaan yang sama masuk bersamaan
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Enrollment{}, ErrAlreadyEnrolled
		}
		return model.Enrollment{}, fmt.Errorf("menyimpan KRS: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Enrollment{}, fmt.Errorf("commit transaction: %w", err)
	}
	return e, nil
}

// FindOwnerUserID mencari id akun (users.id) pemilik sebuah baris KRS.
func (r *enrollmentPostgresRepository) FindOwnerUserID(ctx context.Context, enrollmentID int) (int, error) {
	var userID int
	err := r.pool.QueryRow(ctx,
		`SELECT s.user_id
		 FROM enrollments e
		 JOIN students s ON s.id = e.student_id
		 WHERE e.id = $1`, enrollmentID,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("mengambil pemilik KRS: %w", err)
	}
	return userID, nil
}

// Delete menghapus satu baris KRS. Kuota otomatis kembali karena dihitung dari tabel enrollments.
func (r *enrollmentPostgresRepository) Delete(ctx context.Context, enrollmentID int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1`, enrollmentID)
	if err != nil {
		return fmt.Errorf("menghapus KRS: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}