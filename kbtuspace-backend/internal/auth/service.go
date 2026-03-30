package auth

import (
	"errors"
	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/hash"
	"kbtuspace-backend/pkg/jwt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterUser(input models.RegisterInput) error {
	hashedPassword, err := hash.HashPassword(input.Password)
	if err != nil {
		return err
	}

	user := &models.User{
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Role:         "student",
	}

	return s.repo.CreateUser(user)
}

func (s *Service) LoginUser(input models.LoginInput) (string, error) {
	user, err := s.repo.GetUserByEmail(input.Email)
	if err != nil {
		return "", errors.New("user not found or invalid credentials")
	}

	if !hash.CheckPasswordHash(input.Password, user.PasswordHash) {
		return "", errors.New("user not found or invalid credentials")
	}

	token, err := jwt.GenerateToken(user.ID, user.Role, user.FacultyID)
	if err != nil {
		return "", err
	}

	return token, nil
}
