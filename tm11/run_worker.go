package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"latihan-fiber/tm11/app/repository"
	"latihan-fiber/tm11/config"
	"latihan-fiber/tm11/database"
	"latihan-fiber/tm11/mailer"
	"latihan-fiber/tm11/worker"
)

func runWorker() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := config.NewLogger("./logs/worker.log")
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	mailSvc := mailer.NewFileMailer(
		"./logs/mail.log",
		time.Duration(config.GetEnvInt("MAIL_DELAY_MS", 1500))*time.Millisecond,
		config.GetEnvFloat("MAIL_FAIL_RATE", 0.2),
	)

	userRepo := repository.NewUserRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	outboxRepo := repository.NewOutboxRepository(pool)
	notificationRepo := repository.NewNotificationRepository(pool)
	txManager := repository.NewTxManager(pool)

	deps := worker.Deps{
		Tx:            txManager,
		Outbox:        outboxRepo,
		Notifications: notificationRepo,
		Users:         userRepo,
		Courses:       courseRepo,
		Mailer:        mailSvc,
		Logger:        logger,
	}

	if err := worker.Run(ctx, deps); err != nil && err != context.Canceled {
		logger.Error("worker gagal", "error", err)
		log.Fatalf("worker error: %v", err)
	}
}
