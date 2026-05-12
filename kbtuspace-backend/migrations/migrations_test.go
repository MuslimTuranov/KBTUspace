package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationsUpDownFilesExist(t *testing.T) {
	files, err := filepath.Glob("*.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Fatal("expected up migrations")
	}

	for _, upFile := range files {
		downFile := strings.Replace(upFile, ".up.sql", ".down.sql", 1)

		if _, err := os.Stat(downFile); err != nil {
			t.Fatalf("missing down migration for %s", upFile)
		}
	}
}

func TestMigrationsUpDownNotEmpty(t *testing.T) {
	files, err := filepath.Glob("*.sql")
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Fatal("expected migration files")
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}

		if info.Size() == 0 {
			t.Fatalf("migration file is empty: %s", file)
		}
	}
}
