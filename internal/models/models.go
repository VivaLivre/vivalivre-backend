package models

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Bathroom struct {
	ID                  int             `json:"id"`
	Name                string          `json:"name"`
	Address             string          `json:"address"`
	Latitude            float64         `json:"latitude"`
	Longitude           float64         `json:"longitude"`
	IsAccessible        bool            `json:"is_accessible"`
	HasChangingTable    bool            `json:"has_changing_table"`
	IsFree              bool            `json:"is_free"`
	OperatingHours      json.RawMessage `json:"operating_hours"`
	Observations        *string         `json:"observations,omitempty"`
	Comment             *string         `json:"comment,omitempty"`
	Status              string          `json:"status"`
	CleanlinessRating   float64         `json:"cleanliness_rating"`
	AccessibilityRating float64         `json:"accessibility_rating"`
	PhotoURL            string          `json:"photo_url,omitempty"`
	Distance            float64         `json:"distance,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
}

type HealthEntry struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Symptoms    []string  `json:"symptoms"`
	EntryDate   time.Time `json:"entry_date"`
}

type CreateHealthEntryRequest struct {
	Type        string   `json:"type" binding:"required"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Symptoms    []string `json:"symptoms"`
}

type AuthRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// BathroomReview representa uma avaliação de banheiro
type BathroomReview struct {
	ID                  int       `json:"id"`
	BathroomID          int       `json:"bathroom_id"`
	UserID              int       `json:"user_id"`
	Rating              int       `json:"rating"`
	Title               *string   `json:"title,omitempty"`
	Comment             *string   `json:"comment,omitempty"`
	CleanlinessRating   *int      `json:"cleanliness_rating,omitempty"`
	AccessibilityRating *int      `json:"accessibility_rating,omitempty"`
	SpaciosunessRating  *int      `json:"spaciousness_rating,omitempty"`
	HelpfulCount        int       `json:"helpful_count"`
	UnhelpfulCount      int       `json:"unhelpful_count"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CreateReviewRequest é o payload para criar uma avaliação
type CreateReviewRequest struct {
	Rating              int     `json:"rating" binding:"required,min=1,max=5"`
	Title               *string `json:"title" binding:"omitempty,max=100"`
	Comment             *string `json:"comment" binding:"omitempty,max=500"`
	CleanlinessRating   *int    `json:"cleanliness_rating" binding:"omitempty,min=1,max=5"`
	AccessibilityRating *int    `json:"accessibility_rating" binding:"omitempty,min=1,max=5"`
	SpaciosunessRating  *int    `json:"spaciousness_rating" binding:"omitempty,min=1,max=5"`
}

// UpdateReviewRequest é o payload para atualizar uma avaliação
type UpdateReviewRequest struct {
	Rating              *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	Title               *string `json:"title" binding:"omitempty,max=100"`
	Comment             *string `json:"comment" binding:"omitempty,max=500"`
	CleanlinessRating   *int    `json:"cleanliness_rating" binding:"omitempty,min=1,max=5"`
	AccessibilityRating *int    `json:"accessibility_rating" binding:"omitempty,min=1,max=5"`
	SpaciosunessRating  *int    `json:"spaciousness_rating" binding:"omitempty,min=1,max=5"`
}

// BathroomRatingStats representa estatísticas de ratings
type BathroomRatingStats struct {
	BathroomID         int            `json:"bathroom_id"`
	TotalReviews       int            `json:"total_reviews"`
	AverageRating      float64        `json:"average_rating"`
	AvgCleanliness     float64        `json:"avg_cleanliness"`
	AvgAccessibility   float64        `json:"avg_accessibility"`
	AvgSpaciosuneness  float64        `json:"avg_spaciousness"`
	RatingDistribution map[int]int    `json:"rating_distribution"`
}

// HelpfulVoteRequest é o payload para votar em utilidade
type HelpfulVoteRequest struct {
	IsHelpful bool `json:"is_helpful" binding:"required"`
}

// BathroomReport representa um report feito por um utilizador
type BathroomReport struct {
	ID          string    `json:"id"`
	BathroomID  int       `json:"bathroom_id"`
	BathroomName string   `json:"bathroom_name,omitempty"`
	UserID      int       `json:"user_id"`
	UserEmail   string    `json:"user_email,omitempty"`
	Reason      string    `json:"reason"`
	Description *string   `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateReportRequest é o payload para criar um report
type CreateReportRequest struct {
	Reason      string  `json:"reason" binding:"required"`
	Description *string `json:"description,omitempty"`
}

// BathroomSuggestion representa uma sugestão de alteração
type BathroomSuggestion struct {
	ID               string                 `json:"id"`
	BathroomID       int                    `json:"bathroom_id"`
	BathroomName     string                 `json:"bathroom_name,omitempty"`
	UserID           int                    `json:"user_id"`
	UserEmail        string                 `json:"user_email,omitempty"`
	SuggestedUpdates json.RawMessage        `json:"suggested_updates"`
	Status           string                 `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
}

// CreateSuggestionRequest é o payload para criar uma sugestão
type CreateSuggestionRequest struct {
	SuggestedUpdates json.RawMessage        `json:"suggested_updates" binding:"required"`
}
// UpdateReportStatusRequest é o payload para atualizar o status do report
type UpdateReportStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateSuggestionStatusRequest é o payload para atualizar o status da sugestão
type UpdateSuggestionStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateBathroomAdminRequest represents a partial update for a bathroom by an admin
type UpdateBathroomAdminRequest struct {
	Name             *string          `json:"name,omitempty"`
	Address          *string          `json:"address,omitempty"`
	Latitude         *float64         `json:"latitude,omitempty"`
	Longitude        *float64         `json:"longitude,omitempty"`
	IsAccessible     *bool            `json:"is_accessible,omitempty"`
	HasChangingTable *bool            `json:"has_changing_table,omitempty"`
	IsFree           *bool            `json:"is_free,omitempty"`
	OperatingHours   *json.RawMessage `json:"operating_hours,omitempty"`
	Observations     *string          `json:"observations,omitempty"`
	Status           *string          `json:"status,omitempty"`
	PhotoUrl         *string          `json:"photo_url,omitempty"`
}

// PaginationMeta holds pagination metadata
type PaginationMeta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// PaginatedBathroomsResponse is the response for paginated bathrooms
type PaginatedBathroomsResponse struct {
	Data []Bathroom     `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
