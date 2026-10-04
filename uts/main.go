package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/uts/config"
	"latihan-fiber/uts/database"
	"latihan-fiber/uts/helper"
)

func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatal("gagal konek database: ", err)
	}
	defer pool.Close()

	app := fiber.New(config.NewFiberConfig())

	// endpoint sementara untuk ngecek server dan database
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		if err := pool.Ping(context.Background()); err != nil {
			return helper.Fail(c, 503, "database tidak dapat dihubungi", nil)
		}
		return helper.Ok(c, "server dan database berjalan", nil, nil)
	})

	// SEMENTARA, untuk ngecek bentuk error 422. Hapus setelah dites.
	app.Get("/api/v1/tes-error", func(c *fiber.Ctx) error {
		return helper.Fail(c, 422, "Validasi gagal", map[string][]string{
			"nim": {"NIM sudah terdaftar"},
		})
	})

	port := config.GetEnv("APP_PORT", "3000")
	log.Fatal(app.Listen(":" + port))
}