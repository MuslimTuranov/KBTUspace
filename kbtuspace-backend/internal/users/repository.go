package users

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

func (r *Repository) GetByID(id int) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = $1`

	if err := r.db.Get(&user, query, id); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetAll() ([]models.User, error) {
	users := []models.User{}
	query := `SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users ORDER BY id ASC`
	if err := r.db.Select(&users, query); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE email = $1`

	if err := r.db.Get(&user, query, email); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateProfile(userID int, input models.UpdateProfileInput) (*models.User, error) {
	current, err := r.GetByID(userID)
	if err != nil {
		return nil, err
	}

	email := current.Email
	if input.Email != nil {
		email = *input.Email
	}

	facultyID := current.FacultyID
	if input.FacultyID != nil {
		if *input.FacultyID <= 0 {
			facultyID = nil
		} else {
			facultyID = input.FacultyID
		}
	}

	query := `
		UPDATE users
		SET email = $2, faculty_id = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at
	`

	var updated models.User
	if err := r.db.Get(&updated, query, userID, email, facultyID); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *Repository) AdminUpdate(userID int, input models.AdminUpdateUserInput) (*models.User, error) {
	current, err := r.GetByID(userID)
	if err != nil {
		return nil, err
	}

	role := current.Role
	if input.Role != nil {
		role = *input.Role
	}

	facultyID := current.FacultyID
	if input.FacultyID != nil {
		if *input.FacultyID <= 0 {
			facultyID = nil
		} else {
			facultyID = input.FacultyID
		}
	}

	isBanned := current.IsBanned
	if input.IsBanned != nil {
		isBanned = *input.IsBanned
	}

	query := `
		UPDATE users
		SET role = $2, faculty_id = $3, is_banned = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at
	`

	var updated models.User
	if err := r.db.Get(&updated, query, userID, role, facultyID, isBanned); err != nil {
		return nil, err
	}
	return &updated, nil
}
