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
		INSERT INTO posts (author_id, faculty_id, title, content, image_url, is_pinned, scope, status, approved_by, approved_at, rejection_reason, event_date, location, capacity, current_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 0)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		event.AuthorID,
		event.FacultyID,
		event.Title,
		event.Content,
		event.ImageURL,
		event.IsPinned,
		event.Scope,
		event.Status,
		event.ApprovedBy,
		event.ApprovedAt,
		event.RejectionReason,
		event.EventDate,
		event.Location,
		event.Capacity,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

func (r *Repository) GetAll(facultyID *int) ([]models.Post, error) {
	events := []models.Post{}

	baseQuery := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, scope, status, approved_by, approved_at, rejection_reason, event_date, location, capacity, current_count, created_at, updated_at
		FROM posts
		WHERE event_date IS NOT NULL
		  AND status = 'approved'
	`

	if facultyID != nil {
		baseQuery += " AND (scope = 'global' OR faculty_id = $1) ORDER BY event_date ASC"
		err := r.db.Select(&events, baseQuery, *facultyID)
		return events, err
	}

	baseQuery += " AND scope = 'global' ORDER BY event_date ASC"
	err := r.db.Select(&events, baseQuery)
	return events, err
}

func (r *Repository) GetByID(id int, includeUnapproved bool) (*models.Post, error) {
	var event models.Post

	query := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, scope, status, approved_by, approved_at, rejection_reason, event_date, location, capacity, current_count, created_at, updated_at
		FROM posts
		WHERE id = $1 AND event_date IS NOT NULL
	`
	if !includeUnapproved {
		query += " AND status = 'approved'"
	}

	err := r.db.Get(&event, query, id)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *Repository) Update(event *models.Post, isAdmin bool, actorFacultyID *int) error {
	query := `
		WITH target AS (
			SELECT 1
			FROM posts
			WHERE id = $1 AND event_date IS NOT NULL
		), updated AS (
			UPDATE posts
			SET faculty_id = $2, title = $3, content = $4, image_url = $5, is_pinned = $6, scope = $7, status = $8, approved_by = $9, approved_at = $10, rejection_reason = $11, event_date = $12, location = $13, capacity = $14, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND event_date IS NOT NULL AND ($15 OR faculty_id = $16)
			RETURNING 1
		)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM updated) THEN 'updated'
			WHEN EXISTS (SELECT 1 FROM target) THEN 'forbidden'
			ELSE 'not_found'
		END
	`

	var status string
	err := r.db.Get(
		&status,
		query,
		event.ID,
		event.FacultyID,
		event.Title,
		event.Content,
		event.ImageURL,
		event.IsPinned,
		event.Scope,
		event.Status,
		event.ApprovedBy,
		event.ApprovedAt,
		event.RejectionReason,
		event.EventDate,
		event.Location,
		event.Capacity,
		isAdmin,
		actorFacultyID,
	)
	if err != nil {
		return err
	}

	switch status {
	case "updated":
		return nil
	case "forbidden":
		return ErrForbidden
	default:
		return sql.ErrNoRows
	}
}

func (r *Repository) Delete(id int, isAdmin bool, actorFacultyID *int) error {
	query := `
		WITH target AS (
			SELECT 1
			FROM posts
			WHERE id = $1 AND event_date IS NOT NULL
		), deleted AS (
			DELETE FROM posts
			WHERE id = $1 AND event_date IS NOT NULL AND ($2 OR faculty_id = $3)
			RETURNING 1
		)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM deleted) THEN 'deleted'
			WHEN EXISTS (SELECT 1 FROM target) THEN 'forbidden'
			ELSE 'not_found'
		END
	`

	var status string
	err := r.db.Get(&status, query, id, isAdmin, actorFacultyID)
	if err != nil {
		return err
	}

	switch status {
	case "deleted":
		return nil
	case "forbidden":
		return ErrForbidden
	default:
		return sql.ErrNoRows
	}
}

func (r *Repository) Register(userID int, eventID int) error {
	tx, err := r.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var event struct {
		Capacity     int `db:"capacity"`
		CurrentCount int `db:"current_count"`
	}
	query := `
		SELECT capacity, current_count
		FROM posts
		WHERE id = $1 AND event_date IS NOT NULL AND status = 'approved'
		FOR UPDATE
	`

	if err := tx.Get(&event, query, eventID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}

	if event.CurrentCount >= event.Capacity {
		return ErrEventFull
	}

	if _, err := tx.Exec(`UPDATE posts SET current_count = current_count + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, eventID); err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO registrations (user_id, event_id, status, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`,
		userID,
		eventID,
		models.RegistrationStatusRegistered,
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

func (r *Repository) ListPendingGlobal() ([]models.Post, error) {
	events := []models.Post{}
	query := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, scope, status, approved_by, approved_at, rejection_reason, event_date, location, capacity, current_count, created_at, updated_at
		FROM posts
		WHERE event_date IS NOT NULL AND scope = 'global' AND status = 'pending'
		ORDER BY created_at DESC
	`

	err := r.db.Select(&events, query)
	return events, err
}

func (r *Repository) Approve(id int, adminID int) error {
	query := `
		UPDATE posts
		SET status = 'approved', approved_by = $2, approved_at = CURRENT_TIMESTAMP, rejection_reason = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND event_date IS NOT NULL AND scope = 'global' AND status = 'pending'
	`

	result, err := r.db.Exec(query, id, adminID)
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

func (r *Repository) Reject(id int, reason string) error {
	query := `
		UPDATE posts
		SET status = 'rejected', approved_by = NULL, approved_at = NULL, rejection_reason = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND event_date IS NOT NULL AND scope = 'global' AND status = 'pending'
	`

	result, err := r.db.Exec(query, id, reason)
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
