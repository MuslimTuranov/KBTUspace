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
		INSERT INTO posts (author_id, faculty_id, title, content, image_url, is_pinned, event_date, location, capacity)
		VALUES ($1, $2, $3, $4, $5, $6, NULL, NULL, 0)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		post.AuthorID,
		post.FacultyID,
		post.Title,
		post.Content,
		post.ImageURL,
		post.IsPinned,
	).Scan(&post.ID, &post.CreatedAt)
}

func (r *Repository) GetAll(facultyID *int) ([]models.Post, error) {
	posts := []models.Post{}

	baseQuery := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, event_date, location, capacity, created_at
		FROM posts
		WHERE event_date IS NULL
	`

	if facultyID != nil {
		baseQuery += " AND faculty_id = $1 ORDER BY is_pinned DESC, created_at DESC"
		err := r.db.Select(&posts, baseQuery, *facultyID)
		return posts, err
	}

	baseQuery += " ORDER BY is_pinned DESC, created_at DESC"
	err := r.db.Select(&posts, baseQuery)
	return posts, err
}

func (r *Repository) GetByID(id int) (*models.Post, error) {
	var post models.Post

	query := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, event_date, location, capacity, created_at
		FROM posts
		WHERE id = $1 AND event_date IS NULL
	`

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
			SET faculty_id = $2, title = $3, content = $4, image_url = $5, is_pinned = $6
			WHERE id = $1 AND event_date IS NULL AND ($7 OR author_id = $8)
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

