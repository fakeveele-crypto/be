package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"latihan-fiber/tm9/app/repository"
	"latihan-fiber/tm9/app/service"
	"latihan-fiber/tm9/config"
	"latihan-fiber/tm9/database"
	"latihan-fiber/tm9/helper"
	"latihan-fiber/tm9/middleware"
	"latihan-fiber/tm9/route"
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

	userRepo := repository.NewUserRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	supervisorRepo := repository.NewSupervisorRepository(pool)
	txManager := repository.NewTxManager(pool)
	tokenRepo := repository.NewTokenRepository(pool)

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET tidak diisi atau terlalu pendek (minimal 32 karakter)")
	}

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	roleRepository := repository.NewRoleRepository(pool)

	// Pemetaan role ke permission dibaca SEKALI saat aplikasi menyala.
	// Konsekuensinya: perubahan hak akses di database baru berlaku setelah
	// aplikasi dijalankan ulang. Ini keputusan sadar, bukan kelalaian.
	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		log.Fatalf("gagal memuat permission: %v", err)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	log.Printf("permission dimuat untuk role: %v\n", permissions.KnownRoles())

	studentService := service.NewStudentService(studentRepo, permissions)
	userService := service.NewUserService(userRepo, permissions)
	courseService := service.NewCourseService(courseRepo, txManager)
	supervisionService := service.NewSupervisionService(supervisorRepo, txManager)

	refreshTTL := time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour
	authService := service.NewAuthService(userRepo, tokenRepo, jwtManager, refreshTTL)

	// 4. Inisialisasi Fiber & Logger
	logger := config.NewLogger()
	app := fiber.New(config.NewFiberConfig(logger))
	app.Use(requestid.New())
	app.Use(middleware.RequestLogger(logger))

	// 5. Daftarkan Routes
	route.SetupRoutes(app, studentService, authService, userService, courseService, supervisionService, jwtManager, permissions, pool)

	// 6. Jalankan Server
	port := config.GetEnv("APP_PORT", "3000")
	log.Printf("Server berjalan di port %s...\n", port)

	err = app.Listen(":" + port)
	if err != nil {
		log.Fatal("Gagal menyalakan server: ", err)
	}
}
