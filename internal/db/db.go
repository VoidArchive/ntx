package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ntx/internal/csv"
	"ntx/internal/money"

	"github.com/pressly/goose/v3"
	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection and provides high-level operations
type DB struct {
	conn    *sql.DB
	queries *Queries
}

// NewDB creates a new database connection and runs migrations
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure SQLite for better performance and consistency
	if _, err := conn.Exec(`
		PRAGMA foreign_keys = ON;
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA cache_size = 1000;
		PRAGMA temp_store = memory;
	`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to configure database: %w", err)
	}

	db := &DB{
		conn:    conn,
		queries: New(conn),
	}

	return db, nil
}

// RunMigrations executes all pending migrations using Goose
func (db *DB) RunMigrations(migrationDir string) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db.conn, migrationDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Queries returns the SQLC generated queries struct
func (db *DB) Queries() *Queries {
	return db.queries
}

// WithTx executes a function within a database transaction
func (db *DB) WithTx(fn func(*Queries) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := db.queries.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ConvertCSVTransaction converts a CSV transaction to database format
func ConvertCSVTransaction(csvTx csv.Transaction) CreateTransactionParams {
	// Convert date to string in ISO format
	dateStr := csvTx.Date.Format("2006-01-02")
	
	// Convert price to paisa (integer), handle zero value
	var priceNullInt sql.NullInt64
	if !csvTx.Price.IsZero() {
		priceNullInt = sql.NullInt64{
			Int64: csvTx.Price.Paisa(),
			Valid: true,
		}
	}

	// Handle optional description
	var descNullString sql.NullString
	if csvTx.Description != "" {
		descNullString = sql.NullString{
			String: csvTx.Description,
			Valid:  true,
		}
	}

	return CreateTransactionParams{
		Scrip:           csvTx.Scrip,
		Date:            dateStr,
		Quantity:        int64(csvTx.Quantity),
		Price:           priceNullInt,
		TransactionType: csvTx.TransactionType.String(),
		Description:     descNullString,
	}
}

// ConvertDBTransaction converts a database transaction to CSV transaction format
func ConvertDBTransaction(dbTx Transaction) csv.Transaction {
	// Parse date from string
	date, _ := time.Parse("2006-01-02", dbTx.Date)
	
	// Convert price from paisa (integer) to Money
	var price money.Money
	if dbTx.Price.Valid {
		price = money.NewMoneyFromPaisa(dbTx.Price.Int64)
	}

	// Parse transaction type
	var transactionType csv.TransactionType
	switch dbTx.TransactionType {
	case "IPO":
		transactionType = csv.TransactionTypeIPO
	case "BONUS":
		transactionType = csv.TransactionTypeBonus
	case "RIGHTS":
		transactionType = csv.TransactionTypeRights
	case "MERGER":
		transactionType = csv.TransactionTypeMerger
	case "REARRANGEMENT":
		transactionType = csv.TransactionTypeRearrangement
	default:
		transactionType = csv.TransactionTypeRegular
	}

	description := ""
	if dbTx.Description.Valid {
		description = dbTx.Description.String
	}

	return csv.Transaction{
		Scrip:           dbTx.Scrip,
		Date:            date,
		Quantity:        int(dbTx.Quantity),
		Price:           price,
		TransactionType: transactionType,
		Description:     description,
		// BalanceAfter is not stored in DB, would need to be calculated
	}
}

// InsertCSVTransaction is a convenience method to insert a CSV transaction
func (db *DB) InsertCSVTransaction(csvTx csv.Transaction) (*Transaction, error) {
	params := ConvertCSVTransaction(csvTx)
	tx, err := db.queries.CreateTransaction(context.Background(), params)
	return &tx, err
}

// GetCSVTransactions retrieves all transactions as CSV transaction format
func (db *DB) GetCSVTransactions() ([]csv.Transaction, error) {
	dbTransactions, err := db.queries.GetAllTransactions(context.Background())
	if err != nil {
		return nil, err
	}

	csvTransactions := make([]csv.Transaction, len(dbTransactions))
	for i, dbTx := range dbTransactions {
		csvTransactions[i] = ConvertDBTransaction(dbTx)
	}

	return csvTransactions, nil
}

// GetCSVTransactionsByScripOrdered retrieves transactions for a specific scrip as CSV format
func (db *DB) GetCSVTransactionsByScripOrdered(scrip string) ([]csv.Transaction, error) {
	dbTransactions, err := db.queries.GetTransactionsByScripOrderedByDate(context.Background(), scrip)
	if err != nil {
		return nil, err
	}

	csvTransactions := make([]csv.Transaction, len(dbTransactions))
	for i, dbTx := range dbTransactions {
		csvTransactions[i] = ConvertDBTransaction(dbTx)
	}

	return csvTransactions, nil
}

// Required methods from implementation checklist:

// InsertTransaction inserts a CSV transaction into the database
func (db *DB) InsertTransaction(tx csv.Transaction) error {
	params := ConvertCSVTransaction(tx)
	_, err := db.queries.CreateTransaction(context.Background(), params)
	return err
}

// GetAllTransactions retrieves all transactions as CSV transaction format
func (db *DB) GetAllTransactions() ([]csv.Transaction, error) {
	return db.GetCSVTransactions()
}

// GetTransactionsByScrip retrieves transactions for a specific scrip
func (db *DB) GetTransactionsByScrip(scrip string) ([]csv.Transaction, error) {
	return db.GetCSVTransactionsByScripOrdered(scrip)
}