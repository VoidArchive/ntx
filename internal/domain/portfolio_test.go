package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPortfolio(t *testing.T) {
	portfolio := NewPortfolio()

	assert.NotNil(t, portfolio)
	assert.NotNil(t, portfolio.Holdings)
	assert.NotNil(t, portfolio.RealizedGains)
	assert.NotNil(t, portfolio.Warnings)
	assert.Equal(t, 0, len(portfolio.Holdings))
	assert.Equal(t, 0, len(portfolio.RealizedGains))
	assert.Equal(t, 0, len(portfolio.Warnings))
}

func TestPortfolio_AddSingleBuyTransaction(t *testing.T) {
	portfolio := NewPortfolio()

	transaction := Transaction{
		ID:          1,
		StockSymbol: "ADBL",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionBuy,
		Quantity:    100,
		Price:       NewMoney(500.0),
		Cost:        NewMoney(50000.0),
		Description: "Purchase",
	}

	err := portfolio.AddTransaction(transaction)
	require.NoError(t, err)

	// Verify holding was created
	assert.Equal(t, 1, len(portfolio.Holdings))
	assert.True(t, portfolio.HasHolding("ADBL"))

	queue := portfolio.Holdings["ADBL"]
	assert.Equal(t, 100, queue.TotalShares())
	assert.Equal(t, NewMoney(50000.0), queue.TotalCost())
	assert.Equal(t, NewMoney(500.0), queue.WeightedAverageCost())
}

func TestPortfolio_AddBuyAndSellTransactions(t *testing.T) {
	portfolio := NewPortfolio()

	// Buy transaction
	buyTx := Transaction{
		StockSymbol: "NABIL",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionBuy,
		Quantity:    100,
		Price:       NewMoney(1000.0),
		Cost:        NewMoney(100000.0),
	}

	err := portfolio.AddTransaction(buyTx)
	require.NoError(t, err)

	// Sell transaction
	sellTx := Transaction{
		StockSymbol: "NABIL",
		Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionSell,
		Quantity:    60,
		Price:       NewMoney(1200.0),
		Cost:        NewMoney(72000.0),
	}

	err = portfolio.AddTransaction(sellTx)
	require.NoError(t, err)

	// Verify remaining holdings
	queue := portfolio.Holdings["NABIL"]
	assert.Equal(t, 40, queue.TotalShares())
	assert.Equal(t, NewMoney(40000.0), queue.TotalCost()) // 40 * 1000

	// Verify realized gains
	assert.Equal(t, 1, len(portfolio.RealizedGains))
	gain := portfolio.RealizedGains[0]
	assert.Equal(t, "NABIL", gain.StockSymbol)
	assert.Equal(t, 60, gain.Quantity)
	assert.Equal(t, NewMoney(1200.0), gain.SalePrice)
	assert.Equal(t, NewMoney(1000.0), gain.CostBasis)
	assert.Equal(t, NewMoney(12000.0), gain.GainLoss) // (1200-1000) * 60
}

func TestPortfolio_CompleteSellOff(t *testing.T) {
	portfolio := NewPortfolio()

	// Buy shares
	buyTx := Transaction{
		StockSymbol: "EBL",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionBuy,
		Quantity:    50,
		Price:       NewMoney(800.0),
		Cost:        NewMoney(40000.0),
	}

	err := portfolio.AddTransaction(buyTx)
	require.NoError(t, err)

	// Sell all shares
	sellTx := Transaction{
		StockSymbol: "EBL",
		Date:        time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionSell,
		Quantity:    50,
		Price:       NewMoney(900.0),
		Cost:        NewMoney(45000.0),
	}

	err = portfolio.AddTransaction(sellTx)
	require.NoError(t, err)

	// Verify holding is removed after complete sell-off
	assert.False(t, portfolio.HasHolding("EBL"))
	assert.Equal(t, 0, len(portfolio.Holdings))

	// Verify realized gain is recorded
	assert.Equal(t, 1, len(portfolio.RealizedGains))
	gain := portfolio.RealizedGains[0]
	assert.Equal(t, NewMoney(5000.0), gain.GainLoss) // (900-800) * 50
}

func TestPortfolio_BonusShares(t *testing.T) {
	portfolio := NewPortfolio()

	// Buy shares
	buyTx := Transaction{
		StockSymbol: "HIDCL",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionBuy,
		Quantity:    100,
		Price:       NewMoney(300.0),
		Cost:        NewMoney(30000.0),
	}

	err := portfolio.AddTransaction(buyTx)
	require.NoError(t, err)

	// Bonus shares (20% bonus)
	bonusTx := Transaction{
		StockSymbol: "HIDCL",
		Date:        time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionBonus,
		Quantity:    20,
		Price:       Zero(),
		Cost:        Zero(),
	}

	err = portfolio.AddTransaction(bonusTx)
	require.NoError(t, err)

	// Verify updated holdings
	queue := portfolio.Holdings["HIDCL"]
	assert.Equal(t, 120, queue.TotalShares())                     // 100 + 20
	assert.Equal(t, NewMoney(30000.0), queue.TotalCost())         // Same cost
	assert.Equal(t, NewMoney(250.0), queue.WeightedAverageCost()) // 30000/120 = 250
}

func TestPortfolio_ProcessTransactionsWithAutoSort(t *testing.T) {
	portfolio := NewPortfolio()

	// Create transactions in wrong chronological order
	transactions := []Transaction{
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), // Later date
			Type:        TransactionSell,
			Quantity:    30,
			Price:       NewMoney(550.0),
		},
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), // Earlier date
			Type:        TransactionBuy,
			Quantity:    100,
			Price:       NewMoney(500.0),
		},
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), // Middle date
			Type:        TransactionBuy,
			Quantity:    50,
			Price:       NewMoney(520.0),
		},
	}

	err := portfolio.ProcessTransactions(transactions)
	require.NoError(t, err)

	// Verify transactions were processed in correct order
	queue := portfolio.Holdings["ADBL"]
	assert.Equal(t, 120, queue.TotalShares()) // (100 + 50) - 30

	// Should have warning about out-of-order transactions
	warnings := portfolio.GetWarnings()
	assert.Greater(t, len(warnings), 0)
	assert.Contains(t, warnings[0], "out of chronological order")
}

func TestPortfolio_GetActiveHoldings(t *testing.T) {
	portfolio := NewPortfolio()

	// Add multiple stocks
	transactions := []Transaction{
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    100,
			Price:       NewMoney(500.0),
		},
		{
			StockSymbol: "NABIL",
			Date:        time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    50,
			Price:       NewMoney(1000.0),
		},
	}

	err := portfolio.ProcessTransactions(transactions)
	require.NoError(t, err)

	// Set current prices
	currentPrices := map[string]Money{
		"ADBL":  NewMoney(550.0),
		"NABIL": NewMoney(1100.0),
	}

	holdings := portfolio.GetActiveHoldings(currentPrices)
	assert.Equal(t, 2, len(holdings))

	// Verify holdings are sorted by symbol
	assert.Equal(t, "ADBL", holdings[0].StockSymbol)
	assert.Equal(t, "NABIL", holdings[1].StockSymbol)

	// Verify ADBL holding
	adblHolding := holdings[0]
	assert.Equal(t, 100, adblHolding.TotalShares)
	assert.Equal(t, NewMoney(500.0), adblHolding.WeightedAvgCost)
	assert.Equal(t, NewMoney(50000.0), adblHolding.TotalCost)
	assert.Equal(t, NewMoney(550.0), adblHolding.CurrentPrice)
	assert.Equal(t, NewMoney(55000.0), adblHolding.MarketValue)
	assert.Equal(t, NewMoney(5000.0), adblHolding.UnrealizedGainLoss)
}

func TestPortfolio_GetPortfolioSummary(t *testing.T) {
	portfolio := NewPortfolio()

	// Add transactions with realized gains
	transactions := []Transaction{
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    100,
			Price:       NewMoney(500.0),
		},
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionSell,
			Quantity:    40,
			Price:       NewMoney(600.0),
		},
		{
			StockSymbol: "NABIL",
			Date:        time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    50,
			Price:       NewMoney(1000.0),
		},
	}

	err := portfolio.ProcessTransactions(transactions)
	require.NoError(t, err)

	currentPrices := map[string]Money{
		"ADBL":  NewMoney(550.0),
		"NABIL": NewMoney(1100.0),
	}

	summary := portfolio.GetPortfolioSummary(currentPrices)

	// Total invested: (60 * 500) + (50 * 1000) = 30000 + 50000 = 80000
	expectedInvested := NewMoney(80000.0)
	assert.Equal(t, expectedInvested, summary.TotalInvested)

	// Total market value: (60 * 550) + (50 * 1100) = 33000 + 55000 = 88000
	expectedMarketValue := NewMoney(88000.0)
	assert.Equal(t, expectedMarketValue, summary.TotalMarketValue)

	// Unrealized P/L: 88000 - 80000 = 8000
	expectedUnrealizedPL := NewMoney(8000.0)
	assert.Equal(t, expectedUnrealizedPL, summary.TotalUnrealizedPL)

	// Realized P/L: (600 - 500) * 40 = 4000
	expectedRealizedPL := NewMoney(4000.0)
	assert.Equal(t, expectedRealizedPL, summary.TotalRealizedPL)

	assert.Equal(t, 2, summary.HoldingsCount)
}

func TestPortfolio_GetRealizedGainsSummary(t *testing.T) {
	portfolio := NewPortfolio()

	// Add transactions with different holding periods
	transactions := []Transaction{
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC), // Long-term
			Type:        TransactionBuy,
			Quantity:    100,
			Price:       NewMoney(500.0),
		},
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), // > 1 year later
			Type:        TransactionSell,
			Quantity:    50,
			Price:       NewMoney(600.0),
		},
		{
			StockSymbol: "NABIL",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), // Short-term
			Type:        TransactionBuy,
			Quantity:    50,
			Price:       NewMoney(1000.0),
		},
		{
			StockSymbol: "NABIL",
			Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), // < 1 year later
			Type:        TransactionSell,
			Quantity:    30,
			Price:       NewMoney(1200.0),
		},
	}

	err := portfolio.ProcessTransactions(transactions)
	require.NoError(t, err)

	summary := portfolio.GetRealizedGainsSummary()

	// Long-term gain: (600 - 500) * 50 = 5000
	expectedLongTerm := NewMoney(5000.0)
	assert.Equal(t, expectedLongTerm, summary["long_term_gains"])

	// Short-term gain: (1200 - 1000) * 30 = 6000
	expectedShortTerm := NewMoney(6000.0)
	assert.Equal(t, expectedShortTerm, summary["short_term_gains"])

	// Total gains: 5000 + 6000 = 11000
	expectedTotal := NewMoney(11000.0)
	assert.Equal(t, expectedTotal, summary["total_gains"])

	// Tax calculations
	expectedTaxShortTerm := NewMoney(450.0) // 6000 * 7.5%
	expectedTaxLongTerm := NewMoney(250.0)  // 5000 * 5%
	expectedTotalTax := NewMoney(700.0)     // 450 + 250

	assert.Equal(t, expectedTaxShortTerm, summary["estimated_tax_short_term"])
	assert.Equal(t, expectedTaxLongTerm, summary["estimated_tax_long_term"])
	assert.Equal(t, expectedTotalTax, summary["estimated_total_tax"])

	// Counts
	assert.Equal(t, 1, summary["long_term_count"])
	assert.Equal(t, 1, summary["short_term_count"])
	assert.Equal(t, 2, summary["total_sales"])
}

func TestPortfolio_StockSplit(t *testing.T) {
	portfolio := NewPortfolio()

	// Buy shares
	buyTx := Transaction{
		StockSymbol: "UPPER",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionBuy,
		Quantity:    100,
		Price:       NewMoney(400.0),
	}

	err := portfolio.AddTransaction(buyTx)
	require.NoError(t, err)

	// Apply 2:1 stock split
	splitTx := Transaction{
		StockSymbol: "UPPER",
		Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionSplit,
		Quantity:    2, // 2:1 split
		Price:       Zero(),
	}

	err = portfolio.AddTransaction(splitTx)
	require.NoError(t, err)

	// Verify split results
	queue := portfolio.Holdings["UPPER"]
	assert.Equal(t, 200, queue.TotalShares())                     // Doubled
	assert.Equal(t, NewMoney(40000.0), queue.TotalCost())         // Same total cost
	assert.Equal(t, NewMoney(200.0), queue.WeightedAverageCost()) // Halved price

	// Should have warning about split
	warnings := portfolio.GetWarnings()
	assert.Greater(t, len(warnings), 0)
	assert.Contains(t, warnings[0], "stock split")
}

func TestPortfolio_ErrorCases(t *testing.T) {
	portfolio := NewPortfolio()

	// Try to sell without owning shares
	sellTx := Transaction{
		StockSymbol: "NONEXISTENT",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionSell,
		Quantity:    50,
		Price:       NewMoney(100.0),
	}

	err := portfolio.AddTransaction(sellTx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot sell 50 shares, only 0 available")

	// Buy shares first
	buyTx := Transaction{
		StockSymbol: "TEST",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionBuy,
		Quantity:    30,
		Price:       NewMoney(100.0),
	}

	err = portfolio.AddTransaction(buyTx)
	require.NoError(t, err)

	// Try to sell more than owned
	oversellTx := Transaction{
		StockSymbol: "TEST",
		Date:        time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionSell,
		Quantity:    50,
		Price:       NewMoney(120.0),
	}

	err = portfolio.AddTransaction(oversellTx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot sell 50 shares, only 30 available")

	// Test invalid split ratio
	invalidSplitTx := Transaction{
		StockSymbol: "TEST",
		Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionSplit,
		Quantity:    1, // Invalid split ratio
		Price:       Zero(),
	}

	err = portfolio.AddTransaction(invalidSplitTx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid split ratio")

	// Test unsupported transaction type
	unknownTx := Transaction{
		StockSymbol: "TEST",
		Date:        time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionType("UNKNOWN"),
		Quantity:    10,
		Price:       NewMoney(100.0),
	}

	err = portfolio.AddTransaction(unknownTx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported transaction type")
}

func TestPortfolio_ValidateIntegrity(t *testing.T) {
	portfolio := NewPortfolio()

	// Add normal transactions
	transactions := []Transaction{
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    100,
			Price:       NewMoney(500.0),
		},
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionSell,
			Quantity:    40,
			Price:       NewMoney(550.0),
		},
	}

	err := portfolio.ProcessTransactions(transactions)
	require.NoError(t, err)

	// Validate integrity - should pass
	issues := portfolio.ValidateIntegrity()
	assert.Equal(t, 0, len(issues))
}

func TestPortfolio_GetHoldingBySymbol(t *testing.T) {
	portfolio := NewPortfolio()

	buyTx := Transaction{
		StockSymbol: "ADBL",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Type:        TransactionBuy,
		Quantity:    100,
		Price:       NewMoney(500.0),
	}

	err := portfolio.AddTransaction(buyTx)
	require.NoError(t, err)

	// Get existing holding
	holding, exists := portfolio.GetHoldingBySymbol("ADBL", NewMoney(550.0))
	assert.True(t, exists)
	assert.Equal(t, "ADBL", holding.StockSymbol)
	assert.Equal(t, 100, holding.TotalShares)
	assert.Equal(t, NewMoney(500.0), holding.WeightedAvgCost)

	// Get non-existing holding
	_, exists = portfolio.GetHoldingBySymbol("NONEXISTENT", NewMoney(100.0))
	assert.False(t, exists)
}

func TestPortfolio_GetSymbols(t *testing.T) {
	portfolio := NewPortfolio()

	transactions := []Transaction{
		{
			StockSymbol: "NABIL",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    50,
			Price:       NewMoney(1000.0),
		},
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    100,
			Price:       NewMoney(500.0),
		},
	}

	err := portfolio.ProcessTransactions(transactions)
	require.NoError(t, err)

	symbols := portfolio.GetSymbols()
	assert.Equal(t, 2, len(symbols))
	assert.Equal(t, []string{"ADBL", "NABIL"}, symbols) // Should be sorted
}

func TestPortfolio_GetRealizedGainsFiltered(t *testing.T) {
	portfolio := NewPortfolio()

	transactions := []Transaction{
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    100,
			Price:       NewMoney(500.0),
		},
		{
			StockSymbol: "NABIL",
			Date:        time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    50,
			Price:       NewMoney(1000.0),
		},
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionSell,
			Quantity:    40,
			Price:       NewMoney(550.0),
		},
		{
			StockSymbol: "NABIL",
			Date:        time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC),
			Type:        TransactionSell,
			Quantity:    20,
			Price:       NewMoney(1100.0),
		},
	}

	err := portfolio.ProcessTransactions(transactions)
	require.NoError(t, err)

	// Get all realized gains
	allGains := portfolio.GetRealizedGains("")
	assert.Equal(t, 2, len(allGains))

	// Get gains filtered by symbol
	adblGains := portfolio.GetRealizedGains("ADBL")
	assert.Equal(t, 1, len(adblGains))
	assert.Equal(t, "ADBL", adblGains[0].StockSymbol)

	nabilGains := portfolio.GetRealizedGains("NABIL")
	assert.Equal(t, 1, len(nabilGains))
	assert.Equal(t, "NABIL", nabilGains[0].StockSymbol)
}

func TestPortfolio_WarningsManagement(t *testing.T) {
	portfolio := NewPortfolio()

	// Process out-of-order transactions to generate warnings
	transactions := []Transaction{
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionSell,
			Quantity:    30,
			Price:       NewMoney(550.0),
		},
		{
			StockSymbol: "ADBL",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Type:        TransactionBuy,
			Quantity:    100,
			Price:       NewMoney(500.0),
		},
	}

	err := portfolio.ProcessTransactions(transactions)
	require.NoError(t, err)

	// Should have warnings
	warnings := portfolio.GetWarnings()
	assert.Greater(t, len(warnings), 0)

	// Clear warnings
	portfolio.ClearWarnings()
	clearedWarnings := portfolio.GetWarnings()
	assert.Equal(t, 0, len(clearedWarnings))
}
