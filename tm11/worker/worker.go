package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"latihan-fiber/tm11/app/repository"
	"latihan-fiber/tm11/config"
	"latihan-fiber/tm11/mailer"
	"latihan-fiber/tm11/queue"
)

type Deps struct {
	Tx            *repository.TxManager
	Outbox        repository.OutboxRepository
	Notifications repository.NotificationRepository
	Users         repository.UserRepository
	Courses       repository.CourseRepository
	Mailer        mailer.Mailer
	Logger        *slog.Logger
}

func Run(ctx context.Context, deps Deps) error {
	redisOpt := config.RedisClientOpt()
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	server := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: config.GetEnvInt("WORKER_CONCURRENCY", 10),
		Queues: map[string]int{
			queue.QueueEmail: 1,
		},
		RetryDelayFunc:  RetryDelay,
		ShutdownTimeout: 15 * time.Second,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retryCount, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			if errors.Is(err, asynq.SkipRetry) || (maxRetry > 0 && retryCount >= maxRetry) {
				deps.Logger.Warn("task akan diarsip", "task_type", task.Type(), "retry_count", retryCount, "max_retry", maxRetry, "error", err)
				return
			}
			deps.Logger.Warn("task gagal", "task_type", task.Type(), "retry_count", retryCount, "max_retry", maxRetry, "error", err)
		}),
	})

	mux := asynq.NewServeMux()
	handler := NewEmailHandler(
		deps.Tx,
		deps.Notifications,
		deps.Users,
		deps.Courses,
		deps.Mailer,
		deps.Logger,
	)
	mux.HandleFunc(queue.TypeEnrollmentEmail, handler.ProcessEnrollmentEmail)

	if err := server.Start(mux); err != nil {
		deps.Logger.Warn("server asynq gagal mulai, relay tetap berjalan dan akan mencoba lagi", "error", err)
	}

	relay := NewRelay(deps.Tx, deps.Outbox, client, deps.Logger)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		relay.Run(ctx)
	}()

	<-ctx.Done()
	deps.Logger.Info("berhenti menerima task baru, menunggu task yang sedang berjalan selesai")
	wg.Wait()
	server.Shutdown()
	return nil
}
