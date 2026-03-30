package auth

import (
	"kbtuspace-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, role) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at`

	err := r.db.QueryRow(query, user.Email, user.PasswordHash, user.Role).Scan(&user.ID, &user.CreatedAt)
	return err
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, role, faculty_id, created_at FROM users WHERE email = $1`

	err := r.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
