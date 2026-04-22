package reports

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

func (r *Repository) GetTarget(id int, targetType string) (*models.Post, error) {
	var post models.Post

	query := `
		SELECT id, author_id, faculty_id, title, content, image_url, is_pinned, scope, status, approved_by, approved_at, rejection_reason, event_date, location, capacity, current_count, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	switch targetType {
	case models.ReportTargetPost:
		query += " AND event_date IS NULL"
	case models.ReportTargetEvent:
		query += " AND event_date IS NOT NULL"
	default:
		return nil, ErrInvalidTargetType
	}

	if err := r.db.Get(&post, query, id); err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *Repository) HasPendingDuplicate(reporterID, targetPostID int) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM reports
			WHERE reporter_id = $1
			  AND target_post_id = $2
			  AND status = 'pending'
		)
	`

	err := r.db.Get(&exists, query, reporterID, targetPostID)
	return exists, err
}

func (r *Repository) Create(report *models.Report) error {
	query := `
		INSERT INTO reports (reporter_id, target_post_id, target_type, reason, status, review_note, reviewed_by, reviewed_at)
		VALUES ($1, $2, $3, $4, $5, NULL, NULL, NULL)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		report.ReporterID,
		report.TargetPostID,
		report.TargetType,
		report.Reason,
		report.Status,
	).Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt)
}

func (r *Repository) List(status string) ([]models.Report, error) {
	reports := []models.Report{}

	query := `
		SELECT
			r.id,
			r.reporter_id,
			r.target_post_id,
			r.target_type,
			r.reason,
			r.status,
			r.review_note,
			r.reviewed_by,
			r.reviewed_at,
			r.created_at,
			r.updated_at,
			p.title AS target_title,
			p.author_id AS target_author_id
		FROM reports r
		JOIN posts p ON p.id = r.target_post_id
		WHERE r.status = $1
		ORDER BY r.created_at DESC
	`

	if err := r.db.Select(&reports, query, status); err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *Repository) Close(id int, status, reviewNote string, adminID int) error {
	query := `
		UPDATE reports
		SET status = $2,
			review_note = $3,
			reviewed_by = $4,
			reviewed_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND status = 'pending'
	`

	result, err := r.db.Exec(query, id, status, reviewNote, adminID)
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
