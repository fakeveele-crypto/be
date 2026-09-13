package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm5/app/model"
	"latihan-fiber/tm5/app/repository"
	"latihan-fiber/tm5/helper"
)

const refreshTokenBytes = 32

type AuthService struct {
	users      repository.UserRepository
	tokens     repository.TokenRepository
	jwt        *helper.JWTManager
	refreshTTL time.Duration
}

func NewAuthService(
	users repository.UserRepository,
	tokens repository.TokenRepository,
	jwtManager *helper.JWTManager,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users: users, tokens: tokens, jwt: jwtManager, refreshTTL: refreshTTL,
	}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "format data tidak valid"})
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if errs := ValidateRegister(req); len(errs) > 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"message": "validasi gagal",
			"errors":  errs,
		})
	}

	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "gagal memproses password"})
	}

	created, err := s.users.Create(c.UserContext(), model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     "user",
		IsActive: true,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "pendaftaran berhasil",
		"data":    created,
	})
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "format data tidak valid"})
	}

	user, err := s.users.FindByUsername(c.UserContext(), strings.TrimSpace(req.Username))
	if err != nil {
		// Username tidak ditemukan. Tetap jalankan pemeriksaan palsu agar
		// waktu tanggapnya mirip kasus password salah (cegah user enumeration).
		helper.VerifyDummyPassword(req.Password)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "username atau password salah"})
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "username atau password salah"})
	}

	if !user.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "akun dinonaktifkan"})
	}

	pair, err := s.issueTokenPair(c, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "gagal membuat token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "login berhasil",
		"data":    pair,
	})
}

// Refresh menukar refresh token yang masih aktif dengan sepasang token baru.
// Refresh token lama langsung dicabut (rotasi) sehingga hanya berguna sekali;
// bila token yang sama dipakai lagi, itu tanda kuat token sudah dicuri.
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "format data tidak valid"})
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "refresh_token wajib diisi"})
	}

	hash := helper.SHA256Hex(req.RefreshToken)
	stored, err := s.tokens.FindActive(c.UserContext(), hash)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "refresh token tidak valid atau sudah kedaluwarsa",
		})
	}

	user, err := s.users.FindByID(c.UserContext(), stored.UserID)
	if err != nil || !user.IsActive {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "akun tidak dapat dipakai"})
	}

	// ROTASI: token lama langsung dicabut sebelum menerbitkan yang baru.
	if err := s.tokens.Revoke(c.UserContext(), hash); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "gagal memperbarui token"})
	}

	pair, err := s.issueTokenPair(c, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "gagal membuat token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "token berhasil diperbarui",
		"data":    pair,
	})
}

// Logout mencabut refresh token yang dikirim. Dari sudut pandang client,
// logout selalu dianggap berhasil walau token yang dikirim sudah tidak aktif.
func (s *AuthService) Logout(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "format data tidak valid"})
	}
	if strings.TrimSpace(req.RefreshToken) != "" {
		_ = s.tokens.Revoke(c.UserContext(), helper.SHA256Hex(req.RefreshToken))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "logout berhasil"})
}

// Me mengembalikan profil pemakai yang identitasnya sudah diverifikasi oleh
// middleware RequireAuth dan disimpan di Locals.
func (s *AuthService) Me(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "belum terautentikasi"})
	}
	user, err := s.users.FindByID(c.UserContext(), authUser.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user tidak ditemukan"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "gagal mengambil profil"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "profil berhasil diambil",
		"data":    user,
	})
}

// issueTokenPair membuat access token dan refresh token sekaligus.
func (s *AuthService) issueTokenPair(c *fiber.Ctx, user model.User) (model.TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return model.TokenPair{}, err
	}

	refreshToken, err := helper.RandomToken(refreshTokenBytes)
	if err != nil {
		return model.TokenPair{}, err
	}

	err = s.tokens.Save(c.UserContext(), model.RefreshToken{
		UserID:    user.ID,
		TokenHash: helper.SHA256Hex(refreshToken),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	})
	if err != nil {
		return model.TokenPair{}, err
	}

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
	}, nil
}