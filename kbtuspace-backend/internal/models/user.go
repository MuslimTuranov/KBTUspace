package models

import "time"

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	FacultyID    *int      `json:"faculty_id" db:"faculty_id"`
	IsBanned     bool      `json:"is_banned" db:"is_banned"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileInput struct {
	Email     *string `json:"email,omitempty" binding:"omitempty,email"`
	FacultyID *int    `json:"faculty_id,omitempty"`
}

type AdminUpdateUserInput struct {
	Role      *string `json:"role,omitempty" binding:"omitempty,oneof=student organizer admin"`
	FacultyID *int    `json:"faculty_id,omitempty"`
	IsBanned  *bool   `json:"is_banned,omitempty"`
}
