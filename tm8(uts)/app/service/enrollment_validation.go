package service

import (
	"strconv"
	"strings"

	"latihan-fiber/uts/app/model"
)

// ValidateEnrollment memeriksa body POST /enrollments.
func ValidateEnrollment(req model.CreateEnrollmentRequest) map[string][]string {
	errs := map[string][]string{}

	if req.CourseID < 1 {
		errs["course_id"] = append(errs["course_id"], "course_id wajib diisi")
	}

	if req.TahunAkademik == "" {
		errs["tahun_akademik"] = append(errs["tahun_akademik"], "Tahun akademik wajib diisi")
	} else if !isTahunAkademik(req.TahunAkademik) {
		errs["tahun_akademik"] = append(errs["tahun_akademik"],
			"Format tahun akademik harus seperti 2026/2027-Ganjil atau 2026/2027-Genap")
	}

	return errs
}

// isTahunAkademik memeriksa bentuk 2026/2027-Ganjil (tahun kedua = tahun pertama + 1).
func isTahunAkademik(s string) bool {
	bagian := strings.Split(s, "-")
	if len(bagian) != 2 {
		return false
	}
	if bagian[1] != "Ganjil" && bagian[1] != "Genap" {
		return false
	}

	tahun := strings.Split(bagian[0], "/")
	if len(tahun) != 2 || len(tahun[0]) != 4 || len(tahun[1]) != 4 {
		return false
	}
	awal, err1 := strconv.Atoi(tahun[0])
	akhir, err2 := strconv.Atoi(tahun[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return akhir == awal+1
}