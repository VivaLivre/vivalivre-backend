package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
)

// CreateBathroomReport handles user reports for a bathroom
func CreateBathroomReport(c *gin.Context) {
	// Extract user ID from AuthMiddleware
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(int)

	// Extract bathroom ID from URL parameter
	bathroomIDStr := c.Param("bathroom_id")
	bathroomID, err := strconv.Atoi(bathroomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bathroom ID"})
		return
	}

	var req models.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	query := `
		INSERT INTO bathroom_reports (bathroom_id, user_id, reason, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, status, created_at
	`

	var report models.BathroomReport
	report.BathroomID = bathroomID
	report.UserID = userID
	report.Reason = req.Reason
	report.Description = req.Description

	err = db.QueryRow(c.Request.Context(), query, bathroomID, userID, req.Reason, req.Description).
		Scan(&report.ID, &report.Status, &report.CreatedAt)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// CreateBathroomSuggestion handles user suggestions for a bathroom
func CreateBathroomSuggestion(c *gin.Context) {
	// Extract user ID from AuthMiddleware
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(int)

	// Extract bathroom ID from URL parameter
	bathroomIDStr := c.Param("bathroom_id")
	bathroomID, err := strconv.Atoi(bathroomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bathroom ID"})
		return
	}

	var req models.CreateSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	query := `
		INSERT INTO bathroom_suggestions (bathroom_id, user_id, suggested_updates)
		VALUES ($1, $2, $3)
		RETURNING id, status, created_at
	`

	var suggestion models.BathroomSuggestion
	suggestion.BathroomID = bathroomID
	suggestion.UserID = userID
	suggestion.SuggestedUpdates = req.SuggestedUpdates

	err = db.QueryRow(c.Request.Context(), query, bathroomID, userID, req.SuggestedUpdates).
		Scan(&suggestion.ID, &suggestion.Status, &suggestion.CreatedAt)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create suggestion"})
		return
	}

	c.JSON(http.StatusCreated, suggestion)
}
