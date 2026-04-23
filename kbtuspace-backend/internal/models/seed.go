package models

import (
	"log/slog"

	"kbtuspace-backend/pkg/hash"

	"github.com/jmoiron/sqlx"
)

// SeedDefaults inserts a default admin user on first run if no admin exists.
// Default credentials: admin@kbtu.kz / Admin1234!!
func SeedDefaults(db *sqlx.DB) error {
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM users WHERE role = 'admin'`); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := hash.HashPassword("Admin1234!!")
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, role)
		VALUES ('admin@kbtu.kz', $1, 'admin')
		ON CONFLICT (email) DO NOTHING
	`, hashedPassword)
	if err != nil {
		return err
	}

	slog.Info("Default admin created", slog.String("email", "admin@kbtu.kz"), slog.String("password", "Admin1234!!"))
	return nil
}
