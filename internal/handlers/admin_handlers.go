package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gabrieljose2004/vivalivre-backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// GetPendingBathrooms fetches all bathrooms where status = 'pending'.
// Ordered by created_at ASC.
func GetPendingBathrooms(c *gin.Context) {
	db := database.GetDB()
	query := `
		SELECT id, name, address, photo_url, is_accessible, has_changing_table, is_free, comment, operating_hours, observations, created_at
		FROM bathrooms
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, query)
	if err != nil {
		log.Printf("GetPendingBathrooms: failed to query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending bathrooms"})
		return
	}
	defer rows.Close()

	var bathrooms []models.Bathroom
	for rows.Next() {
		var b models.Bathroom
		var photoUrl *string
		var isAccessible *bool
		var hasChangingTable *bool
		var isFree *bool
		var operatingHours []byte
		var observations *string

		err := rows.Scan(
			&b.ID,
			&b.Name,
			&b.Address,
			&photoUrl,
			&isAccessible,
			&hasChangingTable,
			&isFree,
			&b.Comment,
			&operatingHours,
			&observations,
			&b.CreatedAt,
		)
		if err != nil {
			log.Printf("GetPendingBathrooms: failed to scan row: %v", err)
			continue
		}

		if photoUrl != nil {
			b.PhotoURL = *photoUrl
		}
		if isAccessible != nil {
			b.IsAccessible = *isAccessible
		}
		if hasChangingTable != nil {
			b.HasChangingTable = *hasChangingTable
		}
		if isFree != nil {
			b.IsFree = *isFree
		}
		if operatingHours != nil {
			b.OperatingHours = operatingHours
		}
		if observations != nil {
			b.Observations = observations
		}
		b.Status = "pending" // implicitly known
		bathrooms = append(bathrooms, b)
	}

	if err = rows.Err(); err != nil {
		log.Printf("GetPendingBathrooms: rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing pending bathrooms"})
		return
	}

	// Return empty array instead of null if no records
	if bathrooms == nil {
		bathrooms = []models.Bathroom{}
	}

	c.JSON(http.StatusOK, bathrooms)
}

// UpdateBathroomStatus updates the status of a specific bathroom.
func UpdateBathroomStatus(c *gin.Context) {
	bathroomID := c.Param("id")

	var requestBody struct {
		Status       string  `json:"status" binding:"required"`
		Observations *string `json:"observations"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body. Expected 'status'."})
		return
	}

	// Validate status
	if requestBody.Status != "approved" && requestBody.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be 'approved' or 'rejected'."})
		return
	}

	db := database.GetDB()
	var query string
	var tag pgconn.CommandTag
	var err error

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if requestBody.Observations != nil {
		query = `
			UPDATE bathrooms
			SET status = $1, observations = $2
			WHERE id = $3
		`
		tag, err = db.Exec(ctx, query, requestBody.Status, *requestBody.Observations, bathroomID)
	} else {
		query = `
			UPDATE bathrooms
			SET status = $1
			WHERE id = $2
		`
		tag, err = db.Exec(ctx, query, requestBody.Status, bathroomID)
	}
	if err != nil {
		log.Printf("UpdateBathroomStatus: failed to update status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bathroom status"})
		return
	}

	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bathroom not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// GetAllReports fetches all reports with bathroom and user info
func GetAllReports(c *gin.Context) {
	db := database.GetDB()
	query := `
		SELECT r.id, r.bathroom_id, b.name as bathroom_name, r.user_id, u.email as user_email, r.reason, r.description, r.status, r.created_at
		FROM bathroom_reports r
		JOIN bathrooms b ON r.bathroom_id = b.id
		JOIN users u ON r.user_id = u.id
		ORDER BY r.created_at DESC
	`

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, query)
	if err != nil {
		log.Printf("GetAllReports: failed to query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
		return
	}
	defer rows.Close()

	var reports []models.BathroomReport
	for rows.Next() {
		var r models.BathroomReport
		if err := rows.Scan(&r.ID, &r.BathroomID, &r.BathroomName, &r.UserID, &r.UserEmail, &r.Reason, &r.Description, &r.Status, &r.CreatedAt); err != nil {
			log.Printf("GetAllReports: scan error: %v", err)
			continue
		}
		reports = append(reports, r)
	}

	if reports == nil {
		reports = []models.BathroomReport{}
	}

	c.JSON(http.StatusOK, reports)
}

// UpdateReportStatus updates the status of a specific report
func UpdateReportStatus(c *gin.Context) {
	reportID := c.Param("id")

	var req models.UpdateReportStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	query := `UPDATE bathroom_reports SET status = $1 WHERE id = $2`

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cmdTag, err := db.Exec(ctx, query, req.Status, reportID)
	if err != nil {
		log.Printf("UpdateReportStatus: failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report status"})
		return
	}

	if cmdTag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report status updated"})
}

// GetAllSuggestions fetches all suggestions with bathroom and user info
func GetAllSuggestions(c *gin.Context) {
	db := database.GetDB()
	query := `
		SELECT s.id, s.bathroom_id, b.name as bathroom_name, s.user_id, u.email as user_email, s.suggested_updates, s.status, s.created_at
		FROM bathroom_suggestions s
		JOIN bathrooms b ON s.bathroom_id = b.id
		JOIN users u ON s.user_id = u.id
		ORDER BY s.created_at DESC
	`

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, query)
	if err != nil {
		log.Printf("GetAllSuggestions: failed to query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
		return
	}
	defer rows.Close()

	var suggestions []models.BathroomSuggestion
	for rows.Next() {
		var s models.BathroomSuggestion
		if err := rows.Scan(&s.ID, &s.BathroomID, &s.BathroomName, &s.UserID, &s.UserEmail, &s.SuggestedUpdates, &s.Status, &s.CreatedAt); err != nil {
			log.Printf("GetAllSuggestions: scan error: %v", err)
			continue
		}
		suggestions = append(suggestions, s)
	}

	if suggestions == nil {
		suggestions = []models.BathroomSuggestion{}
	}

	c.JSON(http.StatusOK, suggestions)
}

// GetAdminBathrooms list bathrooms with pagination and search
func GetAdminBathrooms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var countQuery string
	var dataQuery string
	var args []interface{}
	var countArgs []interface{}

	if search != "" {
		countQuery = "SELECT COUNT(*) FROM bathrooms WHERE name ILIKE $1 OR address ILIKE $1"
		dataQuery = `
			SELECT id, name, address, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_accessible, has_changing_table, is_free, operating_hours, observations, comment, status, photo_url, created_at
			FROM bathrooms
			WHERE name ILIKE $1 OR address ILIKE $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		searchTerm := "%" + search + "%"
		countArgs = append(countArgs, searchTerm)
		args = append(args, searchTerm, limit, offset)
	} else {
		countQuery = "SELECT COUNT(*) FROM bathrooms"
		dataQuery = `
			SELECT id, name, address, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_accessible, has_changing_table, is_free, operating_hours, observations, comment, status, photo_url, created_at
			FROM bathrooms
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = append(args, limit, offset)
	}

	var total int
	err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		log.Printf("GetAdminBathrooms: failed to count bathrooms: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count bathrooms"})
		return
	}

	rows, err := db.Query(ctx, dataQuery, args...)
	if err != nil {
		log.Printf("GetAdminBathrooms: failed to fetch bathrooms: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bathrooms"})
		return
	}
	defer rows.Close()

	var bathrooms []models.Bathroom
	for rows.Next() {
		var b models.Bathroom
		var photoUrl *string
		var observations *string
		var comment *string
		var operatingHours []byte
		var isAccessible *bool
		var hasChangingTable *bool
		var isFree *bool

		err := rows.Scan(
			&b.ID,
			&b.Name,
			&b.Address,
			&b.Latitude,
			&b.Longitude,
			&isAccessible,
			&hasChangingTable,
			&isFree,
			&operatingHours,
			&observations,
			&comment,
			&b.Status,
			&photoUrl,
			&b.CreatedAt,
		)
		if err != nil {
			log.Printf("GetAdminBathrooms: failed to scan row: %v", err)
			continue
		}

		if photoUrl != nil {
			b.PhotoURL = *photoUrl
		}
		if observations != nil {
			b.Observations = observations
		}
		if comment != nil {
			b.Comment = comment
		}
		if operatingHours != nil {
			b.OperatingHours = operatingHours
		}
		if isAccessible != nil {
			b.IsAccessible = *isAccessible
		}
		if hasChangingTable != nil {
			b.HasChangingTable = *hasChangingTable
		}
		if isFree != nil {
			b.IsFree = *isFree
		}
		bathrooms = append(bathrooms, b)
	}

	if bathrooms == nil {
		bathrooms = []models.Bathroom{}
	}

	totalPages := (total + limit - 1) / limit

	response := models.PaginatedBathroomsResponse{
		Data: bathrooms,
		Meta: models.PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}

	log.Printf("GetAdminBathrooms: total=%d, returning %d items", total, len(bathrooms))

	c.JSON(http.StatusOK, response)
}

// CreateAdminBathroom allows admin to create a pre-approved bathroom
func CreateAdminBathroom(c *gin.Context) {
	name := c.PostForm("name")
	address := c.PostForm("address")
	latStr := c.PostForm("latitude")
	lngStr := c.PostForm("longitude")
	isAccessibleStr := c.PostForm("is_accessible")
	hasChangingTableStr := c.PostForm("has_changing_table")
	isFreeStr := c.PostForm("is_free")
	operatingHoursStr := c.PostForm("operating_hours")
	observations := c.PostForm("observations")
	photoUrl := c.PostForm("photo_url") // Could be passed as text

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

	// Check for file upload
	file, header, err := c.Request.FormFile("photo")
	var finalPhotoUrl *string
	if err == nil {
		defer file.Close()
		contentType := header.Header.Get("Content-Type")
		if isAllowedImageType(contentType) {
			const maxFileSize = 10 << 20 // 10 MB
			if header.Size <= maxFileSize {
				uploadedUrl, err := storage.UploadToSupabase("bathroom_photos", file, header.Filename, contentType)
				if err == nil {
					finalPhotoUrl = &uploadedUrl
				} else {
					log.Printf("CreateAdminBathroom: failed to upload photo: %v", err)
				}
			}
		}
	} else if photoUrl != "" {
		finalPhotoUrl = &photoUrl
	}

	if operatingHoursStr == "" {
		operatingHoursStr = `{"type": "unknown"}`
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	query := `
		INSERT INTO bathrooms (name, address, location, is_accessible, has_changing_table, is_free, operating_hours, observations, photo_url, status)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($4, $3), 4326), $5, $6, $7, $8::jsonb, $9, $10, 'approved')
		RETURNING id, created_at, status
	`
	
	var b models.Bathroom
	b.Name = name
	b.Address = address
	b.Latitude = lat
	b.Longitude = lng
	b.IsAccessible = isAccessible
	b.HasChangingTable = hasChangingTable
	b.IsFree = isFree

	err = db.QueryRow(ctx, query,
		name,
		address,
		lat,
		lng,
		isAccessible,
		hasChangingTable,
		isFree,
		operatingHoursStr,
		nilIfEmpty(observations),
		finalPhotoUrl,
	).Scan(&b.ID, &b.CreatedAt, &b.Status)

	if err != nil {
		log.Printf("CreateAdminBathroom: failed to insert: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bathroom: " + err.Error()})
		return
	}

	if finalPhotoUrl != nil {
		b.PhotoURL = *finalPhotoUrl
	}
	c.JSON(http.StatusCreated, b)
}

// UpdateAdminBathroom updates a bathroom partially
func UpdateAdminBathroom(c *gin.Context) {
	bathroomID := c.Param("id")

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var setClauses []string
	var args []interface{}
	argId := 1

	formValues := c.Request.MultipartForm.Value

	if val, ok := formValues["name"]; ok && len(val) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argId))
		args = append(args, val[0])
		argId++
	}
	if val, ok := formValues["address"]; ok && len(val) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("address = $%d", argId))
		args = append(args, val[0])
		argId++
	}

	latStr := ""
	lngStr := ""
	if val, ok := formValues["latitude"]; ok && len(val) > 0 {
		latStr = val[0]
	}
	if val, ok := formValues["longitude"]; ok && len(val) > 0 {
		lngStr = val[0]
	}

	if latStr != "" && lngStr != "" {
		lat, _ := strconv.ParseFloat(latStr, 64)
		lng, _ := strconv.ParseFloat(lngStr, 64)
		setClauses = append(setClauses, fmt.Sprintf("location = ST_SetSRID(ST_MakePoint($%d, $%d), 4326)", argId+1, argId))
		args = append(args, lat, lng)
		argId += 2
	} else if latStr != "" {
		lat, _ := strconv.ParseFloat(latStr, 64)
		setClauses = append(setClauses, fmt.Sprintf("location = ST_SetSRID(ST_MakePoint(ST_X(location::geometry), $%d), 4326)", argId))
		args = append(args, lat)
		argId++
	} else if lngStr != "" {
		lng, _ := strconv.ParseFloat(lngStr, 64)
		setClauses = append(setClauses, fmt.Sprintf("location = ST_SetSRID(ST_MakePoint($%d, ST_Y(location::geometry)), 4326)", argId))
		args = append(args, lng)
		argId++
	}

	if val, ok := formValues["is_accessible"]; ok && len(val) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("is_accessible = $%d", argId))
		args = append(args, parseBool(val[0]))
		argId++
	}
	if val, ok := formValues["has_changing_table"]; ok && len(val) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("has_changing_table = $%d", argId))
		args = append(args, parseBool(val[0]))
		argId++
	}
	if val, ok := formValues["is_free"]; ok && len(val) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("is_free = $%d", argId))
		args = append(args, parseBool(val[0]))
		argId++
	}
	if val, ok := formValues["operating_hours"]; ok && len(val) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("operating_hours = $%d::jsonb", argId))
		args = append(args, val[0])
		argId++
	}
	if val, ok := formValues["observations"]; ok && len(val) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("observations = $%d", argId))
		args = append(args, val[0])
		argId++
	}
	if val, ok := formValues["status"]; ok && len(val) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argId))
		args = append(args, val[0])
		argId++
	}

	// Handle Photo
	var newPhotoUrl string
	var deleteOldPhoto bool

	file, header, err := c.Request.FormFile("photo")
	if err == nil {
		defer file.Close()
		contentType := header.Header.Get("Content-Type")
		if isAllowedImageType(contentType) {
			const maxFileSize = 10 << 20 // 10 MB
			if header.Size <= maxFileSize {
				uploadedUrl, err := storage.UploadToSupabase("bathroom_photos", file, header.Filename, contentType)
				if err == nil {
					setClauses = append(setClauses, fmt.Sprintf("photo_url = $%d", argId))
					args = append(args, uploadedUrl)
					argId++
					newPhotoUrl = uploadedUrl
					deleteOldPhoto = true
				}
			}
		}
	} else if val, ok := formValues["photo_url"]; ok && len(val) > 0 {
		setClauses = append(setClauses, fmt.Sprintf("photo_url = $%d", argId))
		args = append(args, val[0])
		argId++
		newPhotoUrl = val[0]
		deleteOldPhoto = true
	}

	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields provided to update"})
		return
	}

	// Fetch old photo URL before update if we are going to replace it
	var oldPhotoUrl *string
	if deleteOldPhoto {
		_ = db.QueryRow(ctx, "SELECT photo_url FROM bathrooms WHERE id = $1", bathroomID).Scan(&oldPhotoUrl)
	}

	args = append(args, bathroomID)
	query := fmt.Sprintf("UPDATE bathrooms SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argId)

	tag, err := db.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("UpdateAdminBathroom: failed to update: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bathroom"})
		return
	}

	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bathroom not found"})
		return
	}

	// Delete old photo if it was updated and it was a Supabase URL
	if deleteOldPhoto && oldPhotoUrl != nil && *oldPhotoUrl != "" && *oldPhotoUrl != newPhotoUrl {
		if strings.Contains(*oldPhotoUrl, "/storage/v1/object/public/bathroom_photos/") {
			parts := strings.Split(*oldPhotoUrl, "/")
			objectName := parts[len(parts)-1]
			if objectName != "" {
				go func(obj string) {
					err := storage.DeleteFromSupabase("bathroom_photos", obj)
					if err != nil {
						log.Printf("Failed to delete old photo %s: %v", obj, err)
					} else {
						log.Printf("Successfully deleted old photo: %s", obj)
					}
				}(objectName)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bathroom updated successfully"})
}

// DeleteAdminBathroom deletes a bathroom entirely
func DeleteAdminBathroom(c *gin.Context) {
	bathroomID := c.Param("id")

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	query := "DELETE FROM bathrooms WHERE id = $1"
	tag, err := db.Exec(ctx, query, bathroomID)

	if err != nil {
		log.Printf("DeleteAdminBathroom: failed to delete: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bathroom"})
		return
	}

	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bathroom not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bathroom deleted successfully"})
}

// UpdateSuggestionStatus updates the status of a specific suggestion
func UpdateSuggestionStatus(c *gin.Context) {
	suggestionID := c.Param("id")

	var req models.UpdateSuggestionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	query := `UPDATE bathroom_suggestions SET status = $1 WHERE id = $2`

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cmdTag, err := db.Exec(ctx, query, req.Status, suggestionID)
	if err != nil {
		log.Printf("UpdateSuggestionStatus: failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suggestion status"})
		return
	}

	if cmdTag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suggestion not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Suggestion status updated"})
}
