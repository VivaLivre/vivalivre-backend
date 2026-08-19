package admin

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gabrieljose2004/vivalivre-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type AdminUserHandler struct {
	service services.AdminUserService
}

func NewAdminUserHandler(service services.AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{service: service}
}

func (h *AdminUserHandler) GetAdminUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := strings.TrimSpace(c.Query("search"))

	ctx := c.Request.Context()
	response, err := h.service.GetAdminUsers(ctx, page, limit, search)
	if err != nil {
		log.Printf("GetAdminUsers: failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AdminUserHandler) UpdateAdminUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body. Expected 'status'."})
		return
	}

	ctx := c.Request.Context()
	err = h.service.UpdateAdminUserStatus(ctx, userID, req.Status)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err.Error() == "invalid status" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be 'active', 'suspended', or 'banned'."})
			return
		}
		log.Printf("UpdateAdminUserStatus: failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("User status updated to '%s'", req.Status)})
}
