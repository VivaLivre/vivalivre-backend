package models

import (
	"encoding/json"
	"time"
)

// BathroomReport representa um report feito por um utilizador
type BathroomReport struct {
	ID           string    `json:"id"`
	BathroomID   int       `json:"bathroom_id"`
	BathroomName string    `json:"bathroom_name,omitempty"`
	UserID       int       `json:"user_id"`
	UserEmail    string    `json:"user_email,omitempty"`
	Reason       string    `json:"reason"`
	Description  *string   `json:"description,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateReportRequest é o payload para criar um report
type CreateReportRequest struct {
	Reason      string  `json:"reason" binding:"required,max=150"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// UpdateReportStatusRequest é o payload para atualizar o status do report
type UpdateReportStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// BathroomSuggestion representa uma sugestão de alteração
type BathroomSuggestion struct {
	ID               string          `json:"id"`
	BathroomID       int             `json:"bathroom_id"`
	BathroomName     string          `json:"bathroom_name,omitempty"`
	UserID           int             `json:"user_id"`
	UserEmail        string          `json:"user_email,omitempty"`
	SuggestedUpdates json.RawMessage `json:"suggested_updates"`
	Status           string          `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
}

// CreateSuggestionRequest é o payload para criar uma sugestão
type CreateSuggestionRequest struct {
	SuggestedUpdates json.RawMessage `json:"suggested_updates" binding:"required"`
}

// UpdateSuggestionStatusRequest é o payload para atualizar o status da sugestão
type UpdateSuggestionStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
