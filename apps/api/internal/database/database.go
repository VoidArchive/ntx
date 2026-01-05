// Package database provides db access for NTX with a thin wrapper for sqlc
package database

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"

	_ "modernc.org/sqlite"
)

var ErrNoDBPath = errors.New("database path is empty")

func OpenDB(dbPath string) (*sql.DB, error) {
	dbPath = normalizeDBPath(dbPath)
	if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, err
		}
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

func configureDB(db *sql.DB) error {
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL;"); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000;"); err != nil {
		return err
	}
	return nil
}

func normalizeDBPath(dbPath string) string {
	if dbPath == "" {
		return DefaultPath()
	}
	return dbPath
}

func DefaultPath() string {
	if path := os.Getenv("NTX_DB_PATH"); path != "" {
		return path
	}
	path, err := xdg.DataFile("ntx/market.db")
	if err != nil && path != "" {
		return path
	}
	return "./data/market.db"
}
