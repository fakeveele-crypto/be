package model

// User adalah satu baris tabel users.
// Password diberi tanda "-" supaya tidak pernah ikut terkirim di JSON.
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

// LoginRequest adalah isi body endpoint login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse adalah data yang dikirim balik setelah login berhasil.
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	User        User   `json:"user"`
}

// AuthUser adalah identitas pemakai yang sudah lolos pengecekan token.
type AuthUser struct {
	UserID int
	Role   string
}

// StudentProfile adalah data singkat mahasiswa untuk endpoint /auth/me.
type StudentProfile struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}

// MeResponse adalah isi response endpoint /auth/me.
// Student hanya terisi kalau role-nya mahasiswa.
type MeResponse struct {
	ID      int             `json:"id"`
	Email   string          `json:"email"`
	Role    string          `json:"role"`
	Student *StudentProfile `json:"student,omitempty"`
}