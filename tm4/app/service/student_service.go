package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm4/app/model"
	"latihan-fiber/tm4/app/repository"
)

type StudentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

// 1. GetStudents 
func (s *StudentService) GetStudents(c *fiber.Ctx) error {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 {
		limit = 10
	}

	students, total, err := s.repo.FindAll(c.UserContext(), search, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	return c.JSON(model.WebResponse{
		Data: students,
		Meta: &model.Meta{
			Search: search, Page: page, Limit: limit, Total: total, TotalPages: totalPages,
		},
	})
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

	return c.JSON(model.WebResponse{Data: student})
}

// 3. CreateStudent 
func (s *StudentService) CreateStudent(c *fiber.Ctx) error {
	var req model.Student
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	student, err := s.repo.Create(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{Data: student})
}

// 4. UpdateStudent 
func (s *StudentService) UpdateStudent(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	var req model.Student
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}
	req.ID = id

	student, err := s.repo.Update(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(model.WebResponse{Data: student})
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