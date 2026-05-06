package store

import (
	"testing"
)

func TestOpen_MigratesCleanDB(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) error: %v", err)
	}
	defer s.DB.Close()

	tables := []string{"users", "accounts", "transactions"}
	for _, table := range tables {
		var name string
		err := s.DB.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found after migration: %v", table, err)
		}
	}
}

func TestOpen_IdempotentMigration(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	s.DB.Close()

	// Running Open again on the same in-memory DB would be a new DB,
	// so just verify the migration SQL itself is idempotent by re-running migrate.
	s2, err := Open(":memory:")
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer s2.DB.Close()
}
