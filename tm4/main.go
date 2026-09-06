package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"

	// Sesuaikan path import dengan nama module di go.mod milikmu
	"latihan-fiber/tm4/app/repository"
	"latihan-fiber/tm4/app/service"
	"latihan-fiber/tm4/config"
	"latihan-fiber/tm4/database"
	"latihan-fiber/tm4/route"
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

	// 4. Inisialisasi Fiber & Logger
	app := fiber.New(config.NewFiberConfig())
	config.SetupLogger(app)

	// 5. Daftarkan Routes
	route.SetupRoutes(app, studentService, pool)

	// 6. Jalankan Server
	port := config.GetEnv("APP_PORT", "3000")
	log.Printf("Server berjalan di port %s...\n", port)

	err = app.Listen(":" + port)
	if err != nil {
		log.Fatal("Gagal menyalakan server: ", err)
	}
}