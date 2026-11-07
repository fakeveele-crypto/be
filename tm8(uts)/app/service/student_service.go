package service

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/app/model"
	"latihan-fiber/uts/app/repository"
	"latihan-fiber/uts/helper"
)

type StudentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

// serverError mencatat error asli di terminal, lalu mengirim pesan umum ke client
// (supaya detail error tidak bocor ke luar).
func serverError(c *fiber.Ctx, konteks string, err error) error {
	log.Println(konteks+":", err)
	return helper.Fail(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil)
}

// getIDParam mengambil :id dari alamat dan memastikan berupa angka positif.
func getIDParam(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// GetStudents: GET /api/v1/students (admin)
func (s *StudentService) GetStudents(c *fiber.Ctx) error {
	errs := map[string][]string{}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		errs["page"] = append(errs["page"], "Page harus berupa angka positif")
	}

	perPage, err := strconv.Atoi(c.Query("per_page", "10"))
	if err != nil || perPage < 1 {
		errs["per_page"] = append(errs["per_page"], "Per_page harus berupa angka positif")
	} else if perPage > 50 {
		perPage = 50 // batas maksimal sesuai soal
	}

	angkatan := 0
	if raw := c.Query("angkatan"); raw != "" {
		angkatan, err = strconv.Atoi(raw)
		if err != nil || angkatan < 1 {
			errs["angkatan"] = append(errs["angkatan"], "Angkatan harus berupa angka")
		}
	}

	sortBy := c.Query("sort")
	if sortBy != "" && sortBy != "nama" && sortBy != "-ipk_terakhir" {
		errs["sort"] = append(errs["sort"], "Sort hanya boleh nama atau -ipk_terakhir")
	}

	if len(errs) > 0 {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errs)
	}

	filter := model.StudentFilter{
		Page:     page,
		PerPage:  perPage,
		Prodi:    strings.TrimSpace(c.Query("prodi")),
		Angkatan: angkatan,
		Search:   strings.TrimSpace(c.Query("search")),
		Sort:     sortBy,
	}

	students, total, err := s.repo.FindAll(c.UserContext(), filter)
	if err != nil {
		return serverError(c, "gagal mengambil daftar mahasiswa", err)
	}

	lastPage := (total + perPage - 1) / perPage
	if lastPage < 1 {
		lastPage = 1
	}
	meta := &model.Meta{CurrentPage: page, PerPage: perPage, Total: total, LastPage: lastPage}

	return helper.Ok(c, "Data mahasiswa berhasil diambil", students, meta)
}

// CreateStudent: POST /api/v1/students (admin)
func (s *StudentService) CreateStudent(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Format data tidak valid", nil)
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Nama = strings.TrimSpace(req.Nama)
	req.Email = strings.TrimSpace(req.Email)
	req.Prodi = strings.TrimSpace(req.Prodi)

	if errs := ValidateCreateStudent(req, time.Now().Year()); len(errs) > 0 {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errs)
	}

	// password awal mahasiswa = NIM-nya, disimpan dalam bentuk hash
	hash, err := helper.HashPassword(req.NIM)
	if err != nil {
		return serverError(c, "gagal hash password", err)
	}

	student, err := s.repo.Create(c.UserContext(), req, hash)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateNIM) {
			return helper.Fail(c, fiber.StatusUnprocessableEntity, "Validasi gagal",
				map[string][]string{"nim": {"NIM sudah terdaftar"}})
		}
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return helper.Fail(c, fiber.StatusUnprocessableEntity, "Validasi gagal",
				map[string][]string{"email": {"Email sudah terdaftar"}})
		}
		return serverError(c, "gagal menambah mahasiswa", err)
	}

	return helper.Created(c, "Mahasiswa berhasil ditambahkan", student)
}

// GetStudentByID: GET /api/v1/students/:id (admin, atau mahasiswa untuk data sendiri)
func (s *StudentService) GetStudentByID(c *fiber.Ctx) error {
	id, ok := getIDParam(c)
	if !ok {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan", nil)
	}

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
	}

	// cari dulu (404 kalau tidak ada), baru cek hak akses (403)
	student, err := s.repo.FindByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan", nil)
		}
		return serverError(c, "gagal mengambil mahasiswa", err)
	}

	if !CanAccessStudent(authUser.Role, authUser.UserID, student.UserID) {
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak boleh mengakses data mahasiswa lain", nil)
	}

	courses, err := s.repo.FindEnrollments(c.UserContext(), student.ID)
	if err != nil {
		return serverError(c, "gagal mengambil KRS", err)
	}

	totalSKS := 0
	for _, mk := range courses {
		totalSKS += mk.SKS
	}

	detail := model.StudentDetail{
		Student:    student,
		MataKuliah: courses,
		TotalSKS:   totalSKS,
		BatasSKS:   BatasSKS(student.IPK),
	}
	return helper.Ok(c, "Detail mahasiswa berhasil diambil", detail, nil)
}

// UpdateStudent: PUT /api/v1/students/:id (admin)
func (s *StudentService) UpdateStudent(c *fiber.Ctx) error {
	id, ok := getIDParam(c)
	if !ok {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan", nil)
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Format data tidak valid", nil)
	}
	req.Nama = strings.TrimSpace(req.Nama)
	req.Prodi = strings.TrimSpace(req.Prodi)

	if errs := ValidateUpdateStudent(req, time.Now().Year()); len(errs) > 0 {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errs)
	}

	student, err := s.repo.Update(c.UserContext(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan", nil)
		}
		return serverError(c, "gagal memperbarui mahasiswa", err)
	}

	return helper.Ok(c, "Data mahasiswa berhasil diperbarui", student, nil)
}

// DeleteStudent: DELETE /api/v1/students/:id (admin, soft delete)
func (s *StudentService) DeleteStudent(c *fiber.Ctx) error {
	id, ok := getIDParam(c)
	if !ok {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan", nil)
	}

	if err := s.repo.SoftDelete(c.UserContext(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan", nil)
		}
		return serverError(c, "gagal menghapus mahasiswa", err)
	}

	return helper.NoContent(c)
}