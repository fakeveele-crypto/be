package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"latihan-fiber/tm11/app/service"
	"latihan-fiber/tm11/helper"
	"latihan-fiber/tm11/middleware"
)

func SetupRoutes(
	app *fiber.App,
	studentService *service.StudentService,
	authService *service.AuthService,
	userService *service.UserService,
	courseService *service.CourseService,
	notificationService *service.NotificationService,
	supervisionService *service.SupervisionService,
	reviewService *service.ReviewService,
	portfolioService *service.PortfolioService,
	jwtManager *helper.JWTManager,
	perms *helper.PermissionSet,
	pool *pgxpool.Pool,
	mongoClient *mongo.Client,
) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(pool, mongoClient))

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

	// --- portofolio mahasiswa (MongoDB, Modul 10 tugas mandiri) ---
	// Hak akses diperiksa di dalam service: sama seperti data student,
	// bergantung pada siapa pemilik student-nya.
	studentsGroup.Get("/:id/portfolio", portfolioService.Get)
	studentsGroup.Put("/:id/portfolio", portfolioService.Replace)
	studentsGroup.Post("/:id/portfolio/projects", portfolioService.AddProject)
	studentsGroup.Delete("/:id/portfolio/projects/:projectId", portfolioService.RemoveProject)

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

	// --- ulasan kelas (MongoDB, Modul 10) ---
	courses.Get("/:id/reviews", reviewService.ListByCourse)
	courses.Get("/:id/reviews/summary", reviewService.Summary)
	courses.Post("/:id/reviews", reviewService.Create)
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

	reviews := api.Group("/reviews", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	reviews.Get("/:id", reviewService.Get)
	reviews.Patch("/:id", reviewService.Update)
	reviews.Delete("/:id", reviewService.Delete)
	reviews.Post("/:id/replies", reviewService.Reply)
	reviews.Post("/:id/helpful", reviewService.MarkHelpful)

	// --- statistik lintas portofolio (MongoDB, Modul 10 tugas mandiri) ---
	portfolios := api.Group("/portfolios", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	portfolios.Get("/stats/technologies", portfolioService.TopTechnologies)

	notifications := api.Group("/notifications", middleware.RequireJSON, middleware.RequireAuth(jwtManager))
	notifications.Get("/", notificationService.Mine)

	// Harus paling akhir: hanya request yang tidak cocok dengan route mana pun yang sampai ke sini.
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})
}

// healthCheck kini memeriksa DUA database. Aplikasi dianggap sehat hanya
// bila keduanya dapat dihubungi, dan pesan kegagalannya menyebut yang mana.
func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}
		if err := mongoClient.Ping(ctx, nil); err != nil {
			return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}
