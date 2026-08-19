package models

import (
	"encoding/json"
	"time"
)

// Bathroom represents a bathroom location
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
	AverageRating       float64         `json:"average_rating"`
	ReviewsCount        int             `json:"reviews_count"`
	PhotoURL            string          `json:"photo_url,omitempty"`
	Distance            float64         `json:"distance,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
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

// PaginatedBathroomsResponse is the response for paginated bathrooms
type PaginatedBathroomsResponse struct {
	Data []Bathroom     `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
