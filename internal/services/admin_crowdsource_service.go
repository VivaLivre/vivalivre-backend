package services

import (
	"context"

	"github.com/gabrieljose2004/vivalivre-backend/internal/models"
	"github.com/gabrieljose2004/vivalivre-backend/internal/repositories"
)

type AdminCrowdsourceService interface {
	GetAllReports(ctx context.Context, page, limit int) ([]models.BathroomReport, int, error)
	UpdateReportStatus(ctx context.Context, reportID, status string) error
	GetAllSuggestions(ctx context.Context, page, limit int) ([]models.BathroomSuggestion, int, error)
	UpdateSuggestionStatus(ctx context.Context, suggestionID, status string) error
}

type adminCrowdsourceService struct {
	repo repositories.AdminCrowdsourceRepository
}

func NewAdminCrowdsourceService(repo repositories.AdminCrowdsourceRepository) AdminCrowdsourceService {
	return &adminCrowdsourceService{repo: repo}
}

func (s *adminCrowdsourceService) GetAllReports(ctx context.Context, page, limit int) ([]models.BathroomReport, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return s.repo.GetAllReports(ctx, page, limit)
}

func (s *adminCrowdsourceService) UpdateReportStatus(ctx context.Context, reportID, status string) error {
	return s.repo.UpdateReportStatus(ctx, reportID, status)
}

func (s *adminCrowdsourceService) GetAllSuggestions(ctx context.Context, page, limit int) ([]models.BathroomSuggestion, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return s.repo.GetAllSuggestions(ctx, page, limit)
}

func (s *adminCrowdsourceService) UpdateSuggestionStatus(ctx context.Context, suggestionID, status string) error {
	return s.repo.UpdateSuggestionStatus(ctx, suggestionID, status)
}
