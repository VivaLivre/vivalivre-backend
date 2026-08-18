package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gabrieljose2004/vivalivre-backend/internal/auth"
	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gabrieljose2004/vivalivre-backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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

	comorbiditiesJSON, err := json.Marshal(req.Comorbidities)
	if err != nil || len(req.Comorbidities) == 0 {
		comorbiditiesJSON = []byte("[]")
	}

	db := database.GetDB()
	var user models.User
	query := `INSERT INTO users (name, email, password_hash, cpf, date_of_birth, gender, weight, height, clinical_condition, comorbidities) 
	          VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, $10) 
	          RETURNING id, name, email, avatar_url, height, weight, date_of_birth::text, clinical_condition, comorbidities, created_at`
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var comorbsBytes []byte
	err = db.QueryRow(ctx, query,
		req.Name, req.Email, hash, req.CPF, req.DateOfBirth, req.Gender, req.Weight, req.Height, req.ClinicalCondition, string(comorbiditiesJSON),
	).Scan(
		&user.ID, &user.Name, &user.Email, &user.AvatarURL, &user.Height, &user.Weight, &user.DateOfBirth, &user.ClinicalCondition, &comorbsBytes, &user.CreatedAt,
	)

	if err == nil {
		_ = json.Unmarshal(comorbsBytes, &user.Comorbidities)
	}

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
	var status *string
	query := `SELECT id, name, email, password_hash, status, avatar_url, height, weight, birth_date, condition, created_at FROM users WHERE email = $1`
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := db.QueryRow(ctx, query, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &hash, &status, &user.AvatarURL, &user.Height, &user.Weight, &user.BirthDate, &user.Condition, &user.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if status != nil && (*status == "banned" || *status == "suspended") {
		c.JSON(http.StatusForbidden, gin.H{"error": "O utilizador está banido ou suspenso."})
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
	query := `SELECT id, name, email, avatar_url, height, weight, birth_date, condition, created_at FROM users WHERE id = $1`
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	err := db.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.AvatarURL, &user.Height, &user.Weight, &user.BirthDate, &user.Condition, &user.CreatedAt,
	)
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
		SELECT id, name, address, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_accessible, has_changing_table, is_free, photo_url, status, operating_hours, observations, created_at,
		ST_Distance(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance
		FROM bathrooms
		WHERE ST_DWithin(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $3)
		AND status = 'approved'
		ORDER BY distance
		LIMIT 50;
	`
	
	// Note: ST_MakePoint takes (longitude, latitude) -> ($1, $2) must be (lng, lat)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, query, lng, lat, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch nearby bathrooms"})
		return
	}
	defer rows.Close()

	bathrooms := []models.Bathroom{}
	for rows.Next() {
		var b models.Bathroom
		var photoURL *string
		var operatingHours []byte
		var observations *string
		if err := rows.Scan(
			&b.ID, &b.Name, &b.Address, &b.Latitude, &b.Longitude, 
			&b.IsAccessible, &b.HasChangingTable, &b.IsFree, 
			&photoURL, &b.Status, &operatingHours, &observations, &b.CreatedAt, &b.Distance,
		); err != nil {
			log.Printf("GetNearbyBathrooms Scan error: %v", err)
			continue
		}
		if photoURL != nil {
			b.PhotoURL = *photoURL
		}
		if operatingHours != nil {
			b.OperatingHours = operatingHours
		}
		if observations != nil {
			b.Observations = observations
		}
		bathrooms = append(bathrooms, b)
	}

	c.JSON(http.StatusOK, bathrooms)
}

// GetHealthEntries returns health data for the logged in user
func GetHealthEntries(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	dateFilter := c.Query("date")

	query := `SELECT id, user_id, type, severity, description, COALESCE(symptoms, '{}'), entry_date FROM health_entries WHERE user_id = $1`
	var args []interface{}
	args = append(args, userID)

	if dateFilter == "today" {
		query += ` AND entry_date >= CURRENT_DATE AND entry_date < CURRENT_DATE + INTERVAL '1 day'`
	} else if dateFilter != "" {
		query += ` AND entry_date >= $2::date AND entry_date < $2::date + INTERVAL '1 day'`
		args = append(args, dateFilter)
	}
	query += ` ORDER BY entry_date DESC`

	db := database.GetDB()
	rows, err := db.Query(context.Background(), query, args...)
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

// GoogleLogin handles authentication via Google Sign-In
func GoogleLogin(c *gin.Context) {
	var req models.GoogleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	payload, err := auth.VerifyGoogleToken(ctx, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google token: " + err.Error()})
		return
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token payload does not contain a valid email"})
		return
	}

	name, _ := payload.Claims["name"].(string)
	if name == "" {
		name = email
	}

	db := database.GetDB()
	var user models.User
	var status string

	// Buscar utilizador na base de dados por email
	querySelect := `SELECT id, name, email, status, avatar_url, height, weight, birth_date, condition, created_at FROM users WHERE email = $1`
	err = db.QueryRow(ctx, querySelect, email).Scan(
		&user.ID, &user.Name, &user.Email, &status, &user.AvatarURL, &user.Height, &user.Weight, &user.BirthDate, &user.Condition, &user.CreatedAt,
	)

	if err == nil {
		// O utilizador existe
		if status == "banned" || status == "suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "O utilizador está banido ou suspenso."})
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
		return
	}

	// Se for erro de NoRows, criamos um novo utilizador
	if err != nil && !errors.Is(err, pgx.ErrNoRows) && !errors.Is(err, context.Canceled) {
		log.Printf("GoogleLogin db select error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao consultar base de dados"})
		return
	}

	// Criar novo utilizador (password_hash fica NULL)
	queryInsert := `INSERT INTO users (name, email, password_hash) VALUES ($1, $2, NULL) RETURNING id, name, email, avatar_url, height, weight, birth_date, condition, created_at`
	err = db.QueryRow(ctx, queryInsert, name, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.AvatarURL, &user.Height, &user.Weight, &user.BirthDate, &user.Condition, &user.CreatedAt,
	)
	if err != nil {
		log.Printf("GoogleLogin db insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível criar o utilizador"})
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

// UpdateProfile handles updating the user's profile information
func UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilizador não autenticado."})
		return
	}

	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O email é obrigatório."})
		return
	}

	var height *int
	if hStr := c.PostForm("height"); hStr != "" {
		if h, err := strconv.Atoi(hStr); err == nil {
			height = &h
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Altura inválida."})
			return
		}
	}

	var weight *float64
	if wStr := c.PostForm("weight"); wStr != "" {
		if w, err := strconv.ParseFloat(wStr, 64); err == nil {
			weight = &w
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Peso inválido."})
			return
		}
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if email is already in use by another user
	var count int
	errCount := db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1 AND id != $2", email, userID).Scan(&count)
	if errCount == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Este email já está a ser utilizado por outro utilizador."})
		return
	}

	var avatarURL *string
	file, header, err := c.Request.FormFile("photo")
	if err == nil {
		defer file.Close()
		contentType := header.Header.Get("Content-Type")
		if !isAllowedImageType(contentType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de imagem não suportado. Use JPEG, PNG ou WebP."})
			return
		}

		const maxFileSize = 5 << 20 // 5 MB
		if header.Size > maxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A foto excede o tamanho máximo de 5 MB."})
			return
		}

		// Upload photo to Supabase Storage using the "avatars" bucket directly
		url, uploadErr := storage.UploadToSupabase("avatars", file, header.Filename, contentType)
		if uploadErr != nil {
			log.Printf("UpdateProfile: failed to upload photo: %v", uploadErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao enviar a foto de perfil para o Supabase."})
			return
		}
		avatarURL = &url
	}

	var query string
	var errUpdate error
	if avatarURL != nil {
		query = `UPDATE users SET email = $1, height = $2, weight = $3, avatar_url = $4 WHERE id = $5`
		_, errUpdate = db.Exec(ctx, query, email, height, weight, *avatarURL, userID)
	} else {
		query = `UPDATE users SET email = $1, height = $2, weight = $3 WHERE id = $4`
		_, errUpdate = db.Exec(ctx, query, email, height, weight, userID)
	}

	if errUpdate != nil {
		log.Printf("UpdateProfile database error: %v", errUpdate)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar perfil na base de dados."})
		return
	}

	var user models.User
	querySelect := `SELECT id, name, email, avatar_url, height, weight, birth_date, condition, created_at FROM users WHERE id = $1`
	err = db.QueryRow(ctx, querySelect, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.AvatarURL, &user.Height, &user.Weight, &user.BirthDate, &user.Condition, &user.CreatedAt,
	)
	if err != nil {
		log.Printf("UpdateProfile select error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar dados atualizados do utilizador."})
		return
	}

	c.JSON(http.StatusOK, user)
}

