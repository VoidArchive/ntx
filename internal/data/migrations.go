package data

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunMigrations applies all pending migrations using Goose
func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// MigrationStatus returns the current migration status
func MigrationStatus(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	return goose.Status(db, "migrations")
}

// MigrateDown rolls back the last migration
func MigrateDown(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	return goose.Down(db, "migrations")
}