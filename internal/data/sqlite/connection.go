package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
	"ntx/internal/data/repository"
	"ntx/internal/security"
)

// Manager implements the DatabaseManager interface for SQLite
type Manager struct {
	db           *sql.DB
	dbPath       string
	credentials  *security.Credentials
	mu           sync.RWMutex
	migrator     *Migrator
	portfolioRepo *PortfolioRepository
	marketRepo   *MarketDataRepository
	txRepo       *TransactionRepository
}

// Config holds SQLite database configuration
type Config struct {
	DatabasePath string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLifetime time.Duration
	BusyTimeout     time.Duration
	WALMode         bool
	ForeignKeys     bool
}

// DefaultConfig returns default SQLite configuration
func DefaultConfig(configDir string) *Config {
	return &Config{
		DatabasePath:    filepath.Join(configDir, "data.db"),
		MaxOpenConns:    1, // SQLite performs better with single connection
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		BusyTimeout:     30 * time.Second,
		WALMode:         true,  // Better concurrency
		ForeignKeys:     true,  // Enforce referential integrity
	}
}

// NewManager creates a new SQLite database manager
func NewManager(config *Config, credentials *security.Credentials) *Manager {
	return &Manager{
		dbPath:      config.DatabasePath,
		credentials: credentials,
		migrator:    NewMigrator(),
	}
}

// Connect establishes connection to the SQLite database
func (m *Manager) Connect(ctx context.Context, databasePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Use provided path or default
	if databasePath != "" {
		m.dbPath = databasePath
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(m.dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return repository.NewConnectionError(
			"failed to create database directory",
			fmt.Errorf("mkdir %s: %w", dbDir, err),
		)
	}

	// Build connection string with SQLite-specific options
	connStr := m.buildConnectionString()

	// Open database connection
	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return repository.NewConnectionError(
			"failed to open database connection",
			err,
		)
	}

	// Configure connection pool
	db.SetMaxOpenConns(1) // SQLite works best with single connection
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return repository.NewConnectionError(
			"failed to ping database",
			err,
		)
	}

	// Set SQLite-specific pragmas
	if err := m.configureSQLite(ctx, db); err != nil {
		db.Close()
		return repository.NewConnectionError(
			"failed to configure SQLite",
			err,
		)
	}

	m.db = db

	// Initialize repository implementations
	m.portfolioRepo = NewPortfolioRepository(db)
	m.marketRepo = NewMarketDataRepository(db)
	m.txRepo = NewTransactionRepository(db)

	// Set proper file permissions
	if err := os.Chmod(m.dbPath, 0600); err != nil {
		return repository.NewConnectionError(
			"failed to set database file permissions",
			err,
		)
	}

	return nil
}

// buildConnectionString constructs the SQLite connection string
func (m *Manager) buildConnectionString() string {
	// Base connection string
	connStr := m.dbPath

	// Add SQLite-specific parameters
	params := []string{
		"_pragma=foreign_keys=1",    // Enable foreign key constraints
		"_pragma=journal_mode=WAL",  // Use WAL mode for better concurrency
		"_pragma=synchronous=NORMAL", // Balance safety and performance
		"_pragma=cache_size=10000",  // Increase cache size
		"_pragma=temp_store=memory", // Store temp tables in memory
		"_pragma=busy_timeout=30000", // 30 second busy timeout
	}

	for i, param := range params {
		if i == 0 {
			connStr += "?" + param
		} else {
			connStr += "&" + param
		}
	}

	return connStr
}

// configureSQLite sets SQLite-specific configuration
func (m *Manager) configureSQLite(ctx context.Context, db *sql.DB) error {
	// Additional pragma statements that need to be executed per connection
	pragmas := []string{
		"PRAGMA optimize",           // Enable query optimizer
		"PRAGMA trusted_schema=OFF", // Security: disable trusted schema
	}

	for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("failed to execute pragma %s: %w", pragma, err)
		}
	}

	return nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return nil
	}

	// Run PRAGMA optimize before closing
	if _, err := m.db.Exec("PRAGMA optimize"); err != nil {
		// Log but don't fail on optimization error
	}

	err := m.db.Close()
	m.db = nil
	m.portfolioRepo = nil
	m.marketRepo = nil
	m.txRepo = nil

	return err
}

// Ping tests the database connection
func (m *Manager) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return repository.NewConnectionError("database not connected", nil)
	}

	return m.db.PingContext(ctx)
}

// RunMigrations runs database schema migrations
func (m *Manager) RunMigrations(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return repository.NewConnectionError("database not connected", nil)
	}

	return m.migrator.Run(ctx, m.db)
}

// GetSchemaVersion returns the current schema version
func (m *Manager) GetSchemaVersion(ctx context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return 0, repository.NewConnectionError("database not connected", nil)
	}

	return m.migrator.GetVersion(ctx, m.db)
}

// BeginTx starts a database transaction
func (m *Manager) BeginTx(ctx context.Context) (repository.Tx, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return nil, repository.NewConnectionError("database not connected", nil)
	}

	sqlTx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, repository.NewTransactionError("failed to begin transaction", err)
	}

	return &Transaction{
		tx:           sqlTx,
		portfolioRepo: NewPortfolioRepository(sqlTx),
		marketRepo:   NewMarketDataRepository(sqlTx),
		txRepo:       NewTransactionRepository(sqlTx),
	}, nil
}

// Backup creates a backup of the database
func (m *Manager) Backup(ctx context.Context, backupPath string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return repository.NewConnectionError("database not connected", nil)
	}

	// Ensure backup directory exists
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return repository.NewBackupError(
			"failed to create backup directory",
			fmt.Errorf("mkdir %s: %w", backupDir, err),
		)
	}

	// Create a temporary file for atomic backup
	tempPath := backupPath + ".tmp"

	// Use SQLite VACUUM INTO for creating backup
	query := fmt.Sprintf("VACUUM INTO '%s'", tempPath)
	if _, err := m.db.ExecContext(ctx, query); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return repository.NewBackupError("failed to create backup", err)
	}

	// Atomically move temp file to final location
	if err := os.Rename(tempPath, backupPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return repository.NewBackupError("failed to finalize backup", err)
	}

	// Set proper permissions on backup file
	if err := os.Chmod(backupPath, 0600); err != nil {
		return repository.NewBackupError("failed to set backup file permissions", err)
	}

	return nil
}

// Restore restores database from backup
func (m *Manager) Restore(ctx context.Context, backupPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return repository.NewBackupError("backup file not found", err)
	}

	// Close current connection if open
	if m.db != nil {
		m.db.Close()
		m.db = nil
	}

	// Create backup of current database
	currentBackup := m.dbPath + ".pre-restore-" + time.Now().Format("20060102-150405")
	if _, err := os.Stat(m.dbPath); err == nil {
		if err := os.Rename(m.dbPath, currentBackup); err != nil {
			return repository.NewBackupError("failed to backup current database", err)
		}
	}

	// Copy backup to database location
	if err := copyFile(backupPath, m.dbPath); err != nil {
		// Try to restore original file
		os.Rename(currentBackup, m.dbPath)
		return repository.NewBackupError("failed to restore from backup", err)
	}

	// Set proper permissions
	if err := os.Chmod(m.dbPath, 0600); err != nil {
		return repository.NewBackupError("failed to set restored database permissions", err)
	}

	// Reconnect to database
	if err := m.Connect(ctx, m.dbPath); err != nil {
		// Try to restore original file
		os.Remove(m.dbPath)
		os.Rename(currentBackup, m.dbPath)
		return repository.NewBackupError("failed to connect to restored database", err)
	}

	// Clean up temporary backup
	os.Remove(currentBackup)

	return nil
}

// GetStats returns database statistics
func (m *Manager) GetStats(ctx context.Context) (*repository.DatabaseStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return nil, repository.NewConnectionError("database not connected", nil)
	}

	stats := &repository.DatabaseStats{
		RecordCounts: make(map[string]int64),
	}

	// Get database file size
	if info, err := os.Stat(m.dbPath); err == nil {
		stats.DatabaseSize = info.Size()
	}

	// Get table count and record counts
	tables := []string{"portfolio", "market_data", "transactions", "corporate_actions"}
	stats.TableCount = len(tables)

	for _, table := range tables {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := m.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			// Table might not exist yet, ignore error
			count = 0
		}
		stats.RecordCounts[table] = count
	}

	// Get schema version
	if version, err := m.migrator.GetVersion(ctx, m.db); err == nil {
		stats.SchemaVersion = version
	}

	stats.ConnectionCount = 1 // SQLite single connection

	return stats, nil
}

// Vacuum runs VACUUM command to reclaim space and optimize database
func (m *Manager) Vacuum(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return repository.NewConnectionError("database not connected", nil)
	}

	_, err := m.db.ExecContext(ctx, "VACUUM")
	if err != nil {
		return repository.NewInternalError("failed to vacuum database", err)
	}

	return nil
}

// Repository access methods
func (m *Manager) Portfolio() repository.PortfolioRepository {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.portfolioRepo
}

func (m *Manager) MarketData() repository.MarketDataRepository {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.marketRepo
}

func (m *Manager) Transactions() repository.TransactionRepository {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.txRepo
}

// Helper function to copy files
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	buf := make([]byte, 64*1024) // 64KB buffer
	for {
		n, err := source.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

