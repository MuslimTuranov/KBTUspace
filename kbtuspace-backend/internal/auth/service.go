package auth

import (
	"errors"

	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/hash"
	"kbtuspace-backend/pkg/jwt"
)

type Service struct {
	repo      *Repository
	jwtSecret []byte
}

func NewService(repo *Repository, jwtSecret []byte) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) RegisterUser(input models.RegisterInput) error {
	if len(input.Email) == 0 || len(input.Password) == 0 {
		return errors.New("email and password are required")
	}

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
	if len(input.Email) == 0 || len(input.Password) == 0 {
		return "", ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(input.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if user.IsBanned {
		return "", ErrUserBanned
	}

	if !hash.CheckPasswordHash(input.Password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	token, err := jwt.GenerateToken(user.ID, user.Role, user.FacultyID, s.jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}
