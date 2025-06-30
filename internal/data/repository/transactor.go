package repository

import (
	"context"
	"database/sql"
	"fmt"
	"ntx/internal/database"
)

type transactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) Transactor {
	return &transactor{
		db: db,
	}
}

// WithTx executes a function within a database transaction
// Creates a new Repository instance with the transaction context
func (t *transactor) WithTx(ctx context.Context, fn func(ctx context.Context, repo *Repository) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create queries instance with transaction
	queries := database.New(tx)
	
	// Create repository instance with transaction queries
	repo := &Repository{
		Portfolio:       NewPortfolioRepository(queries),
		Holding:         NewHoldingRepository(queries),
		Transaction:     NewTransactionRepository(queries),
		CorporateAction: NewCorporateActionRepository(queries),
	}

	// Execute the function with the transactional repository
	if err := fn(ctx, repo); err != nil {
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}