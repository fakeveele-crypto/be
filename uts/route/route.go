package route

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/uts/app/service"
	"latihan-fiber/uts/helper"
	"latihan-fiber/uts/middleware"
)

func SetupRoutes(app *fiber.App, authService *service.AuthService, courseService *service.CourseService, studentService *service.StudentService, jwtManager *helper.JWTManager, pool *pgxpool.Pool) {
	api := app.Group("/api/v1")

	// endpoint sementara untuk ngecek server dan database
	api.Get("/health", func(c *fiber.Ctx) error {
		if err := pool.Ping(context.Background()); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi", nil)
		}
		return helper.Ok(c, "server dan database berjalan", nil, nil)
	})

	// endpoint auth
	authGroup := api.Group("/auth")
	authGroup.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	authGroup.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

	// mata kuliah, boleh dibuka semua role yang sudah login
	api.Get("/courses", middleware.RequireAuth(jwtManager), courseService.GetCourses)

	// mahasiswa: semua endpoint wajib login, sebagian khusus admin
	students := api.Group("/students", middleware.RequireAuth(jwtManager))
	students.Get("/", middleware.RequireRole("admin"), studentService.GetStudents)
	students.Post("/", middleware.RequireRole("admin"), studentService.CreateStudent)
	students.Get("/:id", studentService.GetStudentByID) // admin atau pemilik data, dicek di service
	students.Put("/:id", middleware.RequireRole("admin"), studentService.UpdateStudent)
	students.Delete("/:id", middleware.RequireRole("admin"), studentService.DeleteStudent)
}