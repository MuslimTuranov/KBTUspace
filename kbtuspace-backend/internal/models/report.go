package models

import "time"

const (
	ReportStatusPending  = "pending"
	ReportStatusClosed   = "closed"
	ReportStatusRejected = "rejected"

	ReportTargetPost  = "post"
	ReportTargetEvent = "event"
)

type CreateReportInput struct {
	TargetType string `json:"target_type" binding:"required,oneof=post event"`
	TargetID   int    `json:"target_id" binding:"required,min=1"`
	Reason     string `json:"reason" binding:"required,min=3,max=1000"`
}

type CloseReportInput struct {
	Status     string `json:"status" binding:"required,oneof=closed rejected"`
	ReviewNote string `json:"review_note" binding:"required,min=3,max=1000"`
}

type Report struct {
	ID             int        `db:"id" json:"id"`
	ReporterID     int        `db:"reporter_id" json:"reporter_id"`
	TargetPostID   int        `db:"target_post_id" json:"target_post_id"`
	TargetType     string     `db:"target_type" json:"target_type"`
	Reason         string     `db:"reason" json:"reason"`
	Status         string     `db:"status" json:"status"`
	ReviewNote     *string    `db:"review_note" json:"review_note,omitempty"`
	ReviewedBy     *int       `db:"reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time `db:"reviewed_at" json:"reviewed_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
	TargetTitle    string     `db:"target_title" json:"target_title"`
	TargetAuthorID int        `db:"target_author_id" json:"target_author_id"`
}
