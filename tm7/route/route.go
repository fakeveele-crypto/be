package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/tm7/app/service"
	"latihan-fiber/tm7/helper"
	"latihan-fiber/tm7/middleware"
)

func SetupRoutes(
	app *fiber.App,
	studentService *service.StudentService,
	authService *service.AuthService,
	userService *service.UserService,
	jwtManager *helper.JWTManager,
	perms *helper.PermissionSet,
	pool *pgxpool.Pool,
) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	api := app.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("database tidak dapat dihubungi")
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "server dan database berjalan",
		})
	})

	authGroup := api.Group("/auth", middleware.RequireJSON)
	authGroup.Post("/register", authService.Register)
	authGroup.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	authGroup.Post("/refresh", authService.Refresh)
	authGroup.Post("/logout", authService.Logout)
	authGroup.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

	// Seluruh endpoint students wajib membawa access token yang sah.
	studentsGroup := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(jwtManager))

	// Hak-hak berikut dapat diputuskan TANPA melihat isi data, sehingga
	// cukup diperiksa di sini, sebelum masuk ke service.
	studentsGroup.Get("/", middleware.RequirePermission(perms, "student:list"), studentService.GetStudents)
	studentsGroup.Post("/", middleware.RequirePermission(perms, "student:create"), studentService.CreateStudent)
	studentsGroup.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), studentService.DeleteStudent)

	// Hak-hak berikut BERGANTUNG pada siapa pemilik datanya, sehingga
	// pemeriksaannya dilakukan di dalam service, bukan di sini.
	studentsGroup.Get("/:id", studentService.GetStudentByID)
	studentsGroup.Put("/:id", studentService.UpdateStudent)
	studentsGroup.Patch("/:id", studentService.PatchStudent)

	usersGroup := api.Group("/users", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	usersGroup.Get("/", middleware.RequirePermission(perms, "user:list"), userService.List)
	usersGroup.Get("/:id", userService.Get)

	// Harus paling akhir: hanya request yang tidak cocok dengan route mana pun yang sampai ke sini.
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})
}