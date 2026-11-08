package model

import "time"

// Notification adalah buku tanda terima yang dibaca oleh client.
// Field outbox_id dan user_id disembunyikan agar client tidak melihat
// detail internals yang hanya dipakai oleh worker untuk idempotency.
type Notification struct {
	ID        int64     `json:"id"`
	OutboxID  int64     `json:"-"`
	UserID    int       `json:"-"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
