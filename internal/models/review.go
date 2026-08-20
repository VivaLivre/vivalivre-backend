package models

import "time"

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
	SpaciousnessRating  *int      `json:"spaciousness_rating,omitempty"`
	HelpfulCount        int       `json:"helpful_count"`
	UnhelpfulCount      int       `json:"unhelpful_count"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	UserName            string    `json:"user_name,omitempty"`
	UserAvatar          *string   `json:"user_avatar,omitempty"`
}

// CreateReviewRequest é o payload para criar uma avaliação
type CreateReviewRequest struct {
	Rating              int     `json:"rating" binding:"required,min=1,max=5"`
	Title               *string `json:"title" binding:"omitempty,max=100"`
	Comment             *string `json:"comment" binding:"omitempty,max=500"`
	CleanlinessRating   *int    `json:"cleanliness_rating" binding:"omitempty,min=1,max=5"`
	AccessibilityRating *int    `json:"accessibility_rating" binding:"omitempty,min=1,max=5"`
	SpaciousnessRating  *int    `json:"spaciousness_rating" binding:"omitempty,min=1,max=5"`
}

// UpdateReviewRequest é o payload para atualizar uma avaliação
type UpdateReviewRequest struct {
	Rating              *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	Title               *string `json:"title" binding:"omitempty,max=100"`
	Comment             *string `json:"comment" binding:"omitempty,max=500"`
	CleanlinessRating   *int    `json:"cleanliness_rating" binding:"omitempty,min=1,max=5"`
	AccessibilityRating *int    `json:"accessibility_rating" binding:"omitempty,min=1,max=5"`
	SpaciousnessRating  *int    `json:"spaciousness_rating" binding:"omitempty,min=1,max=5"`
}

// BathroomRatingStats representa estatísticas de ratings
type BathroomRatingStats struct {
	BathroomID         int         `json:"bathroom_id"`
	TotalReviews       int         `json:"total_reviews"`
	AverageRating      float64     `json:"average_rating"`
	AvgCleanliness     float64     `json:"avg_cleanliness"`
	AvgAccessibility   float64     `json:"avg_accessibility"`
	AvgSpaciousness  float64     `json:"avg_spaciousness"`
	RatingDistribution map[int]int `json:"rating_distribution"`
}

// HelpfulVoteRequest é o payload para votar em utilidade
type HelpfulVoteRequest struct {
	IsHelpful *bool `json:"is_helpful" binding:"required"`
}
