package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                 string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPass               string
	DBName               string
	DBSSLMode            string
	RedisURL             string
	JWTSecret            string
	Environment          string
	CORSAllowedOrigins   []string
	DefaultAdminEmail    string
	DefaultAdminPassword string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                 getEnv("PORT", "8080"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPass:               getEnv("DB_PASSWORD", ""),
		DBName:               getEnv("DB_NAME", ""),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		Environment:          getEnv("ENVIRONMENT", "development"),
		CORSAllowedOrigins:   getCSVEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,http://localhost:5173,http://127.0.0.1:3000"),
		DefaultAdminEmail:    getEnv("DEFAULT_ADMIN_EMAIL", "admin@kbtu.kz"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DBUser == "" {
		return errors.New("DB_USER is required")
	}

	if c.DBName == "" {
		return errors.New("DB_NAME is required")
	}

	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters (got %d)", len(c.JWTSecret))
	}

	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("PORT must be a valid number: %w", err)
	}

	if c.Environment != "development" && c.Environment != "staging" && c.Environment != "production" {
		return fmt.Errorf("ENVIRONMENT must be one of: development, staging, production (got %s)", c.Environment)
	}

	if len(c.CORSAllowedOrigins) == 0 {
		return errors.New("CORS_ALLOWED_ORIGINS must contain at least one origin")
	}

	if c.DefaultAdminPassword != "" && len(c.DefaultAdminPassword) < 8 {
		return fmt.Errorf("DEFAULT_ADMIN_PASSWORD must be at least 8 characters (got %d)", len(c.DefaultAdminPassword))
	}

	return nil
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getCSVEnv(key, defaultVal string) []string {
	raw := getEnv(key, defaultVal)
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
