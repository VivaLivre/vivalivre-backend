package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// CreateReview cria uma nova avaliação
// POST /api/bathrooms/:bathroom_id/reviews
func CreateReview(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	bathroomIDStr := c.Param("bathroom_id")
	bathroomID, err := strconv.Atoi(bathroomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bathroom ID"})
		return
	}

	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Inserir review
	query := `
		INSERT INTO bathroom_reviews 
		(bathroom_id, user_id, rating, title, comment, cleanliness_rating, accessibility_rating, spaciousness_rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, helpful_count, unhelpful_count, status, created_at, updated_at
	`

	var review models.BathroomReview
	review.BathroomID = bathroomID
	review.UserID = userID
	review.Rating = req.Rating
	review.Title = req.Title
	review.Comment = req.Comment
	review.CleanlinessRating = req.CleanlinessRating
	review.AccessibilityRating = req.AccessibilityRating
	review.SpaciousnessRating = req.SpaciousnessRating

	err = db.QueryRow(ctx, query,
		bathroomID, userID, req.Rating, req.Title, req.Comment,
		req.CleanlinessRating, req.AccessibilityRating, req.SpaciousnessRating,
	).Scan(&review.ID, &review.HelpfulCount, &review.UnhelpfulCount, &review.Status, &review.CreatedAt, &review.UpdatedAt)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "You have already reviewed this bathroom"})
			return
		}
		log.Printf("CreateReview error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// ListReviews lista avaliações de um banheiro
// GET /api/bathrooms/:bathroom_id/reviews?sort=recent&limit=10&offset=0
func ListReviews(c *gin.Context) {
	bathroomIDStr := c.Param("bathroom_id")
	bathroomID, err := strconv.Atoi(bathroomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bathroom ID"})
		return
	}

	sort := c.DefaultQuery("sort", "recent")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Determinar ordenação
	orderBy := "r.created_at DESC"
	if sort == "helpful" {
		orderBy = "r.helpful_count DESC, r.created_at DESC"
	} else if sort == "rating" {
		orderBy = "r.rating DESC, r.created_at DESC"
	}

	// Buscar reviews com contagem e média usando Window Functions
	query := `
		SELECT r.id, r.bathroom_id, r.user_id, r.rating, r.title, r.comment,
		       r.cleanliness_rating, r.accessibility_rating, r.spaciousness_rating,
		       r.helpful_count, r.unhelpful_count, r.status, r.created_at, r.updated_at,
		       COUNT(*) OVER() as total_count,
		       AVG(r.rating) OVER() as avg_rating
		FROM bathroom_reviews r
		WHERE r.bathroom_id = $1 AND (r.status = 'approved' OR r.status IS NULL)
		ORDER BY ` + orderBy + `
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(ctx, query, bathroomID, limit, offset)
	if err != nil {
		log.Printf("ListReviews query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}
	defer rows.Close()

	reviews := []models.BathroomReview{}
	var total int
	var avgRating float64

	for rows.Next() {
		var r models.BathroomReview
		err := rows.Scan(
			&r.ID, &r.BathroomID, &r.UserID, &r.Rating, &r.Title, &r.Comment,
			&r.CleanlinessRating, &r.AccessibilityRating, &r.SpaciousnessRating,
			&r.HelpfulCount, &r.UnhelpfulCount, &r.Status, &r.CreatedAt, &r.UpdatedAt,
			&total, &avgRating,
		)
		if err != nil {
			continue
		}
		reviews = append(reviews, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews":        reviews,
		"total":          total,
		"average_rating": avgRating,
	})
}

// GetReview obtém uma avaliação específica
// GET /api/reviews/:review_id
func GetReview(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, bathroom_id, user_id, rating, title, comment,
		       cleanliness_rating, accessibility_rating, spaciousness_rating,
		       helpful_count, unhelpful_count, status, created_at, updated_at
		FROM bathroom_reviews
		WHERE id = $1
	`

	var review models.BathroomReview
	err = db.QueryRow(ctx, query, reviewID).Scan(
		&review.ID, &review.BathroomID, &review.UserID, &review.Rating, &review.Title, &review.Comment,
		&review.CleanlinessRating, &review.AccessibilityRating, &review.SpaciousnessRating,
		&review.HelpfulCount, &review.UnhelpfulCount, &review.Status, &review.CreatedAt, &review.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}
	if err != nil {
		log.Printf("GetReview error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// GetRatingStats retorna estatísticas de ratings de um banheiro
// GET /api/bathrooms/:bathroom_id/rating-stats
func GetRatingStats(c *gin.Context) {
	bathroomIDStr := c.Param("bathroom_id")
	bathroomID, err := strconv.Atoi(bathroomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bathroom ID"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Buscar estatísticas agregadas
	query := `
		SELECT 
			COUNT(*) as total_reviews,
			COALESCE(AVG(rating), 0) as average_rating,
			COALESCE(AVG(cleanliness_rating), 0) as avg_cleanliness,
			COALESCE(AVG(accessibility_rating), 0) as avg_accessibility,
			COALESCE(AVG(spaciousness_rating), 0) as avg_spaciousness
		FROM bathroom_reviews
		WHERE bathroom_id = $1 AND (status = 'approved' OR status IS NULL)
	`

	var stats models.BathroomRatingStats
	stats.BathroomID = bathroomID

	err = db.QueryRow(ctx, query, bathroomID).Scan(
		&stats.TotalReviews,
		&stats.AverageRating,
		&stats.AvgCleanliness,
		&stats.AvgAccessibility,
		&stats.AvgSpaciousness,
	)

	if err != nil {
		log.Printf("GetRatingStats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	// Buscar distribuição de ratings
	distQuery := `
		SELECT rating, COUNT(*) as count
		FROM bathroom_reviews
		WHERE bathroom_id = $1 AND (status = 'approved' OR status IS NULL)
		GROUP BY rating
	`

	rows, err := db.Query(ctx, distQuery, bathroomID)
	if err == nil {
		defer rows.Close()
		stats.RatingDistribution = make(map[int]int)
		for rows.Next() {
			var rating, count int
			if err := rows.Scan(&rating, &count); err == nil {
				stats.RatingDistribution[rating] = count
			}
		}
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateReview atualiza uma avaliação
// PUT /api/reviews/:review_id
func UpdateReview(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Verificar ownership
	var ownerID int
	checkQuery := `SELECT user_id FROM bathroom_reviews WHERE id = $1`
	err = db.QueryRow(ctx, checkQuery, reviewID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check review"})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own reviews"})
		return
	}

	// Atualizar review
	updateQuery := `
		UPDATE bathroom_reviews
		SET rating = COALESCE($1, rating),
		    title = COALESCE($2, title),
		    comment = COALESCE($3, comment),
		    cleanliness_rating = COALESCE($4, cleanliness_rating),
		    accessibility_rating = COALESCE($5, accessibility_rating),
		    spaciousness_rating = COALESCE($6, spaciousness_rating),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING id, bathroom_id, user_id, rating, title, comment,
		          cleanliness_rating, accessibility_rating, spaciousness_rating,
		          helpful_count, unhelpful_count, status, created_at, updated_at
	`

	var review models.BathroomReview
	err = db.QueryRow(ctx, updateQuery,
		req.Rating, req.Title, req.Comment,
		req.CleanlinessRating, req.AccessibilityRating, req.SpaciousnessRating,
		reviewID,
	).Scan(
		&review.ID, &review.BathroomID, &review.UserID, &review.Rating, &review.Title, &review.Comment,
		&review.CleanlinessRating, &review.AccessibilityRating, &review.SpaciousnessRating,
		&review.HelpfulCount, &review.UnhelpfulCount, &review.Status, &review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		log.Printf("UpdateReview error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// DeleteReview deleta uma avaliação
// DELETE /api/reviews/:review_id
func DeleteReview(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Verificar ownership e deletar
	deleteQuery := `DELETE FROM bathroom_reviews WHERE id = $1 AND user_id = $2`
	result, err := db.Exec(ctx, deleteQuery, reviewID, userID)
	if err != nil {
		log.Printf("DeleteReview error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found or not authorized"})
		return
	}

	c.Status(http.StatusNoContent)
}

// VoteHelpful registra um voto de utilidade em uma avaliação
// POST /api/reviews/:review_id/helpful
func VoteHelpful(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var req models.HelpfulVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// UPSERT do voto
	voteQuery := `
		INSERT INTO review_helpful_votes (review_id, user_id, is_helpful)
		VALUES ($1, $2, $3)
		ON CONFLICT (review_id, user_id) 
		DO UPDATE SET is_helpful = EXCLUDED.is_helpful
	`

	_, err = db.Exec(ctx, voteQuery, reviewID, userID, req.IsHelpful)
	if err != nil {
		log.Printf("VoteHelpful error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	// Atualizar contadores na review
	updateQuery := `
		UPDATE bathroom_reviews
		SET 
			helpful_count = (SELECT COUNT(*) FROM review_helpful_votes WHERE review_id = $1 AND is_helpful = true),
			unhelpful_count = (SELECT COUNT(*) FROM review_helpful_votes WHERE review_id = $1 AND is_helpful = false)
		WHERE id = $1
		RETURNING helpful_count, unhelpful_count
	`

	var helpfulCount, unhelpfulCount int
	err = db.QueryRow(ctx, updateQuery, reviewID).Scan(&helpfulCount, &unhelpfulCount)
	if err != nil {
		log.Printf("VoteHelpful update counts error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update counts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"helpful_count":   helpfulCount,
		"unhelpful_count": unhelpfulCount,
	})
}
