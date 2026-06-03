package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
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
