package models

import "time"

// User represents a registered user
type User struct {
	ID                int        `json:"id"`
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	PasswordHash      string     `json:"-"`
	AvatarURL         *string    `json:"avatar_url,omitempty"`
	Height            *float64   `json:"height,omitempty"`
	Weight            *float64   `json:"weight,omitempty"`
	BirthDate         *time.Time `json:"birth_date,omitempty"`
	DateOfBirth       *string    `json:"date_of_birth,omitempty"`
	CPF               *string    `json:"cpf,omitempty"`
	Gender            *string    `json:"gender,omitempty"`
	ClinicalCondition *string    `json:"clinical_condition,omitempty"`
	Comorbidities     []string   `json:"comorbidities,omitempty"`
	Role              string     `json:"role,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// AdminUser represents a user as seen by an admin
type AdminUser struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateUserStatusRequest is the payload for updating a user's status
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// PaginatedUsersResponse is the response for paginated users
type PaginatedUsersResponse struct {
	Data []AdminUser    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
