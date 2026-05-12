package models

import (
	"errors"
	"testing"
	"time"

	"kbtuspace-backend/pkg/config"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func setupSeedTest(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	return sqlxDB, mock, func() {
		_ = sqlxDB.Close()
		_ = db.Close()
	}
}

func seedConfig(env string, password string) *config.Config {
	return &config.Config{
		DefaultAdminEmail:    "admin@kbtu.kz",
		DefaultAdminPassword: password,
		Environment:          env,
	}
}

func TestSeedDefaultsCreatesAdminWithPassword(t *testing.T) {
	db, mock, cleanup := setupSeedTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE role = 'admin'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs("admin@kbtu.kz", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := SeedDefaults(db, seedConfig("development", "admin123"))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSeedDefaultsNoAdminWithoutPasswordDevelopment(t *testing.T) {
	db, mock, cleanup := setupSeedTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE role = 'admin'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	err := SeedDefaults(db, seedConfig("development", ""))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSeedDefaultsErrorWithoutPasswordProduction(t *testing.T) {
	db, mock, cleanup := setupSeedTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE role = 'admin'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	err := SeedDefaults(db, seedConfig("production", ""))

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSeedDefaultsDoesNotCreateSecondAdmin(t *testing.T) {
	db, mock, cleanup := setupSeedTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE role = 'admin'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	err := SeedDefaults(db, seedConfig("development", "admin123"))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSeedDefaultsDBError(t *testing.T) {
	db, mock, cleanup := setupSeedTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE role = 'admin'`).
		WillReturnError(errors.New("db error"))

	err := SeedDefaults(db, seedConfig("development", "admin123"))

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSeedDefaultsInsertUsesHashedPassword(t *testing.T) {
	db, mock, cleanup := setupSeedTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE role = 'admin'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs("admin@kbtu.kz", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(time.Now().Unix(), 1))

	err := SeedDefaults(db, seedConfig("development", "admin123"))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
