package faculties

import (
	"context"

	"kbtuspace-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAllFaculties(ctx context.Context) ([]models.Faculty, error) {
	var fs []models.Faculty
	query := `SELECT id, name FROM faculties ORDER BY name ASC`
	if err := r.db.SelectContext(ctx, &fs, query); err != nil {
		return nil, err
	}
	return fs, nil
}
