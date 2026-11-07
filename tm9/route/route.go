package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/tm9/app/service"
	"latihan-fiber/tm9/helper"
	"latihan-fiber/tm9/middleware"
)

func SetupRoutes(
	app *fiber.App,
	studentService *service.StudentService,
	authService *service.AuthService,
	userService *service.UserService,
	courseService *service.CourseService,
	supervisionService *service.SupervisionService,
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

	// --- kelas dan KRS (Modul 9) ---
	courses := api.Group("/courses", middleware.RequireJSON, middleware.RequireAuth(jwtManager))

	// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai
	// urutan pendaftaran; bila dibalik, "mine" akan dianggap sebagai :id.
	courses.Get("/", courseService.List)
	courses.Get("/mine", courseService.Mine)
	courses.Get("/:id", courseService.Get)
	courses.Post("/:id/enroll", courseService.Enroll)
	courses.Delete("/:id/enroll", courseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), courseService.UpdateQuota)

	enrollments := api.Group("/enrollments", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	enrollments.Post("/swap", courseService.Swap)

	// --- pembimbing tugas akhir (Modul 9, tugas D.2) ---
	supervisors := api.Group("/supervisors", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	supervisors.Get("/", supervisionService.List)
	supervisors.Get("/:id", supervisionService.Get)
	supervisors.Post("/:id/assign", middleware.RequirePermission(perms, "supervision:manage"), supervisionService.Assign)
	supervisors.Patch("/:id", middleware.RequirePermission(perms, "supervision:manage"), supervisionService.UpdateQuota)

	supervisions := api.Group("/supervisions", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	supervisions.Post("/switch", middleware.RequirePermission(perms, "supervision:manage"), supervisionService.Switch)

	// Harus paling akhir: hanya request yang tidak cocok dengan route mana pun yang sampai ke sini.
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})
}
