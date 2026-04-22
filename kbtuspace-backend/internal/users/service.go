package users

import (
	"database/sql"
	"errors"

	"kbtuspace-backend/internal/auth"
	"kbtuspace-backend/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetProfile(userID int) (*models.User, error) {
	return s.repo.GetByID(userID)
}

func (s *Service) UpdateProfile(userID int, input models.UpdateProfileInput) (*models.User, error) {
	user, err := s.repo.UpdateProfile(userID, input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, auth.ParseDatabaseError(err)
	}
	return user, nil
}

func (s *Service) AdminUpdateUser(userID int, input models.AdminUpdateUserInput) (*models.User, error) {
	user, err := s.repo.AdminUpdate(userID, input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, auth.ParseDatabaseError(err)
	}
	return user, nil
}
