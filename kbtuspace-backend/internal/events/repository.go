package events

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"kbtuspace-backend/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repository struct {
	db *sqlx.DB
}

type EventAccessMeta struct {
	AuthorID  int  `db:"author_id"`
	FacultyID *int `db:"faculty_id"`
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

func (r *Repository) GetAll(facultyID *int, role string, globalOnly bool) ([]models.Post, error) {
	events := []models.Post{}
	cutoff := time.Now().AddDate(0, 0, -30)

	baseQuery := `
		SELECT p.id, p.author_id, u.email AS author_email, p.faculty_id, p.title, p.content, p.image_url, p.is_pinned, p.scope, p.status, p.approved_by, p.approved_at, p.rejection_reason, p.event_date, p.location, p.capacity, p.current_count, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE event_date IS NOT NULL
		  AND status = 'approved'
		  AND event_date >= $1
	`

	if globalOnly {
		baseQuery += " AND scope = 'global' ORDER BY p.event_date ASC"
		err := r.db.Select(&events, baseQuery, cutoff)
		return events, err
	}

	if facultyID != nil {
		baseQuery += " AND scope = 'faculty' AND p.faculty_id = $2 ORDER BY p.event_date ASC"
		err := r.db.Select(&events, baseQuery, cutoff, *facultyID)
		return events, err
	}

	return events, nil
}

func (r *Repository) GetByID(id int, includeUnapproved bool, actorFacultyID *int) (*models.Post, error) {
	var event models.Post

	query := `
		SELECT p.id, p.author_id, u.email AS author_email, p.faculty_id, p.title, p.content, p.image_url, p.is_pinned, p.scope, p.status, p.approved_by, p.approved_at, p.rejection_reason, p.event_date, p.location, p.capacity, p.current_count, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.id = $1 AND event_date IS NOT NULL
	`
	if !includeUnapproved {
		query += `
			AND status = 'approved'
		`
	}

	args := []interface{}{id}
	err := r.db.Get(&event, query, args...)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *Repository) Update(event *models.Post, actorID int, isAdmin bool, actorFacultyID *int) error {
	query := `
		WITH target AS (
			SELECT 1
			FROM posts
			WHERE id = $1 AND event_date IS NOT NULL
		), updated AS (
			UPDATE posts
			SET faculty_id = $2, title = $3, content = $4, image_url = $5, is_pinned = $6, scope = $7, status = $8, approved_by = $9, approved_at = $10, rejection_reason = $11, event_date = $12, location = $13, capacity = $14, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND event_date IS NOT NULL AND ($15 OR author_id = $16 OR faculty_id = $17)
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
		actorID,
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

func (r *Repository) Delete(id int, actorID int, isAdmin bool, actorFacultyID *int) error {
	query := `
		WITH target AS (
			SELECT 1
			FROM posts
			WHERE id = $1 AND event_date IS NOT NULL
		), deleted AS (
			DELETE FROM posts
			WHERE id = $1 AND event_date IS NOT NULL AND ($2 OR author_id = $3 OR faculty_id = $4)
			RETURNING 1
		)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM deleted) THEN 'deleted'
			WHEN EXISTS (SELECT 1 FROM target) THEN 'forbidden'
			ELSE 'not_found'
		END
	`

	var status string
	err := r.db.Get(&status, query, id, isAdmin, actorID, actorFacultyID)
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

func (r *Repository) Register(userID int, eventID int, actorFacultyID *int, isAdmin bool) error {
	tx, err := r.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var event struct {
		Capacity     int    `db:"capacity"`
		CurrentCount int    `db:"current_count"`
		Scope        string `db:"scope"`
		FacultyID    *int   `db:"faculty_id"`
	}
	query := `
		SELECT capacity, current_count, scope, faculty_id
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

	if event.Scope == models.ContentScopeFaculty && !isAdmin {
		if actorFacultyID == nil || event.FacultyID == nil || *actorFacultyID != *event.FacultyID {
			return ErrCrossFacultyAccess
		}
	}

	var existingStatus string
	err = tx.Get(&existingStatus, `
		SELECT status FROM registrations WHERE user_id = $1 AND event_id = $2 FOR UPDATE
	`, userID, eventID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if err == nil {
		if existingStatus != models.RegistrationStatusCancelled {
			return ErrAlreadyRegistered
		}
		if _, err := tx.Exec(`
			UPDATE registrations SET status = $1, updated_at = CURRENT_TIMESTAMP
			WHERE user_id = $2 AND event_id = $3
		`, models.RegistrationStatusRegistered, userID, eventID); err != nil {
			return err
		}
	} else {
		if _, err := tx.Exec(`
			INSERT INTO registrations (user_id, event_id, status, updated_at)
			VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		`, userID, eventID, models.RegistrationStatusRegistered); err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				return ErrAlreadyRegistered
			}
			return err
		}
	}

	if _, err := tx.Exec(`
		UPDATE posts SET current_count = current_count + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1
	`, eventID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) CancelRegistration(userID, eventID int) error {
	tx, err := r.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var dummy int
	if err := tx.Get(&dummy, `
		SELECT 1 FROM posts WHERE id = $1 AND event_date IS NOT NULL FOR UPDATE
	`, eventID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}

	var existingStatus string
	if err := tx.Get(&existingStatus, `
		SELECT status FROM registrations WHERE user_id = $1 AND event_id = $2 FOR UPDATE
	`, userID, eventID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotRegistered
		}
		return err
	}

	if existingStatus != models.RegistrationStatusRegistered {
		return ErrNotRegistered
	}

	if _, err := tx.Exec(`
		UPDATE registrations SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND event_id = $2
	`, userID, eventID); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		UPDATE posts SET current_count = GREATEST(current_count - 1, 0), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, eventID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) MarkAttended(userID, eventID int) error {
	result, err := r.db.Exec(`
		UPDATE registrations
		SET status = 'attended', updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND event_id = $2 AND status = 'registered'
	`, userID, eventID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotRegistered
	}
	return nil
}

func (r *Repository) IsUserRegistered(userID, eventID int) (bool, error) {
	var registered bool
	err := r.db.Get(&registered, `
		SELECT EXISTS(
			SELECT 1
			FROM registrations
			WHERE user_id = $1
			  AND event_id = $2
			  AND status = 'registered'
		)
	`, userID, eventID)
	return registered, err
}

func (r *Repository) GetEventAccessMeta(eventID int) (*EventAccessMeta, error) {
	var meta EventAccessMeta
	if err := r.db.Get(&meta, `
		SELECT author_id, faculty_id
		FROM posts
		WHERE id = $1 AND event_date IS NOT NULL
	`, eventID); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (r *Repository) ListPendingGlobal() ([]models.Post, error) {
	events := []models.Post{}
	query := `
		SELECT p.id, p.author_id, u.email AS author_email, p.faculty_id, p.title, p.content, p.image_url, p.is_pinned, p.scope, p.status, p.approved_by, p.approved_at, p.rejection_reason, p.event_date, p.location, p.capacity, p.current_count, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE event_date IS NOT NULL AND scope = 'global' AND status = 'pending'
		ORDER BY p.created_at DESC
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
