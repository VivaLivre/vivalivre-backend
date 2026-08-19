package repositories

import (
	"context"
	"strings"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminUserRepository interface {
	GetAdminUsers(ctx context.Context, page, limit int, search string) ([]models.AdminUser, int, error)
	UpdateAdminUserStatus(ctx context.Context, userID int, status string) error
}

type adminUserRepository struct {
	db *pgxpool.Pool
}

func NewAdminUserRepository(db *pgxpool.Pool) AdminUserRepository {
	return &adminUserRepository{db: db}
}

func (r *adminUserRepository) GetAdminUsers(ctx context.Context, page, limit int, search string) ([]models.AdminUser, int, error) {
	offset := (page - 1) * limit
	var total int
	var users []models.AdminUser

	if search != "" {
		likePattern := "%" + strings.ToLower(search) + "%"
		if err := r.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM users WHERE LOWER(name) LIKE $1 OR LOWER(email) LIKE $1`,
			likePattern).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err := r.db.Query(ctx,
			`SELECT id, name, email, COALESCE(role,'user'), COALESCE(status,'active'), created_at
			 FROM users
			 WHERE LOWER(name) LIKE $1 OR LOWER(email) LIKE $1
			 ORDER BY created_at DESC
			 LIMIT $2 OFFSET $3`,
			likePattern, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()
		for rows.Next() {
			var u models.AdminUser
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt); err != nil {
				continue
			}
			users = append(users, u)
		}
	} else {
		if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err := r.db.Query(ctx,
			`SELECT id, name, email, COALESCE(role,'user'), COALESCE(status,'active'), created_at
			 FROM users
			 ORDER BY created_at DESC
			 LIMIT $1 OFFSET $2`,
			limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()
		for rows.Next() {
			var u models.AdminUser
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt); err != nil {
				continue
			}
			users = append(users, u)
		}
	}

	if users == nil {
		users = []models.AdminUser{}
	}

	return users, total, nil
}

func (r *adminUserRepository) UpdateAdminUserStatus(ctx context.Context, userID int, status string) error {
	tag, err := r.db.Exec(ctx, `UPDATE users SET status = $1 WHERE id = $2`, status, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}
