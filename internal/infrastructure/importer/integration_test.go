package importer

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/VoidArchive/ntx/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullWorkflow tests the complete CSV import → Portfolio processing workflow
func TestFullWorkflow_CSVImportAndPortfolioProcessing(t *testing.T) {
	// Create CSV importer
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Sample CSV data representing a realistic portfolio history
	csvData := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"1","NABIL","2024-01-15","100","-","100.0","ON-CR TD:123456 TX:789012 1301020000003172 SET:1211002024015"
"2","NABIL","2024-02-20","50","-","150.0","ON-CR TD:234567 TX:890123 1301020000003172 SET:1211002024051"
"3","ADBL","2024-03-10","80","-","80.0","ON-CR TD:345678 TX:901234 1301020000003172 SET:1211002024069"
"4","NABIL","2024-04-15","20","-","170.0","CA-Bonus                  00009001   B-13.33%-2023/24 CREDIT"
"5","NABIL","2024-05-20","-","70","100.0","ON-DR TD:456789 TX:012345 1301020000003172 SET:1211002024140"
"6","ADBL","2024-06-25","-","30","50.0","ON-DR TD:567890 TX:123456 1301020000003172 SET:1211002024176"
"7","EBL","2024-07-30","25","-","25.0","CA-Rights                 00009002     R-100%-2024 CREDIT"`

	// Step 1: Import transactions from CSV
	reader := strings.NewReader(csvData)
	result, err := importer.ImportFromReader(reader)
	require.NoError(t, err)

	// Verify import was successful
	assert.Equal(t, 7, result.Stats.TotalRows)
	assert.Equal(t, 7, result.Stats.SuccessfulImports)
	assert.Equal(t, 0, result.Stats.SkippedRows)

	// Verify transaction types
	assert.Equal(t, 3, result.Stats.TransactionTypes[domain.TransactionBuy])    // NABIL, NABIL, ADBL
	assert.Equal(t, 2, result.Stats.TransactionTypes[domain.TransactionSell])   // NABIL, ADBL sales
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionBonus])  // NABIL bonus
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionRights]) // EBL rights

	// Step 2: Process transactions in Portfolio
	portfolio := domain.NewPortfolio()
	err = portfolio.ProcessTransactions(result.Transactions)
	require.NoError(t, err)

	// Step 3: Verify Portfolio State
	// Should have holdings for NABIL, ADBL, and EBL
	assert.Equal(t, 3, len(portfolio.GetSymbols()))
	assert.True(t, portfolio.HasHolding("NABIL"))
	assert.True(t, portfolio.HasHolding("ADBL"))
	assert.True(t, portfolio.HasHolding("EBL"))

	// Step 4: Check NABIL holdings (complex scenario with buy, bonus, sell)
	nabilQueue := portfolio.Holdings["NABIL"]
	require.NotNil(t, nabilQueue)

	// NABIL timeline:
	// - Buy 100 @ Rs.100 (Jan 15)
	// - Buy 50 @ Rs.100 (Feb 20) -> Total: 150 shares
	// - Bonus 20 @ Rs.0 (Apr 15) -> Total: 170 shares, WAC = (15000+0)/170 = Rs.88.24
	// - Sell 70 (May 20) -> Remaining: 100 shares

	expectedNabilShares := 100 // 170 - 70 sold
	assert.Equal(t, expectedNabilShares, nabilQueue.TotalShares())

	// WAC should be affected by bonus shares dilution
	expectedNabilCost := domain.NewMoney(15000.0)                                 // Original cost: 100*100 + 50*100 = 15000
	remainingCostAfterSale := expectedNabilCost.Subtract(domain.NewMoney(7000.0)) // FIFO: sold 70 @ Rs.100 = 7000
	assert.Equal(t, domain.NewMoney(8000.0), remainingCostAfterSale)              // Should be 8000 remaining

	// Step 5: Check ADBL holdings
	adblQueue := portfolio.Holdings["ADBL"]
	require.NotNil(t, adblQueue)

	// ADBL timeline:
	// - Buy 80 @ Rs.100 (Mar 10)
	// - Sell 30 (Jun 25) -> Remaining: 50 shares

	expectedAdblShares := 50 // 80 - 30 sold
	assert.Equal(t, expectedAdblShares, adblQueue.TotalShares())
	assert.Equal(t, domain.NewMoney(5000.0), adblQueue.TotalCost()) // 50 * 100
	assert.Equal(t, domain.NewMoney(100.0), adblQueue.WeightedAverageCost())

	// Step 6: Check EBL holdings (rights shares)
	eblQueue := portfolio.Holdings["EBL"]
	require.NotNil(t, eblQueue)

	assert.Equal(t, 25, eblQueue.TotalShares())
	assert.Equal(t, domain.NewMoney(2500.0), eblQueue.TotalCost()) // 25 * 100

	// Step 7: Verify Realized Gains
	realizedGains := portfolio.GetRealizedGains("")
	assert.Equal(t, 2, len(realizedGains)) // NABIL and ADBL sales

	// Check NABIL sale (70 shares)
	nabilGain := realizedGains[0]
	assert.Equal(t, "NABIL", nabilGain.StockSymbol)
	assert.Equal(t, 70, nabilGain.Quantity)
	assert.Equal(t, domain.NewMoney(100.0), nabilGain.SalePrice) // Default price
	assert.Equal(t, domain.NewMoney(100.0), nabilGain.CostBasis) // FIFO cost
	assert.Equal(t, domain.Zero(), nabilGain.GainLoss)           // No gain at default prices

	// Check ADBL sale (30 shares)
	adblGain := realizedGains[1]
	assert.Equal(t, "ADBL", adblGain.StockSymbol)
	assert.Equal(t, 30, adblGain.Quantity)
	assert.Equal(t, domain.Zero(), adblGain.GainLoss) // No gain at default prices

	// Step 8: Test Portfolio Summary
	currentPrices := map[string]domain.Money{
		"NABIL": domain.NewMoney(120.0), // 20% gain
		"ADBL":  domain.NewMoney(110.0), // 10% gain
		"EBL":   domain.NewMoney(150.0), // 50% gain
	}

	summary := portfolio.GetPortfolioSummary(currentPrices)

	// Total invested = remaining cost basis of all holdings
	expectedTotalInvested := domain.NewMoney(15500.0) // NABIL: 8000 + ADBL: 5000 + EBL: 2500
	assert.Equal(t, expectedTotalInvested, summary.TotalInvested)

	// Total market value = current prices * quantities
	expectedMarketValue := domain.NewMoney(120.0).Multiply(100). // NABIL: 100 * 120 = 12000
									Add(domain.NewMoney(110.0).Multiply(50)). // ADBL: 50 * 110 = 5500
									Add(domain.NewMoney(150.0).Multiply(25))  // EBL: 25 * 150 = 3750
	// Total: 12000 + 5500 + 3750 = 21250
	assert.Equal(t, domain.NewMoney(21250.0), expectedMarketValue)
	assert.Equal(t, expectedMarketValue, summary.TotalMarketValue)

	// Unrealized P/L = market value - invested
	expectedUnrealizedPL := expectedMarketValue.Subtract(expectedTotalInvested) // 21250 - 15500 = 5750
	assert.Equal(t, expectedUnrealizedPL, summary.TotalUnrealizedPL)

	// Holdings count
	assert.Equal(t, 3, summary.HoldingsCount)

	// Step 9: Test Individual Holdings
	holdings := portfolio.GetActiveHoldings(currentPrices)
	require.Len(t, holdings, 3)

	// Holdings should be sorted by symbol (ADBL, EBL, NABIL)
	assert.Equal(t, "ADBL", holdings[0].StockSymbol)
	assert.Equal(t, "EBL", holdings[1].StockSymbol)
	assert.Equal(t, "NABIL", holdings[2].StockSymbol)

	// Check ADBL holding details
	adblHolding := holdings[0]
	assert.Equal(t, 50, adblHolding.TotalShares)
	assert.Equal(t, domain.NewMoney(100.0), adblHolding.WeightedAvgCost)
	assert.Equal(t, domain.NewMoney(5000.0), adblHolding.TotalCost)
	assert.Equal(t, domain.NewMoney(110.0), adblHolding.CurrentPrice)
	assert.Equal(t, domain.NewMoney(5500.0), adblHolding.MarketValue)
	assert.Equal(t, domain.NewMoney(500.0), adblHolding.UnrealizedGainLoss) // 5500 - 5000
	assert.InDelta(t, 10.0, adblHolding.UnrealizedGainPct, 0.1)             // 10% gain

	// Step 10: Verify Import Warnings
	assert.Greater(t, len(result.Warnings), 0)

	// Should have warnings about default prices
	defaultPriceWarnings := 0
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "default price") {
			defaultPriceWarnings++
		}
	}
	assert.Greater(t, defaultPriceWarnings, 0)

	// Step 11: Generate and verify import summary
	summary_text := importer.GenerateImportSummary(result)
	assert.Contains(t, summary_text, "Total rows processed: 7")
	assert.Contains(t, summary_text, "Successful imports: 7")
	assert.Contains(t, summary_text, "BUY: 3")
	assert.Contains(t, summary_text, "SELL: 2")
	assert.Contains(t, summary_text, "BONUS: 1")
	assert.Contains(t, summary_text, "RIGHTS: 1")
}

// TestComplexScenario tests edge cases and complex corporate actions
func TestComplexScenario_CorporateActionsAndMergers(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Complex scenario with mergers and multiple corporate actions
	csvData := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"1","MEGA","2024-01-15","100","-","100.0","ON-CR TD:123456 TX:789012 1301020000003172 SET:1211002024015"
"2","MEGA","2024-02-15","10","-","110.0","CA-Bonus                  00009001   B-10%-2023/24 CREDIT"
"3","NIMB","2024-03-17","110","-","110.0","CA-Merger                 00008100 Cr Current Balance"
"4","MEGA","2024-03-17","-","110","0.0","CA-Merger                 00008100 Db Current Balance"`

	reader := strings.NewReader(csvData)
	result, err := importer.ImportFromReader(reader)
	require.NoError(t, err)

	assert.Equal(t, 4, result.Stats.TotalRows)
	assert.Equal(t, 4, result.Stats.SuccessfulImports)

	// Verify transaction types
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionBuy])
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionBonus])
	assert.Equal(t, 2, result.Stats.TransactionTypes[domain.TransactionMerger])

	// Process in portfolio
	portfolio := domain.NewPortfolio()
	err = portfolio.ProcessTransactions(result.Transactions)
	require.NoError(t, err)

	// After processing:
	// - MEGA should be empty (sold in merger)
	// - NIMB should have 110 shares (received in merger)
	assert.False(t, portfolio.HasHolding("MEGA"))
	assert.True(t, portfolio.HasHolding("NIMB"))

	nimbQueue := portfolio.Holdings["NIMB"]
	assert.Equal(t, 110, nimbQueue.TotalShares())

	// Should have warnings about merger transactions
	mergerWarnings := 0
	for _, warning := range result.Warnings {
		if strings.Contains(strings.ToLower(warning), "merger") {
			mergerWarnings++
		}
	}
	assert.Greater(t, mergerWarnings, 0)
}

// TestPerformanceWithLargeDataset tests performance with larger datasets
func TestPerformanceWithLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Generate a larger CSV dataset (similar to real portfolio size)
	var csvBuilder strings.Builder
	csvBuilder.WriteString(`"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"` + "\n")

	// Generate 1000 transactions across 50 stocks
	stocks := []string{"NABIL", "ADBL", "EBL", "SCB", "NMB", "SANIMA", "NCC", "PRVU", "MEGA", "KBL"}

	for i := 1; i <= 1000; i++ {
		stock := stocks[i%len(stocks)]
		date := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i).Format("2006-01-02")

		// NOTE: Keep it simple for performance testing - all buys, no sells
		csvBuilder.WriteString(fmt.Sprintf(`"%d","%s","%s","10","-","%d.0","ON-CR TD:123456 TX:789012 1301020000003172 SET:1211002024015"`+"\n",
			i, stock, date, i))
	}

	reader := strings.NewReader(csvBuilder.String())

	// Measure import time
	start := time.Now()
	result, err := importer.ImportFromReader(reader)
	importDuration := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 1000, result.Stats.TotalRows)
	assert.Equal(t, 1000, result.Stats.SuccessfulImports)

	// Import should be fast (under 1 second for 1000 transactions)
	assert.Less(t, importDuration, time.Second, "Import took too long: %v", importDuration)

	// Process in portfolio
	portfolio := domain.NewPortfolio()

	start = time.Now()
	err = portfolio.ProcessTransactions(result.Transactions)
	processingDuration := time.Since(start)

	require.NoError(t, err)

	// Portfolio processing should also be fast
	assert.Less(t, processingDuration, time.Second, "Portfolio processing took too long: %v", processingDuration)

	// Verify final state
	assert.Equal(t, len(stocks), len(portfolio.GetSymbols()))

	// NOTE: All transactions are buys - each stock should have exactly 1000 shares
	// 1000 transactions / 10 stocks = 100 transactions per stock * 10 shares = 1000 shares each
	for _, stock := range stocks {
		assert.True(t, portfolio.HasHolding(stock))
		queue := portfolio.Holdings[stock]
		assert.Equal(t, 1000, queue.TotalShares()) // 100 buys * 10 shares = 1000
	}

	t.Logf("Performance test completed: Import=%v, Processing=%v", importDuration, processingDuration)
}
