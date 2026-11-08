package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"latihan-fiber/tm11/app/model"
	"latihan-fiber/tm11/app/repository"
	"latihan-fiber/tm11/mailer"
	"latihan-fiber/tm11/queue"
)

type EmailHandler struct {
	tx            *repository.TxManager
	notifications repository.NotificationRepository
	users         repository.UserRepository
	courses       repository.CourseRepository
	mailer        mailer.Mailer
	logger        *slog.Logger
}

func NewEmailHandler(
	tx *repository.TxManager,
	notifications repository.NotificationRepository,
	users repository.UserRepository,
	courses repository.CourseRepository,
	mail mailer.Mailer,
	logger *slog.Logger,
) *EmailHandler {
	return &EmailHandler{
		tx:            tx,
		notifications: notifications,
		users:         users,
		courses:       courses,
		mailer:        mail,
		logger:        logger,
	}
}

func (h *EmailHandler) ProcessEnrollmentEmail(ctx context.Context, t *asynq.Task) error {
	var payload queue.EnrollmentEmailTask
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("%w: payload email rusak: %v", asynq.SkipRetry, err)
	}

	user, err := h.users.FindByID(ctx, payload.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: user tidak ditemukan", asynq.SkipRetry)
		}
		return err
	}

	course, err := h.courses.FindByID(ctx, payload.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: kelas tidak ditemukan", asynq.SkipRetry)
		}
		return err
	}

	title := "Pendaftaran kelas berhasil"
	message := fmt.Sprintf("%s berhasil mendaftar di %s (%s)", user.Username, course.Code, course.Name)
	subject := title
	body := message

	var txErr error
	txErr = h.tx.WithTx(ctx, func(tx repository.DBTX) error {
		notificationRepo := h.notifications.WithTx(tx)
		inserted, err := notificationRepo.InsertIfAbsent(ctx, model.Notification{
			OutboxID: payload.OutboxID,
			UserID:   user.ID,
			Title:    title,
			Message:  message,
		})
		if err != nil {
			return err
		}
		if !inserted {
			h.logger.Info("email dilewati karena notifikasi sudah tercatat", "outbox_id", payload.OutboxID)
			return nil
		}

		// Catat nomor dulu, lalu kirim email. Dalam satu transaction, dua worker
		// yang sama-sama mengambil tugas yang sama akan menunggu di lock yang sama,
		// sehingga notifikasi tidak bisa dibuat dua kali untuk satu outbox.
		// Celah at-least-once masih ada: email bisa terkirim sebelum commit
		// transaction, lalu commit gagal. Dalam situasi itu email dapat
		// terkirim dua kali dan tidak ada cara menghilangkannya seluruhnya.
		if err := h.mailer.Send(ctx, user.Email, subject, body); err != nil {
			return err
		}
		return nil
	})

	if txErr != nil {
		if errors.Is(txErr, mailer.ErrPermanent) {
			return fmt.Errorf("%w: %v", asynq.SkipRetry, txErr)
		}
		return txErr
	}
	return nil
}
