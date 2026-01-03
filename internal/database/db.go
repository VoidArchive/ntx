// Package database provides database access for NTX.
package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	_ "modernc.org/sqlite"
)

// OpenDB opens the database at the given path.
// Creates parent directories if needed.
func OpenDB(dbPath string) (*sql.DB, error) {
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := configureDB(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

// DefaultCLIPath returns the default database path for the CLI.
// Uses XDG data directory: ~/.local/share/ntx/ntx.db
func DefaultCLIPath() (string, error) {
	return xdg.DataFile("ntx/ntx.db")
}

// DefaultServerPath returns the database path for the server.
// Uses NTX_DB_PATH env var, or ~/.local/share/ntx/market.db for development.
func DefaultServerPath() string {
	if path := os.Getenv("NTX_DB_PATH"); path != "" {
		return path
	}
	path, err := xdg.DataFile("ntx/market.db")
	if err != nil {
		return "./market.db"
	}
	return path
}

func OpenTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	if err := configureDB(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func configureDB(db *sql.DB) error {
	ctx := context.Background()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		return err
	}
	// WAL mode not needed for in-memory, but harmless
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL;"); err != nil {
		return err
	}

	return nil
}
