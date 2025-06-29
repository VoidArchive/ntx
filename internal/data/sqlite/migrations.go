package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"ntx/internal/data/repository"
)

// Migration represents a database schema migration
type Migration struct {
	Version     int
	Description string
	Up          string
	Down        string
}

// Migrator manages database schema migrations
type Migrator struct {
	migrations []Migration
}

// NewMigrator creates a new migration manager
func NewMigrator() *Migrator {
	m := &Migrator{}
	m.registerMigrations()
	return m
}

// registerMigrations registers all available migrations
func (m *Migrator) registerMigrations() {
	m.migrations = []Migration{
		{
			Version:     1,
			Description: "Initial schema: portfolio and market_data tables",
			Up: `
				-- Create schema_migrations table to track migrations
				CREATE TABLE IF NOT EXISTS schema_migrations (
					version INTEGER PRIMARY KEY,
					applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					description TEXT
				);

				-- Portfolio holdings table (using integer paisa storage)
				CREATE TABLE portfolio (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					symbol TEXT NOT NULL UNIQUE,
					quantity INTEGER NOT NULL CHECK (quantity >= 0),
					avg_cost INTEGER NOT NULL CHECK (avg_cost > 0), -- paisa per share
					purchase_date DATE,
					notes TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);

				-- Create index on symbol for fast lookups
				CREATE INDEX idx_portfolio_symbol ON portfolio(symbol);
				CREATE INDEX idx_portfolio_purchase_date ON portfolio(purchase_date);

				-- Market data table (using integer paisa storage)
				CREATE TABLE market_data (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					symbol TEXT NOT NULL,
					last_price INTEGER NOT NULL CHECK (last_price > 0), -- paisa
					change_amount INTEGER DEFAULT 0, -- paisa
					change_percent INTEGER DEFAULT 0, -- basis points
					volume INTEGER DEFAULT 0 CHECK (volume >= 0),
					timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
					UNIQUE(symbol, timestamp)
				);

				-- Create indexes for market data
				CREATE INDEX idx_market_data_symbol ON market_data(symbol);
				CREATE INDEX idx_market_data_timestamp ON market_data(timestamp);
				CREATE INDEX idx_market_data_symbol_timestamp ON market_data(symbol, timestamp DESC);

				-- Trigger to update portfolio updated_at timestamp
				CREATE TRIGGER portfolio_updated_at 
				AFTER UPDATE ON portfolio
				BEGIN
					UPDATE portfolio SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
				END;
			`,
			Down: `
				DROP TRIGGER IF EXISTS portfolio_updated_at;
				DROP INDEX IF EXISTS idx_market_data_symbol_timestamp;
				DROP INDEX IF EXISTS idx_market_data_timestamp;
				DROP INDEX IF EXISTS idx_market_data_symbol;
				DROP INDEX IF EXISTS idx_portfolio_purchase_date;
				DROP INDEX IF EXISTS idx_portfolio_symbol;
				DROP TABLE IF EXISTS market_data;
				DROP TABLE IF EXISTS portfolio;
				DROP TABLE IF EXISTS schema_migrations;
			`,
		},
		{
			Version:     2,
			Description: "Add transactions table for trade history",
			Up: `
				-- Transactions table for tracking buy/sell history
				CREATE TABLE transactions (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					type TEXT NOT NULL CHECK (type IN ('buy', 'sell', 'bonus', 'rights', 'split')),
					symbol TEXT NOT NULL,
					quantity INTEGER NOT NULL CHECK (quantity > 0),
					price INTEGER NOT NULL CHECK (price > 0), -- paisa per share
					total_amount INTEGER NOT NULL CHECK (total_amount > 0), -- paisa
					fees INTEGER DEFAULT 0 CHECK (fees >= 0), -- paisa
					date DATE NOT NULL,
					notes TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);

				-- Create indexes for transactions
				CREATE INDEX idx_transactions_symbol ON transactions(symbol);
				CREATE INDEX idx_transactions_date ON transactions(date DESC);
				CREATE INDEX idx_transactions_type ON transactions(type);
				CREATE INDEX idx_transactions_symbol_date ON transactions(symbol, date DESC);

				-- Trigger to update transactions updated_at timestamp
				CREATE TRIGGER transactions_updated_at 
				AFTER UPDATE ON transactions
				BEGIN
					UPDATE transactions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
				END;
			`,
			Down: `
				DROP TRIGGER IF EXISTS transactions_updated_at;
				DROP INDEX IF EXISTS idx_transactions_symbol_date;
				DROP INDEX IF EXISTS idx_transactions_type;
				DROP INDEX IF EXISTS idx_transactions_date;
				DROP INDEX IF EXISTS idx_transactions_symbol;
				DROP TABLE IF EXISTS transactions;
			`,
		},
		{
			Version:     3,
			Description: "Add corporate actions table for NEPSE-specific features",
			Up: `
				-- Corporate actions table for tracking bonus shares, dividends, etc.
				CREATE TABLE corporate_actions (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					symbol TEXT NOT NULL,
					action_type TEXT NOT NULL CHECK (action_type IN ('bonus', 'dividend', 'rights', 'split')),
					announcement_date DATE,
					record_date DATE,
					ex_date DATE,
					ratio TEXT, -- e.g., "1:5" for bonus shares
					dividend_amount INTEGER, -- paisa for dividend actions
					processed BOOLEAN DEFAULT FALSE,
					processed_date DATETIME,
					notes TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);

				-- Create indexes for corporate actions
				CREATE INDEX idx_corporate_actions_symbol ON corporate_actions(symbol);
				CREATE INDEX idx_corporate_actions_type ON corporate_actions(action_type);
				CREATE INDEX idx_corporate_actions_processed ON corporate_actions(processed);
				CREATE INDEX idx_corporate_actions_ex_date ON corporate_actions(ex_date);
				CREATE INDEX idx_corporate_actions_symbol_date ON corporate_actions(symbol, ex_date DESC);

				-- Trigger to update corporate_actions updated_at timestamp
				CREATE TRIGGER corporate_actions_updated_at 
				AFTER UPDATE ON corporate_actions
				BEGIN
					UPDATE corporate_actions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
				END;
			`,
			Down: `
				DROP TRIGGER IF EXISTS corporate_actions_updated_at;
				DROP INDEX IF EXISTS idx_corporate_actions_symbol_date;
				DROP INDEX IF EXISTS idx_corporate_actions_ex_date;
				DROP INDEX IF EXISTS idx_corporate_actions_processed;
				DROP INDEX IF EXISTS idx_corporate_actions_type;
				DROP INDEX IF EXISTS idx_corporate_actions_symbol;
				DROP TABLE IF EXISTS corporate_actions;
			`,
		},
		{
			Version:     4,
			Description: "Add backup tracking and optimize schema",
			Up: `
				-- Backup tracking table
				CREATE TABLE backup_history (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					backup_path TEXT NOT NULL,
					backup_size INTEGER,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					restored_at DATETIME,
					notes TEXT
				);

				-- Create index for backup history
				CREATE INDEX idx_backup_history_created_at ON backup_history(created_at DESC);

				-- Add validation constraints and optimize existing tables
				
				-- Update portfolio table constraints
				CREATE TABLE portfolio_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					symbol TEXT NOT NULL UNIQUE COLLATE NOCASE,
					quantity INTEGER NOT NULL CHECK (quantity >= 0),
					avg_cost INTEGER NOT NULL CHECK (avg_cost > 0),
					purchase_date DATE,
					notes TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);

				-- Copy data from old portfolio table
				INSERT INTO portfolio_new SELECT * FROM portfolio;

				-- Drop old table and rename new one
				DROP TABLE portfolio;
				ALTER TABLE portfolio_new RENAME TO portfolio;

				-- Recreate indexes and triggers for portfolio
				CREATE INDEX idx_portfolio_symbol ON portfolio(symbol);
				CREATE INDEX idx_portfolio_purchase_date ON portfolio(purchase_date);
				
				CREATE TRIGGER portfolio_updated_at 
				AFTER UPDATE ON portfolio
				BEGIN
					UPDATE portfolio SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
				END;

				-- Add cleanup view for stale market data
				CREATE VIEW stale_market_data AS
				SELECT symbol, MAX(timestamp) as latest_timestamp
				FROM market_data 
				WHERE timestamp < datetime('now', '-1 day')
				GROUP BY symbol;
			`,
			Down: `
				DROP VIEW IF EXISTS stale_market_data;
				DROP TRIGGER IF EXISTS portfolio_updated_at;
				DROP INDEX IF EXISTS idx_portfolio_purchase_date;
				DROP INDEX IF EXISTS idx_portfolio_symbol;
				DROP INDEX IF EXISTS idx_backup_history_created_at;
				DROP TABLE IF EXISTS backup_history;
				
				-- Note: Reverting portfolio table changes is complex and risky
				-- In production, consider making this irreversible
			`,
		},
	}

	// Sort migrations by version to ensure correct order
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})
}

// Run executes all pending migrations
func (m *Migrator) Run(ctx context.Context, db *sql.DB) error {
	// Ensure schema_migrations table exists
	if err := m.ensureMigrationsTable(ctx, db); err != nil {
		return repository.NewMigrationError("failed to ensure migrations table", err)
	}

	// Get current schema version
	currentVersion, err := m.GetVersion(ctx, db)
	if err != nil {
		return repository.NewMigrationError("failed to get current schema version", err)
	}

	// Run pending migrations
	for _, migration := range m.migrations {
		if migration.Version <= currentVersion {
			continue // Skip already applied migrations
		}

		if err := m.runMigration(ctx, db, migration); err != nil {
			return repository.NewMigrationError(
				fmt.Sprintf("failed to run migration %d: %s", migration.Version, migration.Description),
				err,
			)
		}
	}

	return nil
}

// ensureMigrationsTable creates the schema_migrations table if it doesn't exist
func (m *Migrator) ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			description TEXT
		)
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// runMigration executes a single migration
func (m *Migrator) runMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	// Start transaction for migration
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration in schema_migrations table
	recordQuery := `
		INSERT OR REPLACE INTO schema_migrations (version, applied_at, description)
		VALUES (?, ?, ?)
	`
	if _, err := tx.ExecContext(ctx, recordQuery, migration.Version, time.Now(), migration.Description); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// GetVersion returns the current schema version
func (m *Migrator) GetVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	query := "SELECT COALESCE(MAX(version), 0) FROM schema_migrations"
	
	err := db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		// If table doesn't exist, version is 0
		if err.Error() == "no such table: schema_migrations" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to query schema version: %w", err)
	}

	return version, nil
}

// Rollback rolls back to a specific version (be careful with this!)
func (m *Migrator) Rollback(ctx context.Context, db *sql.DB, targetVersion int) error {
	currentVersion, err := m.GetVersion(ctx, db)
	if err != nil {
		return repository.NewMigrationError("failed to get current version", err)
	}

	if targetVersion >= currentVersion {
		return repository.NewMigrationError("target version must be less than current version", nil)
	}

	// Find migrations to rollback (in reverse order)
	var migrationsToRollback []Migration
	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if migration.Version > targetVersion && migration.Version <= currentVersion {
			migrationsToRollback = append(migrationsToRollback, migration)
		}
	}

	// Execute rollback migrations
	for _, migration := range migrationsToRollback {
		if err := m.rollbackMigration(ctx, db, migration); err != nil {
			return repository.NewMigrationError(
				fmt.Sprintf("failed to rollback migration %d", migration.Version),
				err,
			)
		}
	}

	return nil
}

// rollbackMigration executes a single migration rollback
func (m *Migrator) rollbackMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute rollback SQL
	if _, err := tx.ExecContext(ctx, migration.Down); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Remove migration record
	removeQuery := "DELETE FROM schema_migrations WHERE version = ?"
	if _, err := tx.ExecContext(ctx, removeQuery, migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	return nil
}

// GetAppliedMigrations returns a list of applied migrations
func (m *Migrator) GetAppliedMigrations(ctx context.Context, db *sql.DB) ([]Migration, error) {
	query := `
		SELECT version, applied_at, description 
		FROM schema_migrations 
		ORDER BY version
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var applied []Migration
	for rows.Next() {
		var migration Migration
		var appliedAt time.Time
		
		if err := rows.Scan(&migration.Version, &appliedAt, &migration.Description); err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}
		
		applied = append(applied, migration)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration rows: %w", err)
	}

	return applied, nil
}

// GetPendingMigrations returns a list of pending migrations
func (m *Migrator) GetPendingMigrations(ctx context.Context, db *sql.DB) ([]Migration, error) {
	currentVersion, err := m.GetVersion(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	var pending []Migration
	for _, migration := range m.migrations {
		if migration.Version > currentVersion {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

