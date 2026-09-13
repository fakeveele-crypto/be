package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm5/app/repository"
	"latihan-fiber/tm5/app/service"
	"latihan-fiber/tm5/config"
	"latihan-fiber/tm5/database"
	"latihan-fiber/tm5/helper"
	"latihan-fiber/tm5/route"
)

func main() {
	// 1. Load Environment Variable
	config.LoadEnv()

	// 2. Inisialisasi Database Pool
	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	// 3. Inject Repository & Service
	studentRepo := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepo)

	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)

	jwtManager := helper.NewJWTManager(
		config.GetEnv("JWT_SECRET", ""),
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		15*time.Minute,
	)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtManager, 7*24*time.Hour)

	// 4. Inisialisasi Fiber & Logger
	app := fiber.New(config.NewFiberConfig())
	config.SetupLogger(app)

	// 5. Daftarkan Routes
	route.SetupRoutes(app, studentService, authService, pool)

	// 6. Jalankan Server
	port := config.GetEnv("APP_PORT", "3000")
	log.Printf("Server berjalan di port %s...\n", port)

	err = app.Listen(":" + port)
	if err != nil {
		log.Fatal("Gagal menyalakan server: ", err)
	}
}