package repository

import (
	"context"
	"time"

	"ntx/internal/data/models"
)

// PortfolioRepository defines the interface for portfolio data operations
type PortfolioRepository interface {
	// Portfolio CRUD operations
	CreateHolding(ctx context.Context, holding *models.Holding) error
	GetHolding(ctx context.Context, id int64) (*models.Holding, error)
	GetHoldingBySymbol(ctx context.Context, symbol string) (*models.Holding, error)
	GetAllHoldings(ctx context.Context) ([]models.Holding, error)
	UpdateHolding(ctx context.Context, holding *models.Holding) error
	DeleteHolding(ctx context.Context, id int64) error
	DeleteHoldingBySymbol(ctx context.Context, symbol string) error

	// Portfolio aggregation operations
	GetPortfolioValue(ctx context.Context) (models.Money, error)
	GetPortfolioMetrics(ctx context.Context) (*models.PortfolioMetrics, error)
	GetHoldingSummaries(ctx context.Context) ([]models.HoldingSummary, error)

	// Bulk operations
	CreateHoldings(ctx context.Context, holdings []models.Holding) error
	UpdateHoldings(ctx context.Context, holdings []models.Holding) error
}

// MarketDataRepository defines the interface for market data operations
type MarketDataRepository interface {
	// Market data CRUD operations
	UpsertMarketData(ctx context.Context, data *models.MarketData) error
	GetMarketData(ctx context.Context, symbol string) (*models.MarketData, error)
	GetMarketDataBatch(ctx context.Context, symbols []string) (map[string]*models.MarketData, error)
	GetAllMarketData(ctx context.Context) ([]models.MarketData, error)
	DeleteMarketData(ctx context.Context, symbol string) error

	// Historical data operations
	GetHistoricalPrices(ctx context.Context, symbol string, from, to time.Time) ([]models.MarketData, error)
	GetLatestPrices(ctx context.Context, symbols []string) (map[string]models.Money, error)
	
	// Data cleanup operations
	CleanupStaleData(ctx context.Context, olderThan time.Time) error
	GetDataAge(ctx context.Context, symbol string) (time.Duration, error)
}

// TransactionRepository defines the interface for transaction history
type TransactionRepository interface {
	// Transaction CRUD operations
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	GetTransaction(ctx context.Context, id int64) (*Transaction, error)
	GetTransactionsBySymbol(ctx context.Context, symbol string) ([]Transaction, error)
	GetTransactionsByDateRange(ctx context.Context, from, to time.Time) ([]Transaction, error)
	GetAllTransactions(ctx context.Context) ([]Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction) error
	DeleteTransaction(ctx context.Context, id int64) error

	// Transaction analysis
	GetTransactionSummary(ctx context.Context, symbol string) (*TransactionSummary, error)
	CalculateAverageCost(ctx context.Context, symbol string) (models.Money, error)
}

// DatabaseManager defines the interface for database management operations
type DatabaseManager interface {
	// Connection management
	Connect(ctx context.Context, databasePath string) error
	Close() error
	Ping(ctx context.Context) error

	// Migration management
	RunMigrations(ctx context.Context) error
	GetSchemaVersion(ctx context.Context) (int, error)

	// Transaction management
	BeginTx(ctx context.Context) (Tx, error)
	
	// Backup and restore
	Backup(ctx context.Context, backupPath string) error
	Restore(ctx context.Context, backupPath string) error

	// Health and diagnostics
	GetStats(ctx context.Context) (*DatabaseStats, error)
	Vacuum(ctx context.Context) error
}

// Tx represents a database transaction
type Tx interface {
	// Repository access within transaction
	Portfolio() PortfolioRepository
	MarketData() MarketDataRepository
	Transactions() TransactionRepository

	// Transaction control
	Commit() error
	Rollback() error
}

// Transaction represents a portfolio transaction (buy/sell)
type Transaction struct {
	ID          int64             `json:"id" db:"id"`
	Type        TransactionType   `json:"type" db:"type"`
	Symbol      string            `json:"symbol" db:"symbol"`
	Quantity    models.Quantity   `json:"quantity" db:"quantity"`
	Price       models.Money      `json:"price" db:"price"`
	TotalAmount models.Money      `json:"total_amount" db:"total_amount"`
	Fees        models.Money      `json:"fees" db:"fees"`
	Date        time.Time         `json:"date" db:"date"`
	Notes       string            `json:"notes" db:"notes"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeBuy    TransactionType = "buy"
	TransactionTypeSell   TransactionType = "sell"
	TransactionTypeBonus  TransactionType = "bonus"
	TransactionTypeRights TransactionType = "rights"
	TransactionTypeSplit  TransactionType = "split"
)

// TransactionSummary provides aggregated transaction information
type TransactionSummary struct {
	Symbol          string          `json:"symbol"`
	TotalBought     models.Quantity `json:"total_bought"`
	TotalSold       models.Quantity `json:"total_sold"`
	NetQuantity     models.Quantity `json:"net_quantity"`
	TotalCost       models.Money    `json:"total_cost"`
	TotalSales      models.Money    `json:"total_sales"`
	AverageBuyPrice models.Money    `json:"average_buy_price"`
	TotalFees       models.Money    `json:"total_fees"`
	FirstBuyDate    time.Time       `json:"first_buy_date"`
	LastTradeDate   time.Time       `json:"last_trade_date"`
}

// DatabaseStats provides database health and usage statistics
type DatabaseStats struct {
	DatabaseSize    int64     `json:"database_size"`
	TableCount      int       `json:"table_count"`
	RecordCounts    map[string]int64 `json:"record_counts"`
	LastVacuum      time.Time `json:"last_vacuum"`
	LastBackup      time.Time `json:"last_backup"`
	SchemaVersion   int       `json:"schema_version"`
	ConnectionCount int       `json:"connection_count"`
}

// RepositoryError represents domain-specific repository errors
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

// ErrorType represents the type of repository error
type ErrorType string

const (
	ErrorTypeNotFound        ErrorType = "not_found"
	ErrorTypeAlreadyExists   ErrorType = "already_exists"
	ErrorTypeInvalidData     ErrorType = "invalid_data"
	ErrorTypeConstraintViolation ErrorType = "constraint_violation"
	ErrorTypeConnectionError ErrorType = "connection_error"
	ErrorTypeTransactionError ErrorType = "transaction_error"
	ErrorTypeMigrationError  ErrorType = "migration_error"
	ErrorTypeBackupError     ErrorType = "backup_error"
	ErrorTypeInternal        ErrorType = "internal"
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

// Helper function to convert various types to string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int64:
		return string(rune(val))
	case int:
		return string(rune(val))
	default:
		return "unknown"
	}
}