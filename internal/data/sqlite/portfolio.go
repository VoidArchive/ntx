package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ntx/internal/data/models"
	"ntx/internal/data/repository"
)

// PortfolioRepository implements the repository.PortfolioRepository interface for SQLite
type PortfolioRepository struct {
	db QueryExecutor
}

// QueryExecutor interface allows using both sql.DB and sql.Tx
type QueryExecutor interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// NewPortfolioRepository creates a new portfolio repository
func NewPortfolioRepository(db QueryExecutor) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

// CreateHolding creates a new portfolio holding
func (r *PortfolioRepository) CreateHolding(ctx context.Context, holding *models.Holding) error {
	if holding == nil || !holding.IsValid() {
		return repository.NewInvalidDataError("invalid holding data", nil)
	}

	query := `
		INSERT INTO portfolio (symbol, quantity, avg_cost, purchase_date, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if holding.CreatedAt.IsZero() {
		holding.CreatedAt = now
	}
	if holding.UpdatedAt.IsZero() {
		holding.UpdatedAt = now
	}

	result, err := r.db.ExecContext(ctx, query,
		holding.Symbol,
		holding.Quantity.Int64(),
		holding.AvgCost.Paisa(),
		holding.PurchaseDate,
		holding.Notes,
		holding.CreatedAt,
		holding.UpdatedAt,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			return repository.NewAlreadyExistsError("holding", holding.Symbol)
		}
		return repository.NewInternalError("failed to create holding", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return repository.NewInternalError("failed to get insert ID", err)
	}

	holding.ID = id
	return nil
}

// GetHolding retrieves a holding by ID
func (r *PortfolioRepository) GetHolding(ctx context.Context, id int64) (*models.Holding, error) {
	query := `
		SELECT id, symbol, quantity, avg_cost, purchase_date, notes, created_at, updated_at
		FROM portfolio
		WHERE id = ?
	`

	holding := &models.Holding{}
	var purchaseDate sql.NullTime
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&holding.ID,
		&holding.Symbol,
		&holding.Quantity,
		&holding.AvgCost,
		&purchaseDate,
		&notes,
		&holding.CreatedAt,
		&holding.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.NewNotFoundError("holding", id)
		}
		return nil, repository.NewInternalError("failed to get holding", err)
	}

	if purchaseDate.Valid {
		holding.PurchaseDate = purchaseDate.Time
	}
	if notes.Valid {
		holding.Notes = notes.String
	}

	// Convert integer values back to proper types
	holding.Quantity = models.Quantity(holding.Quantity)
	holding.AvgCost = models.Money(holding.AvgCost)

	return holding, nil
}

// GetHoldingBySymbol retrieves a holding by symbol
func (r *PortfolioRepository) GetHoldingBySymbol(ctx context.Context, symbol string) (*models.Holding, error) {
	query := `
		SELECT id, symbol, quantity, avg_cost, purchase_date, notes, created_at, updated_at
		FROM portfolio
		WHERE symbol = ? COLLATE NOCASE
	`

	holding := &models.Holding{}
	var purchaseDate sql.NullTime
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, symbol).Scan(
		&holding.ID,
		&holding.Symbol,
		&holding.Quantity,
		&holding.AvgCost,
		&purchaseDate,
		&notes,
		&holding.CreatedAt,
		&holding.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.NewNotFoundError("holding", symbol)
		}
		return nil, repository.NewInternalError("failed to get holding by symbol", err)
	}

	if purchaseDate.Valid {
		holding.PurchaseDate = purchaseDate.Time
	}
	if notes.Valid {
		holding.Notes = notes.String
	}

	// Convert integer values back to proper types
	holding.Quantity = models.Quantity(holding.Quantity)
	holding.AvgCost = models.Money(holding.AvgCost)

	return holding, nil
}

// GetAllHoldings retrieves all portfolio holdings
func (r *PortfolioRepository) GetAllHoldings(ctx context.Context) ([]models.Holding, error) {
	query := `
		SELECT id, symbol, quantity, avg_cost, purchase_date, notes, created_at, updated_at
		FROM portfolio
		ORDER BY symbol
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, repository.NewInternalError("failed to query holdings", err)
	}
	defer rows.Close()

	var holdings []models.Holding
	for rows.Next() {
		holding := models.Holding{}
		var purchaseDate sql.NullTime
		var notes sql.NullString

		err := rows.Scan(
			&holding.ID,
			&holding.Symbol,
			&holding.Quantity,
			&holding.AvgCost,
			&purchaseDate,
			&notes,
			&holding.CreatedAt,
			&holding.UpdatedAt,
		)

		if err != nil {
			return nil, repository.NewInternalError("failed to scan holding row", err)
		}

		if purchaseDate.Valid {
			holding.PurchaseDate = purchaseDate.Time
		}
		if notes.Valid {
			holding.Notes = notes.String
		}

		// Convert integer values back to proper types
		holding.Quantity = models.Quantity(holding.Quantity)
		holding.AvgCost = models.Money(holding.AvgCost)

		holdings = append(holdings, holding)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewInternalError("error iterating holding rows", err)
	}

	return holdings, nil
}

// UpdateHolding updates an existing holding
func (r *PortfolioRepository) UpdateHolding(ctx context.Context, holding *models.Holding) error {
	if holding == nil || !holding.IsValid() {
		return repository.NewInvalidDataError("invalid holding data", nil)
	}

	query := `
		UPDATE portfolio 
		SET quantity = ?, avg_cost = ?, purchase_date = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	holding.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		holding.Quantity.Int64(),
		holding.AvgCost.Paisa(),
		holding.PurchaseDate,
		holding.Notes,
		holding.UpdatedAt,
		holding.ID,
	)

	if err != nil {
		return repository.NewInternalError("failed to update holding", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError("holding", holding.ID)
	}

	return nil
}

// DeleteHolding deletes a holding by ID
func (r *PortfolioRepository) DeleteHolding(ctx context.Context, id int64) error {
	query := "DELETE FROM portfolio WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return repository.NewInternalError("failed to delete holding", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError("holding", id)
	}

	return nil
}

// DeleteHoldingBySymbol deletes a holding by symbol
func (r *PortfolioRepository) DeleteHoldingBySymbol(ctx context.Context, symbol string) error {
	query := "DELETE FROM portfolio WHERE symbol = ? COLLATE NOCASE"

	result, err := r.db.ExecContext(ctx, query, symbol)
	if err != nil {
		return repository.NewInternalError("failed to delete holding by symbol", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError("holding", symbol)
	}

	return nil
}

// GetPortfolioValue calculates the total portfolio value
func (r *PortfolioRepository) GetPortfolioValue(ctx context.Context) (models.Money, error) {
	query := `
		SELECT COALESCE(SUM(p.quantity * COALESCE(m.last_price, p.avg_cost)), 0) as total_value
		FROM portfolio p
		LEFT JOIN (
			SELECT DISTINCT symbol, 
				   FIRST_VALUE(last_price) OVER (PARTITION BY symbol ORDER BY timestamp DESC) as last_price
			FROM market_data
		) m ON p.symbol = m.symbol COLLATE NOCASE
	`

	var totalValue int64
	err := r.db.QueryRowContext(ctx, query).Scan(&totalValue)
	if err != nil {
		return models.Money(0), repository.NewInternalError("failed to calculate portfolio value", err)
	}

	return models.Money(totalValue), nil
}

// GetPortfolioMetrics calculates comprehensive portfolio metrics
func (r *PortfolioRepository) GetPortfolioMetrics(ctx context.Context) (*models.PortfolioMetrics, error) {
	// Get holdings with current prices
	holdings, err := r.getHoldingsWithPrices(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate metrics
	var totalValue, totalCost models.Money
	for _, holding := range holdings {
		holding.CalculateTotalCost()
		if !holding.CurrentPrice.IsZero() {
			holding.CalculateMarketValue(holding.CurrentPrice)
		} else {
			holding.CalculateMarketValue(holding.AvgCost) // Fallback to avg cost
		}
		
		totalValue = totalValue.Add(holding.MarketValue)
		totalCost = totalCost.Add(holding.TotalCost)
	}

	totalGainLoss := totalValue.Subtract(totalCost)
	var totalGainLossPerc models.Percentage
	if !totalCost.IsZero() {
		totalGainLossPerc = models.CalculatePercentageChange(totalCost, totalValue)
	}

	return &models.PortfolioMetrics{
		TotalValue:        totalValue,
		TotalCost:         totalCost,
		TotalGainLoss:     totalGainLoss,
		TotalGainLossPerc: totalGainLossPerc,
		DayChange:         models.Money(0), // TODO: Implement with price history
		DayChangePerc:     models.Percentage(0), // TODO: Implement with price history
		PortfolioCount:    len(holdings),
		LastUpdated:       time.Now(),
	}, nil
}

// GetHoldingSummaries returns summarized holding information
func (r *PortfolioRepository) GetHoldingSummaries(ctx context.Context) ([]models.HoldingSummary, error) {
	holdings, err := r.getHoldingsWithPrices(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate total portfolio value for allocation percentages
	metrics, err := r.GetPortfolioMetrics(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]models.HoldingSummary, len(holdings))
	for i, holding := range holdings {
		holding.UpdateCalculations(holding.CurrentPrice)
		
		var allocationPerc models.Percentage
		if !metrics.TotalValue.IsZero() {
			allocationFloat := (holding.MarketValue.Rupees() / metrics.TotalValue.Rupees()) * 100
			allocationPerc = models.NewPercentageFromFloat(allocationFloat)
		}

		summaries[i] = models.HoldingSummary{
			Symbol:         holding.Symbol,
			Quantity:       holding.Quantity,
			AvgCost:        holding.AvgCost,
			CurrentPrice:   holding.CurrentPrice,
			MarketValue:    holding.MarketValue,
			TotalCost:      holding.TotalCost,
			UnrealizedPnL:  holding.UnrealizedPnL,
			GainLossPerc:   holding.GainLossPerc,
			AllocationPerc: allocationPerc,
			DayChange:      models.Money(0), // TODO: Implement
			DayChangePerc:  models.Percentage(0), // TODO: Implement
		}
	}

	return summaries, nil
}

// CreateHoldings creates multiple holdings in a single transaction
func (r *PortfolioRepository) CreateHoldings(ctx context.Context, holdings []models.Holding) error {
	if len(holdings) == 0 {
		return nil
	}

	query := `
		INSERT INTO portfolio (symbol, quantity, avg_cost, purchase_date, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	for i := range holdings {
		holding := &holdings[i]
		if !holding.IsValid() {
			return repository.NewInvalidDataError(fmt.Sprintf("invalid holding data for symbol %s", holding.Symbol), nil)
		}

		if holding.CreatedAt.IsZero() {
			holding.CreatedAt = now
		}
		if holding.UpdatedAt.IsZero() {
			holding.UpdatedAt = now
		}

		result, err := r.db.ExecContext(ctx, query,
			holding.Symbol,
			holding.Quantity.Int64(),
			holding.AvgCost.Paisa(),
			holding.PurchaseDate,
			holding.Notes,
			holding.CreatedAt,
			holding.UpdatedAt,
		)

		if err != nil {
			if isUniqueConstraintError(err) {
				return repository.NewAlreadyExistsError("holding", holding.Symbol)
			}
			return repository.NewInternalError("failed to create holding", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return repository.NewInternalError("failed to get insert ID", err)
		}

		holding.ID = id
	}

	return nil
}

// UpdateHoldings updates multiple holdings
func (r *PortfolioRepository) UpdateHoldings(ctx context.Context, holdings []models.Holding) error {
	if len(holdings) == 0 {
		return nil
	}

	query := `
		UPDATE portfolio 
		SET quantity = ?, avg_cost = ?, purchase_date = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	for _, holding := range holdings {
		if !holding.IsValid() {
			return repository.NewInvalidDataError(fmt.Sprintf("invalid holding data for symbol %s", holding.Symbol), nil)
		}

		_, err := r.db.ExecContext(ctx, query,
			holding.Quantity.Int64(),
			holding.AvgCost.Paisa(),
			holding.PurchaseDate,
			holding.Notes,
			now,
			holding.ID,
		)

		if err != nil {
			return repository.NewInternalError("failed to update holding", err)
		}
	}

	return nil
}

// getHoldingsWithPrices retrieves holdings with current market prices
func (r *PortfolioRepository) getHoldingsWithPrices(ctx context.Context) ([]models.Holding, error) {
	query := `
		SELECT p.id, p.symbol, p.quantity, p.avg_cost, p.purchase_date, p.notes, 
		       p.created_at, p.updated_at, COALESCE(m.last_price, p.avg_cost) as current_price
		FROM portfolio p
		LEFT JOIN (
			SELECT DISTINCT symbol, 
				   FIRST_VALUE(last_price) OVER (PARTITION BY symbol ORDER BY timestamp DESC) as last_price
			FROM market_data
		) m ON p.symbol = m.symbol COLLATE NOCASE
		ORDER BY p.symbol
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, repository.NewInternalError("failed to query holdings with prices", err)
	}
	defer rows.Close()

	var holdings []models.Holding
	for rows.Next() {
		holding := models.Holding{}
		var purchaseDate sql.NullTime
		var notes sql.NullString
		var currentPrice int64

		err := rows.Scan(
			&holding.ID,
			&holding.Symbol,
			&holding.Quantity,
			&holding.AvgCost,
			&purchaseDate,
			&notes,
			&holding.CreatedAt,
			&holding.UpdatedAt,
			&currentPrice,
		)

		if err != nil {
			return nil, repository.NewInternalError("failed to scan holding with price", err)
		}

		if purchaseDate.Valid {
			holding.PurchaseDate = purchaseDate.Time
		}
		if notes.Valid {
			holding.Notes = notes.String
		}

		// Convert integer values back to proper types
		holding.Quantity = models.Quantity(holding.Quantity)
		holding.AvgCost = models.Money(holding.AvgCost)
		holding.CurrentPrice = models.Money(currentPrice)

		holdings = append(holdings, holding)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewInternalError("error iterating holding rows", err)
	}

	return holdings, nil
}

// Helper function to check for unique constraint errors
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return len(errStr) > 0 && (
		fmt.Sprintf("%s", errStr) == "UNIQUE constraint failed" ||
		fmt.Sprintf("%s", errStr) == "constraint failed" ||
		fmt.Sprintf("%s", errStr) == "UNIQUE constraint failed: portfolio.symbol")
}