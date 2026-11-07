package service

import (
	"net/mail"

	"latihan-fiber/uts/app/model"
)

// ValidateLogin memeriksa isi body login.
// Hasilnya berupa peta: nama field -> daftar pesan error. Kalau kosong, berarti valid.
func ValidateLogin(req model.LoginRequest) map[string][]string {
	errs := map[string][]string{}

	if req.Email == "" {
		errs["email"] = append(errs["email"], "Email wajib diisi")
	} else if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email {
		// ParseAddress juga menerima bentuk "Nama <a@b.com>", jadi dicocokkan lagi
		errs["email"] = append(errs["email"], "Format email tidak valid")
	}

	if req.Password == "" {
		errs["password"] = append(errs["password"], "Password wajib diisi")
	} else if len(req.Password) < 8 {
		errs["password"] = append(errs["password"], "Password minimal 8 karakter")
	}

	return errs
}