package route
import (
	"context"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/tm4/app/service"
	"latihan-fiber/tm4/middleware"
)

func SetupRoutes(app *fiber.App, studentService *service.StudentService, pool *pgxpool.Pool) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

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

	studentsGroup := api.Group("/students", middleware.RequireJSON)
	studentsGroup.Get("/", studentService.GetStudents)
	studentsGroup.Get("/:id", studentService.GetStudentByID)
	studentsGroup.Post("/", studentService.CreateStudent)
	studentsGroup.Put("/:id", studentService.UpdateStudent)
	studentsGroup.Patch("/:id", studentService.UpdateStudent)
	studentsGroup.Delete("/:id", studentService.DeleteStudent)
}