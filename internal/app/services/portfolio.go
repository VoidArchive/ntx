package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"ntx/internal/data/models"
	"ntx/internal/database"
	"time"
)

// PortfolioService handles portfolio management and real-time P&L calculations.
// It integrates with the database using SQLC queries and provides accurate financial metrics.
type PortfolioService struct {
	dbManager *database.Manager
	logger    *slog.Logger
}

// PortfolioData represents complete portfolio information with real-time calculations
type PortfolioData struct {
	TotalValue       models.Money      `json:"total_value"`
	TotalCost        models.Money      `json:"total_cost"`
	TotalGain        models.Money      `json:"total_gain"`
	TotalGainPercent models.Percentage `json:"total_gain_percent"`
	DayChange        models.Money      `json:"day_change"`
	DayChangePercent models.Percentage `json:"day_change_percent"`
	Holdings         []HoldingData     `json:"holdings"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// HoldingData represents a single holding with real-time calculations
type HoldingData struct {
	ID                int64             `json:"id"`
	Symbol            string            `json:"symbol"`
	Quantity          models.Quantity   `json:"quantity"`
	AvgCost           models.Money      `json:"avg_cost"`
	CurrentPrice      models.Money      `json:"current_price"`
	PreviousClose     models.Money      `json:"previous_close"`
	TotalCost         models.Money      `json:"total_cost"`
	CurrentValue      models.Money      `json:"current_value"`
	UnrealizedGain    models.Money      `json:"unrealized_gain"`
	GainPercent       models.Percentage `json:"gain_percent"`
	DayChange         models.Money      `json:"day_change"`
	DayChangePercent  models.Percentage `json:"day_change_percent"`
	AllocationPercent models.Percentage `json:"allocation_percent"`
	PurchaseDate      time.Time         `json:"purchase_date"`
	Notes             string            `json:"notes"`
}

// NewPortfolioService creates a new portfolio service
func NewPortfolioService(dbManager *database.Manager, logger *slog.Logger) *PortfolioService {
	return &PortfolioService{
		dbManager: dbManager,
		logger:    logger,
	}
}

// GetPortfolioData calculates and returns complete portfolio data with real-time P&L
func (ps *PortfolioService) GetPortfolioData(ctx context.Context) (*PortfolioData, error) {
	ps.logger.Debug("Calculating real-time portfolio data")

	// Get portfolio holdings with current prices
	holdingsData, err := ps.dbManager.Queries().GetPortfolioWithCurrentPrices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio holdings: %w", err)
	}

	if len(holdingsData) == 0 {
		ps.logger.Info("No portfolio holdings found")
		return &PortfolioData{
			Holdings:    []HoldingData{},
			LastUpdated: time.Now(),
		}, nil
	}

	// Calculate portfolio metrics
	var totalValue, totalCost, totalGain, dayChange models.Money
	holdings := make([]HoldingData, 0, len(holdingsData))

	for _, dbHolding := range holdingsData {
		holding := ps.calculateHoldingMetrics(ctx, dbHolding)
		holdings = append(holdings, holding)

		totalValue = totalValue.Add(holding.CurrentValue)
		totalCost = totalCost.Add(holding.TotalCost)
		totalGain = totalGain.Add(holding.UnrealizedGain)
		dayChange = dayChange.Add(holding.DayChange)
	}

	// Calculate portfolio-level percentages
	var totalGainPercent, dayChangePercent models.Percentage
	if !totalCost.IsZero() {
		totalGainPercent = models.CalculatePercentageChange(totalCost, totalValue)
	}
	if !totalValue.IsZero() {
		previousValue := totalValue.Subtract(dayChange)
		if !previousValue.IsZero() {
			dayChangePercent = models.CalculatePercentageChange(previousValue, totalValue)
		}
	}

	// Calculate allocation percentages for each holding
	for i := range holdings {
		if !totalValue.IsZero() {
			allocationFloat := (holdings[i].CurrentValue.Rupees() / totalValue.Rupees()) * 100
			holdings[i].AllocationPercent = models.NewPercentageFromFloat(allocationFloat)
		}
	}

	portfolioData := &PortfolioData{
		TotalValue:       totalValue,
		TotalCost:        totalCost,
		TotalGain:        totalGain,
		TotalGainPercent: totalGainPercent,
		DayChange:        dayChange,
		DayChangePercent: dayChangePercent,
		Holdings:         holdings,
		LastUpdated:      time.Now(),
	}

	ps.logger.Info("Portfolio data calculated",
		"total_value", totalValue.FormattedString(),
		"total_gain", totalGain.FormattedString(),
		"day_change", dayChange.FormattedString(),
		"holdings_count", len(holdings))

	return portfolioData, nil
}

// calculateHoldingMetrics calculates all metrics for a single holding
func (ps *PortfolioService) calculateHoldingMetrics(ctx context.Context, dbHolding database.GetPortfolioWithCurrentPricesRow) HoldingData {
	// Convert database types to domain models
	quantity := models.Quantity(dbHolding.Quantity)
	avgCost := models.Money(dbHolding.AvgCost)
	currentPrice := models.Money(dbHolding.CurrentPrice)

	// Get previous close price for day change calculation
	var previousClose models.Money
	if prevPrice, err := ps.dbManager.Queries().GetPreviousClosePrice(ctx, dbHolding.Symbol); err == nil {
		previousClose = models.Money(prevPrice)
	} else {
		// If no previous close available, use current price (no day change)
		previousClose = currentPrice
		ps.logger.Debug("No previous close price found", "symbol", dbHolding.Symbol)
	}

	// Calculate financial metrics
	totalCost := avgCost.MultiplyByQuantity(quantity)
	currentValue := currentPrice.MultiplyByQuantity(quantity)
	unrealizedGain := currentValue.Subtract(totalCost)

	var gainPercent models.Percentage
	if !totalCost.IsZero() {
		gainPercent = models.CalculatePercentageChange(totalCost, currentValue)
	}

	// Calculate day change
	previousValue := previousClose.MultiplyByQuantity(quantity)
	dayChangeAmount := currentValue.Subtract(previousValue)

	var dayChangePercent models.Percentage
	if !previousValue.IsZero() {
		dayChangePercent = models.CalculatePercentageChange(previousValue, currentValue)
	}

	// Handle time fields (now returned as strings)
	var purchaseDate time.Time
	if dbHolding.PurchaseDate != "" {
		if parsed, err := time.Parse("2006-01-02 15:04:05", dbHolding.PurchaseDate); err == nil {
			purchaseDate = parsed
		} else if parsed, err := time.Parse("2006-01-02", dbHolding.PurchaseDate); err == nil {
			purchaseDate = parsed
		}
		// If parsing fails, purchaseDate remains zero value
	}

	var notes string
	if dbHolding.Notes.Valid {
		notes = dbHolding.Notes.String
	}

	return HoldingData{
		ID:               dbHolding.ID,
		Symbol:           dbHolding.Symbol,
		Quantity:         quantity,
		AvgCost:          avgCost,
		CurrentPrice:     currentPrice,
		PreviousClose:    previousClose,
		TotalCost:        totalCost,
		CurrentValue:     currentValue,
		UnrealizedGain:   unrealizedGain,
		GainPercent:      gainPercent,
		DayChange:        dayChangeAmount,
		DayChangePercent: dayChangePercent,
		PurchaseDate:     purchaseDate,
		Notes:            notes,
	}
}

// GetHoldingSymbols returns all symbols in the portfolio for market data watching
func (ps *PortfolioService) GetHoldingSymbols(ctx context.Context) ([]string, error) {
	holdings, err := ps.dbManager.Queries().GetAllHoldings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio symbols: %w", err)
	}

	symbols := make([]string, len(holdings))
	for i, holding := range holdings {
		symbols[i] = holding.Symbol
	}

	return symbols, nil
}

// GetPortfolioSummary returns basic portfolio summary metrics
func (ps *PortfolioService) GetPortfolioSummary(ctx context.Context) (*database.GetPortfolioSummaryRow, error) {
	summary, err := ps.dbManager.Queries().GetPortfolioSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio summary: %w", err)
	}

	return &summary, nil
}

// AddHolding adds a new holding to the portfolio
func (ps *PortfolioService) AddHolding(ctx context.Context, symbol string, quantity models.Quantity, avgCost models.Money, purchaseDate time.Time, notes string) error {
	now := time.Now()

	_, err := ps.dbManager.Queries().CreateHolding(ctx, database.CreateHoldingParams{
		Symbol:       symbol,
		Quantity:     quantity,
		AvgCost:      avgCost,
		PurchaseDate: sql.NullTime{Time: purchaseDate, Valid: true},
		Notes:        sql.NullString{String: notes, Valid: notes != ""},
		CreatedAt:    sql.NullTime{Time: now, Valid: true},
		UpdatedAt:    sql.NullTime{Time: now, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to add holding: %w", err)
	}

	ps.logger.Info("Added portfolio holding",
		"symbol", symbol,
		"quantity", quantity.String(),
		"avg_cost", avgCost.FormattedString())

	return nil
}

// RemoveHolding removes a holding from the portfolio
func (ps *PortfolioService) RemoveHolding(ctx context.Context, symbol string) error {
	err := ps.dbManager.Queries().DeleteHoldingBySymbol(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to remove holding: %w", err)
	}

	ps.logger.Info("Removed portfolio holding", "symbol", symbol)
	return nil
}

// UpdateHolding updates an existing holding
func (ps *PortfolioService) UpdateHolding(ctx context.Context, id int64, quantity models.Quantity, avgCost models.Money, purchaseDate time.Time, notes string) error {
	err := ps.dbManager.Queries().UpdateHolding(ctx, database.UpdateHoldingParams{
		ID:           id,
		Quantity:     quantity,
		AvgCost:      avgCost,
		PurchaseDate: sql.NullTime{Time: purchaseDate, Valid: true},
		Notes:        sql.NullString{String: notes, Valid: notes != ""},
		UpdatedAt:    sql.NullTime{Time: time.Now(), Valid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to update holding: %w", err)
	}

	ps.logger.Info("Updated portfolio holding", "id", id)
	return nil
}
