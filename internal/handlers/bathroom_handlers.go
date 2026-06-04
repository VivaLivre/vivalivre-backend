package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/storage"
	"github.com/gin-gonic/gin"
)

// RequestBathroom handles the POST /api/bathrooms/request endpoint.
// It receives a multipart/form-data with bathroom suggestion data and a photo,
// uploads the photo to Supabase Storage, and inserts the bathroom with status 'pending'.
func RequestBathroom(c *gin.Context) {
	// --- 1. Extrair userID e validar spam (Rate Limiting) ---
	
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilizador não autenticado."})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Contar quantos banheiros este user criou nas últimas 24 horas
	var recentCount int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM bathrooms WHERE user_id = $1 AND created_at > NOW() - INTERVAL '24 hours'", userID).Scan(&recentCount)
	if err != nil {
		log.Printf("RequestBathroom: failed to count recent bathrooms: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao verificar limites do utilizador."})
		return
	}

	if recentCount >= 10 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Limite diário atingido. Você pode abrir até 10 requisições a cada 24 horas."})
		return
	}

	// --- 2. Parse & validate form fields ---

	name := c.PostForm("name")
	address := c.PostForm("address")
	latStr := c.PostForm("latitude")
	lngStr := c.PostForm("longitude")
	isAccessibleStr := c.PostForm("is_accessible")
	hasChangingTableStr := c.PostForm("has_changing_table")
	isFreeStr := c.PostForm("is_free")
	comment := c.PostForm("comment") // optional
	operatingHoursStr := c.PostForm("operating_hours") // optional

	// Required text fields
	if name == "" || address == "" || latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Os campos name, address, latitude e longitude são obrigatórios."})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Latitude inválida."})
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Longitude inválida."})
		return
	}

	isAccessible := parseBool(isAccessibleStr)
	hasChangingTable := parseBool(hasChangingTableStr)
	isFree := parseBool(isFreeStr)

	// --- 2. Extract and validate the photo file ---

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Uma foto é obrigatória para a sugestão."})
		return
	}
	defer file.Close()

	// Validate MIME type
	contentType := header.Header.Get("Content-Type")
	if !isAllowedImageType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de imagem não suportado. Use JPEG, PNG ou WebP."})
		return
	}

	// Validate file size (max 10 MB)
	const maxFileSize = 10 << 20 // 10 MB
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A foto excede o tamanho máximo de 10 MB."})
		return
	}

	// --- 3. Upload photo to Supabase Storage ---

	photoURL, err := storage.UploadToSupabase("bathroom_photos", file, header.Filename, contentType)
	if err != nil {
		log.Printf("RequestBathroom: failed to upload photo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao enviar a foto. Tente novamente."})
		return
	}

	if operatingHoursStr == "" {
		operatingHoursStr = `{"type": "unknown"}`
	}

	// --- 4. Insert into PostgreSQL ---

	query := `
		INSERT INTO bathrooms (name, address, location, is_accessible, has_changing_table, is_free, comment, photo_url, operating_hours, user_id)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`

	var bathroomID int
	var createdAt time.Time
	err = db.QueryRow(ctx, query,
		name,             // $1
		address,          // $2
		lng,              // $3 — ST_MakePoint(longitude, latitude)
		lat,              // $4
		isAccessible,     // $5
		hasChangingTable, // $6
		isFree,           // $7
		nilIfEmpty(comment), // $8
		photoURL,         // $9
		operatingHoursStr, // $10
		userID,           // $11
	).Scan(&bathroomID, &createdAt)

	if err != nil {
		log.Printf("RequestBathroom: failed to insert bathroom: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao registar a sugestão. Tente novamente."})
		return
	}

	// --- 5. Return success response ---

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Sugestão enviada com sucesso! Será analisada pela equipa.",
		"id":        bathroomID,
		"status":    "pending",
		"photo_url": photoURL,
		"created_at": createdAt,
	})
}

// --- Helper functions ---

// parseBool converts common boolean string representations to bool.
func parseBool(s string) bool {
	s = strconv.FormatBool(s == "true" || s == "1" || s == "yes")
	result, _ := strconv.ParseBool(s)
	return result
}

// isAllowedImageType checks if the MIME type is an accepted image format.
func isAllowedImageType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/jpg":
		return true
	}
	return false
}

// nilIfEmpty returns nil for empty strings, used for optional DB fields.
func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
