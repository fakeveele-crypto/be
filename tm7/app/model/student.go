package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	Nim       string    `json:"nim"` 
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateStudentRequest struct {
	NIM   string  `json:"nim" validate:"required,nim"`
	Name  string  `json:"name" validate:"required,max=100"`
	Grade float64 `json:"grade" validate:"gte=0,lte=4"`
}

type ReplaceStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,nim"`
	Name     string  `json:"name" validate:"required,max=100"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=4"`
	IsActive bool    `json:"is_active"`
}

type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty" validate:"omitnil,nim"`
	Name     *string  `json:"name,omitempty" validate:"omitnil,min=1,max=100"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,gte=0,lte=4"`
	IsActive *bool    `json:"is_active,omitempty"`
}

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	Search     string `json:"search,omitempty"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Total      int    `json:"total"`
	TotalPages int    `json:"total_pages"`
}