package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"latihan-fiber/tm11/app/model"
)

type OutboxRepository interface {
	WithTx(tx DBTX) OutboxRepository
	Insert(ctx context.Context, eventType string, payload []byte) (int64, error)
	FindPending(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	MarkProcessed(ctx context.Context, id int64) error
}

type outboxPostgresRepository struct{ db DBTX }

func NewOutboxRepository(db DBTX) OutboxRepository {
	return &outboxPostgresRepository{db: db}
}

func (r *outboxPostgresRepository) WithTx(tx DBTX) OutboxRepository {
	return &outboxPostgresRepository{db: tx}
}

func scanOutbox(row pgx.Row) (model.OutboxEvent, error) {
	var event model.OutboxEvent
	err := row.Scan(&event.ID, &event.EventType, &event.Payload, &event.CreatedAt)
	return event, err
}

func (r *outboxPostgresRepository) Insert(ctx context.Context, eventType string, payload []byte) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO outbox (event_type, payload) VALUES ($1, $2) RETURNING id`,
		eventType, payload,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("mencatat surat keluar: %w", err)
	}
	return id, nil
}

func (r *outboxPostgresRepository) FindPending(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	// FOR UPDATE SKIP LOCKED mencegah dua worker yang sama-sama menangkap
	// surat yang sama. Worker pertama memegang baris, worker lain melompati
	// row itu dan mengambil item lain, sehingga tidak terjadi perebutan task.
	rows, err := r.db.Query(ctx,
		`SELECT id, event_type, payload, created_at
		   FROM outbox
		  WHERE processed_at IS NULL
		  ORDER BY id
		  LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, fmt.Errorf("mengambil surat keluar: %w", err)
	}
	defer rows.Close()

	result := make([]model.OutboxEvent, 0)
	for rows.Next() {
		var event model.OutboxEvent
		if err := rows.Scan(&event.ID, &event.EventType, &event.Payload, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca surat keluar: %w", err)
		}
		result = append(result, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mengambil surat keluar: %w", err)
	}
	return result, nil
}

func (r *outboxPostgresRepository) MarkProcessed(ctx context.Context, id int64) error {
	cmd, err := r.db.Exec(ctx,
		`UPDATE outbox SET processed_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menandai surat keluar selesai: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("surat keluar tidak ditemukan")
	}
	return nil
}
