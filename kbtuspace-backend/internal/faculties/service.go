package faculties

import (
	"context"

	"kbtuspace-backend/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAllFaculties(ctx context.Context) ([]models.Faculty, error) {
	return s.repo.GetAllFaculties(ctx)
}
