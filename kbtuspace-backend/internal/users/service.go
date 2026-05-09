package users

import (
	"database/sql"
	"errors"

	"kbtuspace-backend/internal/auth"
	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/hash"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetProfile(userID int) (*models.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	normalizeAdminFaculty(user)
	return user, nil
}

func (s *Service) GetAllUsers() ([]models.User, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	for i := range users {
		normalizeAdminFaculty(&users[i])
	}
	return users, nil
}

func (s *Service) UpdateProfile(userID int, input models.UpdateProfileInput) (*models.User, error) {
	if input.Email != nil {
		if err := auth.ValidateKBTUEmail(*input.Email); err != nil {
			return nil, err
		}
	}

	user, err := s.repo.UpdateProfile(userID, input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, auth.ParseDatabaseError(err)
	}
	normalizeAdminFaculty(user)
	return user, nil
}

func (s *Service) ChangePassword(userID int, input models.ChangePasswordInput) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.ErrUserNotFound
		}
		return err
	}

	if !hash.CheckPasswordHash(input.CurrentPassword, user.PasswordHash) {
		return auth.ErrInvalidPassword
	}

	passwordHash, err := hash.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(userID, passwordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *Service) AdminUpdateUser(userID int, input models.AdminUpdateUserInput) (*models.User, error) {
	user, err := s.repo.AdminUpdate(userID, input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, auth.ParseDatabaseError(err)
	}
	normalizeAdminFaculty(user)
	return user, nil
}

func normalizeAdminFaculty(user *models.User) {
	if user != nil && user.Role == "admin" {
		user.FacultyID = nil
	}
}
