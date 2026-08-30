package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var allowedSort = map[string]bool{
	"id":    true,
	"name":  true,
	"grade": true,
}

var students = []Student{
	{ID: 1, NIM: "434241088", Name: "Valerina", Grade: 3.8, IsActive: true},
	{ID: 2, NIM: "434241002", Name: "Alisya", Grade: 3.5, IsActive: true},
	{ID: 3, NIM: "434241003", Name: "Lailia", Grade: 3.6, IsActive: true},
}
var nextID = 4

// requireJSON menolak request POST/PUT/PATCH yang Content-Type-nya bukan JSON (415).
func requireJSON(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut || c.Method() == fiber.MethodPatch {
		if !strings.HasPrefix(c.Get("Content-Type"), fiber.MIMEApplicationJSON) {
			return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json", nil)
		}
	}
	return c.Next()
}

func getStudents(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	sortField := c.Query("sort", "id")
	order := strings.ToLower(c.Query("order", "asc"))

	var isActiveFilter *bool
	if raw := c.Query("is_active", ""); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			isActiveFilter = &v
		}
	}

	minGrade, hasMinGrade := 0.0, false
	if raw := c.Query("min_grade", ""); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			minGrade = v
			hasMinGrade = true
		}
	}

	maxGrade, hasMaxGrade := 0.0, false
	if raw := c.Query("max_grade", ""); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			maxGrade = v
			hasMaxGrade = true
		}
	}

	if !allowedSort[sortField] {
		sortField = "id"
	}
	if order != "desc" {
		order = "asc"
	}
	if limit > 100 {
		limit = 100
	}

	filtered := []Student{}
	for _, s := range students {
		if search != "" && !strings.Contains(strings.ToLower(s.Name), strings.ToLower(search)) {
			continue
		}
		if isActiveFilter != nil && s.IsActive != *isActiveFilter {
			continue
		}
		if hasMinGrade && s.Grade < minGrade {
			continue
		}
		if hasMaxGrade && s.Grade > maxGrade {
			continue
		}
		filtered = append(filtered, s)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		var less bool
		switch sortField {
		case "name":
			less = filtered[i].Name < filtered[j].Name
		case "grade":
			less = filtered[i].Grade < filtered[j].Grade
		default: // "id"
			less = filtered[i].ID < filtered[j].ID
		}
		if order == "desc" {
			return !less
		}
		return less
	})

	start := (page - 1) * limit
	end := start + limit

	var result []Student
	if start < len(filtered) {
		if end > len(filtered) {
			end = len(filtered)
		}
		result = filtered[start:end]
	} else {
		result = []Student{}
	}

	meta := &Meta{
		Page:       page,
		Limit:      limit,
		Total:      len(filtered),
		TotalPages: (len(filtered) + limit - 1) / limit,
	}

	return ok(c, "daftar student berhasil diambil", result, meta)
}

func getStudentByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka", nil)
	}

	for _, s := range students {
		if s.ID == id {
			return ok(c, "student berhasil ditemukan", s, nil)
		}
	}

	return fail(c, fiber.StatusNotFound, "student tidak ditemukan", nil)
}

func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "format body salah", err.Error())
	}

	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 4 {
		errs["grade"] = "harus di antara 0 dan 4"
	}
	if len(errs) > 0 {
		return fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	for _, s := range students {
		if s.NIM == req.NIM {
			return fail(c, fiber.StatusConflict, "NIM sudah terdaftar", nil)
		}
	}

	newStudent := Student{
		ID:       nextID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
	}

	students = append(students, newStudent)
	nextID++

	location := fmt.Sprintf("/api/v1/students/%d", newStudent.ID)

	return created(c, "student berhasil dibuat", newStudent, location)
}

func updateStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka", nil)
	}

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "format body salah", err.Error())
	}

	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 4 {
		errs["grade"] = "harus di antara 0 dan 4"
	}
	if len(errs) > 0 {
		return fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	for _, other := range students {
		if other.ID != id && other.NIM == req.NIM {
			return fail(c, fiber.StatusConflict, "NIM sudah dipakai student lain", nil)
		}
	}

	for i, s := range students {
		if s.ID == id {
			students[i].NIM = req.NIM
			students[i].Name = req.Name
			students[i].Grade = req.Grade
			students[i].IsActive = req.IsActive
			return ok(c, "student berhasil diperbarui total", students[i], nil)
		}
	}
	return fail(c, fiber.StatusNotFound, "student tidak ditemukan", nil)
}

func patchStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka", nil)
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "format body salah", err.Error())
	}

	errs := map[string]string{}
	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			errs["nim"] = "tidak boleh kosong"
		} else {
			for _, other := range students {
				if other.ID != id && other.NIM == *req.NIM {
					return fail(c, fiber.StatusConflict, "NIM sudah dipakai student lain", nil)
				}
			}
		}
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		errs["name"] = "tidak boleh kosong"
	}
	if req.Grade != nil && (*req.Grade < 0 || *req.Grade > 4) {
		errs["grade"] = "harus di antara 0 dan 4"
	}
	if len(errs) > 0 {
		return fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	for i, s := range students {
		if s.ID == id {
			if req.NIM != nil {
				students[i].NIM = *req.NIM
			}
			if req.Name != nil {
				students[i].Name = *req.Name
			}
			if req.Grade != nil {
				students[i].Grade = *req.Grade
			}
			if req.IsActive != nil {
				students[i].IsActive = *req.IsActive
			}
			return ok(c, "student berhasil diperbarui sebagian", students[i], nil)
		}
	}
	return fail(c, fiber.StatusNotFound, "student tidak ditemukan", nil)
}

func deleteStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka", nil)
	}

	for i, s := range students {
		if s.ID == id {
			students = append(students[:i], students[i+1:]...)
			return noContent(c)
		}
	}
	return fail(c, fiber.StatusNotFound, "student tidak ditemukan", nil)
}