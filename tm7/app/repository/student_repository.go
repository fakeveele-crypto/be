package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/tm7/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

type StudentRepository interface {
	FindAll(ctx context.Context, search string, page, limit int) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Student, error)
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

func (r *studentPostgresRepository) FindAll(ctx context.Context, search string, page, limit int) ([]model.Student, int, error) {
	where := "WHERE 1=1"
	args := []any{}

	if search != "" {
		where += fmt.Sprintf(" AND LOWER(name) ILIKE $%d", len(args)+1)
		args = append(args, "%"+search+"%")
	}

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total data: %w", err)
	}

	offset := (page - 1) * limit
	sqlText := fmt.Sprintf(`
		SELECT id, nim, name, grade, is_active, owner_id, created_at 
		FROM students %s
		ORDER BY id ASC 
		LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	
	args = append(args, limit, offset)
	
	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.Nim, &s.Name, &s.Grade, &s.IsActive, &s.OwnerID, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal membaca baris: %w", err)
		}
		hasil = append(hasil, s)
	}
	return hasil, total, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx, `
		SELECT id, nim, name, grade, is_active, owner_id, created_at 
		FROM students WHERE id = $1`, id).
		Scan(&s.ID, &s.Nim, &s.Name, &s.Grade, &s.IsActive, &s.OwnerID, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("gagal mencari by ID: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Create(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO students (nim, name, grade, is_active, owner_id) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, created_at`, 
		s.Nim, s.Name, s.Grade, s.IsActive, s.OwnerID).
		Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("gagal menambah data: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE students 
		SET nim = $1, name = $2, grade = $3, is_active = $4 
		WHERE id = $5 
		RETURNING id, nim, name, grade, is_active, created_at`,
		s.Nim, s.Name, s.Grade, s.IsActive, s.ID).
		Scan(&s.ID, &s.Nim, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("gagal mengubah data: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM students WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("gagal menghapus data: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *studentPostgresRepository) FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Student, error) {
	args := []any{}
	where := " WHERE 1 = 1"

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND name ILIKE $%d", len(args))
	}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND is_active = $%d", len(args))
	}
	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)-1, len(args))
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(`
		SELECT id, nim, name, grade, is_active, owner_id, created_at
		FROM students%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d`, where, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar students: %w", err)
	}
	defer rows.Close()

	result := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.Nim, &s.Name, &s.Grade, &s.IsActive, &s.OwnerID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca row student: %w", err)
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query: %w", err)
	}
	return result, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}