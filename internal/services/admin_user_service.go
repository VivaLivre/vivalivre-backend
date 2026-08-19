package services

import (
	"context"
	"errors"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gabrieljose2004/vivalivre-backend/internal/repositories"
)

type AdminUserService interface {
	GetAdminUsers(ctx context.Context, page, limit int, search string) (models.PaginatedUsersResponse, error)
	UpdateAdminUserStatus(ctx context.Context, userID int, status string) error
}

type adminUserService struct {
	repo repositories.AdminUserRepository
}

func NewAdminUserService(repo repositories.AdminUserRepository) AdminUserService {
	return &adminUserService{repo: repo}
}

func (s *adminUserService) GetAdminUsers(ctx context.Context, page, limit int, search string) (models.PaginatedUsersResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := s.repo.GetAdminUsers(ctx, page, limit, search)
	if err != nil {
		return models.PaginatedUsersResponse{}, err
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	return models.PaginatedUsersResponse{
		Data: users,
		Meta: models.PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *adminUserService) UpdateAdminUserStatus(ctx context.Context, userID int, status string) error {
	validStatuses := map[string]bool{"active": true, "suspended": true, "banned": true}
	if !validStatuses[status] {
		return errors.New("invalid status")
	}
	return s.repo.UpdateAdminUserStatus(ctx, userID, status)
}
