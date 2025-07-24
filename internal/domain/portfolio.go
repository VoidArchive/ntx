package domain

import (
	"fmt"
	"sort"
	"time"
)

// Portfolio manages multiple stock holdings and provides portfolio-wide operations
type Portfolio struct {
	Holdings      map[string]*FIFOQueue // symbol -> FIFO queue
	RealizedGains []RealizedGain        // All historical sales
	Warnings      []string              // Validation warnings
	LastUpdated   time.Time
}

// PortfolioError represents errors that occur during portfolio operations
type PortfolioError struct {
	Operation string
	Symbol    string
	Message   string
}

func (e *PortfolioError) Error() string {
	if e.Symbol != "" {
		return fmt.Sprintf("%s [%s]: %s", e.Operation, e.Symbol, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Operation, e.Message)
}

// NewPortfolio creates a new empty portfolio
func NewPortfolio() *Portfolio {
	return &Portfolio{
		Holdings:      make(map[string]*FIFOQueue),
		RealizedGains: make([]RealizedGain, 0),
		Warnings:      make([]string, 0),
		LastUpdated:   time.Now(),
	}
}

// AddTransaction processes a single transaction and updates the portfolio
func (p *Portfolio) AddTransaction(transaction Transaction) error {
	symbol := transaction.StockSymbol

	// Ensure FIFO queue exists for this symbol
	if p.Holdings[symbol] == nil {
		p.Holdings[symbol] = NewFIFOQueue(symbol)
	}

	queue := p.Holdings[symbol]

	switch transaction.Type {
	case TransactionBuy, TransactionBonus, TransactionRights:
		return p.processBuyTransaction(queue, transaction)

	case TransactionSell:
		return p.processSellTransaction(queue, transaction)

	case TransactionSplit:
		return p.processSplitTransaction(queue, transaction)

	case TransactionDividend:
		// For now, just track dividends without affecting holdings
		p.addWarning(fmt.Sprintf("Dividend transaction for %s recorded but not processed in holdings", symbol))
		return nil

	case TransactionMerger:
		p.addWarning(fmt.Sprintf("Merger transaction for %s requires manual handling", symbol))
		return nil

	default:
		return &PortfolioError{
			Operation: "AddTransaction",
			Symbol:    symbol,
			Message:   fmt.Sprintf("unsupported transaction type: %s", transaction.Type),
		}
	}
}

// processBuyTransaction handles buy, bonus, and rights transactions
func (p *Portfolio) processBuyTransaction(queue *FIFOQueue, transaction Transaction) error {
	err := queue.Buy(transaction.Quantity, transaction.Price, transaction.Date)
	if err != nil {
		return &PortfolioError{
			Operation: "Buy",
			Symbol:    transaction.StockSymbol,
			Message:   err.Error(),
		}
	}

	p.LastUpdated = time.Now()
	return nil
}

// processSellTransaction handles sell transactions
func (p *Portfolio) processSellTransaction(queue *FIFOQueue, transaction Transaction) error {
	// Check if we have enough shares to sell
	if transaction.Quantity > queue.TotalShares() {
		return &PortfolioError{
			Operation: "Sell",
			Symbol:    transaction.StockSymbol,
			Message:   fmt.Sprintf("cannot sell %d shares, only %d available", transaction.Quantity, queue.TotalShares()),
		}
	}

	result, err := queue.Sell(transaction.Quantity, transaction.Price, transaction.Date)
	if err != nil {
		return &PortfolioError{
			Operation: "Sell",
			Symbol:    transaction.StockSymbol,
			Message:   err.Error(),
		}
	}

	// Add realized gains to portfolio history
	p.RealizedGains = append(p.RealizedGains, result.RealizedGains...)

	// Remove holding if completely sold out
	if queue.IsEmpty() {
		delete(p.Holdings, transaction.StockSymbol)
	}

	p.LastUpdated = time.Now()
	return nil
}

// processSplitTransaction handles stock splits
func (p *Portfolio) processSplitTransaction(queue *FIFOQueue, transaction Transaction) error {
	// For stock splits, the quantity in transaction represents the multiplier
	// e.g., transaction.Quantity = 2 means 2:1 split (double shares, halve price)
	if transaction.Quantity <= 1 {
		return &PortfolioError{
			Operation: "Split",
			Symbol:    transaction.StockSymbol,
			Message:   fmt.Sprintf("invalid split ratio: %d", transaction.Quantity),
		}
	}

	// Multiply all lot quantities and divide prices
	for i := range queue.Lots {
		lot := &queue.Lots[i]
		lot.Quantity *= transaction.Quantity
		lot.Price = lot.Price.Divide(transaction.Quantity)
	}

	p.addWarning(fmt.Sprintf("Applied %d:1 stock split for %s", transaction.Quantity, transaction.StockSymbol))
	p.LastUpdated = time.Now()
	return nil
}

// ProcessTransactions handles multiple transactions with auto-sorting and validation
func (p *Portfolio) ProcessTransactions(transactions []Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	// Sort transactions by date (chronological order)
	sortedTransactions := make([]Transaction, len(transactions))
	copy(sortedTransactions, transactions)

	sort.Slice(sortedTransactions, func(i, j int) bool {
		return sortedTransactions[i].Date.Before(sortedTransactions[j].Date)
	})

	// Check for out-of-order transactions in original list
	p.validateTransactionOrder(transactions, sortedTransactions)

	// Process each transaction in chronological order
	for _, transaction := range sortedTransactions {
		err := p.AddTransaction(transaction)
		if err != nil {
			return err
		}
	}

	return nil
}

// validateTransactionOrder checks if transactions were out of chronological order
func (p *Portfolio) validateTransactionOrder(original, sorted []Transaction) {
	outOfOrder := 0
	for i, orig := range original {
		if !orig.Date.Equal(sorted[i].Date) || orig.StockSymbol != sorted[i].StockSymbol {
			outOfOrder++
		}
	}

	if outOfOrder > 0 {
		p.addWarning(fmt.Sprintf("Found %d transactions out of chronological order (auto-corrected)", outOfOrder))
	}
}

// GetActiveHoldings returns all current stock holdings with current prices
func (p *Portfolio) GetActiveHoldings(currentPrices map[string]Money) []Holding {
	holdings := make([]Holding, 0, len(p.Holdings))

	for symbol, queue := range p.Holdings {
		currentPrice := currentPrices[symbol]
		if currentPrice.IsZero() {
			// Use a default price and warn
			currentPrice = NewMoney(100.0) // Default price
			p.addWarning(fmt.Sprintf("No current price available for %s, using default Rs.100", symbol))
		}

		holding := queue.GetHolding(currentPrice)
		holdings = append(holdings, holding)
	}

	// Sort holdings by symbol for consistent display
	sort.Slice(holdings, func(i, j int) bool {
		return holdings[i].StockSymbol < holdings[j].StockSymbol
	})

	return holdings
}

// GetHoldingBySymbol returns the holding for a specific stock symbol
func (p *Portfolio) GetHoldingBySymbol(symbol string, currentPrice Money) (*Holding, bool) {
	queue, exists := p.Holdings[symbol]
	if !exists {
		return nil, false
	}

	holding := queue.GetHolding(currentPrice)
	return &holding, true
}

// GetPortfolioSummary calculates aggregate portfolio metrics
func (p *Portfolio) GetPortfolioSummary(currentPrices map[string]Money) PortfolioSummary {
	var totalInvested, totalMarketValue, totalRealizedPL Money

	// Calculate totals from current holdings
	for symbol, queue := range p.Holdings {
		totalInvested = totalInvested.Add(queue.TotalCost())

		currentPrice := currentPrices[symbol]
		if currentPrice.IsZero() {
			currentPrice = NewMoney(100.0) // Default price
		}

		marketValue := currentPrice.Multiply(queue.TotalShares())
		totalMarketValue = totalMarketValue.Add(marketValue)
	}

	// Calculate total realized P/L
	for _, gain := range p.RealizedGains {
		totalRealizedPL = totalRealizedPL.Add(gain.GainLoss)
	}

	// Calculate unrealized P/L and percentage
	totalUnrealizedPL := totalMarketValue.Subtract(totalInvested)
	var unrealizedPLPct float64
	if !totalInvested.IsZero() {
		unrealizedPLPct = totalInvested.PercentageChange(totalMarketValue)
	}

	return PortfolioSummary{
		TotalInvested:     totalInvested,
		TotalMarketValue:  totalMarketValue,
		TotalUnrealizedPL: totalUnrealizedPL,
		TotalRealizedPL:   totalRealizedPL,
		UnrealizedPLPct:   unrealizedPLPct,
		HoldingsCount:     len(p.Holdings),
		LastUpdated:       p.LastUpdated,
	}
}

// GetRealizedGains returns all realized gains, optionally filtered by symbol
func (p *Portfolio) GetRealizedGains(symbol string) []RealizedGain {
	if symbol == "" {
		// Return all realized gains
		gains := make([]RealizedGain, len(p.RealizedGains))
		copy(gains, p.RealizedGains)
		return gains
	}

	// Filter by symbol
	filteredGains := make([]RealizedGain, 0)
	for _, gain := range p.RealizedGains {
		if gain.StockSymbol == symbol {
			filteredGains = append(filteredGains, gain)
		}
	}
	return filteredGains
}

// GetRealizedGainsSummary calculates realized gains summary with tax implications
func (p *Portfolio) GetRealizedGainsSummary() map[string]any {
	var totalGains, shortTermGains, longTermGains Money
	shortTermCount, longTermCount := 0, 0

	for _, gain := range p.RealizedGains {
		totalGains = totalGains.Add(gain.GainLoss)

		if gain.IsLongTerm {
			longTermGains = longTermGains.Add(gain.GainLoss)
			longTermCount++
		} else {
			shortTermGains = shortTermGains.Add(gain.GainLoss)
			shortTermCount++
		}
	}

	// Calculate estimated taxes (Nepal rates: 7.5% short-term, 5% long-term)
	estimatedTaxShortTerm := shortTermGains.Percentage(7.5)
	estimatedTaxLongTerm := longTermGains.Percentage(5.0)
	estimatedTotalTax := estimatedTaxShortTerm.Add(estimatedTaxLongTerm)

	return map[string]any{
		"total_gains":              totalGains,
		"short_term_gains":         shortTermGains,
		"long_term_gains":          longTermGains,
		"short_term_count":         shortTermCount,
		"long_term_count":          longTermCount,
		"estimated_tax_short_term": estimatedTaxShortTerm,
		"estimated_tax_long_term":  estimatedTaxLongTerm,
		"estimated_total_tax":      estimatedTotalTax,
		"total_sales":              len(p.RealizedGains),
	}
}

// GetWarnings returns all validation warnings
func (p *Portfolio) GetWarnings() []string {
	warnings := make([]string, len(p.Warnings))
	copy(warnings, p.Warnings)
	return warnings
}

// ClearWarnings removes all warnings
func (p *Portfolio) ClearWarnings() {
	p.Warnings = make([]string, 0)
}

// addWarning adds a warning message to the portfolio
func (p *Portfolio) addWarning(message string) {
	p.Warnings = append(p.Warnings, message)
}

// ValidateIntegrity performs consistency checks on the portfolio
func (p *Portfolio) ValidateIntegrity() []string {
	issues := make([]string, 0)

	// Check for negative holdings (shouldn't happen with proper FIFO)
	for symbol, queue := range p.Holdings {
		if queue.TotalShares() < 0 {
			issues = append(issues, fmt.Sprintf("Negative holdings detected for %s: %d shares", symbol, queue.TotalShares()))
		}

		if queue.TotalCost().IsNegative() {
			issues = append(issues, fmt.Sprintf("Negative cost basis detected for %s: %s", symbol, queue.TotalCost().String()))
		}
	}

	// Check for realized gains without corresponding transactions
	for _, gain := range p.RealizedGains {
		if gain.Quantity <= 0 {
			issues = append(issues, fmt.Sprintf("Invalid realized gain quantity for %s: %d", gain.StockSymbol, gain.Quantity))
		}
	}

	return issues
}

// GetSymbols returns all stock symbols in the portfolio
func (p *Portfolio) GetSymbols() []string {
	symbols := make([]string, 0, len(p.Holdings))
	for symbol := range p.Holdings {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	return symbols
}

// HasHolding checks if portfolio has holdings for a given symbol
func (p *Portfolio) HasHolding(symbol string) bool {
	queue, exists := p.Holdings[symbol]
	return exists && !queue.IsEmpty()
}
