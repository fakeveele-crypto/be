package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm10/app/model"
	"latihan-fiber/tm10/app/repository"
	"latihan-fiber/tm10/helper"
)

// topTechnologiesLimit adalah jumlah teknologi pada statistik.
const topTechnologiesLimit = 10

// PortfolioService, seperti ReviewService, memakai dua database: data
// mahasiswa dibaca dari PostgreSQL, portofolionya disimpan di MongoDB.
type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
	perms      *helper.PermissionSet
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
	perms *helper.PermissionSet,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students, perms: perms}
}

// authorize memastikan mahasiswanya ada di PostgreSQL (MongoDB tidak punya
// FOREIGN KEY, jadi service yang harus memeriksa) lalu memeriksa hak akses
// dengan aturan yang SAMA seperti data student: pemilik selalu boleh, selain
// itu butuh permission ":any" yang sesuai.
func (s *PortfolioService) authorize(c *fiber.Ctx, anyPermission, denyMessage string) (int, error) {
	studentID, valid := helper.ParamID(c)
	if !valid {
		return 0, helper.BadRequest("id student harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.students.FindByID(ctx, studentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, helper.NotFound("student tidak ditemukan")
		}
		return 0, helper.Internal(err)
	}
	current, ok := helper.CurrentUser(c)
	if !ok {
		return 0, helper.Unauthorized("belum terautentikasi")
	}
	if !CanAccessStudent(current, student.OwnerID, s.perms, anyPermission) {
		return 0, helper.Forbidden(denyMessage)
	}
	return studentID, nil
}

// ---------- GET /students/:id/portfolio ----------

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	studentID, err := s.authorize(c, "student:read:any", "tidak berhak melihat portofolio ini")
	if err != nil {
		return err
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	p, err := s.portfolios.FindByStudent(ctx, studentID)
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusOK, "portofolio ditemukan", p)
}

// ---------- PUT /students/:id/portfolio ----------

func (s *PortfolioService) Replace(c *fiber.Ctx) error {
	studentID, err := s.authorize(c, "student:update:any", "tidak berhak mengubah portofolio ini")
	if err != nil {
		return err
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Body didekode ke struct bertipe tetap, bukan map[string]any. Itulah
	// pertahanan terhadap NoSQL injection: {"headline": {"$ne": ""}} ditolak
	// dengan 400 sebelum menyentuh MongoDB.
	var req model.ReplacePortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)
	for i := range req.Projects {
		normalizeProject(&req.Projects[i])
	}
	for i := range req.Certificates {
		req.Certificates[i].Name = strings.TrimSpace(req.Certificates[i].Name)
		req.Certificates[i].Issuer = strings.TrimSpace(req.Certificates[i].Issuer)
		req.Certificates[i].Date = strings.TrimSpace(req.Certificates[i].Date)
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	p, created, err := s.portfolios.Replace(ctx, studentID, req)
	if err != nil {
		return translatePortfolioError(err)
	}
	if created {
		return helper.Created(c, "portofolio berhasil dibuat", p,
			"/api/v1/students/"+strconv.Itoa(studentID)+"/portfolio")
	}
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diganti", p)
}

// ---------- POST /students/:id/portfolio/projects ----------

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	studentID, err := s.authorize(c, "student:update:any", "tidak berhak mengubah portofolio ini")
	if err != nil {
		return err
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.ProjectInput
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	normalizeProject(&req)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	p, err := s.portfolios.AddProject(ctx, studentID, req)
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan", p)
}

// ---------- DELETE /students/:id/portfolio/projects/:projectId ----------

func (s *PortfolioService) RemoveProject(c *fiber.Ctx) error {
	studentID, err := s.authorize(c, "student:update:any", "tidak berhak mengubah portofolio ini")
	if err != nil {
		return err
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.portfolios.RemoveProject(ctx, studentID, c.Params("projectId")); err != nil {
		return translatePortfolioError(err)
	}
	return helper.NoContent(c)
}

// ---------- GET /portfolios/stats/technologies ----------

func (s *PortfolioService) TopTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	top, err := s.portfolios.TopTechnologies(ctx, topTechnologiesLimit)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", top)
}

// normalizeProject merapikan input proyek sebelum divalidasi. Teknologi
// dinormalkan seperti tag ulasan: "Go", "go", dan " GO " dihitung sebagai
// satu teknologi pada statistik.
func normalizeProject(p *model.ProjectInput) {
	p.Title = strings.TrimSpace(p.Title)
	p.URL = strings.TrimSpace(p.URL)
	p.Technologies = NormalizeTags(p.Technologies)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("id proyek tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NewError(fiber.StatusNotFound, "PORTFOLIO_NOT_FOUND",
			"portofolio belum dibuat, buat dulu dengan PUT /students/:id/portfolio")
	case errors.Is(err, repository.ErrProjectNotFound):
		return helper.NewError(fiber.StatusNotFound, "PROJECT_NOT_FOUND", "proyek tidak ditemukan")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT",
			"jumlah proyek sudah mencapai batas "+strconv.Itoa(model.MaxProjectsPerPortfolio))
	default:
		return helper.Internal(err)
	}
}
