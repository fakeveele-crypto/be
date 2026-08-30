package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	api := app.Group("/api/v1")
	studentsGroup := api.Group("/students", requireJSON)

	studentsGroup.Get("/", getStudents)
	studentsGroup.Get("/:id", getStudentByID)
	studentsGroup.Post("/", createStudent)
	studentsGroup.Put("/:id", updateStudent)
	studentsGroup.Patch("/:id", patchStudent)
	studentsGroup.Delete("/:id", deleteStudent)

	log.Println("Server berjalan di port 3000...")
	err := app.Listen(":3000")
	if err != nil {
		log.Fatal("Gagal menyalakan server: ", err)
	}
}