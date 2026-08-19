package services

import (
	"context"
	"errors"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gabrieljose2004/vivalivre-backend/internal/repositories"
)

type AdminService interface {
	GetDashboardMetrics(ctx context.Context) (models.DashboardOverview, error)
	GetPendingBathrooms(ctx context.Context) ([]models.Bathroom, error)
	UpdateBathroomStatus(ctx context.Context, id string, status string, observations *string) error
}

type adminService struct {
	repo repositories.AdminRepository
}

func NewAdminService(repo repositories.AdminRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) GetDashboardMetrics(ctx context.Context) (models.DashboardOverview, error) {
	return s.repo.GetDashboardMetrics(ctx)
}

func (s *adminService) GetPendingBathrooms(ctx context.Context) ([]models.Bathroom, error) {
	return s.repo.GetPendingBathrooms(ctx)
}

func (s *adminService) UpdateBathroomStatus(ctx context.Context, id string, status string, observations *string) error {
	// Business logic / Validations
	if status != "approved" && status != "rejected" {
		return errors.New("invalid status")
	}
	return s.repo.UpdateBathroomStatus(ctx, id, status, observations)
}
