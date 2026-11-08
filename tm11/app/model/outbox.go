package model

import "time"

const EventEnrollmentEmail = "enrollment.email"

type OutboxEvent struct {
	ID        int64     `json:"id"`
	EventType string    `json:"event_type"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type EnrollmentEmailPayload struct {
	UserID   int `json:"user_id"`
	CourseID int `json:"course_id"`
}
