package database

import (
	"path/filepath"
	"testing"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// OpenTestDB gives a test its own migrated SQLite file, so tests never touch the real market.db.
func OpenTestDB(t testing.TB) *sqlc.Queries {
	t.Helper()
	db, err := OpenDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return sqlc.New(db)
}
