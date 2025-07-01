package data

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Database wraps the SQL connection with NTX-specific configuration
type Database struct {
	*sql.DB
	path string
}

// NewDatabase creates a new SQLite database connection
func NewDatabase() (*Database, error) {
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite connection with optimized pragmas
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=temp_store(memory)&_pragma=mmap_size(268435456)&_pragma=foreign_keys(1)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to ping database: %w, and failed to close database: %w", err, closeErr)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{
		DB:   db,
		path: dbPath,
	}, nil
}

// Path returns the database file path
func (d *Database) Path() string {
	return d.path
}

// getDatabasePath returns the standard database file location
func getDatabasePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".local", "share", "ntx", "portfolio.db"), nil
}
