package auth

import (
	"errors"
	"strings"

	"github.com/lib/pq"
)

var (
	ErrDuplicateEmail   = errors.New("email already exists")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// ParseDatabaseError преобразует database ошибки в application ошибки
func ParseDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	// Проверяем pq ошибки
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			if strings.Contains(pgErr.Message, "email") {
				return ErrDuplicateEmail
			}
		}
	}

	return err
}
