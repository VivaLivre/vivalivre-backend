package admin

import (
	"net/http"

	"github.com/gabrieljose2004/vivalivre-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service services.AdminService
}

func NewAdminHandler(service services.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) GetDashboardOverview(c *gin.Context) {
	ctx := c.Request.Context()
	metrics, err := h.service.GetDashboardMetrics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard metrics"})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (h *AdminHandler) GetPendingBathrooms(c *gin.Context) {
	ctx := c.Request.Context()
	bathrooms, err := h.service.GetPendingBathrooms(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending bathrooms"})
		return
	}
	c.JSON(http.StatusOK, bathrooms)
}

func (h *AdminHandler) UpdateBathroomStatus(c *gin.Context) {
	bathroomID := c.Param("id")

	var requestBody struct {
		Status       string  `json:"status" binding:"required"`
		Observations *string `json:"observations"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body. Expected 'status'."})
		return
	}

	ctx := c.Request.Context()
	err := h.service.UpdateBathroomStatus(ctx, bathroomID, requestBody.Status, requestBody.Observations)
	if err != nil {
		if err.Error() == "invalid status" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be 'approved' or 'rejected'."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bathroom status"})
		return
	}

	// Aqui deveríamos lidar com "Delete orphaned photo if rejected" 
	// mas para não complicar por agora (que já está a ser processado),
	// devolvemos o StatusOK. A remoção de foto passará para a camada de serviço futuramente.

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}
