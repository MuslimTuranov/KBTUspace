package models

import "time"

type Post struct {
	ID        int        `db:"id" json:"id"`
	AuthorID  int        `db:"author_id" json:"author_id"`
	FacultyID *int       `db:"faculty_id" json:"faculty_id,omitempty"`
	Title     string     `db:"title" json:"title"`
	Content   string     `db:"content" json:"content"`
	ImageURL  *string    `db:"image_url" json:"image_url,omitempty"`
	IsPinned  bool       `db:"is_pinned" json:"is_pinned"`
	EventDate *time.Time `db:"event_date" json:"event_date,omitempty"`
	Location  *string    `db:"location" json:"location,omitempty"`
	Capacity  int        `db:"capacity" json:"capacity,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

type CreatePostInput struct {
	FacultyID *int    `json:"faculty_id,omitempty"`
	Title     string  `json:"title" binding:"required,min=3,max=255"`
	Content   string  `json:"content" binding:"required,min=10,max=5000"`
	ImageURL  *string `json:"image_url" binding:"omitempty,url"`
}

type UpdatePostInput struct {
	FacultyID *int    `json:"faculty_id,omitempty"`
	Title     string  `json:"title" binding:"required,min=3,max=255"`
	Content   string  `json:"content" binding:"required,min=10,max=5000"`
	ImageURL  *string `json:"image_url" binding:"omitempty,url"`
	IsPinned  bool    `json:"is_pinned"`
}
