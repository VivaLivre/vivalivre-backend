package repositories

import (
	"context"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository interface {
	GetDashboardMetrics(ctx context.Context) (models.DashboardOverview, error)
	GetPendingBathrooms(ctx context.Context) ([]models.Bathroom, error)
	UpdateBathroomStatus(ctx context.Context, id string, status string, observations *string) error
}

type adminRepository struct {
	db *pgxpool.Pool
}

func NewAdminRepository(db *pgxpool.Pool) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetDashboardMetrics(ctx context.Context) (models.DashboardOverview, error) {
	var metrics models.DashboardOverview

	// Queries ignoram erros para métricas básicas
	r.db.QueryRow(ctx, "SELECT count(*) FROM bathrooms").Scan(&metrics.TotalLocations)
	r.db.QueryRow(ctx, "SELECT count(*) FROM bathrooms WHERE created_at >= date_trunc('month', current_date)").Scan(&metrics.LocationsThisMonth)
	r.db.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&metrics.ActiveUsers)
	r.db.QueryRow(ctx, "SELECT count(*) FROM users WHERE created_at >= date_trunc('month', current_date)").Scan(&metrics.UsersThisMonth)
	r.db.QueryRow(ctx, "SELECT count(*) FROM bathrooms WHERE status = 'pending'").Scan(&metrics.PendingSuggestions)

	r.db.QueryRow(ctx, "SELECT count(*) FROM bathroom_reviews WHERE status = 'pending'").Scan(&metrics.PendingReviews)
	r.db.QueryRow(ctx, "SELECT count(*) FROM bathrooms WHERE status = 'approved' AND created_at >= current_date").Scan(&metrics.ApprovalsToday)

	metrics.ApprovalRate = 100.0 // Simplification
	metrics.WeeklyActivity = []map[string]interface{}{}
	metrics.RecentActivities = []map[string]interface{}{}

	return metrics, nil
}

func (r *adminRepository) GetPendingBathrooms(ctx context.Context) ([]models.Bathroom, error) {
	query := `
		SELECT id, name, address, photo_url, is_accessible, has_changing_table, is_free, comment, operating_hours, observations, created_at
		FROM bathrooms
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bathrooms []models.Bathroom
	for rows.Next() {
		var b models.Bathroom
		var photoUrl, observations *string
		var isAccessible, hasChangingTable, isFree *bool
		var operatingHours []byte

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
		b.Status = "pending"
		bathrooms = append(bathrooms, b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if bathrooms == nil {
		bathrooms = []models.Bathroom{}
	}

	return bathrooms, nil
}

func (r *adminRepository) UpdateBathroomStatus(ctx context.Context, id string, status string, observations *string) error {
	var err error
	if observations != nil {
		query := `
			UPDATE bathrooms
			SET status = $1, observations = $2
			WHERE id = $3
		`
		_, err = r.db.Exec(ctx, query, status, *observations, id)
	} else {
		query := `
			UPDATE bathrooms
			SET status = $1
			WHERE id = $2
		`
		_, err = r.db.Exec(ctx, query, status, id)
	}
	return err
}
