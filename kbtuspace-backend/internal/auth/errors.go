package auth

import (
	"errors"
	"strings"

	"github.com/lib/pq"
)

var (
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserBanned         = errors.New("user is banned")
)

func ParseDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if strings.Contains(pgErr.Message, "email") {
				return ErrDuplicateEmail
			}
		}
	}

	return err
}
