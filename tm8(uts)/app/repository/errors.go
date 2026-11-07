package repository

import (
	"errors"
	"fmt"
)

// ErrNotFound dipakai kalau data yang dicari tidak ada.
var ErrNotFound = errors.New("data tidak ditemukan")

// Dipakai waktu menambah mahasiswa dengan NIM atau email yang sudah terdaftar.
var (
	ErrDuplicateNIM   = errors.New("nim sudah terdaftar")
	ErrDuplicateEmail = errors.New("email sudah terdaftar")
)

// Dipakai waktu mahasiswa mengambil mata kuliah (KRS).
var (
	ErrCourseNotFound  = errors.New("mata kuliah tidak ditemukan")
	ErrAlreadyEnrolled = errors.New("mata kuliah sudah diambil pada tahun akademik ini")
	ErrCourseFull      = errors.New("kuota mata kuliah sudah penuh")
)

// SKSExceededError dikembalikan kalau total SKS akan melebihi batas.
// Dibuat sebagai struct supaya angka-angkanya bisa dipakai untuk pesan error.
type SKSExceededError struct {
	Batas     int // batas SKS mahasiswa
	Terpakai  int // SKS yang sudah diambil
	SKSMatkul int // SKS mata kuliah yang mau diambil
}

func (e *SKSExceededError) Error() string {
	return fmt.Sprintf("total SKS melebihi batas %d (terpakai %d, mata kuliah %d)",
		e.Batas, e.Terpakai, e.SKSMatkul)
}