package main

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm3/app/model"
	"latihan-fiber/tm3/app/repository"
)

// 1. Bikin struct untuk nyimpen repository
type StudentHandler struct {
	repo repository.StudentRepository
}

func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{repo: repo}
}

// 2. Semua fungsi diubah jadi method milik StudentHandler
func (h *StudentHandler) getStudents(c *fiber.Ctx) error {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 { page = 1 }
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 { limit = 10 }

	// Panggil FindAll dari repository
	students, total, err := h.repo.FindAll(c.UserContext(), search, page, limit)
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

func (h *StudentHandler) getStudentByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	// Panggil FindByID dari repository
	student, err := h.repo.FindByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()}) // Status 404
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()}) // Status 500
	}

	return c.JSON(model.WebResponse{Data: student})
}

func (h *StudentHandler) createStudent(c *fiber.Ctx) error {
	var req model.Student
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	// Panggil Create dari repository
	student, err := h.repo.Create(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()}) // Status 409
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{Data: student})
}

func (h *StudentHandler) updateStudent(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	var req model.Student
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}
	req.ID = id

	// Panggil Update dari repository
	student, err := h.repo.Update(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()}) // Status 404
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()}) // Status 409
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(model.WebResponse{Data: student})
}

func (h *StudentHandler) deleteStudent(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	// Panggil Delete dari repository
	err = h.repo.Delete(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()}) // Status 404
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}