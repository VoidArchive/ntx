package sqlite

import (
	"context"
	"database/sql"
	"time"

	"ntx/internal/data/models"
	"ntx/internal/data/repository"
)

// TransactionRepository implements the repository.TransactionRepository interface for SQLite
type TransactionRepository struct {
	db QueryExecutor
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db QueryExecutor) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateTransaction creates a new transaction record
func (r *TransactionRepository) CreateTransaction(ctx context.Context, transaction *repository.Transaction) error {
	if transaction == nil || !isValidTransaction(transaction) {
		return repository.NewInvalidDataError("invalid transaction data", nil)
	}

	query := `
		INSERT INTO transactions (type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = now
	}
	if transaction.UpdatedAt.IsZero() {
		transaction.UpdatedAt = now
	}

	result, err := r.db.ExecContext(ctx, query,
		string(transaction.Type),
		transaction.Symbol,
		transaction.Quantity.Int64(),
		transaction.Price.Paisa(),
		transaction.TotalAmount.Paisa(),
		transaction.Fees.Paisa(),
		transaction.Date,
		transaction.Notes,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)

	if err != nil {
		return repository.NewInternalError("failed to create transaction", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return repository.NewInternalError("failed to get insert ID", err)
	}

	transaction.ID = id
	return nil
}

// GetTransaction retrieves a transaction by ID
func (r *TransactionRepository) GetTransaction(ctx context.Context, id int64) (*repository.Transaction, error) {
	query := `
		SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
		FROM transactions
		WHERE id = ?
	`

	transaction := &repository.Transaction{}
	var transactionType string
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transactionType,
		&transaction.Symbol,
		&transaction.Quantity,
		&transaction.Price,
		&transaction.TotalAmount,
		&transaction.Fees,
		&transaction.Date,
		&notes,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.NewNotFoundError("transaction", id)
		}
		return nil, repository.NewInternalError("failed to get transaction", err)
	}

	transaction.Type = repository.TransactionType(transactionType)
	if notes.Valid {
		transaction.Notes = notes.String
	}

	// Convert integer values back to proper types
	transaction.Quantity = models.Quantity(transaction.Quantity)
	transaction.Price = models.Money(transaction.Price)
	transaction.TotalAmount = models.Money(transaction.TotalAmount)
	transaction.Fees = models.Money(transaction.Fees)

	return transaction, nil
}

// GetTransactionsBySymbol retrieves all transactions for a specific symbol
func (r *TransactionRepository) GetTransactionsBySymbol(ctx context.Context, symbol string) ([]repository.Transaction, error) {
	query := `
		SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
		FROM transactions
		WHERE symbol = ? COLLATE NOCASE
		ORDER BY date DESC, created_at DESC
	`

	return r.queryTransactions(ctx, query, symbol)
}

// GetTransactionsByDateRange retrieves transactions within a date range
func (r *TransactionRepository) GetTransactionsByDateRange(ctx context.Context, from, to time.Time) ([]repository.Transaction, error) {
	query := `
		SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
		FROM transactions
		WHERE date BETWEEN ? AND ?
		ORDER BY date DESC, created_at DESC
	`

	return r.queryTransactions(ctx, query, from, to)
}

// GetAllTransactions retrieves all transactions
func (r *TransactionRepository) GetAllTransactions(ctx context.Context) ([]repository.Transaction, error) {
	query := `
		SELECT id, type, symbol, quantity, price, total_amount, fees, date, notes, created_at, updated_at
		FROM transactions
		ORDER BY date DESC, created_at DESC
	`

	return r.queryTransactions(ctx, query)
}

// UpdateTransaction updates an existing transaction
func (r *TransactionRepository) UpdateTransaction(ctx context.Context, transaction *repository.Transaction) error {
	if transaction == nil || !isValidTransaction(transaction) {
		return repository.NewInvalidDataError("invalid transaction data", nil)
	}

	query := `
		UPDATE transactions 
		SET type = ?, symbol = ?, quantity = ?, price = ?, total_amount = ?, fees = ?, date = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	transaction.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		string(transaction.Type),
		transaction.Symbol,
		transaction.Quantity.Int64(),
		transaction.Price.Paisa(),
		transaction.TotalAmount.Paisa(),
		transaction.Fees.Paisa(),
		transaction.Date,
		transaction.Notes,
		transaction.UpdatedAt,
		transaction.ID,
	)

	if err != nil {
		return repository.NewInternalError("failed to update transaction", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError("transaction", transaction.ID)
	}

	return nil
}

// DeleteTransaction deletes a transaction by ID
func (r *TransactionRepository) DeleteTransaction(ctx context.Context, id int64) error {
	query := "DELETE FROM transactions WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return repository.NewInternalError("failed to delete transaction", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError("transaction", id)
	}

	return nil
}

// GetTransactionSummary provides aggregated transaction information for a symbol
func (r *TransactionRepository) GetTransactionSummary(ctx context.Context, symbol string) (*repository.TransactionSummary, error) {
	query := `
		SELECT 
			symbol,
			COALESCE(SUM(CASE WHEN type IN ('buy', 'bonus', 'rights') THEN quantity ELSE 0 END), 0) as total_bought,
			COALESCE(SUM(CASE WHEN type = 'sell' THEN quantity ELSE 0 END), 0) as total_sold,
			COALESCE(SUM(CASE WHEN type IN ('buy', 'bonus', 'rights') THEN quantity ELSE -quantity END), 0) as net_quantity,
			COALESCE(SUM(CASE WHEN type IN ('buy', 'bonus', 'rights') THEN total_amount ELSE 0 END), 0) as total_cost,
			COALESCE(SUM(CASE WHEN type = 'sell' THEN total_amount ELSE 0 END), 0) as total_sales,
			COALESCE(SUM(fees), 0) as total_fees,
			MIN(CASE WHEN type = 'buy' THEN date END) as first_buy_date,
			MAX(date) as last_trade_date
		FROM transactions
		WHERE symbol = ? COLLATE NOCASE
		GROUP BY symbol
	`

	summary := &repository.TransactionSummary{}
	var firstBuyDate, lastTradeDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, symbol).Scan(
		&summary.Symbol,
		&summary.TotalBought,
		&summary.TotalSold,
		&summary.NetQuantity,
		&summary.TotalCost,
		&summary.TotalSales,
		&summary.TotalFees,
		&firstBuyDate,
		&lastTradeDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.NewNotFoundError("transaction summary", symbol)
		}
		return nil, repository.NewInternalError("failed to get transaction summary", err)
	}

	// Convert integer values back to proper types
	summary.TotalBought = models.Quantity(summary.TotalBought)
	summary.TotalSold = models.Quantity(summary.TotalSold)
	summary.NetQuantity = models.Quantity(summary.NetQuantity)
	summary.TotalCost = models.Money(summary.TotalCost)
	summary.TotalSales = models.Money(summary.TotalSales)
	summary.TotalFees = models.Money(summary.TotalFees)

	if firstBuyDate.Valid {
		summary.FirstBuyDate = firstBuyDate.Time
	}
	if lastTradeDate.Valid {
		summary.LastTradeDate = lastTradeDate.Time
	}

	// Calculate average buy price
	if summary.TotalBought.IsPositive() && summary.TotalCost.IsPositive() {
		summary.AverageBuyPrice = summary.TotalCost.DivideByQuantity(summary.TotalBought)
	}

	return summary, nil
}

// CalculateAverageCost calculates the average cost for a symbol based on buy transactions
func (r *TransactionRepository) CalculateAverageCost(ctx context.Context, symbol string) (models.Money, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type IN ('buy', 'bonus', 'rights') THEN quantity ELSE 0 END), 0) as total_quantity,
			COALESCE(SUM(CASE WHEN type IN ('buy', 'bonus', 'rights') THEN total_amount ELSE 0 END), 0) as total_cost
		FROM transactions
		WHERE symbol = ? COLLATE NOCASE
	`

	var totalQuantity, totalCost int64
	err := r.db.QueryRowContext(ctx, query, symbol).Scan(&totalQuantity, &totalCost)
	if err != nil {
		return models.Money(0), repository.NewInternalError("failed to calculate average cost", err)
	}

	if totalQuantity == 0 {
		return models.Money(0), nil
	}

	avgCost := models.Money(totalCost).DivideByQuantity(models.Quantity(totalQuantity))
	return avgCost, nil
}

// queryTransactions is a helper function to execute transaction queries
func (r *TransactionRepository) queryTransactions(ctx context.Context, query string, args ...interface{}) ([]repository.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, repository.NewInternalError("failed to query transactions", err)
	}
	defer rows.Close()

	var transactions []repository.Transaction
	for rows.Next() {
		transaction := repository.Transaction{}
		var transactionType string
		var notes sql.NullString

		err := rows.Scan(
			&transaction.ID,
			&transactionType,
			&transaction.Symbol,
			&transaction.Quantity,
			&transaction.Price,
			&transaction.TotalAmount,
			&transaction.Fees,
			&transaction.Date,
			&notes,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)

		if err != nil {
			return nil, repository.NewInternalError("failed to scan transaction row", err)
		}

		transaction.Type = repository.TransactionType(transactionType)
		if notes.Valid {
			transaction.Notes = notes.String
		}

		// Convert integer values back to proper types
		transaction.Quantity = models.Quantity(transaction.Quantity)
		transaction.Price = models.Money(transaction.Price)
		transaction.TotalAmount = models.Money(transaction.TotalAmount)
		transaction.Fees = models.Money(transaction.Fees)

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewInternalError("error iterating transaction rows", err)
	}

	return transactions, nil
}

// isValidTransaction validates transaction data
func isValidTransaction(tx *repository.Transaction) bool {
	return tx.Symbol != "" &&
		tx.Quantity.IsPositive() &&
		tx.Price.IsPositive() &&
		tx.TotalAmount.IsPositive() &&
		!tx.Date.IsZero() &&
		isValidTransactionType(tx.Type)
}

// isValidTransactionType checks if transaction type is valid
func isValidTransactionType(txType repository.TransactionType) bool {
	switch txType {
	case repository.TransactionTypeBuy,
		repository.TransactionTypeSell,
		repository.TransactionTypeBonus,
		repository.TransactionTypeRights,
		repository.TransactionTypeSplit:
		return true
	default:
		return false
	}
}