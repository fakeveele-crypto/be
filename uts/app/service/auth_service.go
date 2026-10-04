package service

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/app/model"
	"latihan-fiber/uts/app/repository"
	"latihan-fiber/uts/helper"
)

type AuthService struct {
	users repository.UserRepository
	jwt   *helper.JWTManager
}

func NewAuthService(users repository.UserRepository, jwtManager *helper.JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwtManager}
}

// Login: POST /api/v1/auth/login
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Format data tidak valid", nil)
	}
	req.Email = strings.TrimSpace(req.Email)

	// 1. cek isi body, kalau salah kirim 422
	if errs := ValidateLogin(req); len(errs) > 0 {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errs)
	}

	// 2. cari user berdasarkan email
	user, err := s.users.FindByEmail(c.UserContext(), req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah", nil)
		}
		log.Println("login gagal:", err)
		return helper.Fail(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil)
	}

	// 3. cocokkan password dengan hash di database
	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah", nil)
	}

	// 4. buat access token
	token, err := s.jwt.GenerateAccess(user)
	if err != nil {
		log.Println("gagal membuat token:", err)
		return helper.Fail(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil)
	}

	return helper.Ok(c, "Login berhasil", model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwt.AccessTTL().Seconds()),
		User:        user,
	}, nil)
}

// Me: GET /api/v1/auth/me
func (s *AuthService) Me(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
	}

	user, err := s.users.FindByID(c.UserContext(), authUser.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusUnauthorized, "Akun tidak ditemukan", nil)
		}
		log.Println("me gagal:", err)
		return helper.Fail(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil)
	}

	data := model.MeResponse{ID: user.ID, Email: user.Email, Role: user.Role}

	// kalau mahasiswa, tambahkan data dari tabel students
	if user.Role == "mahasiswa" {
		profile, err := s.users.FindStudentProfile(c.UserContext(), user.ID)
		if err != nil {
			log.Println("me gagal ambil mahasiswa:", err)
			return helper.Fail(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil)
		}
		data.Student = &profile
	}

	return helper.Ok(c, "Data pengguna berhasil diambil", data, nil)
}