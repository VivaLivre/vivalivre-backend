package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
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
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.DefaultQuery("radius", "5000") // Default 5km

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing lat or lng parameters"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lat parameter"})
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lng parameter"})
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid radius parameter"})
		return
	}

	db := database.GetDB()
	query := `
		SELECT id, name, address, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_accessible, has_changing_table, is_free, photo_url, status, created_at,
		ST_Distance(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance
		FROM bathrooms
		WHERE ST_DWithin(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $3)
		AND status = 'approved'
		ORDER BY distance
		LIMIT 50;
	`
	
	// Note: ST_MakePoint takes (longitude, latitude) -> ($1, $2) must be (lng, lat)
	rows, err := db.Query(context.Background(), query, lng, lat, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch nearby bathrooms"})
		return
	}
	defer rows.Close()

	bathrooms := []models.Bathroom{}
	for rows.Next() {
		var b models.Bathroom
		var photoURL *string
		if err := rows.Scan(
			&b.ID, &b.Name, &b.Address, &b.Latitude, &b.Longitude, 
			&b.IsAccessible, &b.HasChangingTable, &b.IsFree, 
			&photoURL, &b.Status, &b.CreatedAt, &b.Distance,
		); err != nil {
			log.Printf("GetNearbyBathrooms Scan error: %v", err)
			continue
		}
		if photoURL != nil {
			b.PhotoURL = *photoURL
		}
		bathrooms = append(bathrooms, b)
	}

	c.JSON(http.StatusOK, bathrooms)
}

// GetHealthEntries returns health data for the logged in user
func GetHealthEntries(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	
	db := database.GetDB()
	rows, err := db.Query(context.Background(), `SELECT id, user_id, type, severity, description, COALESCE(symptoms, '{}'), entry_date FROM health_entries WHERE user_id = $1 ORDER BY entry_date DESC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch health entries"})
		return
	}
	defer rows.Close()

	var entries []models.HealthEntry
	for rows.Next() {
		var entry models.HealthEntry
		if err := rows.Scan(&entry.ID, &entry.UserID, &entry.Type, &entry.Severity, &entry.Description, &entry.Symptoms, &entry.EntryDate); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, entries)
}

// CreateHealthEntry handles creating a new health entry
func CreateHealthEntry(c *gin.Context) {
	userID := c.MustGet("userID").(int)

	var req models.CreateHealthEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Symptoms == nil {
		req.Symptoms = []string{}
	}

	db := database.GetDB()
	query := `
		INSERT INTO health_entries (user_id, type, severity, description, symptoms) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, user_id, type, severity, description, COALESCE(symptoms, '{}'), entry_date
	`
	var entry models.HealthEntry
	err := db.QueryRow(context.Background(), query, userID, req.Type, req.Severity, req.Description, req.Symptoms).
		Scan(&entry.ID, &entry.UserID, &entry.Type, &entry.Severity, &entry.Description, &entry.Symptoms, &entry.EntryDate)
	
	if err != nil {
		log.Printf("CreateHealthEntry ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create health entry"})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// DeleteHealthEntry deletes a specific health entry
func DeleteHealthEntry(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	entryID := c.Param("id")

	db := database.GetDB()
	query := `DELETE FROM health_entries WHERE id = $1 AND user_id = $2`
	result, err := db.Exec(context.Background(), query, entryID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete health entry"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Health entry not found or not authorized"})
		return
	}

	c.Status(http.StatusNoContent)
}
