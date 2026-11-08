package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/uts/app/model"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindStudentProfile(ctx context.Context, userID int) (model.StudentProfile, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

// FindByEmail dipakai saat login.
// Mahasiswa yang sudah di-soft delete tidak ditemukan, jadi tidak bisa login.
func (r *userPostgresRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password, u.role
		 FROM users u
		 LEFT JOIN students s ON s.user_id = u.id
		 WHERE LOWER(u.email) = LOWER($1) AND s.deleted_at IS NULL`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

// FindByID dipakai di /auth/me, aturan soft delete-nya sama dengan FindByEmail.
func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password, u.role
		 FROM users u
		 LEFT JOIN students s ON s.user_id = u.id
		 WHERE u.id = $1 AND s.deleted_at IS NULL`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

// FindStudentProfile mengambil data mahasiswa milik satu user (untuk /auth/me).
func (r *userPostgresRepository) FindStudentProfile(ctx context.Context, userID int) (model.StudentProfile, error) {
	var p model.StudentProfile
	err := r.pool.QueryRow(ctx,
		`SELECT nim, nama, prodi, angkatan FROM students WHERE user_id = $1`, userID,
	).Scan(&p.NIM, &p.Nama, &p.Prodi, &p.Angkatan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.StudentProfile{}, ErrNotFound
		}
		return model.StudentProfile{}, fmt.Errorf("mengambil data mahasiswa: %w", err)
	}
	return p, nil
}