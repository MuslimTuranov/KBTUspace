package models

import "time"

type Event struct {
	ID           int       `db:"id" json:"id"`
	AuthorID     int       `db:"author_id" json:"author_id"`
	FacultyID    *int      `db:"faculty_id" json:"faculty_id"`
	Title        string    `db:"title" json:"title"`
	Description  string    `db:"description" json:"description"`
	ImageURL     *string   `db:"image_url" json:"image_url,omitempty"`
	EventDate    time.Time `db:"event_date" json:"event_date"`
	Location     string    `db:"location" json:"location"`
	Capacity     int       `db:"capacity" json:"capacity"`
	CurrentCount int       `db:"current_count" json:"current_count"`
	IsPinned     bool      `db:"is_pinned" json:"is_pinned"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type CreateEventInput struct {
	FacultyID   *int      `json:"faculty_id,omitempty"`
	Title       string    `json:"title" binding:"required,min=3,max=255"`
	Description string    `json:"description" binding:"required,min=10,max=5000"`
	ImageURL    *string   `json:"image_url" binding:"omitempty,url"`
	EventDate   time.Time `json:"event_date" binding:"required"`
	Location    string    `json:"location" binding:"required,min=3,max=255"`
	Capacity    int       `json:"capacity" binding:"required,min=1,max=10000"`
	Scope       string    `json:"scope" binding:"omitempty,oneof=faculty global"`
}

type UpdateEventInput struct {
	FacultyID   *int      `json:"faculty_id,omitempty"`
	Title       string    `json:"title" binding:"required,min=3,max=255"`
	Description string    `json:"description" binding:"required,min=10,max=5000"`
	ImageURL    *string   `json:"image_url" binding:"omitempty,url"`
	EventDate   time.Time `json:"event_date" binding:"required"`
	Location    string    `json:"location" binding:"required,min=3,max=255"`
	Capacity    int       `json:"capacity" binding:"required,min=1,max=10000"`
	Scope       string    `json:"scope" binding:"omitempty,oneof=faculty global"`
}

type EventRegistration struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	EventID   int       `db:"event_id" json:"event_id"`
	Status    string    `db:"status" json:"status"` // registered, cancelled, attended
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
