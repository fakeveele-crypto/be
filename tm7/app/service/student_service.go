package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm7/app/model"
	"latihan-fiber/tm7/app/repository"
	"latihan-fiber/tm7/helper"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

// 1. GetStudents memakai cursor pagination.
func (s *StudentService) GetStudents(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	// Baris tambahan hasil limit+1 dipotong di sini. Ia hanya penanda bahwa
	// masih ada halaman berikutnya, bukan bagian dari halaman ini.
	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar student berhasil diambil", rows, meta)
}

// 2. GetStudentByID 
func (s *StudentService) GetStudentByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	student, err := s.repo.FindByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	current, _ := helper.CurrentUser(c)
	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:read:any") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "tidak berhak mengakses data ini"})
	}

	return c.JSON(model.WebResponse{Success: true, Data: student})
}

// translateStudentError mengubah error repository menjadi AppError.
func translateStudentError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("NIM sudah terdaftar")
	default:
		return helper.Internal(err)
	}
}

// 3. CreateStudent
func (s *StudentService) CreateStudent(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	student, err := s.repo.Create(c.UserContext(), model.Student{
		Nim: req.NIM, Name: req.Name, Grade: req.Grade,
		IsActive: true, OwnerID: current.UserID,
	})
	if err != nil {
		return translateStudentError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{Success: true, Data: student})
}

// 4. UpdateStudent (PUT: mengganti seluruh data)
func (s *StudentService) UpdateStudent(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return helper.BadRequest("id harus berupa angka")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	existing, err := s.repo.FindByID(c.UserContext(), id)
	if err != nil {
		return translateStudentError(err)
	}

	current, _ := helper.CurrentUser(c)
	if !CanAccessStudent(current, existing.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data ini")
	}

	student, err := s.repo.Update(c.UserContext(), model.Student{
		ID: id, Nim: req.NIM, Name: req.Name, Grade: req.Grade,
		IsActive: req.IsActive, OwnerID: existing.OwnerID,
	})
	if err != nil {
		return translateStudentError(err)
	}

	return c.JSON(model.WebResponse{Success: true, Data: student})
}

// PatchStudent (PATCH: hanya mengubah field yang dikirim)
func (s *StudentService) PatchStudent(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return helper.BadRequest("id harus berupa angka")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if req.NIM != nil {
		v := strings.TrimSpace(*req.NIM)
		req.NIM = &v
	}
	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		req.Name = &v
	}

	if IsEmptyPatchStudent(req) {
		return helper.BadRequest("body PATCH harus memuat setidaknya satu field")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	existing, err := s.repo.FindByID(c.UserContext(), id)
	if err != nil {
		return translateStudentError(err)
	}

	current, _ := helper.CurrentUser(c)
	if !CanAccessStudent(current, existing.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data ini")
	}

	student, err := s.repo.Update(c.UserContext(), ApplyPatchStudent(existing, req))
	if err != nil {
		return translateStudentError(err)
	}

	return c.JSON(model.WebResponse{Success: true, Data: student})
}

// 5. DeleteStudent 
func (s *StudentService) DeleteStudent(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	err = s.repo.Delete(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}