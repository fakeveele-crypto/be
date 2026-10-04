package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/app/repository"
	"latihan-fiber/uts/app/service"
	"latihan-fiber/uts/config"
	"latihan-fiber/uts/database"
	"latihan-fiber/uts/helper"
	"latihan-fiber/uts/route"
)

func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatal("gagal konek database: ", err)
	}
	defer pool.Close()

	// pengaturan JWT dari file .env
	secret := config.GetEnv("JWT_SECRET", "")
	if secret == "" {
		log.Fatal("JWT_SECRET belum diisi di .env")
	}
	ttl := time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60)) * time.Minute
	jwtManager := helper.NewJWTManager(secret, config.GetEnv("JWT_ISSUER", "siakad-mini"), ttl)

	// rangkai: repository -> service
	userRepo := repository.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo, jwtManager)
	courseRepo := repository.NewCourseRepository(pool)
	courseService := service.NewCourseService(courseRepo)

	app := fiber.New(config.NewFiberConfig())
	route.SetupRoutes(app, authService, courseService, jwtManager, pool)

	port := config.GetEnv("APP_PORT", "3000")
	log.Fatal(app.Listen(":" + port))
}