package services

import (
	"context"
	"database/sql"
	"fmt"
	"ntx/internal/data/repository"
	"ntx/internal/database"
	"ntx/internal/portfolio/models"
	"time"
)

// CalculatorService handles portfolio calculations and metrics
type CalculatorService struct {
	repo *repository.Repository
}

// NewCalculatorService creates a new calculator service
func NewCalculatorService(repo *repository.Repository) *CalculatorService {
	return &CalculatorService{
		repo: repo,
	}
}

// CalculatePortfolioStats computes comprehensive portfolio statistics
func (c *CalculatorService) CalculatePortfolioStats(ctx context.Context, portfolioID int64) (*models.PortfolioStats, error) {
	// Get portfolio details
	portfolioData, err := c.repo.Portfolio.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	portfolio := c.convertPortfolio(portfolioData)

	// Get portfolio basic stats from database
	dbStats, err := c.repo.Portfolio.GetStats(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio stats: %w", err)
	}

	// Get transaction stats
	transactionStats, err := c.repo.Transaction.GetPortfolioStats(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction stats: %w", err)
	}

	// Create portfolio stats
	stats := &models.PortfolioStats{
		Portfolio:     *portfolio,
		HoldingCount:  dbStats.HoldingCount,
		TotalCost:     models.NewMoneyFromPaisa(dbStats.TotalCostPaisa.(int64)),
		TotalValue:    models.NewMoneyFromPaisa(dbStats.TotalValuePaisa.(int64)),
		TotalInvested: models.NewMoneyFromPaisa(int64(transactionStats.TotalInvestedPaisa.Float64)),
		TotalRealized: models.NewMoneyFromPaisa(int64(transactionStats.TotalRealizedPaisa.Float64)),
		TotalFees:     models.NewMoneyFromPaisa(int64(transactionStats.TotalFeesPaisa.Float64)),
	}

	// Calculate derived metrics
	stats.CalculateUnrealizedMetrics()
	stats.CalculateRealizedMetrics()

	return stats, nil
}

// CalculateHoldingMetrics computes metrics for a single holding
func (c *CalculatorService) CalculateHoldingMetrics(ctx context.Context, holdingID int64) (*models.Holding, error) {
	// Get holding with calculated values
	holdingValue, err := c.repo.Holding.GetValue(ctx, holdingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get holding value: %w", err)
	}

	holding := c.convertHoldingWithValue(holdingValue)
	holding.CalculateMetrics()

	return holding, nil
}

// CalculateHoldingsForPortfolio computes metrics for all holdings in a portfolio
func (c *CalculatorService) CalculateHoldingsForPortfolio(ctx context.Context, portfolioID int64) ([]models.Holding, error) {
	// Get holdings with values
	holdingsData, err := c.repo.Holding.ListWithValue(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings with values: %w", err)
	}

	holdings := make([]models.Holding, len(holdingsData))
	for i, data := range holdingsData {
		holding := c.convertHoldingWithValueRow(&data)
		holding.CalculateMetrics()
		holdings[i] = *holding
	}

	return holdings, nil
}

// CalculateTransactionImpact calculates how a new transaction affects holdings
func (c *CalculatorService) CalculateTransactionImpact(transaction *models.Transaction, currentHolding *models.Holding) *TransactionImpact {
	// Calculate impact
	impact := &TransactionImpact{
		Symbol:            transaction.Symbol,
		Transaction:       *transaction,
		CurrentHolding:    currentHolding,
		TransactionAmount: transaction.TotalAmount(),
	}

	if currentHolding != nil {
		if transaction.IsBuy() {
			impact.CalculateBuyImpact()
		} else {
			impact.CalculateSellImpact()
		}
	} else {
		// New holding from buy transaction
		if transaction.IsBuy() {
			impact.NewHolding = &models.Holding{
				PortfolioID: transaction.PortfolioID,
				Symbol:      transaction.Symbol,
				Quantity:    transaction.Quantity,
				AverageCost: transaction.Price,
			}
			impact.NewHolding.CalculateMetrics()
		}
	}

	return impact
}

// CalculateAverageCost calculates new average cost after a transaction
func (c *CalculatorService) CalculateAverageCost(currentQuantity int64, currentAvgCost models.Money, newQuantity int64, newPrice models.Money) models.Money {
	if currentQuantity == 0 {
		return newPrice
	}

	totalCost := currentAvgCost.MultiplyInt(currentQuantity).Add(newPrice.MultiplyInt(newQuantity))
	totalQuantity := currentQuantity + newQuantity

	if totalQuantity == 0 {
		return models.Zero()
	}

	return totalCost.DivideInt(totalQuantity)
}

// TransactionImpact represents the impact of a transaction on holdings
type TransactionImpact struct {
	Symbol            string           `json:"symbol"`
	Transaction       models.Transaction `json:"transaction"`
	CurrentHolding    *models.Holding  `json:"current_holding,omitempty"`
	NewHolding        *models.Holding  `json:"new_holding,omitempty"`
	TransactionAmount models.Money     `json:"transaction_amount"`
	RealizedPnL       models.Money     `json:"realized_pnl,omitempty"`
}

// CalculateBuyImpact calculates impact of a buy transaction
func (ti *TransactionImpact) CalculateBuyImpact() {
	if ti.CurrentHolding == nil {
		return
	}

	// Calculate new average cost
	currentCost := ti.CurrentHolding.AverageCost.MultiplyInt(ti.CurrentHolding.Quantity)
	newCost := ti.Transaction.Price.MultiplyInt(ti.Transaction.Quantity)
	totalCost := currentCost.Add(newCost)
	totalQuantity := ti.CurrentHolding.Quantity + ti.Transaction.Quantity

	newAvgCost := totalCost.DivideInt(totalQuantity)

	ti.NewHolding = &models.Holding{
		PortfolioID: ti.CurrentHolding.PortfolioID,
		Symbol:      ti.Symbol,
		Quantity:    totalQuantity,
		AverageCost: newAvgCost,
		LastPrice:   ti.CurrentHolding.LastPrice,
	}
	ti.NewHolding.CalculateMetrics()
}

// CalculateSellImpact calculates impact of a sell transaction
func (ti *TransactionImpact) CalculateSellImpact() {
	if ti.CurrentHolding == nil {
		return
	}

	// Calculate realized P/L
	costBasis := ti.CurrentHolding.AverageCost.MultiplyInt(ti.Transaction.Quantity)
	saleProceeds := ti.Transaction.NetAmount()
	ti.RealizedPnL = saleProceeds.Sub(costBasis)

	// Calculate remaining holding
	newQuantity := ti.CurrentHolding.Quantity - ti.Transaction.Quantity
	if newQuantity > 0 {
		ti.NewHolding = &models.Holding{
			PortfolioID: ti.CurrentHolding.PortfolioID,
			Symbol:      ti.Symbol,
			Quantity:    newQuantity,
			AverageCost: ti.CurrentHolding.AverageCost, // Average cost remains same
			LastPrice:   ti.CurrentHolding.LastPrice,
		}
		ti.NewHolding.CalculateMetrics()
	}
}

// Conversion functions from database types to domain models
func (c *CalculatorService) convertPortfolio(p *database.Portfolios) *models.Portfolio {
	currency := "NPR"
	if p.Currency.Valid {
		currency = p.Currency.String
	}
	
	createdAt := time.Now()
	if p.CreatedAt.Valid {
		createdAt = p.CreatedAt.Time
	}
	
	updatedAt := time.Now()
	if p.UpdatedAt.Valid {
		updatedAt = p.UpdatedAt.Time
	}

	return &models.Portfolio{
		ID:          p.ID,
		Name:        p.Name,
		Description: nullStringToPtr(p.Description),
		Currency:    currency,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func (c *CalculatorService) convertHolding(h *database.Holdings) *models.Holding {
	var lastPrice *models.Money
	if h.LastPricePaisa.Valid {
		price := models.NewMoneyFromPaisa(h.LastPricePaisa.Int64)
		lastPrice = &price
	}

	createdAt := time.Now()
	if h.CreatedAt.Valid {
		createdAt = h.CreatedAt.Time
	}
	
	updatedAt := time.Now()
	if h.UpdatedAt.Valid {
		updatedAt = h.UpdatedAt.Time
	}

	return &models.Holding{
		ID:          h.ID,
		PortfolioID: h.PortfolioID,
		Symbol:      h.Symbol,
		Quantity:    h.Quantity,
		AverageCost: models.NewMoneyFromPaisa(h.AverageCostPaisa),
		LastPrice:   lastPrice,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func (c *CalculatorService) convertHoldingWithValue(h *database.GetHoldingValueRow) *models.Holding {
	var lastPrice *models.Money
	if h.LastPricePaisa.Valid {
		price := models.NewMoneyFromPaisa(h.LastPricePaisa.Int64)
		lastPrice = &price
	}

	createdAt := time.Now()
	if h.CreatedAt.Valid {
		createdAt = h.CreatedAt.Time
	}
	
	updatedAt := time.Now()
	if h.UpdatedAt.Valid {
		updatedAt = h.UpdatedAt.Time
	}

	holding := &models.Holding{
		ID:               h.ID,
		PortfolioID:      h.PortfolioID,
		Symbol:           h.Symbol,
		Quantity:         h.Quantity,
		AverageCost:      models.NewMoneyFromPaisa(h.AverageCostPaisa),
		LastPrice:        lastPrice,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		TotalCost:        models.NewMoneyFromPaisa(h.TotalCostPaisa.(int64)),
		TotalValue:       models.NewMoneyFromPaisa(h.TotalValuePaisa.(int64)),
		UnrealizedPnL:    models.NewMoneyFromPaisa(h.UnrealizedPnlPaisa),
	}

	// Calculate percentage
	holding.UnrealizedPnLPct = holding.TotalValue.PercentageChange(holding.TotalCost)
	return holding
}

func (c *CalculatorService) convertHoldingWithValueRow(h *database.ListHoldingsWithValueRow) *models.Holding {
	var lastPrice *models.Money
	if h.LastPricePaisa.Valid {
		price := models.NewMoneyFromPaisa(h.LastPricePaisa.Int64)
		lastPrice = &price
	}

	createdAt := time.Now()
	if h.CreatedAt.Valid {
		createdAt = h.CreatedAt.Time
	}
	
	updatedAt := time.Now()
	if h.UpdatedAt.Valid {
		updatedAt = h.UpdatedAt.Time
	}

	holding := &models.Holding{
		ID:               h.ID,
		PortfolioID:      h.PortfolioID,
		Symbol:           h.Symbol,
		Quantity:         h.Quantity,
		AverageCost:      models.NewMoneyFromPaisa(h.AverageCostPaisa),
		LastPrice:        lastPrice,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		TotalCost:        models.NewMoneyFromPaisa(h.TotalCostPaisa.(int64)),
		TotalValue:       models.NewMoneyFromPaisa(h.TotalValuePaisa.(int64)),
		UnrealizedPnL:    models.NewMoneyFromPaisa(h.UnrealizedPnlPaisa),
	}

	// Calculate percentage
	holding.UnrealizedPnLPct = holding.TotalValue.PercentageChange(holding.TotalCost)
	return holding
}

// Helper functions
func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}