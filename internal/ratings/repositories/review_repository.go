package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/VivaLivre/vivalivre-backend/internal/ratings/models"
)

// ReviewRepository gerencia operações de reviews no banco de dados
type ReviewRepository struct {
	db *sqlx.DB
}

// NewReviewRepository cria uma nova instância do repository
func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// CreateReview cria uma nova avaliação
func (r *ReviewRepository) CreateReview(ctx context.Context, bathroomID, userID uuid.UUID, req *models.CreateReviewRequest) (*models.BathroomReview, error) {
	review := &models.BathroomReview{
		ID:                  uuid.New(),
		BathroomID:          bathroomID,
		UserID:              userID,
		Rating:              req.Rating,
		Title:               req.Title,
		Comment:             req.Comment,
		CleanlinessRating:   req.CleanlinessRating,
		AccessibilityRating: req.AccessibilityRating,
		SpaciosunessRating:  req.SpaciosunessRating,
		Status:              string(models.ReviewStatusPending),
		CreatedAt:           ctx.Value("now").(time.Time),
		UpdatedAt:           ctx.Value("now").(time.Time),
	}

	query := `
		INSERT INTO bathroom_reviews (
			id, bathroom_id, user_id, rating, title, comment,
			cleanliness_rating, accessibility_rating, spaciousness_rating,
			status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		review.ID, review.BathroomID, review.UserID, review.Rating,
		review.Title, review.Comment, review.CleanlinessRating,
		review.AccessibilityRating, review.SpaciosunessRating,
		review.Status, review.CreatedAt, review.UpdatedAt,
	)

	if err != nil {
		log.Printf("[ReviewRepository] Error creating review: %v", err)
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	log.Printf("[ReviewRepository] Review created: %s", review.ID)
	return review, nil
}

// GetReviewByID busca uma avaliação pelo ID
func (r *ReviewRepository) GetReviewByID(ctx context.Context, reviewID uuid.UUID) (*models.BathroomReview, error) {
	review := &models.BathroomReview{}

	query := `
		SELECT id, bathroom_id, user_id, rating, title, comment,
		       cleanliness_rating, accessibility_rating, spaciousness_rating,
		       helpful_count, unhelpful_count, status, created_at, updated_at
		FROM bathroom_reviews
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, review, query, reviewID)
	if err != nil {
		log.Printf("[ReviewRepository] Error getting review: %v", err)
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	// Buscar fotos
	photos, err := r.GetReviewPhotos(ctx, reviewID)
	if err == nil {
		review.Photos = photos
	}

	return review, nil
}

// GetReviewsByBathroom busca todas as avaliações de um banheiro
func (r *ReviewRepository) GetReviewsByBathroom(ctx context.Context, bathroomID uuid.UUID, sort string, limit, offset int) ([]*models.BathroomReview, int, error) {
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Determinar ordenação
	orderBy := "created_at DESC"
	switch sort {
	case "helpful":
		orderBy = "(helpful_count - unhelpful_count) DESC"
	case "rating":
		orderBy = "rating DESC"
	default:
		orderBy = "created_at DESC"
	}

	// Contar total
	countQuery := `SELECT COUNT(*) FROM bathroom_reviews WHERE bathroom_id = $1 AND status = 'approved'`
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, bathroomID)
	if err != nil {
		log.Printf("[ReviewRepository] Error counting reviews: %v", err)
		return nil, 0, fmt.Errorf("failed to count reviews: %w", err)
	}

	// Buscar reviews
	query := fmt.Sprintf(`
		SELECT id, bathroom_id, user_id, rating, title, comment,
		       cleanliness_rating, accessibility_rating, spaciousness_rating,
		       helpful_count, unhelpful_count, status, created_at, updated_at
		FROM bathroom_reviews
		WHERE bathroom_id = $1 AND status = 'approved'
		ORDER BY %s
		LIMIT $2 OFFSET $3
	`, orderBy)

	var reviews []*models.BathroomReview
	err = r.db.SelectContext(ctx, &reviews, query, bathroomID, limit, offset)
	if err != nil {
		log.Printf("[ReviewRepository] Error getting reviews: %v", err)
		return nil, 0, fmt.Errorf("failed to get reviews: %w", err)
	}

	// Buscar fotos para cada review
	for _, review := range reviews {
		photos, err := r.GetReviewPhotos(ctx, review.ID)
		if err == nil {
			review.Photos = photos
		}
	}

	return reviews, total, nil
}

// UpdateReview atualiza uma avaliação
func (r *ReviewRepository) UpdateReview(ctx context.Context, reviewID uuid.UUID, req *models.UpdateReviewRequest) (*models.BathroomReview, error) {
	query := `
		UPDATE bathroom_reviews
		SET rating = COALESCE($1, rating),
		    title = COALESCE($2, title),
		    comment = COALESCE($3, comment),
		    cleanliness_rating = COALESCE($4, cleanliness_rating),
		    accessibility_rating = COALESCE($5, accessibility_rating),
		    spaciousness_rating = COALESCE($6, spaciousness_rating),
		    updated_at = NOW()
		WHERE id = $7
		RETURNING id, bathroom_id, user_id, rating, title, comment,
		          cleanliness_rating, accessibility_rating, spaciousness_rating,
		          helpful_count, unhelpful_count, status, created_at, updated_at
	`

	review := &models.BathroomReview{}
	err := r.db.GetContext(ctx, review, query,
		req.Rating, req.Title, req.Comment,
		req.CleanlinessRating, req.AccessibilityRating, req.SpaciosunessRating,
		reviewID,
	)

	if err != nil {
		log.Printf("[ReviewRepository] Error updating review: %v", err)
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	log.Printf("[ReviewRepository] Review updated: %s", reviewID)
	return review, nil
}

// DeleteReview deleta uma avaliação
func (r *ReviewRepository) DeleteReview(ctx context.Context, reviewID uuid.UUID) error {
	query := `DELETE FROM bathroom_reviews WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, reviewID)
	if err != nil {
		log.Printf("[ReviewRepository] Error deleting review: %v", err)
		return fmt.Errorf("failed to delete review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	log.Printf("[ReviewRepository] Review deleted: %s", reviewID)
	return nil
}

// GetReviewPhotos busca fotos de uma avaliação
func (r *ReviewRepository) GetReviewPhotos(ctx context.Context, reviewID uuid.UUID) ([]string, error) {
	query := `SELECT photo_url FROM review_photos WHERE review_id = $1 ORDER BY uploaded_at ASC`

	var photos []string
	err := r.db.SelectContext(ctx, &photos, query, reviewID)
	if err != nil {
		log.Printf("[ReviewRepository] Error getting photos: %v", err)
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}

	return photos, nil
}

// AddReviewPhoto adiciona uma foto a uma avaliação
func (r *ReviewRepository) AddReviewPhoto(ctx context.Context, reviewID uuid.UUID, photoURL string) (*models.ReviewPhoto, error) {
	photo := &models.ReviewPhoto{
		ID:        uuid.New(),
		ReviewID:  reviewID,
		PhotoURL:  photoURL,
		UploadedAt: time.Now(),
	}

	query := `
		INSERT INTO review_photos (id, review_id, photo_url, uploaded_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query, photo.ID, photo.ReviewID, photo.PhotoURL, photo.UploadedAt)
	if err != nil {
		log.Printf("[ReviewRepository] Error adding photo: %v", err)
		return nil, fmt.Errorf("failed to add photo: %w", err)
	}

	return photo, nil
}

// VoteHelpful registra um voto de utilidade
func (r *ReviewRepository) VoteHelpful(ctx context.Context, reviewID, userID uuid.UUID, isHelpful bool) error {
	query := `
		INSERT INTO review_helpful_votes (id, review_id, user_id, is_helpful, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (review_id, user_id) DO UPDATE
		SET is_helpful = $4
	`

	_, err := r.db.ExecContext(ctx, query, uuid.New(), reviewID, userID, isHelpful)
	if err != nil {
		log.Printf("[ReviewRepository] Error voting helpful: %v", err)
		return fmt.Errorf("failed to vote: %w", err)
	}

	// Atualizar contadores
	if isHelpful {
		r.db.ExecContext(ctx, `UPDATE bathroom_reviews SET helpful_count = helpful_count + 1 WHERE id = $1`, reviewID)
	} else {
		r.db.ExecContext(ctx, `UPDATE bathroom_reviews SET unhelpful_count = unhelpful_count + 1 WHERE id = $1`, reviewID)
	}

	return nil
}

// GetBathroomRatingStats busca estatísticas de ratings de um banheiro
func (r *ReviewRepository) GetBathroomRatingStats(ctx context.Context, bathroomID uuid.UUID) (*models.BathroomRatingStats, error) {
	stats := &models.BathroomRatingStats{
		BathroomID:         bathroomID,
		RatingDistribution: make(map[int]int),
	}

	query := `
		SELECT bathroom_id, total_reviews, average_rating,
		       avg_cleanliness, avg_accessibility, avg_spaciousness
		FROM bathroom_rating_stats
		WHERE bathroom_id = $1
	`

	err := r.db.GetContext(ctx, stats, query, bathroomID)
	if err != nil {
		log.Printf("[ReviewRepository] Error getting rating stats: %v", err)
		return nil, fmt.Errorf("failed to get rating stats: %w", err)
	}

	// Buscar distribuição de ratings
	distQuery := `
		SELECT rating, COUNT(*) as count
		FROM bathroom_reviews
		WHERE bathroom_id = $1 AND status = 'approved'
		GROUP BY rating
	`

	rows, err := r.db.QueryContext(ctx, distQuery, bathroomID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rating, count int
			if err := rows.Scan(&rating, &count); err == nil {
				stats.RatingDistribution[rating] = count
			}
		}
	}

	return stats, nil
}

// CheckUserReviewExists verifica se o utilizador já avaliou o banheiro
func (r *ReviewRepository) CheckUserReviewExists(ctx context.Context, bathroomID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM bathroom_reviews WHERE bathroom_id = $1 AND user_id = $2)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, bathroomID, userID)
	if err != nil {
		log.Printf("[ReviewRepository] Error checking review existence: %v", err)
		return false, fmt.Errorf("failed to check review: %w", err)
	}

	return exists, nil
}
