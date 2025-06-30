package models

import (
	"time"
)

// Portfolio represents a collection of holdings
type Portfolio struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PortfolioStats contains aggregated portfolio statistics
type PortfolioStats struct {
	Portfolio        Portfolio `json:"portfolio"`
	HoldingCount     int64     `json:"holding_count"`
	TotalCost        Money     `json:"total_cost"`
	TotalValue       Money     `json:"total_value"`
	UnrealizedPnL    Money     `json:"unrealized_pnl"`
	UnrealizedPnLPct float64   `json:"unrealized_pnl_pct"`
	TotalInvested    Money     `json:"total_invested"`
	TotalRealized    Money     `json:"total_realized"`
	RealizedPnL      Money     `json:"realized_pnl"`
	TotalFees        Money     `json:"total_fees"`
}

// CalculateUnrealizedMetrics calculates unrealized P/L and percentage
func (ps *PortfolioStats) CalculateUnrealizedMetrics() {
	ps.UnrealizedPnL = ps.TotalValue.Sub(ps.TotalCost)
	ps.UnrealizedPnLPct = ps.TotalValue.PercentageChange(ps.TotalCost)
}

// CalculateRealizedMetrics calculates realized P/L from trading
func (ps *PortfolioStats) CalculateRealizedMetrics() {
	ps.RealizedPnL = ps.TotalRealized.Sub(ps.TotalInvested).Sub(ps.TotalFees)
}

// TotalPnL returns combined realized + unrealized P/L
func (ps *PortfolioStats) TotalPnL() Money {
	return ps.RealizedPnL.Add(ps.UnrealizedPnL)
}

// TotalPnLPct returns combined P/L percentage
func (ps *PortfolioStats) TotalPnLPct() float64 {
	if ps.TotalInvested.IsZero() {
		return 0
	}
	totalPnL := ps.TotalPnL()
	return totalPnL.PercentageChange(ps.TotalInvested)
}

// IsProfit returns true if portfolio is in profit
func (ps *PortfolioStats) IsProfit() bool {
	return ps.TotalPnL().IsPositive()
}

// Holding represents a stock position in a portfolio
type Holding struct {
	ID               int64     `json:"id"`
	PortfolioID      int64     `json:"portfolio_id"`
	Symbol           string    `json:"symbol"`
	Quantity         int64     `json:"quantity"`
	AverageCost      Money     `json:"average_cost"`
	LastPrice        *Money    `json:"last_price,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	
	// Calculated fields
	TotalCost        Money     `json:"total_cost"`
	TotalValue       Money     `json:"total_value"`
	UnrealizedPnL    Money     `json:"unrealized_pnl"`
	UnrealizedPnLPct float64   `json:"unrealized_pnl_pct"`
}

// CalculateMetrics calculates all holding metrics
func (h *Holding) CalculateMetrics() {
	h.TotalCost = h.AverageCost.MultiplyInt(h.Quantity)
	
	currentPrice := h.AverageCost
	if h.LastPrice != nil {
		currentPrice = *h.LastPrice
	}
	
	h.TotalValue = currentPrice.MultiplyInt(h.Quantity)
	h.UnrealizedPnL = h.TotalValue.Sub(h.TotalCost)
	h.UnrealizedPnLPct = h.TotalValue.PercentageChange(h.TotalCost)
}

// IsProfit returns true if holding is profitable
func (h *Holding) IsProfit() bool {
	return h.UnrealizedPnL.IsPositive()
}

// CurrentPrice returns the current price (last price or average cost)
func (h *Holding) CurrentPrice() Money {
	if h.LastPrice != nil {
		return *h.LastPrice
	}
	return h.AverageCost
}

// Transaction represents a buy/sell transaction
type Transaction struct {
	ID              int64     `json:"id"`
	PortfolioID     int64     `json:"portfolio_id"`
	Symbol          string    `json:"symbol"`
	TransactionType string    `json:"transaction_type"` // "buy" or "sell"
	Quantity        int64     `json:"quantity"`
	Price           Money     `json:"price"`
	Commission      Money     `json:"commission"`
	Tax             Money     `json:"tax"`
	TransactionDate time.Time `json:"transaction_date"`
	Notes           *string   `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// TotalAmount returns total transaction amount including fees
func (t *Transaction) TotalAmount() Money {
	base := t.Price.MultiplyInt(t.Quantity)
	return base.Add(t.Commission).Add(t.Tax)
}

// NetAmount returns net amount (excluding fees)
func (t *Transaction) NetAmount() Money {
	return t.Price.MultiplyInt(t.Quantity)
}

// IsBuy returns true if transaction is a buy
func (t *Transaction) IsBuy() bool {
	return t.TransactionType == "buy"
}

// IsSell returns true if transaction is a sell
func (t *Transaction) IsSell() bool {
	return t.TransactionType == "sell"
}

// CorporateAction represents corporate actions like dividends, bonus shares
type CorporateAction struct {
	ID               int64      `json:"id"`
	Symbol           string     `json:"symbol"`
	ActionType       string     `json:"action_type"` // "bonus", "dividend", "split", "rights"
	AnnouncementDate time.Time  `json:"announcement_date"`
	RecordDate       time.Time  `json:"record_date"`
	ExecutionDate    *time.Time `json:"execution_date,omitempty"`
	RatioFrom        *int64     `json:"ratio_from,omitempty"` // For ratios like 1:5
	RatioTo          *int64     `json:"ratio_to,omitempty"`
	Amount           *Money     `json:"amount,omitempty"` // For dividends
	Notes            *string    `json:"notes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// IsBonus returns true if action is bonus shares
func (ca *CorporateAction) IsBonus() bool {
	return ca.ActionType == "bonus"
}

// IsDividend returns true if action is dividend
func (ca *CorporateAction) IsDividend() bool {
	return ca.ActionType == "dividend"
}

// IsSplit returns true if action is stock split
func (ca *CorporateAction) IsSplit() bool {
	return ca.ActionType == "split"
}

// IsRights returns true if action is rights issue
func (ca *CorporateAction) IsRights() bool {
	return ca.ActionType == "rights"
}

// IsPending returns true if action hasn't been executed
func (ca *CorporateAction) IsPending() bool {
	return ca.ExecutionDate == nil || ca.ExecutionDate.After(time.Now())
}

// GetRatio returns the ratio as a decimal (e.g., 1:5 = 0.2)
func (ca *CorporateAction) GetRatio() float64 {
	if ca.RatioFrom == nil || ca.RatioTo == nil || *ca.RatioTo == 0 {
		return 0
	}
	return float64(*ca.RatioFrom) / float64(*ca.RatioTo)
}