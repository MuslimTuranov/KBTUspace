package events

import (
	"context"
	"database/sql"
	"errors"

	"kbtuspace-backend/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(event *models.Post) error {
	query := `
		INSERT INTO posts (author_id, faculty_id, title, content, image_url, is_pinned, event_date, location, capacity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		event.AuthorID,
		event.FacultyID,
		event.Title,
		event.Content,
		event.ImageURL,
		event.IsPinned,
		event.EventDate,
		event.Location,
		event.Capacity,
	).Scan(&event.ID, &event.CreatedAt)
}

func (r *Repository) GetAll(facultyID *int) ([]models.Post, error) {
	events := []models.Post{}

	baseQuery := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, event_date, location, capacity, created_at
		FROM posts
		WHERE event_date IS NOT NULL
	`

	if facultyID != nil {
		baseQuery += " AND faculty_id = $1 ORDER BY event_date ASC"
		err := r.db.Select(&events, baseQuery, *facultyID)
		return events, err
	}

	baseQuery += " ORDER BY event_date ASC"
	err := r.db.Select(&events, baseQuery)
	return events, err
}

func (r *Repository) GetByID(id int) (*models.Post, error) {
	var event models.Post

	query := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, event_date, location, capacity, created_at
		FROM posts
		WHERE id = $1 AND event_date IS NOT NULL
	`

	err := r.db.Get(&event, query, id)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *Repository) Update(event *models.Post) error {
	query := `
		UPDATE posts
		SET faculty_id = $1, title = $2, content = $3, image_url = $4, is_pinned = $5, event_date = $6, location = $7, capacity = $8
		WHERE id = $9 AND event_date IS NOT NULL
	`

	result, err := r.db.Exec(
		query,
		event.FacultyID,
		event.Title,
		event.Content,
		event.ImageURL,
		event.IsPinned,
		event.EventDate,
		event.Location,
		event.Capacity,
		event.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) Delete(id int) error {
	query := `DELETE FROM posts WHERE id = $1 AND event_date IS NOT NULL`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) Register(userID int, eventID int) error {
	tx, err := r.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var capacity int
	query := `
		SELECT capacity
		FROM posts
		WHERE id = $1 AND event_date IS NOT NULL
		FOR UPDATE
	`

	if err := tx.Get(&capacity, query, eventID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}

	if capacity <= 0 {
		return ErrEventFull
	}

	if _, err := tx.Exec(`UPDATE posts SET capacity = capacity - 1 WHERE id = $1`, eventID); err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO registrations (user_id, event_id, status) VALUES ($1, $2, 'registered')`,
		userID,
		eventID,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrAlreadyRegistered
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

