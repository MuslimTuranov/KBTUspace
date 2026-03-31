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

func (r *Repository) Update(post *models.Post) error {
	query := `
		UPDATE posts
		SET faculty_id = $1, title = $2, content = $3, image_url = $4, is_pinned = $5
		WHERE id = $6 AND event_date IS NULL
	`

	_, err := r.db.Exec(
		query,
		post.FacultyID,
		post.Title,
		post.Content,
		post.ImageURL,
		post.IsPinned,
		post.ID,
	)

	return err
}

func (r *Repository) Delete(id int) error {
	query := `DELETE FROM posts WHERE id = $1 AND event_date IS NULL`
	_, err := r.db.Exec(query, id)
	return err
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

func IsNotFound(err error) bool {
	return err == sql.ErrNoRows
}
