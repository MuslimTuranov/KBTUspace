package posts

import (
	"database/sql"

	"kbtuspace-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(post *models.Post) error {
	query := `
		INSERT INTO posts (author_id, faculty_id, title, content, image_url, is_pinned, scope, status, approved_by, approved_at, rejection_reason, event_date, location, capacity, current_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NULL, NULL, 0, 0)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		post.AuthorID,
		post.FacultyID,
		post.Title,
		post.Content,
		post.ImageURL,
		post.IsPinned,
		post.Scope,
		post.Status,
		post.ApprovedBy,
		post.ApprovedAt,
		post.RejectionReason,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (r *Repository) GetAll(facultyID *int) ([]models.Post, error) {
	posts := []models.Post{}

	baseQuery := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, scope, status, approved_by, approved_at, rejection_reason, event_date, location, capacity, current_count, created_at, updated_at
		FROM posts
		WHERE event_date IS NULL
		  AND status = 'approved'
	`

	if facultyID != nil {
		baseQuery += " AND (scope = 'global' OR faculty_id = $1) ORDER BY is_pinned DESC, created_at DESC"
		err := r.db.Select(&posts, baseQuery, *facultyID)
		return posts, err
	}

	baseQuery += " AND scope = 'global' ORDER BY is_pinned DESC, created_at DESC"
	err := r.db.Select(&posts, baseQuery)
	return posts, err
}

func (r *Repository) GetByID(id int, includeUnapproved bool) (*models.Post, error) {
	var post models.Post

	query := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, scope, status, approved_by, approved_at, rejection_reason, event_date, location, capacity, current_count, created_at, updated_at
		FROM posts
		WHERE id = $1 AND event_date IS NULL
	`
	if !includeUnapproved {
		query += " AND status = 'approved'"
	}

	err := r.db.Get(&post, query, id)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *Repository) Update(post *models.Post, actorID int, isAdmin bool) error {
	query := `
		WITH target AS (
			SELECT 1
			FROM posts
			WHERE id = $1 AND event_date IS NULL
		), updated AS (
			UPDATE posts
			SET faculty_id = $2, title = $3, content = $4, image_url = $5, is_pinned = $6, scope = $7, status = $8, approved_by = $9, approved_at = $10, rejection_reason = $11, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND event_date IS NULL AND ($12 OR author_id = $13)
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
		post.ID,
		post.FacultyID,
		post.Title,
		post.Content,
		post.ImageURL,
		post.IsPinned,
		post.Scope,
		post.Status,
		post.ApprovedBy,
		post.ApprovedAt,
		post.RejectionReason,
		isAdmin,
		actorID,
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

func (r *Repository) Delete(id int, actorID int, isAdmin bool) error {
	query := `
		WITH target AS (
			SELECT 1
			FROM posts
			WHERE id = $1 AND event_date IS NULL
		), deleted AS (
			DELETE FROM posts
			WHERE id = $1 AND event_date IS NULL AND ($2 OR author_id = $3)
			RETURNING 1
		)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM deleted) THEN 'deleted'
			WHEN EXISTS (SELECT 1 FROM target) THEN 'forbidden'
			ELSE 'not_found'
		END
	`

	var status string
	err := r.db.Get(&status, query, id, isAdmin, actorID)
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

func (r *Repository) ExistsByID(id int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1 AND event_date IS NULL)`
	err := r.db.Get(&exists, query, id)
	return exists, err
}

func (r *Repository) GetAuthorID(id int) (int, error) {
	var authorID int
	query := `SELECT author_id FROM posts WHERE id = $1 AND event_date IS NULL`
	err := r.db.Get(&authorID, query, id)
	return authorID, err
}

func (r *Repository) ListPendingGlobal() ([]models.Post, error) {
	posts := []models.Post{}
	query := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, scope, status, approved_by, approved_at, rejection_reason, event_date, location, capacity, current_count, created_at, updated_at
		FROM posts
		WHERE event_date IS NULL AND scope = 'global' AND status = 'pending'
		ORDER BY created_at DESC
	`

	err := r.db.Select(&posts, query)
	return posts, err
}

func (r *Repository) Approve(id int, adminID int) error {
	query := `
		UPDATE posts
		SET status = 'approved', approved_by = $2, approved_at = CURRENT_TIMESTAMP, rejection_reason = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND event_date IS NULL AND scope = 'global' AND status = 'pending'
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
		WHERE id = $1 AND event_date IS NULL AND scope = 'global' AND status = 'pending'
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

func (r *Repository) Pin(id int, isPinned bool, isAdmin bool, actorFacultyID *int) error {
	query := `
		WITH target AS (
			SELECT 1
			FROM posts
			WHERE id = $1 AND event_date IS NULL AND status = 'approved'
		), updated AS (
			UPDATE posts
			SET is_pinned = $2, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
			  AND event_date IS NULL
			  AND status = 'approved'
			  AND ($3 OR (scope = 'faculty' AND faculty_id = $4))
			RETURNING 1
		)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM updated) THEN 'updated'
			WHEN EXISTS (SELECT 1 FROM target) THEN 'forbidden'
			ELSE 'not_found'
		END
	`

	var status string
	err := r.db.Get(&status, query, id, isPinned, isAdmin, actorFacultyID)
	if err != nil {
		return err
	}

	switch status {
	case "updated":
		return nil
	case "forbidden":
		return ErrInvalidPinScope
	default:
		return sql.ErrNoRows
	}
}
