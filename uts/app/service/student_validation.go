package service

import (
	"net/mail"
	"strings"

	"latihan-fiber/uts/app/model"
)

// ValidateCreateStudent memeriksa body POST /students.
// Hasilnya peta: nama field -> daftar pesan error. Kalau kosong, berarti valid.
func ValidateCreateStudent(req model.CreateStudentRequest, tahunSekarang int) map[string][]string {
	errs := map[string][]string{}

	if req.NIM == "" {
		errs["nim"] = append(errs["nim"], "NIM wajib diisi")
	} else if len(req.NIM) != 12 || !isAllDigits(req.NIM) {
		errs["nim"] = append(errs["nim"], "NIM harus 12 digit angka")
	}

	if req.Email == "" {
		errs["email"] = append(errs["email"], "Email wajib diisi")
	} else if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email {
		errs["email"] = append(errs["email"], "Format email tidak valid")
	}

	validateStudentFields(errs, req.Nama, req.Prodi, req.Angkatan, req.IPK, tahunSekarang)
	return errs
}

// ValidateUpdateStudent memeriksa body PUT /students/:id.
func ValidateUpdateStudent(req model.UpdateStudentRequest, tahunSekarang int) map[string][]string {
	errs := map[string][]string{}
	validateStudentFields(errs, req.Nama, req.Prodi, req.Angkatan, req.IPK, tahunSekarang)
	return errs
}

// validateStudentFields memeriksa field yang sama di create dan update.
func validateStudentFields(errs map[string][]string, nama, prodi string, angkatan int, ipk *float64, tahunSekarang int) {
	if strings.TrimSpace(nama) == "" {
		errs["nama"] = append(errs["nama"], "Nama wajib diisi")
	}
	if strings.TrimSpace(prodi) == "" {
		errs["prodi"] = append(errs["prodi"], "Prodi wajib diisi")
	}

	if angkatan == 0 {
		errs["angkatan"] = append(errs["angkatan"], "Angkatan wajib diisi")
	} else if angkatan < 1000 || angkatan > 9999 {
		errs["angkatan"] = append(errs["angkatan"], "Angkatan harus 4 digit")
	} else if angkatan > tahunSekarang {
		errs["angkatan"] = append(errs["angkatan"], "Angkatan tidak boleh melebihi tahun berjalan")
	}

	if ipk != nil && (*ipk < 0 || *ipk > 4) {
		errs["ipk_terakhir"] = append(errs["ipk_terakhir"], "IPK harus antara 0,00 sampai 4,00")
	}
}

// isAllDigits memeriksa apakah teks hanya berisi angka 0-9.
func isAllDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}