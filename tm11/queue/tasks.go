package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// QueueEmail adalah nama rel struk yang dipakai relay dan worker.
// Nama yang sama penting agar task tidak menunggu di antrian yang tidak
// pernah dibaca worker.
const QueueEmail = "email"

// TypeEnrollmentEmail adalah task yang memberitahu worker untuk
// mengirim email pendaftaran.
const TypeEnrollmentEmail = "email:enrollment"

type EnrollmentEmailTask struct {
	OutboxID int64 `json:"outbox_id"`
	UserID   int   `json:"user_id"`
	CourseID int   `json:"course_id"`
}

func NewEnrollmentEmailTask(p EnrollmentEmailTask) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEnrollmentEmail, payload), nil
}
