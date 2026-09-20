package service

import (
	"strings"
	"unicode"

	"latihan-fiber/tm6/app/model"
)

const minPasswordLength = 8

// ValidateRegister memeriksa kelengkapan dan kekuatan data pendaftaran.
func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := map[string]string{}

	username := strings.TrimSpace(req.Username)
	switch {
	case username == "":
		errs["username"] = "wajib diisi"
	case len(username) < 3:
		errs["username"] = "minimal 3 karakter"
	}

	if !strings.Contains(req.Email, "@") {
		errs["email"] = "format email tidak valid"
	}

	if msg := checkPasswordStrength(req.Password); msg != "" {
		errs["password"] = msg
	}

	return errs
}

// checkPasswordStrength memeriksa panjang dan kombinasi karakter password.
func checkPasswordStrength(password string) string {
	if len(password) < minPasswordLength {
		return "minimal 8 karakter"
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	weak := map[string]bool{
		"password1": true, "12345678": true, "qwerty123": true,
	}
	if weak[strings.ToLower(password)] {
		return "password terlalu umum"
	}

	return ""
}