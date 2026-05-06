package models

import "time"

const (
	ContentScopeFaculty = "faculty"
	ContentScopeGlobal  = "global"

	ContentStatusDraft    = "draft"
	ContentStatusPending  = "pending"
	ContentStatusApproved = "approved"
	ContentStatusRejected = "rejected"
)

type Post struct {
	ID              int        `db:"id" json:"id"`
	AuthorID        int        `db:"author_id" json:"author_id"`
	FacultyID       *int       `db:"faculty_id" json:"faculty_id,omitempty"`
	Title           string     `db:"title" json:"title"`
	Content         string     `db:"content" json:"content"`
	ImageURL        *string    `db:"image_url" json:"image_url,omitempty"`
	IsPinned        bool       `db:"is_pinned" json:"is_pinned"`
	Scope           string     `db:"scope" json:"scope"`
	Status          string     `db:"status" json:"status"`
	ApprovedBy      *int       `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `db:"approved_at" json:"approved_at,omitempty"`
	RejectionReason *string    `db:"rejection_reason" json:"rejection_reason,omitempty"`
	EventDate       *time.Time `db:"event_date" json:"event_date,omitempty"`
	Location        *string    `db:"location" json:"location,omitempty"`
	Capacity        int        `db:"capacity" json:"capacity"`
	CurrentCount    int        `db:"current_count" json:"current_count"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

type CreatePostInput struct {
	FacultyID *int    `json:"faculty_id,omitempty"`
	Title     string  `json:"title" binding:"required,min=3,max=255"`
	Content   string  `json:"content" binding:"required,min=10,max=5000"`
	ImageURL  *string `json:"image_url" binding:"omitempty,url"`
	Scope     string  `json:"scope" binding:"omitempty,oneof=faculty global"`
}

type UpdatePostInput struct {
	FacultyID *int    `json:"faculty_id,omitempty"`
	Title     string  `json:"title" binding:"required,min=3,max=255"`
	Content   string  `json:"content" binding:"required,min=10,max=5000"`
	ImageURL  *string `json:"image_url" binding:"omitempty,url"`
	IsPinned  bool    `json:"is_pinned"`
	Scope     string  `json:"scope" binding:"omitempty,oneof=faculty global"`
}

type RejectContentInput struct {
	Reason string `json:"reason" binding:"required,min=3,max=1000"`
}

type PinPostInput struct {
	IsPinned bool `json:"is_pinned"`
}
