package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/VivaLivre/vivalivre-backend/internal/ratings/models"
	"github.com/VivaLivre/vivalivre-backend/internal/ratings/repositories"
)

// ReviewHandler gerencia requisições de reviews
type ReviewHandler struct {
	reviewRepo *repositories.ReviewRepository
}

// NewReviewHandler cria uma nova instância do handler
func NewReviewHandler(reviewRepo *repositories.ReviewRepository) *ReviewHandler {
	return &ReviewHandler{
		reviewRepo: reviewRepo,
	}
}

// CreateReview cria uma nova avaliação
// POST /api/bathrooms/:bathroom_id/reviews
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	bathroomIDStr := c.Param("bathroom_id")
	bathroomID, err := uuid.Parse(bathroomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bathroom ID"})
		return
	}

	// Obter user_id do contexto (JWT middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verificar se utilizador já avaliou este banheiro
	exists, err := h.reviewRepo.CheckUserReviewExists(c.Request.Context(), bathroomID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check review"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already reviewed this bathroom"})
		return
	}

	// Criar review
	review, err := h.reviewRepo.CreateReview(c.Request.Context(), bathroomID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// GetReview obtém uma avaliação específica
// GET /api/reviews/:review_id
func (h *ReviewHandler) GetReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	review, err := h.reviewRepo.GetReviewByID(c.Request.Context(), reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// ListReviews lista avaliações de um banheiro
// GET /api/bathrooms/:bathroom_id/reviews?sort=recent&limit=10&offset=0
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	bathroomIDStr := c.Param("bathroom_id")
	bathroomID, err := uuid.Parse(bathroomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bathroom ID"})
		return
	}

	// Parse query parameters
	sort := c.DefaultQuery("sort", "recent")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	reviews, total, err := h.reviewRepo.GetReviewsByBathroom(c.Request.Context(), bathroomID, sort, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reviews"})
		return
	}

	// Obter estatísticas
	stats, _ := h.reviewRepo.GetBathroomRatingStats(c.Request.Context(), bathroomID)

	c.JSON(http.StatusOK, gin.H{
		"reviews":          reviews,
		"total":            total,
		"average_rating":   stats.AverageRating,
		"rating_stats":     stats,
	})
}

// UpdateReview atualiza uma avaliação
// PUT /api/reviews/:review_id
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	// Obter user_id do contexto
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verificar se review pertence ao utilizador
	review, err := h.reviewRepo.GetReviewByID(c.Request.Context(), reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	if review.UserID.String() != userIDStr.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own reviews"})
		return
	}

	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedReview, err := h.reviewRepo.UpdateReview(c.Request.Context(), reviewID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	c.JSON(http.StatusOK, updatedReview)
}

// DeleteReview deleta uma avaliação
// DELETE /api/reviews/:review_id
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	// Obter user_id do contexto
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verificar se review pertence ao utilizador
	review, err := h.reviewRepo.GetReviewByID(c.Request.Context(), reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	if review.UserID.String() != userIDStr.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own reviews"})
		return
	}

	err = h.reviewRepo.DeleteReview(c.Request.Context(), reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// VoteHelpful marca uma avaliação como útil/inútil
// POST /api/reviews/:review_id/helpful
func (h *ReviewHandler) VoteHelpful(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	// Obter user_id do contexto
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.HelpfulVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.reviewRepo.VoteHelpful(c.Request.Context(), reviewID, userID, req.IsHelpful)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to vote"})
		return
	}

	// Obter review atualizado
	review, _ := h.reviewRepo.GetReviewByID(c.Request.Context(), reviewID)

	c.JSON(http.StatusOK, gin.H{
		"helpful_count":   review.HelpfulCount,
		"unhelpful_count": review.UnhelpfulCount,
	})
}

// GetBathroomRatingStats obtém estatísticas de ratings
// GET /api/bathrooms/:bathroom_id/rating-stats
func (h *ReviewHandler) GetBathroomRatingStats(c *gin.Context) {
	bathroomIDStr := c.Param("bathroom_id")
	bathroomID, err := uuid.Parse(bathroomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bathroom ID"})
		return
	}

	stats, err := h.reviewRepo.GetBathroomRatingStats(c.Request.Context(), bathroomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rating stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// RegisterRoutes registra as rotas de reviews
func RegisterRoutes(router *gin.Engine, handler *ReviewHandler) {
	// Reviews
	bathrooms := router.Group("/api/bathrooms/:bathroom_id")
	{
		bathrooms.POST("/reviews", handler.CreateReview)
		bathrooms.GET("/reviews", handler.ListReviews)
		bathrooms.GET("/rating-stats", handler.GetBathroomRatingStats)
	}

	reviews := router.Group("/api/reviews")
	{
		reviews.GET("/:review_id", handler.GetReview)
		reviews.PUT("/:review_id", handler.UpdateReview)
		reviews.DELETE("/:review_id", handler.DeleteReview)
		reviews.POST("/:review_id/helpful", handler.VoteHelpful)
	}
}
