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
	Capacity  int        `db:"capacity" json:"capacity"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

type CreatePostInput struct {
	FacultyID *int    `json:"faculty_id"`
	Title     string  `json:"title" binding:"required"`
	Content   string  `json:"content" binding:"required"`
	ImageURL  *string `json:"image_url"`
	IsPinned  bool    `json:"is_pinned"`
}

type UpdatePostInput struct {
	FacultyID *int    `json:"faculty_id"`
	Title     string  `json:"title" binding:"required"`
	Content   string  `json:"content" binding:"required"`
	ImageURL  *string `json:"image_url"`
	IsPinned  bool    `json:"is_pinned"`
}

type CreateEventInput struct {
	FacultyID *int      `json:"faculty_id"`
	Title     string    `json:"title" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	ImageURL  *string   `json:"image_url"`
	IsPinned  bool      `json:"is_pinned"`
	EventDate time.Time `json:"event_date" binding:"required"`
	Location  string    `json:"location" binding:"required"`
	Capacity  int       `json:"capacity" binding:"required,min=1"`
}

type UpdateEventInput struct {
	FacultyID *int      `json:"faculty_id"`
	Title     string    `json:"title" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	ImageURL  *string   `json:"image_url"`
	IsPinned  bool      `json:"is_pinned"`
	EventDate time.Time `json:"event_date" binding:"required"`
	Location  string    `json:"location" binding:"required"`
	Capacity  int       `json:"capacity" binding:"required,min=1"`
}
