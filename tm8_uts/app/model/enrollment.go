package model

import "time"

// Enrollment adalah satu baris KRS: mahasiswa mengambil satu mata kuliah.
type Enrollment struct {
	ID            int       `json:"id"`
	CourseID      int       `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateEnrollmentRequest adalah body POST /enrollments.
type CreateEnrollmentRequest struct {
	CourseID      int    `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}