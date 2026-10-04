package service

import (
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/app/repository"
	"latihan-fiber/uts/helper"
)

type CourseService struct {
	repo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

// GetCourses: GET /api/v1/courses
func (s *CourseService) GetCourses(c *fiber.Ctx) error {
	search := strings.TrimSpace(c.Query("search"))
	onlyAvailable := c.Query("available") == "true"

	// semester boleh kosong, tapi kalau diisi harus angka
	semester := 0
	if raw := c.Query("semester"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return helper.Fail(c, fiber.StatusUnprocessableEntity, "Validasi gagal", map[string][]string{
				"semester": {"Semester harus berupa angka positif"},
			})
		}
		semester = n
	}

	courses, err := s.repo.FindAll(c.UserContext(), semester, search, onlyAvailable)
	if err != nil {
		log.Println("gagal mengambil mata kuliah:", err)
		return helper.Fail(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil)
	}

	return helper.Ok(c, "Data mata kuliah berhasil diambil", courses, nil)
}