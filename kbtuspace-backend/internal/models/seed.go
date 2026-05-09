package models

import (
	"errors"
	"log/slog"

	"kbtuspace-backend/pkg/config"
	"kbtuspace-backend/pkg/hash"

	"github.com/jmoiron/sqlx"
)

func SeedDefaults(db *sqlx.DB, cfg *config.Config) error {
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM users WHERE role = 'admin'`); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	adminPassword := cfg.DefaultAdminPassword
	if adminPassword == "" {
		if cfg.Environment == "production" {
			return errors.New("DEFAULT_ADMIN_PASSWORD is required to seed the first admin in production")
		}
		slog.Warn("Default admin was not created because DEFAULT_ADMIN_PASSWORD is empty")
		return nil
	}

	hashedPassword, err := hash.HashPassword(adminPassword)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, 'admin')
		ON CONFLICT (email) DO NOTHING
	`, cfg.DefaultAdminEmail, hashedPassword)
	if err != nil {
		return err
	}

	slog.Info("Default admin created", slog.String("email", cfg.DefaultAdminEmail))
	return nil
}
