package admin

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gabrieljose2004/vivalivre-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type AdminCrowdsourceHandler struct {
	service services.AdminCrowdsourceService
}

func NewAdminCrowdsourceHandler(service services.AdminCrowdsourceService) *AdminCrowdsourceHandler {
	return &AdminCrowdsourceHandler{service: service}
}

func (h *AdminCrowdsourceHandler) GetAllReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	ctx := c.Request.Context()
	reports, total, err := h.service.GetAllReports(ctx, page, limit)
	if err != nil {
		log.Printf("GetAllReports: failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reports,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *AdminCrowdsourceHandler) UpdateReportStatus(c *gin.Context) {
	reportID := c.Param("id")

	var req models.UpdateReportStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	err := h.service.UpdateReportStatus(ctx, reportID, req.Status)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
			return
		}
		log.Printf("UpdateReportStatus: failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report status updated"})
}

func (h *AdminCrowdsourceHandler) GetAllSuggestions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	ctx := c.Request.Context()
	suggestions, total, err := h.service.GetAllSuggestions(ctx, page, limit)
	if err != nil {
		log.Printf("GetAllSuggestions: failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  suggestions,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *AdminCrowdsourceHandler) UpdateSuggestionStatus(c *gin.Context) {
	suggestionID := c.Param("id")

	var req models.UpdateSuggestionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	err := h.service.UpdateSuggestionStatus(ctx, suggestionID, req.Status)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Suggestion not found"})
			return
		}
		log.Printf("UpdateSuggestionStatus: failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suggestion status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Suggestion status updated"})
}
