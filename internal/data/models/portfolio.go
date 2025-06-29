package models

import (
	"time"
)

// Portfolio represents a collection of holdings
type Portfolio struct {
	Holdings      []Holding     `json:"holdings"`
	TotalValue    Money         `json:"total_value"`
	TotalCost     Money         `json:"total_cost"`
	DayChange     Money         `json:"day_change"`
	DayChangePerc Percentage    `json:"day_change_perc"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// Holding represents a single stock position in the portfolio
type Holding struct {
	ID           int64     `json:"id" db:"id"`
	Symbol       string    `json:"symbol" db:"symbol"`
	Quantity     Quantity  `json:"quantity" db:"quantity"`
	AvgCost      Money     `json:"avg_cost" db:"avg_cost"`
	PurchaseDate time.Time `json:"purchase_date" db:"purchase_date"`
	Notes        string    `json:"notes" db:"notes"` // May be encrypted
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	
	// Calculated fields (not stored in database)
	CurrentPrice  Money      `json:"current_price,omitempty"`
	MarketValue   Money      `json:"market_value,omitempty"`
	TotalCost     Money      `json:"total_cost,omitempty"`
	UnrealizedPnL Money      `json:"unrealized_pnl,omitempty"`
	GainLossPerc  Percentage `json:"gain_loss_perc,omitempty"`
}

// MarketData represents current and historical market price data
type MarketData struct {
	ID            int64      `json:"id" db:"id"`
	Symbol        string     `json:"symbol" db:"symbol"`
	LastPrice     Money      `json:"last_price" db:"last_price"`
	ChangeAmount  Money      `json:"change_amount" db:"change_amount"`
	ChangePercent Percentage `json:"change_percent" db:"change_percent"`
	Volume        Quantity   `json:"volume" db:"volume"`
	Timestamp     time.Time  `json:"timestamp" db:"timestamp"`
}

// PortfolioMetrics represents calculated portfolio performance metrics
type PortfolioMetrics struct {
	TotalValue      Money      `json:"total_value"`
	TotalCost       Money      `json:"total_cost"`
	TotalGainLoss   Money      `json:"total_gain_loss"`
	TotalGainLossPerc Percentage `json:"total_gain_loss_perc"`
	DayChange       Money      `json:"day_change"`
	DayChangePerc   Percentage `json:"day_change_perc"`
	PortfolioCount  int        `json:"portfolio_count"`
	LastUpdated     time.Time  `json:"last_updated"`
}

// HoldingSummary represents summarized holding information for display
type HoldingSummary struct {
	Symbol        string     `json:"symbol"`
	Quantity      Quantity   `json:"quantity"`
	AvgCost       Money      `json:"avg_cost"`
	CurrentPrice  Money      `json:"current_price"`
	MarketValue   Money      `json:"market_value"`
	TotalCost     Money      `json:"total_cost"`
	UnrealizedPnL Money      `json:"unrealized_pnl"`
	GainLossPerc  Percentage `json:"gain_loss_perc"`
	AllocationPerc Percentage `json:"allocation_perc"`
	DayChange     Money      `json:"day_change"`
	DayChangePerc Percentage `json:"day_change_perc"`
}

// PortfolioRequest represents a request to add/update portfolio holdings
type PortfolioRequest struct {
	Symbol       string    `json:"symbol" validate:"required,min=2,max=10"`
	Quantity     Quantity  `json:"quantity" validate:"required,min=1"`
	AvgCost      Money     `json:"avg_cost" validate:"required,min=1"`
	PurchaseDate time.Time `json:"purchase_date"`
	Notes        string    `json:"notes,omitempty" validate:"max=500"`
}

// Methods for Holding

// CalculateMarketValue calculates the current market value
func (h *Holding) CalculateMarketValue(currentPrice Money) {
	h.CurrentPrice = currentPrice
	h.MarketValue = currentPrice.MultiplyByQuantity(h.Quantity)
}

// CalculateTotalCost calculates the total cost basis
func (h *Holding) CalculateTotalCost() {
	h.TotalCost = h.AvgCost.MultiplyByQuantity(h.Quantity)
}

// CalculateUnrealizedPnL calculates unrealized profit/loss
func (h *Holding) CalculateUnrealizedPnL() {
	if h.MarketValue.IsZero() || h.TotalCost.IsZero() {
		return
	}
	h.UnrealizedPnL = h.MarketValue.Subtract(h.TotalCost)
	h.GainLossPerc = CalculatePercentageChange(h.TotalCost, h.MarketValue)
}

// UpdateCalculations updates all calculated fields for the holding
func (h *Holding) UpdateCalculations(currentPrice Money) {
	h.CalculateMarketValue(currentPrice)
	h.CalculateTotalCost()
	h.CalculateUnrealizedPnL()
}

// IsValid validates the holding data
func (h *Holding) IsValid() bool {
	return h.Symbol != "" && 
		   h.Quantity.IsPositive() && 
		   h.AvgCost.IsPositive()
}

// Methods for Portfolio

// AddHolding adds a new holding to the portfolio
func (p *Portfolio) AddHolding(holding Holding) {
	p.Holdings = append(p.Holdings, holding)
	p.UpdatedAt = time.Now()
}

// RemoveHolding removes a holding by symbol
func (p *Portfolio) RemoveHolding(symbol string) bool {
	for i, holding := range p.Holdings {
		if holding.Symbol == symbol {
			p.Holdings = append(p.Holdings[:i], p.Holdings[i+1:]...)
			p.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// FindHolding finds a holding by symbol
func (p *Portfolio) FindHolding(symbol string) (*Holding, bool) {
	for i, holding := range p.Holdings {
		if holding.Symbol == symbol {
			return &p.Holdings[i], true
		}
	}
	return nil, false
}

// UpdateHolding updates an existing holding
func (p *Portfolio) UpdateHolding(symbol string, updatedHolding Holding) bool {
	for i, holding := range p.Holdings {
		if holding.Symbol == symbol {
			updatedHolding.ID = holding.ID
			updatedHolding.CreatedAt = holding.CreatedAt
			updatedHolding.UpdatedAt = time.Now()
			p.Holdings[i] = updatedHolding
			p.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// CalculateMetrics calculates portfolio-level metrics
func (p *Portfolio) CalculateMetrics() PortfolioMetrics {
	var totalValue, totalCost, dayChange Money
	
	for _, holding := range p.Holdings {
		totalValue = totalValue.Add(holding.MarketValue)
		totalCost = totalCost.Add(holding.TotalCost)
		// Day change would be calculated based on previous day's prices
		// For now, we'll leave it as zero until we implement price history
	}
	
	totalGainLoss := totalValue.Subtract(totalCost)
	var totalGainLossPerc Percentage
	if !totalCost.IsZero() {
		totalGainLossPerc = CalculatePercentageChange(totalCost, totalValue)
	}
	
	var dayChangePerc Percentage
	if !totalValue.IsZero() && !dayChange.IsZero() {
		dayChangePerc = CalculatePercentageChange(totalValue.Subtract(dayChange), totalValue)
	}
	
	return PortfolioMetrics{
		TotalValue:        totalValue,
		TotalCost:         totalCost,
		TotalGainLoss:     totalGainLoss,
		TotalGainLossPerc: totalGainLossPerc,
		DayChange:         dayChange,
		DayChangePerc:     dayChangePerc,
		PortfolioCount:    len(p.Holdings),
		LastUpdated:       time.Now(),
	}
}

// GetHoldingSummaries returns summarized holding information for display
func (p *Portfolio) GetHoldingSummaries() []HoldingSummary {
	metrics := p.CalculateMetrics()
	summaries := make([]HoldingSummary, len(p.Holdings))
	
	for i, holding := range p.Holdings {
		var allocationPerc Percentage
		if !metrics.TotalValue.IsZero() {
			allocationFloat := (holding.MarketValue.Rupees() / metrics.TotalValue.Rupees()) * 100
			allocationPerc = NewPercentageFromFloat(allocationFloat)
		}
		
		summaries[i] = HoldingSummary{
			Symbol:         holding.Symbol,
			Quantity:       holding.Quantity,
			AvgCost:        holding.AvgCost,
			CurrentPrice:   holding.CurrentPrice,
			MarketValue:    holding.MarketValue,
			TotalCost:      holding.TotalCost,
			UnrealizedPnL:  holding.UnrealizedPnL,
			GainLossPerc:   holding.GainLossPerc,
			AllocationPerc: allocationPerc,
			DayChange:      Money(0), // TODO: Implement with price history
			DayChangePerc:  Percentage(0), // TODO: Implement with price history
		}
	}
	
	return summaries
}

// IsEmpty checks if the portfolio has no holdings
func (p *Portfolio) IsEmpty() bool {
	return len(p.Holdings) == 0
}

// GetSymbols returns a list of all symbols in the portfolio
func (p *Portfolio) GetSymbols() []string {
	symbols := make([]string, len(p.Holdings))
	for i, holding := range p.Holdings {
		symbols[i] = holding.Symbol
	}
	return symbols
}

// Methods for PortfolioRequest

// ToHolding converts a PortfolioRequest to a Holding
func (pr *PortfolioRequest) ToHolding() Holding {
	now := time.Now()
	purchaseDate := pr.PurchaseDate
	if purchaseDate.IsZero() {
		purchaseDate = now
	}
	
	return Holding{
		Symbol:       pr.Symbol,
		Quantity:     pr.Quantity,
		AvgCost:      pr.AvgCost,
		PurchaseDate: purchaseDate,
		Notes:        pr.Notes,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// IsValid validates the portfolio request
func (pr *PortfolioRequest) IsValid() bool {
	return pr.Symbol != "" && 
		   len(pr.Symbol) >= 2 && len(pr.Symbol) <= 10 &&
		   pr.Quantity.IsPositive() && 
		   pr.AvgCost.IsPositive() &&
		   len(pr.Notes) <= 500
}

// Methods for MarketData

// IsRecent checks if market data is recent (within last hour)
func (md *MarketData) IsRecent() bool {
	return time.Since(md.Timestamp) < time.Hour
}

// IsStale checks if market data is stale (older than 6 hours)
func (md *MarketData) IsStale() bool {
	return time.Since(md.Timestamp) > 6*time.Hour
}

// Age returns the age of the market data
func (md *MarketData) Age() time.Duration {
	return time.Since(md.Timestamp)
}