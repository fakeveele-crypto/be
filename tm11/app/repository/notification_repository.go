package repository

import (
	"context"
	"fmt"

	"latihan-fiber/tm11/app/model"
)

type NotificationRepository interface {
	WithTx(tx DBTX) NotificationRepository
	InsertIfAbsent(ctx context.Context, n model.Notification) (bool, error)
	FindByUser(ctx context.Context, userID, limit int) ([]model.Notification, error)
}

type notificationPostgresRepository struct{ db DBTX }

func NewNotificationRepository(db DBTX) NotificationRepository {
	return &notificationPostgresRepository{db: db}
}

func (r *notificationPostgresRepository) WithTx(tx DBTX) NotificationRepository {
	return &notificationPostgresRepository{db: tx}
}

func (r *notificationPostgresRepository) InsertIfAbsent(ctx context.Context, n model.Notification) (bool, error) {
	cmd, err := r.db.Exec(ctx,
		`INSERT INTO notifications (outbox_id, user_id, title, message)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (outbox_id) DO NOTHING`,
		n.OutboxID, n.UserID, n.Title, n.Message,
	)
	if err != nil {
		return false, fmt.Errorf("mencatat notifikasi: %w", err)
	}
	return cmd.RowsAffected() == 1, nil
}

func (r *notificationPostgresRepository) FindByUser(ctx context.Context, userID, limit int) ([]model.Notification, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, outbox_id, user_id, title, message, created_at
		   FROM notifications
		  WHERE user_id = $1
		  ORDER BY id DESC
		  LIMIT $2`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("mengambil notifikasi: %w", err)
	}
	defer rows.Close()

	items := []model.Notification{}
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.OutboxID, &n.UserID, &n.Title, &n.Message, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca notifikasi: %w", err)
		}
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca notifikasi: %w", err)
	}
	return items, nil
}
