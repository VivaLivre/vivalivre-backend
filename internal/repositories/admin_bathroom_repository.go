package repositories

import (
	"context"
	"fmt"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminBathroomRepository interface {
	GetAdminBathrooms(ctx context.Context, page, limit int, search string) (models.PaginatedBathroomsResponse, error)
	GetAdminBathroomByID(ctx context.Context, id string) (models.Bathroom, error)
	CreateAdminBathroom(ctx context.Context, b models.Bathroom, photoUrl *string) (models.Bathroom, error)
	UpdateAdminBathroom(ctx context.Context, id string, b models.Bathroom, photoUrl *string) error
	DeleteAdminBathroom(ctx context.Context, id string) error
}

type adminBathroomRepository struct {
	db *pgxpool.Pool
}

func NewAdminBathroomRepository(db *pgxpool.Pool) AdminBathroomRepository {
	return &adminBathroomRepository{db: db}
}

func (r *adminBathroomRepository) GetAdminBathrooms(ctx context.Context, page, limit int, search string) (models.PaginatedBathroomsResponse, error) {
	var countQuery, dataQuery string
	var args []interface{}
	var countArgs []interface{}
	offset := (page - 1) * limit

	if search != "" {
		countQuery = "SELECT COUNT(*) FROM bathrooms WHERE name ILIKE $1 OR address ILIKE $1"
		dataQuery = `
			SELECT id, name, address, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_accessible, has_changing_table, is_free, operating_hours, observations, comment, status, photo_url, created_at
			FROM bathrooms
			WHERE name ILIKE $1 OR address ILIKE $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		searchTerm := "%" + search + "%"
		countArgs = append(countArgs, searchTerm)
		args = append(args, searchTerm, limit, offset)
	} else {
		countQuery = "SELECT COUNT(*) FROM bathrooms"
		dataQuery = `
			SELECT id, name, address, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_accessible, has_changing_table, is_free, operating_hours, observations, comment, status, photo_url, created_at
			FROM bathrooms
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = append(args, limit, offset)
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return models.PaginatedBathroomsResponse{}, err
	}

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return models.PaginatedBathroomsResponse{}, err
	}
	defer rows.Close()

	var bathrooms []models.Bathroom
	for rows.Next() {
		var b models.Bathroom
		var photoUrl, observations, comment *string
		var isAccessible, hasChangingTable, isFree *bool
		var operatingHours []byte

		err := rows.Scan(
			&b.ID, &b.Name, &b.Address, &b.Latitude, &b.Longitude,
			&isAccessible, &hasChangingTable, &isFree, &operatingHours,
			&observations, &comment, &b.Status, &photoUrl, &b.CreatedAt,
		)
		if err != nil {
			continue
		}

		if photoUrl != nil {
			b.PhotoURL = *photoUrl
		}
		if observations != nil {
			b.Observations = observations
		}
		if comment != nil {
			b.Comment = comment
		}
		if operatingHours != nil {
			b.OperatingHours = operatingHours
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
		bathrooms = append(bathrooms, b)
	}

	if bathrooms == nil {
		bathrooms = []models.Bathroom{}
	}

	totalPages := (total + limit - 1) / limit

	return models.PaginatedBathroomsResponse{
		Data: bathrooms,
		Meta: models.PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

func (r *adminBathroomRepository) GetAdminBathroomByID(ctx context.Context, id string) (models.Bathroom, error) {
	query := `
		SELECT id, name, address, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_accessible, has_changing_table, is_free, operating_hours, observations, comment, status, photo_url, created_at
		FROM bathrooms
		WHERE id = $1
	`
	var b models.Bathroom
	var photoUrl, observations, comment *string
	var isAccessible, hasChangingTable, isFree *bool
	var operatingHours []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.Name, &b.Address, &b.Latitude, &b.Longitude,
		&isAccessible, &hasChangingTable, &isFree, &operatingHours,
		&observations, &comment, &b.Status, &photoUrl, &b.CreatedAt,
	)

	if err != nil {
		return b, err
	}

	if photoUrl != nil {
		b.PhotoURL = *photoUrl
	}
	if observations != nil {
		b.Observations = observations
	}
	if comment != nil {
		b.Comment = comment
	}
	if operatingHours != nil {
		b.OperatingHours = operatingHours
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
	return b, nil
}

func (r *adminBathroomRepository) CreateAdminBathroom(ctx context.Context, b models.Bathroom, photoUrl *string) (models.Bathroom, error) {
	query := `
		INSERT INTO bathrooms (name, address, location, is_accessible, has_changing_table, is_free, operating_hours, observations, photo_url, status)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, $6, $7, $8::jsonb, $9, $10, 'approved')
		RETURNING id, created_at, status
	`
	
	err := r.db.QueryRow(ctx, query,
		b.Name, b.Address, b.Longitude, b.Latitude,
		b.IsAccessible, b.HasChangingTable, b.IsFree,
		b.OperatingHours, b.Observations, photoUrl,
	).Scan(&b.ID, &b.CreatedAt, &b.Status)

	return b, err
}

func (r *adminBathroomRepository) UpdateAdminBathroom(ctx context.Context, id string, b models.Bathroom, photoUrl *string) error {
	query := `
		UPDATE bathrooms
		SET name = $1, address = $2, location = ST_SetSRID(ST_MakePoint($3, $4), 4326), 
		    is_accessible = $5, has_changing_table = $6, is_free = $7, 
		    operating_hours = $8::jsonb, observations = $9, status = $10
	`
	args := []interface{}{
		b.Name, b.Address, b.Longitude, b.Latitude,
		b.IsAccessible, b.HasChangingTable, b.IsFree,
		b.OperatingHours, b.Observations, b.Status,
	}
	
	if photoUrl != nil {
		query += `, photo_url = $11 WHERE id = $12`
		args = append(args, photoUrl, id)
	} else {
		query += ` WHERE id = $11`
		args = append(args, id)
	}

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bathroom not found")
	}
	return nil
}

func (r *adminBathroomRepository) DeleteAdminBathroom(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM bathrooms WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bathroom not found")
	}
	return nil
}
