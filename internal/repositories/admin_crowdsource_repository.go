package repositories

import (
	"context"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminCrowdsourceRepository interface {
	GetAllReports(ctx context.Context, page, limit int) ([]models.BathroomReport, int, error)
	UpdateReportStatus(ctx context.Context, reportID, status string) error
	GetAllSuggestions(ctx context.Context, page, limit int) ([]models.BathroomSuggestion, int, error)
	UpdateSuggestionStatus(ctx context.Context, suggestionID, status string) error
}

type adminCrowdsourceRepository struct {
	db *pgxpool.Pool
}

func NewAdminCrowdsourceRepository(db *pgxpool.Pool) AdminCrowdsourceRepository {
	return &adminCrowdsourceRepository{db: db}
}

func (r *adminCrowdsourceRepository) GetAllReports(ctx context.Context, page, limit int) ([]models.BathroomReport, int, error) {
	offset := (page - 1) * limit
	query := `
		SELECT r.id, r.bathroom_id, b.name as bathroom_name, r.user_id, u.email as user_email, r.reason, r.description, r.status, r.created_at,
		       COUNT(*) OVER() as total_count
		FROM bathroom_reports r
		JOIN bathrooms b ON r.bathroom_id = b.id
		JOIN users u ON r.user_id = u.id
		ORDER BY r.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []models.BathroomReport
	var total int
	for rows.Next() {
		var r models.BathroomReport
		if err := rows.Scan(&r.ID, &r.BathroomID, &r.BathroomName, &r.UserID, &r.UserEmail, &r.Reason, &r.Description, &r.Status, &r.CreatedAt, &total); err != nil {
			continue
		}
		reports = append(reports, r)
	}

	if reports == nil {
		reports = []models.BathroomReport{}
	}
	return reports, total, nil
}

func (r *adminCrowdsourceRepository) UpdateReportStatus(ctx context.Context, reportID, status string) error {
	query := `UPDATE bathroom_reports SET status = $1 WHERE id = $2`
	tag, err := r.db.Exec(ctx, query, status, reportID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *adminCrowdsourceRepository) GetAllSuggestions(ctx context.Context, page, limit int) ([]models.BathroomSuggestion, int, error) {
	offset := (page - 1) * limit
	query := `
		SELECT s.id, s.bathroom_id, b.name as bathroom_name, s.user_id, u.email as user_email, s.suggested_updates, s.status, s.created_at,
		       COUNT(*) OVER() as total_count
		FROM bathroom_suggestions s
		JOIN bathrooms b ON s.bathroom_id = b.id
		JOIN users u ON s.user_id = u.id
		ORDER BY s.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var suggestions []models.BathroomSuggestion
	var total int
	for rows.Next() {
		var s models.BathroomSuggestion
		if err := rows.Scan(&s.ID, &s.BathroomID, &s.BathroomName, &s.UserID, &s.UserEmail, &s.SuggestedUpdates, &s.Status, &s.CreatedAt, &total); err != nil {
			continue
		}
		suggestions = append(suggestions, s)
	}

	if suggestions == nil {
		suggestions = []models.BathroomSuggestion{}
	}
	return suggestions, total, nil
}

func (r *adminCrowdsourceRepository) UpdateSuggestionStatus(ctx context.Context, suggestionID, status string) error {
	query := `UPDATE bathroom_suggestions SET status = $1 WHERE id = $2`
	tag, err := r.db.Exec(ctx, query, status, suggestionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}
