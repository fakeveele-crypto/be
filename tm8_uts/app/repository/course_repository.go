package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/uts/app/model"
)

type CourseRepository interface {
	FindAll(ctx context.Context, semester int, search string, onlyAvailable bool) ([]model.Course, error)
}

type coursePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &coursePostgresRepository{pool: pool}
}

// FindAll mengambil daftar mata kuliah beserta jumlah yang sudah mengambil (terisi).
func (r *coursePostgresRepository) FindAll(ctx context.Context, semester int, search string, onlyAvailable bool) ([]model.Course, error) {
	where := "WHERE 1=1"
	args := []any{}

	if semester > 0 {
		where += fmt.Sprintf(" AND c.semester = $%d", len(args)+1)
		args = append(args, semester)
	}
	if search != "" {
		where += fmt.Sprintf(" AND (c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", len(args)+1, len(args)+1)
		args = append(args, "%"+search+"%")
	}

	// available=true: hanya yang jumlah pengambilnya masih di bawah kuota
	having := ""
	if onlyAvailable {
		having = "HAVING COUNT(e.id) < c.kuota"
	}

	sqlText := fmt.Sprintf(`
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, COUNT(e.id)
		FROM courses c
		LEFT JOIN enrollments e ON e.course_id = c.id
		%s
		GROUP BY c.id
		%s
		ORDER BY c.id ASC`, where, having)

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil mata kuliah: %w", err)
	}
	defer rows.Close()

	hasil := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi); err != nil {
			return nil, fmt.Errorf("gagal membaca baris: %w", err)
		}
		c.SisaKuota = c.Kuota - c.Terisi
		hasil = append(hasil, c)
	}
	return hasil, nil
}