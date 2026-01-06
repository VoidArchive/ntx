package database

import (
	"database/sql"
	"embed"
	"errors"
	"io"
	"log"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*sql
var migrations embed.FS

func AutoMigrate(db *sql.DB) error {
	goose.SetBaseFS(migrations)
	goose.SetLogger(log.New(io.Discard, "", 0))

	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil &&
		!errors.Is(err, goose.ErrNoNextVersion) {
		return err
	}
	return nil
}

func MigrateDown(db *sql.DB) error {
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}
	return goose.Down(db, "migrations")
}
