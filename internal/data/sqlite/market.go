package sqlite

import (
	"context"
	"database/sql"
	"time"

	"ntx/internal/data/models"
	"ntx/internal/data/repository"
)

// MarketDataRepository implements the repository.MarketDataRepository interface for SQLite
type MarketDataRepository struct {
	db QueryExecutor
}

// NewMarketDataRepository creates a new market data repository
func NewMarketDataRepository(db QueryExecutor) *MarketDataRepository {
	return &MarketDataRepository{db: db}
}

// UpsertMarketData inserts or updates market data for a symbol
func (r *MarketDataRepository) UpsertMarketData(ctx context.Context, data *models.MarketData) error {
	if data == nil || data.Symbol == "" {
		return repository.NewInvalidDataError("invalid market data", nil)
	}

	// Use INSERT OR REPLACE to handle upsert
	query := `
		INSERT OR REPLACE INTO market_data 
		(symbol, last_price, change_amount, change_percent, volume, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	result, err := r.db.ExecContext(ctx, query,
		data.Symbol,
		data.LastPrice.Paisa(),
		data.ChangeAmount.Paisa(),
		data.ChangePercent.BasisPoints(),
		data.Volume.Int64(),
		data.Timestamp,
	)

	if err != nil {
		return repository.NewInternalError("failed to upsert market data", err)
	}

	if data.ID == 0 {
		id, err := result.LastInsertId()
		if err != nil {
			return repository.NewInternalError("failed to get insert ID", err)
		}
		data.ID = id
	}

	return nil
}

// GetMarketData retrieves the latest market data for a symbol
func (r *MarketDataRepository) GetMarketData(ctx context.Context, symbol string) (*models.MarketData, error) {
	query := `
		SELECT id, symbol, last_price, change_amount, change_percent, volume, timestamp
		FROM market_data
		WHERE symbol = ? COLLATE NOCASE
		ORDER BY timestamp DESC
		LIMIT 1
	`

	data := &models.MarketData{}
	err := r.db.QueryRowContext(ctx, query, symbol).Scan(
		&data.ID,
		&data.Symbol,
		&data.LastPrice,
		&data.ChangeAmount,
		&data.ChangePercent,
		&data.Volume,
		&data.Timestamp,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.NewNotFoundError("market data", symbol)
		}
		return nil, repository.NewInternalError("failed to get market data", err)
	}

	// Convert integer values back to proper types
	data.LastPrice = models.Money(data.LastPrice)
	data.ChangeAmount = models.Money(data.ChangeAmount)
	data.ChangePercent = models.Percentage(data.ChangePercent)
	data.Volume = models.Quantity(data.Volume)

	return data, nil
}

// GetMarketDataBatch retrieves market data for multiple symbols
func (r *MarketDataRepository) GetMarketDataBatch(ctx context.Context, symbols []string) (map[string]*models.MarketData, error) {
	if len(symbols) == 0 {
		return make(map[string]*models.MarketData), nil
	}

	// Build query with placeholders for symbols
	placeholders := make([]interface{}, len(symbols))
	queryArgs := ""
	for i, symbol := range symbols {
		if i > 0 {
			queryArgs += ","
		}
		queryArgs += "?"
		placeholders[i] = symbol
	}

	query := `
		SELECT id, symbol, last_price, change_amount, change_percent, volume, timestamp
		FROM market_data m1
		WHERE m1.symbol IN (` + queryArgs + `) COLLATE NOCASE
		AND m1.timestamp = (
			SELECT MAX(m2.timestamp) 
			FROM market_data m2 
			WHERE m2.symbol = m1.symbol COLLATE NOCASE
		)
	`

	rows, err := r.db.QueryContext(ctx, query, placeholders...)
	if err != nil {
		return nil, repository.NewInternalError("failed to query market data batch", err)
	}
	defer rows.Close()

	result := make(map[string]*models.MarketData)
	for rows.Next() {
		data := &models.MarketData{}
		err := rows.Scan(
			&data.ID,
			&data.Symbol,
			&data.LastPrice,
			&data.ChangeAmount,
			&data.ChangePercent,
			&data.Volume,
			&data.Timestamp,
		)

		if err != nil {
			return nil, repository.NewInternalError("failed to scan market data row", err)
		}

		// Convert integer values back to proper types
		data.LastPrice = models.Money(data.LastPrice)
		data.ChangeAmount = models.Money(data.ChangeAmount)
		data.ChangePercent = models.Percentage(data.ChangePercent)
		data.Volume = models.Quantity(data.Volume)

		result[data.Symbol] = data
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewInternalError("error iterating market data rows", err)
	}

	return result, nil
}

// GetAllMarketData retrieves all latest market data
func (r *MarketDataRepository) GetAllMarketData(ctx context.Context) ([]models.MarketData, error) {
	query := `
		SELECT id, symbol, last_price, change_amount, change_percent, volume, timestamp
		FROM market_data m1
		WHERE m1.timestamp = (
			SELECT MAX(m2.timestamp) 
			FROM market_data m2 
			WHERE m2.symbol = m1.symbol COLLATE NOCASE
		)
		ORDER BY symbol
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, repository.NewInternalError("failed to query all market data", err)
	}
	defer rows.Close()

	var marketData []models.MarketData
	for rows.Next() {
		data := models.MarketData{}
		err := rows.Scan(
			&data.ID,
			&data.Symbol,
			&data.LastPrice,
			&data.ChangeAmount,
			&data.ChangePercent,
			&data.Volume,
			&data.Timestamp,
		)

		if err != nil {
			return nil, repository.NewInternalError("failed to scan market data row", err)
		}

		// Convert integer values back to proper types
		data.LastPrice = models.Money(data.LastPrice)
		data.ChangeAmount = models.Money(data.ChangeAmount)
		data.ChangePercent = models.Percentage(data.ChangePercent)
		data.Volume = models.Quantity(data.Volume)

		marketData = append(marketData, data)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewInternalError("error iterating market data rows", err)
	}

	return marketData, nil
}

// DeleteMarketData deletes market data for a symbol
func (r *MarketDataRepository) DeleteMarketData(ctx context.Context, symbol string) error {
	query := "DELETE FROM market_data WHERE symbol = ? COLLATE NOCASE"

	result, err := r.db.ExecContext(ctx, query, symbol)
	if err != nil {
		return repository.NewInternalError("failed to delete market data", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError("market data", symbol)
	}

	return nil
}

// GetHistoricalPrices retrieves historical price data for a symbol within a date range
func (r *MarketDataRepository) GetHistoricalPrices(ctx context.Context, symbol string, from, to time.Time) ([]models.MarketData, error) {
	query := `
		SELECT id, symbol, last_price, change_amount, change_percent, volume, timestamp
		FROM market_data
		WHERE symbol = ? COLLATE NOCASE
		AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, symbol, from, to)
	if err != nil {
		return nil, repository.NewInternalError("failed to query historical prices", err)
	}
	defer rows.Close()

	var prices []models.MarketData
	for rows.Next() {
		data := models.MarketData{}
		err := rows.Scan(
			&data.ID,
			&data.Symbol,
			&data.LastPrice,
			&data.ChangeAmount,
			&data.ChangePercent,
			&data.Volume,
			&data.Timestamp,
		)

		if err != nil {
			return nil, repository.NewInternalError("failed to scan historical price row", err)
		}

		// Convert integer values back to proper types
		data.LastPrice = models.Money(data.LastPrice)
		data.ChangeAmount = models.Money(data.ChangeAmount)
		data.ChangePercent = models.Percentage(data.ChangePercent)
		data.Volume = models.Quantity(data.Volume)

		prices = append(prices, data)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewInternalError("error iterating historical price rows", err)
	}

	return prices, nil
}

// GetLatestPrices retrieves the latest prices for multiple symbols
func (r *MarketDataRepository) GetLatestPrices(ctx context.Context, symbols []string) (map[string]models.Money, error) {
	if len(symbols) == 0 {
		return make(map[string]models.Money), nil
	}

	// Build query with placeholders for symbols
	placeholders := make([]interface{}, len(symbols))
	queryArgs := ""
	for i, symbol := range symbols {
		if i > 0 {
			queryArgs += ","
		}
		queryArgs += "?"
		placeholders[i] = symbol
	}

	query := `
		SELECT symbol, last_price
		FROM market_data m1
		WHERE m1.symbol IN (` + queryArgs + `) COLLATE NOCASE
		AND m1.timestamp = (
			SELECT MAX(m2.timestamp) 
			FROM market_data m2 
			WHERE m2.symbol = m1.symbol COLLATE NOCASE
		)
	`

	rows, err := r.db.QueryContext(ctx, query, placeholders...)
	if err != nil {
		return nil, repository.NewInternalError("failed to query latest prices", err)
	}
	defer rows.Close()

	result := make(map[string]models.Money)
	for rows.Next() {
		var symbol string
		var lastPrice int64

		err := rows.Scan(&symbol, &lastPrice)
		if err != nil {
			return nil, repository.NewInternalError("failed to scan latest price row", err)
		}

		result[symbol] = models.Money(lastPrice)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewInternalError("error iterating latest price rows", err)
	}

	return result, nil
}

// CleanupStaleData removes market data older than the specified time
func (r *MarketDataRepository) CleanupStaleData(ctx context.Context, olderThan time.Time) error {
	// Keep the latest record for each symbol, even if it's older than the threshold
	query := `
		DELETE FROM market_data 
		WHERE timestamp < ?
		AND id NOT IN (
			SELECT id FROM (
				SELECT id, ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) as rn
				FROM market_data
			) ranked
			WHERE rn = 1
		)
	`

	_, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return repository.NewInternalError("failed to cleanup stale market data", err)
	}

	return nil
}

// GetDataAge returns the age of the latest market data for a symbol
func (r *MarketDataRepository) GetDataAge(ctx context.Context, symbol string) (time.Duration, error) {
	query := `
		SELECT timestamp
		FROM market_data
		WHERE symbol = ? COLLATE NOCASE
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var timestamp time.Time
	err := r.db.QueryRowContext(ctx, query, symbol).Scan(&timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, repository.NewNotFoundError("market data", symbol)
		}
		return 0, repository.NewInternalError("failed to get data age", err)
	}

	return time.Since(timestamp), nil
}