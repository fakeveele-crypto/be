package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"latihan-fiber/tm11/app/model"
	"latihan-fiber/tm11/app/repository"
	"latihan-fiber/tm11/queue"
)

type Relay struct {
	tx        *repository.TxManager
	outbox    repository.OutboxRepository
	client    *asynq.Client
	logger    *slog.Logger
	redisDown bool
}

func NewRelay(tx *repository.TxManager, outbox repository.OutboxRepository, client *asynq.Client, logger *slog.Logger) *Relay {
	return &Relay{tx: tx, outbox: outbox, client: client, logger: logger}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.processBatch(ctx)
		}
	}
}

func (r *Relay) processBatch(ctx context.Context) {
	batchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	enqueueFailed := false
	streamErr := r.tx.WithTx(batchCtx, func(tx repository.DBTX) error {
		repo := r.outbox.WithTx(tx)
		pending, err := repo.FindPending(batchCtx, 50)
		if err != nil {
			return err
		}

		for _, event := range pending {
			switch event.EventType {
			case model.EventEnrollmentEmail:
				var payload model.EnrollmentEmailPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					r.logger.Error("payload outbox rusak, surat diabaikan", "outbox_id", event.ID, "error", err)
					continue
				}

				task, err := queue.NewEnrollmentEmailTask(queue.EnrollmentEmailTask{
					OutboxID: event.ID,
					UserID:   payload.UserID,
					CourseID: payload.CourseID,
				})
				if err != nil {
					r.logger.Error("gagal membuat task email", "outbox_id", event.ID, "error", err)
					continue
				}

				_, err = r.client.EnqueueContext(batchCtx, task,
					asynq.Queue(queue.QueueEmail),
					asynq.MaxRetry(MaxRetry),
					asynq.TaskID(fmt.Sprintf("outbox-%d", event.ID)),
					asynq.Timeout(30*time.Second),
					asynq.Retention(24*time.Hour),
				)
				if err == nil || errors.Is(err, asynq.ErrTaskIDConflict) {
					if r.redisDown {
						r.redisDown = false
						r.logger.Info("Redis pulih, relay melanjutkan pengiriman task")
					}
					if err := repo.MarkProcessed(batchCtx, event.ID); err != nil {
						return err
					}
					continue
				}

				// enqueue bisa ganda bila enqueue sukses tetapi commit gagal.
				// Pengaman berlapis di sini adalah TaskID pada relay dan UNIQUE
				// outbox_id pada tabel notifications di handler.
				enqueueFailed = true
				if !r.redisDown {
					r.logger.Warn("Redis gagal menerima task, menunda pengiriman", "outbox_id", event.ID, "error", err)
					r.redisDown = true
				}
				return nil
			default:
				r.logger.Error("event_type tidak dikenal", "event_type", event.EventType, "outbox_id", event.ID)
			}
		}

		return nil
	})

	if streamErr != nil || enqueueFailed {
		return
	}
}
