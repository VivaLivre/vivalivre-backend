package services

import (
	"context"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gabrieljose2004/vivalivre-backend/internal/repositories"
)

type AdminBathroomService interface {
	GetAdminBathrooms(ctx context.Context, page, limit int, search string) (models.PaginatedBathroomsResponse, error)
	GetAdminBathroomByID(ctx context.Context, id string) (models.Bathroom, error)
	CreateAdminBathroom(ctx context.Context, b models.Bathroom, photoUrl *string) (models.Bathroom, error)
	UpdateAdminBathroom(ctx context.Context, id string, b models.Bathroom, photoUrl *string) error
	DeleteAdminBathroom(ctx context.Context, id string) error
}

type adminBathroomService struct {
	repo repositories.AdminBathroomRepository
}

func NewAdminBathroomService(repo repositories.AdminBathroomRepository) AdminBathroomService {
	return &adminBathroomService{repo: repo}
}

func (s *adminBathroomService) GetAdminBathrooms(ctx context.Context, page, limit int, search string) (models.PaginatedBathroomsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return s.repo.GetAdminBathrooms(ctx, page, limit, search)
}

func (s *adminBathroomService) GetAdminBathroomByID(ctx context.Context, id string) (models.Bathroom, error) {
	return s.repo.GetAdminBathroomByID(ctx, id)
}

func (s *adminBathroomService) CreateAdminBathroom(ctx context.Context, b models.Bathroom, photoUrl *string) (models.Bathroom, error) {
	return s.repo.CreateAdminBathroom(ctx, b, photoUrl)
}

func (s *adminBathroomService) UpdateAdminBathroom(ctx context.Context, id string, b models.Bathroom, photoUrl *string) error {
	return s.repo.UpdateAdminBathroom(ctx, id, b, photoUrl)
}

func (s *adminBathroomService) DeleteAdminBathroom(ctx context.Context, id string) error {
	return s.repo.DeleteAdminBathroom(ctx, id)
}
