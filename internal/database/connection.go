package database

import (
	"database/sql"
	"fmt"
)

// Service wraps SQLC queries with database connection
type Service struct {
	*Queries
	db *sql.DB
}

// NewService creates a new database service with SQLC queries
func NewService(db *sql.DB) *Service {
	return &Service{
		Queries: New(db),
		db:      db,
	}
}

// Close closes the database connection
func (s *Service) Close() error {
	return s.db.Close()
}

// WithTx executes a function within a database transaction
func (s *Service) WithTx(fn func(*Queries) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit()
}