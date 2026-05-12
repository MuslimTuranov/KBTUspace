package config

import "testing"

func validConfig() *Config {
	return &Config{
		Port:                 "8080",
		DBUser:               "postgres",
		DBName:               "kbtuspace",
		JWTSecret:            "12345678901234567890123456789012",
		Environment:          "development",
		CORSAllowedOrigins:   []string{"http://localhost:3000"},
		DefaultAdminPassword: "admin123",
	}
}

func TestConfigValidEnv(t *testing.T) {
	cfg := validConfig()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestConfigMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Config)
	}{
		{"missing DB_USER", func(c *Config) { c.DBUser = "" }},
		{"missing DB_NAME", func(c *Config) { c.DBName = "" }},
		{"missing JWT_SECRET", func(c *Config) { c.JWTSecret = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.edit(cfg)

			if err := cfg.Validate(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestConfigShortJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWTSecret = "short"

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigInvalidPort(t *testing.T) {
	cfg := validConfig()
	cfg.Port = "abc"

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigInvalidEnvironment(t *testing.T) {
	cfg := validConfig()
	cfg.Environment = "local"

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigEmptyCORSAllowedOrigins(t *testing.T) {
	cfg := validConfig()
	cfg.CORSAllowedOrigins = []string{}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigShortDefaultAdminPassword(t *testing.T) {
	cfg := validConfig()
	cfg.DefaultAdminPassword = "123"

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
