package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/app/model"
	"latihan-fiber/uts/app/repository"
	"latihan-fiber/uts/helper"
)

type EnrollmentService struct {
	repo repository.EnrollmentRepository
}

func NewEnrollmentService(repo repository.EnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{repo: repo}
}

// CreateEnrollment: POST /api/v1/enrollments (mahasiswa)
func (s *EnrollmentService) CreateEnrollment(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Format data tidak valid", nil)
	}
	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)

	if errs := ValidateEnrollment(req); len(errs) > 0 {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errs)
	}

	enrollment, err := s.repo.Enroll(c.UserContext(), authUser.UserID, req.CourseID, req.TahunAkademik, BatasSKS)
	if err != nil {
		var sksErr *repository.SKSExceededError

		switch {
		case errors.Is(err, repository.ErrCourseNotFound):
			return helper.Fail(c, fiber.StatusUnprocessableEntity, "Validasi gagal",
				map[string][]string{"course_id": {"Mata kuliah tidak ditemukan"}})
		case errors.Is(err, repository.ErrAlreadyEnrolled):
			return helper.Fail(c, fiber.StatusConflict,
				"Mata kuliah ini sudah diambil pada tahun akademik yang sama", nil)
		case errors.Is(err, repository.ErrCourseFull):
			return helper.Fail(c, fiber.StatusUnprocessableEntity, "Kuota mata kuliah sudah penuh", nil)
		case errors.As(err, &sksErr):
			sisa := sksErr.Batas - sksErr.Terpakai
			if sisa < 0 {
				sisa = 0
			}
			pesan := fmt.Sprintf(
				"Total SKS melebihi batas. Sisa SKS Anda %d (batas %d), sedangkan mata kuliah ini %d SKS",
				sisa, sksErr.Batas, sksErr.SKSMatkul)
			return helper.Fail(c, fiber.StatusUnprocessableEntity, pesan, nil)
		case errors.Is(err, repository.ErrNotFound):
			return helper.Fail(c, fiber.StatusForbidden, "Akun mahasiswa tidak aktif", nil)
		default:
			return serverError(c, "gagal mengambil mata kuliah", err)
		}
	}

	return helper.Created(c, "Mata kuliah berhasil ditambahkan ke KRS", enrollment)
}

// DeleteEnrollment: DELETE /api/v1/enrollments/:id (mahasiswa, hanya milik sendiri)
func (s *EnrollmentService) DeleteEnrollment(c *fiber.Ctx) error {
	id, ok := getIDParam(c)
	if !ok {
		return helper.Fail(c, fiber.StatusNotFound, "Data KRS tidak ditemukan", nil)
	}

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
	}

	// cari dulu pemiliknya (404 kalau tidak ada), baru cek hak akses (403)
	ownerID, err := s.repo.FindOwnerUserID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Data KRS tidak ditemukan", nil)
		}
		return serverError(c, "gagal mencari KRS", err)
	}

	if !CanAccessStudent(authUser.Role, authUser.UserID, ownerID) {
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak boleh menghapus KRS milik mahasiswa lain", nil)
	}

	if err := s.repo.Delete(c.UserContext(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Data KRS tidak ditemukan", nil)
		}
		return serverError(c, "gagal menghapus KRS", err)
	}

	return helper.NoContent(c)
}