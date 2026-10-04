package repository

import "errors"

// ErrNotFound dipakai kalau data yang dicari tidak ada.
var ErrNotFound = errors.New("data tidak ditemukan")