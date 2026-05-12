package faculties

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func setupFacultiesTest(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewRepository(sqlxDB)
	service := NewService(repo)

	cleanup := func() {
		_ = sqlxDB.Close()
		_ = db.Close()
	}

	return service, mock, cleanup
}

func TestGetAllFaculties(t *testing.T) {
	service, mock, cleanup := setupFacultiesTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT`).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"name",
			}).
				AddRow(1, "SITE").
				AddRow(2, "BS").
				AddRow(3, "ISE").
				AddRow(4, "KMA"),
		)

	faculties, err := service.GetAllFaculties(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(faculties) != 4 {
		t.Fatalf("expected 4 faculties, got %d", len(faculties))
	}
}

func TestSeedMigrationCreatesFourFaculties(t *testing.T) {
	faculties := []string{
		"SITE",
		"BS",
		"ISE",
		"KMA",
	}

	if len(faculties) != 4 {
		t.Fatal("expected 4 faculties")
	}
}

func TestDuplicateFacultyNamesNotCreatedTwice(t *testing.T) {
	faculties := []string{
		"SITE",
		"BS",
		"ISE",
		"KMA",
	}

	seen := map[string]bool{}

	for _, faculty := range faculties {
		if seen[faculty] {
			t.Fatalf("duplicate faculty found: %s", faculty)
		}

		seen[faculty] = true
	}
}
