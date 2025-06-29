// Package database provides database connection management with Goose migrations and SQLC queries
//
// This file implements the database connection manager that integrates with Goose for migrations
// and SQLC for type-safe queries. It maintains compatibility with the existing SQLite configuration
// and backup/restore functionality while modernizing the database layer.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // SQLite driver
)

// DatabaseStats provides database health and usage statistics
// This maintains compatibility with the existing database monitoring functionality.
type DatabaseStats struct {
	DatabaseSize    int64            `json:"database_size"`
	TableCount      int              `json:"table_count"`
	RecordCounts    map[string]int64 `json:"record_counts"`
	LastVacuum      time.Time        `json:"last_vacuum"`
	LastBackup      time.Time        `json:"last_backup"`
	SchemaVersion   int              `json:"schema_version"`
	ConnectionCount int              `json:"connection_count"`
}

// RepositoryError represents domain-specific database errors
// This maintains compatibility with existing error handling while providing modern error types.
type RepositoryError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Cause   error     `json:"cause,omitempty"`
}

func (e *RepositoryError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *RepositoryError) Unwrap() error {
	return e.Cause
}

// ErrorType represents the type of database error
type ErrorType string

const (
	ErrorTypeNotFound            ErrorType = "not_found"
	ErrorTypeAlreadyExists       ErrorType = "already_exists"
	ErrorTypeInvalidData         ErrorType = "invalid_data"
	ErrorTypeConstraintViolation ErrorType = "constraint_violation"
	ErrorTypeConnectionError     ErrorType = "connection_error"
	ErrorTypeTransactionError    ErrorType = "transaction_error"
	ErrorTypeMigrationError      ErrorType = "migration_error"
	ErrorTypeBackupError         ErrorType = "backup_error"
	ErrorTypeInternal            ErrorType = "internal"
)

// Helper functions for creating domain-specific errors

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string, identifier interface{}) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeNotFound,
		Message: resource + " not found: " + toString(identifier),
	}
}

// NewAlreadyExistsError creates an already exists error
func NewAlreadyExistsError(resource string, identifier interface{}) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeAlreadyExists,
		Message: resource + " already exists: " + toString(identifier),
	}
}

// NewInvalidDataError creates an invalid data error
func NewInvalidDataError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeInvalidData,
		Message: message,
		Cause:   cause,
	}
}

// NewConstraintViolationError creates a constraint violation error
func NewConstraintViolationError(constraint string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeConstraintViolation,
		Message: "constraint violation: " + constraint,
		Cause:   cause,
	}
}

// NewConnectionError creates a connection error
func NewConnectionError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeConnectionError,
		Message: message,
		Cause:   cause,
	}
}

// NewInternalError creates an internal error
func NewInternalError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeInternal,
		Message: message,
		Cause:   cause,
	}
}

// NewTransactionError creates a transaction error
func NewTransactionError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeTransactionError,
		Message: message,
		Cause:   cause,
	}
}

// NewMigrationError creates a migration error
func NewMigrationError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeMigrationError,
		Message: message,
		Cause:   cause,
	}
}

// NewBackupError creates a backup error
func NewBackupError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeBackupError,
		Message: message,
		Cause:   cause,
	}
}

// toString converts various types to string representation
func toString(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v)
}

// Manager handles database connections, migrations, and operations using Goose + SQLC
// It maintains compatibility with the existing connection configuration while providing
// modern migration management and type-safe query operations.
type Manager struct {
	db           *sql.DB
	queries      *Queries
	databasePath string
	isConnected  bool
}

// NewManager creates a new database manager with Goose + SQLC integration
// This replaces the custom migration system with industry-standard Goose migrations
// while preserving all existing SQLite configuration and functionality.
func NewManager() *Manager {
	return &Manager{
		isConnected: false,
	}
}

// Connect establishes database connection with SQLite-specific optimizations
// This method preserves all the existing SQLite configuration (WAL mode, busy timeout, etc.)
// while integrating Goose for migration management.
func (m *Manager) Connect(ctx context.Context, databasePath string) error {
	if m.isConnected {
		return NewConnectionError("already connected", nil)
	}

	// Build connection string with SQLite-specific options (same as original)
	connStr := m.buildConnectionString(databasePath)

	// Open database connection
	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return NewConnectionError("failed to open database connection", err)
	}

	// Configure connection pool (SQLite works best with single connection)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Connections never expire

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return NewConnectionError("failed to ping database", err)
	}

	// Configure SQLite-specific settings
	if err := m.configureSQLite(ctx, db); err != nil {
		db.Close()
		return NewConnectionError("failed to configure SQLite", err)
	}

	m.db = db
	m.queries = New(db)
	m.databasePath = databasePath
	m.isConnected = true

	return nil
}

// buildConnectionString creates SQLite connection string with optimizations
// This maintains the same SQLite configuration as the original system for consistency.
func (m *Manager) buildConnectionString(databasePath string) string {
	params := []string{
		"cache=shared",        // Enable shared cache
		"mode=rwc",            // Read-write-create mode
		"_journal_mode=WAL",   // Write-Ahead Logging for better concurrency
		"_synchronous=NORMAL", // Balance between safety and performance
		"_busy_timeout=30000", // 30 second busy timeout
		"_foreign_keys=ON",    // Enable foreign key constraints
		"_cache_size=-64000",  // 64MB cache size (negative = KB)
		"_temp_store=MEMORY",  // Store temporary data in memory
	}

	return fmt.Sprintf("file:%s?%s", databasePath, strings.Join(params, "&"))
}

// configureSQLite applies SQLite-specific settings for optimal performance
// This preserves the existing pragma settings that have been tested in production.
func (m *Manager) configureSQLite(ctx context.Context, db *sql.DB) error {
	pragmas := []string{
		"PRAGMA optimize", // SQLite query optimization
	}

	for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
		}
	}

	return nil
}

// RunMigrations executes all pending Goose migrations
// This replaces the custom migration system with industry-standard Goose migrations
// while maintaining the same automatic migration behavior on startup.
func (m *Manager) RunMigrations(ctx context.Context) error {
	if !m.isConnected {
		return NewConnectionError("not connected to database", nil)
	}

	// Set Goose dialect to SQLite
	goose.SetDialect("sqlite3")

	// Get migrations directory - handle both running from project root and cmd/ntx
	migrationsDir := "internal/data/migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try from parent directory (when running from cmd/ntx)
		migrationsDir = "../../internal/data/migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			return NewMigrationError("migrations directory not found", err)
		}
	}

	// Run migrations up to latest version
	if err := goose.UpContext(ctx, m.db, migrationsDir); err != nil {
		return NewMigrationError("failed to run migrations", err)
	}

	return nil
}

// GetSchemaVersion returns the current schema version using Goose
// This provides compatibility with existing version checking while using Goose's version tracking.
func (m *Manager) GetSchemaVersion(ctx context.Context) (int64, error) {
	if !m.isConnected {
		return 0, NewConnectionError("not connected to database", nil)
	}

	goose.SetDialect("sqlite3")
	version, err := goose.GetDBVersionContext(ctx, m.db)
	if err != nil {
		return 0, NewInternalError("failed to get schema version", err)
	}

	return version, nil
}

// Queries returns the SQLC-generated queries interface
// This provides access to all type-safe database operations, replacing the repository pattern
// with direct access to generated query methods.
func (m *Manager) Queries() *Queries {
	return m.queries
}

// DB returns the underlying sql.DB connection
// This allows for custom transactions and operations not covered by SQLC queries.
func (m *Manager) DB() *sql.DB {
	return m.db
}

// Close closes the database connection and cleans up resources
func (m *Manager) Close() error {
	if !m.isConnected {
		return nil
	}

	err := m.db.Close()
	m.db = nil
	m.queries = nil
	m.isConnected = false
	m.databasePath = ""

	if err != nil {
		return NewConnectionError("failed to close database", err)
	}

	return nil
}

// Ping tests the database connection
func (m *Manager) Ping(ctx context.Context) error {
	if !m.isConnected {
		return NewConnectionError("not connected to database", nil)
	}

	return m.db.PingContext(ctx)
}

// BeginTx starts a new database transaction
// This maintains compatibility with existing transaction handling patterns.
func (m *Manager) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if !m.isConnected {
		return nil, NewConnectionError("not connected to database", nil)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, NewInternalError("failed to begin transaction", err)
	}

	return tx, nil
}

// WithTx creates SQLC queries that use a transaction
// This allows SQLC queries to participate in transactions for atomic operations.
func (m *Manager) WithTx(tx *sql.Tx) *Queries {
	return New(tx)
}

// IsConnected returns whether the database is currently connected
func (m *Manager) IsConnected() bool {
	return m.isConnected
}

// GetDatabasePath returns the path to the database file
func (m *Manager) GetDatabasePath() string {
	return m.databasePath
}

// Vacuum runs VACUUM on the database to reclaim space and optimize performance
// This maintains the existing database maintenance functionality.
func (m *Manager) Vacuum(ctx context.Context) error {
	if !m.isConnected {
		return NewConnectionError("not connected to database", nil)
	}

	_, err := m.db.ExecContext(ctx, "VACUUM")
	if err != nil {
		return NewInternalError("failed to vacuum database", err)
	}

	return nil
}

// GetStats returns database statistics for monitoring and debugging
// This preserves the existing database statistics functionality.
func (m *Manager) GetStats(ctx context.Context) (*DatabaseStats, error) {
	if !m.isConnected {
		return nil, NewConnectionError("not connected to database", nil)
	}

	stats := &DatabaseStats{
		RecordCounts: make(map[string]int64),
	}

	// Get table counts
	tables := []string{"portfolio", "market_data", "transactions", "corporate_actions", "backup_history"}
	for _, table := range tables {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := m.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			// If table doesn't exist, count is 0
			if strings.Contains(err.Error(), "no such table") {
				count = 0
			} else {
				return nil, NewInternalError(fmt.Sprintf("failed to count %s", table), err)
			}
		}
		stats.RecordCounts[table] = count
	}

	// Get schema version
	if version, err := m.GetSchemaVersion(ctx); err == nil {
		stats.SchemaVersion = int(version)
	}

	return stats, nil
}

// Backup creates a backup of the database using SQLite's backup API
// This method maintains compatibility with the existing backup service interface
// while providing efficient SQLite-specific backup functionality.
func (m *Manager) Backup(ctx context.Context, backupPath string) error {
	if !m.isConnected {
		return NewConnectionError("not connected to database", nil)
	}

	// For SQLite, we can use the .backup command or copy the file
	// Since we're using WAL mode, we need to checkpoint first
	if _, err := m.db.ExecContext(ctx, "PRAGMA wal_checkpoint(FULL)"); err != nil {
		return NewBackupError("failed to checkpoint WAL", err)
	}

	// Create backup using SQLite's VACUUM INTO command (available in SQLite 3.27+)
	query := fmt.Sprintf("VACUUM INTO '%s'", backupPath)
	if _, err := m.db.ExecContext(ctx, query); err != nil {
		return NewBackupError("failed to create backup", err)
	}

	return nil
}

// Restore restores the database from a backup file
// This method maintains compatibility with the existing backup service interface
// while providing efficient SQLite-specific restore functionality.
func (m *Manager) Restore(ctx context.Context, backupPath string) error {
	if !m.isConnected {
		return NewConnectionError("not connected to database", nil)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return NewBackupError("backup file not found", err)
	}

	// Close current connection
	if err := m.Close(); err != nil {
		return NewBackupError("failed to close current connection", err)
	}

	// Create a backup of the current database
	currentBackup := m.databasePath + ".pre-restore-" + time.Now().Format("20060102-150405")
	if _, err := os.Stat(m.databasePath); err == nil {
		if err := os.Rename(m.databasePath, currentBackup); err != nil {
			return NewBackupError("failed to backup current database", err)
		}
	}

	// Copy backup file to database location
	if err := copyFile(backupPath, m.databasePath); err != nil {
		// Try to restore original file
		if _, statErr := os.Stat(currentBackup); statErr == nil {
			os.Rename(currentBackup, m.databasePath)
		}
		return NewBackupError("failed to restore from backup", err)
	}

	// Set proper permissions
	if err := os.Chmod(m.databasePath, 0600); err != nil {
		return NewBackupError("failed to set restored database permissions", err)
	}

	// Reconnect to database
	if err := m.Connect(ctx, m.databasePath); err != nil {
		// Try to restore original file
		os.Remove(m.databasePath)
		if _, statErr := os.Stat(currentBackup); statErr == nil {
			os.Rename(currentBackup, m.databasePath)
		}
		return NewBackupError("failed to connect to restored database", err)
	}

	// Clean up temporary backup
	os.Remove(currentBackup)

	return nil
}

// copyFile copies a file from src to dst
// This helper function is used for backup and restore operations to ensure
// atomic file operations with proper error handling.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Ensure all data is written to disk
	return destFile.Sync()
}
