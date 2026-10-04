package repository

import "errors"

// ErrNotFound dipakai kalau data yang dicari tidak ada.
var ErrNotFound = errors.New("data tidak ditemukan")

// Dipakai waktu menambah mahasiswa dengan NIM atau email yang sudah terdaftar.
var (
	ErrDuplicateNIM   = errors.New("nim sudah terdaftar")
	ErrDuplicateEmail = errors.New("email sudah terdaftar")
)