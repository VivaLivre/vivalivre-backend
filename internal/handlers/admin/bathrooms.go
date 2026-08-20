package admin

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gabrieljose2004/vivalivre-backend/internal/services"
	"github.com/gabrieljose2004/vivalivre-backend/internal/storage"
	"github.com/gin-gonic/gin"
)

type AdminBathroomHandler struct {
	service services.AdminBathroomService
}

func NewAdminBathroomHandler(service services.AdminBathroomService) *AdminBathroomHandler {
	return &AdminBathroomHandler{service: service}
}

func (h *AdminBathroomHandler) GetAdminBathrooms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	ctx := c.Request.Context()
	response, err := h.service.GetAdminBathrooms(ctx, page, limit, search)
	if err != nil {
		log.Printf("GetAdminBathrooms: failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bathrooms"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func parseBool(val string) *bool {
	if val == "" {
		return nil
	}
	b := val == "true" || val == "1"
	return &b
}

func isAllowedImageType(contentType string) bool {
	allowed := []string{"image/jpeg", "image/png", "image/webp"}
	for _, a := range allowed {
		if a == contentType {
			return true
		}
	}
	return false
}

func nilIfEmpty(val string) *string {
	if strings.TrimSpace(val) == "" {
		return nil
	}
	return &val
}

func (h *AdminBathroomHandler) CreateAdminBathroom(c *gin.Context) {
	name := c.PostForm("name")
	address := c.PostForm("address")
	latStr := c.PostForm("latitude")
	lngStr := c.PostForm("longitude")
	isAccessibleStr := c.PostForm("is_accessible")
	hasChangingTableStr := c.PostForm("has_changing_table")
	isFreeStr := c.PostForm("is_free")
	operatingHoursStr := c.PostForm("operating_hours")
	observations := c.PostForm("observations")
	photoUrl := c.PostForm("photo_url")

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

	var b models.Bathroom
	b.Name = name
	b.Address = address
	b.Latitude = lat
	b.Longitude = lng
	if isAccessible != nil {
		b.IsAccessible = *isAccessible
	}
	if hasChangingTable != nil {
		b.HasChangingTable = *hasChangingTable
	}
	if isFree != nil {
		b.IsFree = *isFree
	}
	b.OperatingHours = []byte(operatingHoursStr)
	if obs := nilIfEmpty(observations); obs != nil {
		b.Observations = obs
	}

	ctx := c.Request.Context()
	newB, err := h.service.CreateAdminBathroom(ctx, b, finalPhotoUrl)
	if err != nil {
		log.Printf("CreateAdminBathroom: failed to insert: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bathroom"})
		return
	}

	if finalPhotoUrl != nil {
		newB.PhotoURL = *finalPhotoUrl
	}
	c.JSON(http.StatusCreated, newB)
}

func (h *AdminBathroomHandler) UpdateAdminBathroom(c *gin.Context) {
	bathroomID := c.Param("id")
	ctx := c.Request.Context()

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// 1. Fetch current bathroom
	currentB, err := h.service.GetAdminBathroomByID(ctx, bathroomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bathroom not found"})
		return
	}

	formValues := c.Request.MultipartForm.Value

	if val, ok := formValues["name"]; ok && len(val) > 0 {
		currentB.Name = val[0]
	}
	if val, ok := formValues["address"]; ok && len(val) > 0 {
		currentB.Address = val[0]
	}

	latStr := ""
	lngStr := ""
	if val, ok := formValues["latitude"]; ok && len(val) > 0 {
		latStr = val[0]
	}
	if val, ok := formValues["longitude"]; ok && len(val) > 0 {
		lngStr = val[0]
	}

	if latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			currentB.Latitude = lat
		}
	}
	if lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
			currentB.Longitude = lng
		}
	}

	if val, ok := formValues["is_accessible"]; ok && len(val) > 0 {
		if parsed := parseBool(val[0]); parsed != nil {
			currentB.IsAccessible = *parsed
		}
	}
	if val, ok := formValues["has_changing_table"]; ok && len(val) > 0 {
		if parsed := parseBool(val[0]); parsed != nil {
			currentB.HasChangingTable = *parsed
		}
	}
	if val, ok := formValues["is_free"]; ok && len(val) > 0 {
		if parsed := parseBool(val[0]); parsed != nil {
			currentB.IsFree = *parsed
		}
	}
	if val, ok := formValues["operating_hours"]; ok && len(val) > 0 {
		currentB.OperatingHours = []byte(val[0])
	}
	if val, ok := formValues["observations"]; ok && len(val) > 0 {
		currentB.Observations = nilIfEmpty(val[0])
	}
	if val, ok := formValues["status"]; ok && len(val) > 0 {
		currentB.Status = val[0]
	}

	// Handle Photo
	var finalPhotoUrl *string
	file, header, err := c.Request.FormFile("photo")
	if err == nil {
		defer file.Close()
		contentType := header.Header.Get("Content-Type")
		if isAllowedImageType(contentType) {
			const maxFileSize = 10 << 20 // 10 MB
			if header.Size <= maxFileSize {
				uploadedUrl, err := storage.UploadToSupabase("bathroom_photos", file, header.Filename, contentType)
				if err == nil {
					finalPhotoUrl = &uploadedUrl
				}
			}
		}
	}

	if err := h.service.UpdateAdminBathroom(ctx, bathroomID, currentB, finalPhotoUrl); err != nil {
		log.Printf("UpdateAdminBathroom: failed to update: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bathroom"})
		return
	}

	if finalPhotoUrl != nil {
		currentB.PhotoURL = *finalPhotoUrl
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bathroom updated successfully",
		"data":    currentB,
	})
}

func (h *AdminBathroomHandler) DeleteAdminBathroom(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if err := h.service.DeleteAdminBathroom(ctx, id); err != nil {
		if err.Error() == "bathroom not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bathroom not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bathroom"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bathroom deleted"})
}

func deletePhotoByURL(photoURL string, bathroomID string) {
	if photoURL == "" {
		return
	}
	parts := strings.Split(photoURL, "/")
	objectName := parts[len(parts)-1]
	if objectName != "" {
		err := storage.DeleteFromSupabase("bathroom_photos", objectName)
		if err != nil {
			log.Printf("Failed to delete orphaned photo %s for bathroom %s: %v", objectName, bathroomID, err)
		} else {
			log.Printf("Deleted orphaned photo %s for bathroom %s", objectName, bathroomID)
		}
	}
}
