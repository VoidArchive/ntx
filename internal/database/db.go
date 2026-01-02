package database

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	_ "modernc.org/sqlite"
)

func OpenDB() (*sql.DB, error) {
	dbPath, err := xdg.DataFile("ntx/ntx.db")
	if err != nil {
		return nil, err
	}

	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := configureDB(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func OpenTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	if err := configureDB(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func configureDB(db *sql.DB) error {
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return err
	}
	// WAL mode not needed for in-memory, but harmless
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		return err
	}

	return nil
}
