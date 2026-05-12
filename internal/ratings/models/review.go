package models

import (
	"time"

	"github.com/google/uuid"
)

// BathroomReview representa uma avaliação de banheiro
type BathroomReview struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	BathroomID           uuid.UUID  `json:"bathroom_id" db:"bathroom_id"`
	UserID               uuid.UUID  `json:"user_id" db:"user_id"`
	Rating               int        `json:"rating" db:"rating"`
	Title                *string    `json:"title" db:"title"`
	Comment              *string    `json:"comment" db:"comment"`
	CleanlinessRating    *int       `json:"cleanliness_rating" db:"cleanliness_rating"`
	AccessibilityRating  *int       `json:"accessibility_rating" db:"accessibility_rating"`
	SpaciosunessRating   *int       `json:"spaciousness_rating" db:"spaciousness_rating"`
	HelpfulCount         int        `json:"helpful_count" db:"helpful_count"`
	UnhelpfulCount       int        `json:"unhelpful_count" db:"unhelpful_count"`
	Status               string     `json:"status" db:"status"`
	Photos               []string   `json:"photos,omitempty" db:"-"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// ReviewPhoto representa uma foto de uma avaliação
type ReviewPhoto struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ReviewID  uuid.UUID `json:"review_id" db:"review_id"`
	PhotoURL  string    `json:"photo_url" db:"photo_url"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// ReviewHelpfulVote representa um voto de utilidade
type ReviewHelpfulVote struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ReviewID  uuid.UUID `json:"review_id" db:"review_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	IsHelpful bool      `json:"is_helpful" db:"is_helpful"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// BathroomRatingStats representa estatísticas de ratings de um banheiro
type BathroomRatingStats struct {
	BathroomID          uuid.UUID `json:"bathroom_id" db:"bathroom_id"`
	TotalReviews        int       `json:"total_reviews" db:"total_reviews"`
	AverageRating       float64   `json:"average_rating" db:"average_rating"`
	AvgCleanliness      float64   `json:"avg_cleanliness" db:"avg_cleanliness"`
	AvgAccessibility    float64   `json:"avg_accessibility" db:"avg_accessibility"`
	AvgSpaciosuneness   float64   `json:"avg_spaciousness" db:"avg_spaciousness"`
	RatingDistribution  map[int]int `json:"rating_distribution,omitempty" db:"-"`
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

// HelpfulVoteRequest é o payload para votar em utilidade
type HelpfulVoteRequest struct {
	IsHelpful bool `json:"is_helpful" binding:"required"`
}

// ReviewResponse é a resposta com informações do utilizador
type ReviewResponse struct {
	ID                  uuid.UUID `json:"id"`
	BathroomID          uuid.UUID `json:"bathroom_id"`
	User                *UserInfo `json:"user"`
	Rating              int       `json:"rating"`
	Title               *string   `json:"title"`
	Comment             *string   `json:"comment"`
	CleanlinessRating   *int      `json:"cleanliness_rating"`
	AccessibilityRating *int      `json:"accessibility_rating"`
	SpaciosunessRating  *int      `json:"spaciousness_rating"`
	HelpfulCount        int       `json:"helpful_count"`
	UnhelpfulCount      int       `json:"unhelpful_count"`
	Photos              []string  `json:"photos"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// UserInfo contém informações básicas do utilizador
type UserInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ReviewStatus define os possíveis estados de uma avaliação
type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
)

// ListReviewsQuery contém parâmetros de query para listar reviews
type ListReviewsQuery struct {
	Sort   string `form:"sort" binding:"omitempty,oneof=recent helpful rating"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" binding:"omitempty,min=0"`
}
