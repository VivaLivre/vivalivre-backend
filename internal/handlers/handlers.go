package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gabrieljose2004/vivalivre-backend/internal/auth"
	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// Register handles user registration
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	db := database.GetDB()
	var user models.User
	query := `INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id, name, email, created_at`
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = db.QueryRow(ctx, query, req.Name, req.Email, hash).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "Este email já foi utilizado."})
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Tempo esgotado ao criar a conta."})
			return
		}

		log.Printf("failed to register user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível criar a conta."})
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user authentication
func Login(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var user models.User
	var hash string
	query := `SELECT id, name, email, password_hash, created_at FROM users WHERE email = $1`
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := db.QueryRow(ctx, query, req.Email).Scan(&user.ID, &user.Name, &user.Email, &hash, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !auth.CheckPassword(req.Password, hash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetMe returns the current logged in user
func GetMe(c *gin.Context) {
	userID := c.MustGet("userID").(int)

	db := database.GetDB()
	var user models.User
	query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
	err := db.QueryRow(context.Background(), query, userID).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetNearbyBathrooms handles searching for bathrooms using PostGIS
func GetNearbyBathrooms(c *gin.Context) {
	// Example implementation - logic would call get_nearby_bathrooms function
	// Placeholder for now
	c.JSON(http.StatusOK, gin.H{"message": "PostGIS proximity search logic integrated"})
}

// GetHealthEntries returns health data for the logged in user
func GetHealthEntries(c *gin.Context) {
	userID := c.MustGet("userID").(int)

	db := database.GetDB()
	rows, err := db.Query(context.Background(), `SELECT id, user_id, type, description, entry_date FROM health_entries WHERE user_id = $1`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch health entries"})
		return
	}
	defer rows.Close()

	var entries []models.HealthEntry
	for rows.Next() {
		var entry models.HealthEntry
		if err := rows.Scan(&entry.ID, &entry.UserID, &entry.Type, &entry.Description, &entry.EntryDate); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, entries)
}
