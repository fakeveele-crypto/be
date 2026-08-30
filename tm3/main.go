package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/tm3/config"   
	"latihan-fiber/tm3/database" 
)

func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	app := fiber.New()
	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"message": "database tidak dapat dihubungi",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "server dan database berjalan",
		})
	})

	studentsGroup := api.Group("/students", requireJSON)
	studentsGroup.Get("/", getStudents)
	studentsGroup.Get("/:id", getStudentByID)
	studentsGroup.Post("/", createStudent)
	studentsGroup.Put("/:id", updateStudent)
	studentsGroup.Patch("/:id", patchStudent)
	studentsGroup.Delete("/:id", deleteStudent)

	port := config.GetEnv("APP_PORT", "3000")
	log.Printf("Server berjalan di port %s...\n", port)
	
	err = app.Listen(":" + port)
	if err != nil {
		log.Fatal("Gagal menyalakan server: ", err)
	}
}