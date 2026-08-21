package models

import "time"

// AuthRequest is the payload for login
type AuthRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// RegisterRequest is the payload for registration
type RegisterRequest struct {
	Name              string   `json:"name" binding:"required,min=3,max=100"`
	Email             string   `json:"email" binding:"required,email,max=255"`
	Password          string   `json:"password" binding:"required,min=6,max=100"`
	CPF               string   `json:"cpf" binding:"omitempty,len=11,numeric"`
	DateOfBirth       string   `json:"date_of_birth"`
	Gender            string   `json:"gender"`
	Weight            *float64 `json:"weight"`
	Height            *float64 `json:"height"`
	ClinicalCondition string   `json:"clinical_condition" binding:"omitempty,max=150"`
	Comorbidities     []string `json:"comorbidities"`
}

// AuthResponse is the response after successful authentication
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// GoogleAuthRequest is the payload for Google OAuth login
type GoogleAuthRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// HealthEntry represents a health diary entry
type HealthEntry struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Symptoms    []string  `json:"symptoms"`
	EntryDate   time.Time `json:"entry_date"`
}

// CreateHealthEntryRequest is the payload for creating a health entry
type CreateHealthEntryRequest struct {
	Type        string   `json:"type" binding:"required,max=50"`
	Severity    string   `json:"severity" binding:"omitempty,max=50"`
	Description string   `json:"description" binding:"omitempty,max=1000"`
	Symptoms    []string `json:"symptoms"`
}

// UpdateHealthEntryRequest is the payload for updating a health entry
type UpdateHealthEntryRequest struct {
	Type        string   `json:"type" binding:"required,max=50"`
	Severity    string   `json:"severity" binding:"omitempty,max=50"`
	Description string   `json:"description" binding:"omitempty,max=1000"`
	Symptoms    []string `json:"symptoms"`
}
