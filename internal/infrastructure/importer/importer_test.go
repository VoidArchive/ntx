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

func TestCSVImporter_ValidateHeader(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Valid header
	validHeader := []string{
		"S.N", "Scrip", "Transaction Date", "Credit Quantity",
		"Debit Quantity", "Balance After Transaction", "History Description",
	}

	err := importer.validateHeader(validHeader)
	assert.NoError(t, err)

	// Invalid header - wrong number of columns
	invalidHeader1 := []string{"S.N", "Scrip", "Date"}
	err = importer.validateHeader(invalidHeader1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 7 columns, got 3")

	// Invalid header - wrong column names
	invalidHeader2 := []string{
		"ID", "Symbol", "Date", "Credit", "Debit", "Balance", "Description",
	}
	err = importer.validateHeader(invalidHeader2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 'S.N', got 'ID'")
}

func TestCSVImporter_ParseTransactionDate(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Valid date
	date, err := importer.parseTransactionDate("2025-06-22")
	require.NoError(t, err)
	expected := time.Date(2025, 6, 22, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, date)

	// Invalid date format
	_, err = importer.parseTransactionDate("22/06/2025")
	assert.Error(t, err)

	// Invalid date
	_, err = importer.parseTransactionDate("2025-13-45")
	assert.Error(t, err)
}

func TestCSVImporter_DetermineTransactionType(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	testCases := []struct {
		name           string
		record         CSVRecord
		expectedType   domain.TransactionType
		expectedQty    int
		expectWarnings bool
	}{
		{
			name: "Buy Transaction",
			record: CSVRecord{
				CreditQuantity:     "50",
				DebitQuantity:      "-",
				HistoryDescription: "ON-CR TD:870496 TX:746047 1301020000003172 SET:1211002025130",
			},
			expectedType: domain.TransactionBuy,
			expectedQty:  50,
		},
		{
			name: "Sell Transaction",
			record: CSVRecord{
				CreditQuantity:     "-",
				DebitQuantity:      "20",
				HistoryDescription: "ON-DR TD:263417 TX:431885 1301020000003172 SET:1211002025124",
			},
			expectedType: domain.TransactionSell,
			expectedQty:  20,
		},
		{
			name: "Bonus Shares",
			record: CSVRecord{
				CreditQuantity:     "1",
				DebitQuantity:      "-",
				HistoryDescription: "CA-Bonus                  00009458   B-6.5%-2023-24 CREDIT",
			},
			expectedType: domain.TransactionBonus,
			expectedQty:  1,
		},
		{
			name: "Rights Shares",
			record: CSVRecord{
				CreditQuantity:     "25",
				DebitQuantity:      "-",
				HistoryDescription: "CA-Rights                 00006300     R-50%-076/77 CREDIT",
			},
			expectedType: domain.TransactionRights,
			expectedQty:  25,
		},
		{
			name: "IPO Purchase",
			record: CSVRecord{
				CreditQuantity:     "10",
				DebitQuantity:      "-",
				HistoryDescription: "INITIAL PUBLIC OFFERING   00000389         IPO-2080 CREDIT",
			},
			expectedType: domain.TransactionBuy,
			expectedQty:  10,
		},
		{
			name: "Merger Credit",
			record: CSVRecord{
				CreditQuantity:     "5",
				DebitQuantity:      "-",
				HistoryDescription: "CA-Merger                 00009286 Cr Current Balance",
			},
			expectedType:   domain.TransactionMerger,
			expectedQty:    5,
			expectWarnings: true,
		},
		{
			name: "Corporate Rearrangement",
			record: CSVRecord{
				CreditQuantity:     "182",
				DebitQuantity:      "-",
				HistoryDescription: "CA-Rearrangement          00009000   PUR 09-04-2025 CREDIT",
			},
			expectedType:   domain.TransactionOther,
			expectedQty:    182,
			expectWarnings: true,
		},
		{
			name: "Unknown Transaction",
			record: CSVRecord{
				CreditQuantity:     "10",
				DebitQuantity:      "-",
				HistoryDescription: "SOME UNKNOWN TRANSACTION TYPE",
			},
			expectedType:   domain.TransactionOther,
			expectedQty:    10,
			expectWarnings: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txType, quantity, warnings := importer.determineTransactionType(tc.record)

			assert.Equal(t, tc.expectedType, txType, "Transaction type mismatch")
			assert.Equal(t, tc.expectedQty, quantity, "Quantity mismatch")

			if tc.expectWarnings {
				assert.Greater(t, len(warnings), 0, "Expected warnings but got none")
			}
		})
	}
}

func TestCSVImporter_ConvertToTransaction(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	record := CSVRecord{
		SN:                      "1",
		Scrip:                   "API",
		TransactionDate:         "2025-06-22",
		CreditQuantity:          "50",
		DebitQuantity:           "-",
		BalanceAfterTransaction: "121.0",
		HistoryDescription:      "ON-CR TD:870496 TX:746047 1301020000003172 SET:1211002025130",
	}

	transaction, warnings, err := importer.convertToTransaction(record, 1)
	require.NoError(t, err)

	// Verify transaction details
	assert.Equal(t, 1, transaction.ID)
	assert.Equal(t, "API", transaction.StockSymbol)
	assert.Equal(t, domain.TransactionBuy, transaction.Type)
	assert.Equal(t, 50, transaction.Quantity)
	assert.Equal(t, domain.NewMoney(100.0), transaction.Price)
	assert.Equal(t, domain.NewMoney(5000.0), transaction.Cost) // 50 * 100

	expectedDate := time.Date(2025, 6, 22, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedDate, transaction.Date)

	// Should have warning about default price
	assert.Greater(t, len(warnings), 0)
	assert.Contains(t, warnings[0], "default price")
}

func TestCSVImporter_ImportFromReader_ValidData(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	csvData := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"1","API","2025-06-22","50","-","121.0","ON-CR TD:870496 TX:746047 1301020000003172 SET:1211002025130"
"2","BPCL","2025-06-11","-","20","0.0","ON-DR TD:263417 TX:431885 1301020000003172 SET:1211002025124"
"3","SCB","2025-04-21","1","-","25.0","CA-Bonus                  00009458   B-6.5%-2023-24 CREDIT"`

	reader := strings.NewReader(csvData)
	result, err := importer.ImportFromReader(reader)
	require.NoError(t, err)

	// Verify import statistics
	assert.Equal(t, 3, result.Stats.TotalRows)
	assert.Equal(t, 3, result.Stats.SuccessfulImports)
	assert.Equal(t, 0, result.Stats.SkippedRows)
	assert.Equal(t, 2, result.Stats.DefaultPricesUsed) // Buy and bonus transactions

	// Verify transaction types
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionBuy])
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionSell])
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionBonus])

	// Verify transactions
	require.Len(t, result.Transactions, 3)

	// First transaction (Buy)
	tx1 := result.Transactions[0]
	assert.Equal(t, "API", tx1.StockSymbol)
	assert.Equal(t, domain.TransactionBuy, tx1.Type)
	assert.Equal(t, 50, tx1.Quantity)

	// Second transaction (Sell)
	tx2 := result.Transactions[1]
	assert.Equal(t, "BPCL", tx2.StockSymbol)
	assert.Equal(t, domain.TransactionSell, tx2.Type)
	assert.Equal(t, 20, tx2.Quantity)

	// Third transaction (Bonus)
	tx3 := result.Transactions[2]
	assert.Equal(t, "SCB", tx3.StockSymbol)
	assert.Equal(t, domain.TransactionBonus, tx3.Type)
	assert.Equal(t, 1, tx3.Quantity)
	assert.Equal(t, domain.Zero(), tx3.Price) // Bonus shares have zero price
}

func TestCSVImporter_ImportFromReader_InvalidData(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// CSV with invalid date
	csvData := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"1","API","invalid-date","50","-","121.0","ON-CR TD:870496 TX:746047 1301020000003172 SET:1211002025130"
"2","BPCL","2025-06-11","invalid-qty","-","0.0","ON-CR TD:263417 TX:431885 1301020000003172 SET:1211002025124"`

	reader := strings.NewReader(csvData)
	result, err := importer.ImportFromReader(reader)
	require.NoError(t, err)

	// Should have errors for both rows
	assert.Equal(t, 2, result.Stats.TotalRows)
	assert.Equal(t, 0, result.Stats.SuccessfulImports)
	assert.Equal(t, 2, result.Stats.SkippedRows)
	assert.Len(t, result.Errors, 2)

	// Check error messages
	assert.Contains(t, result.Errors[0], "invalid transaction date")
	assert.Contains(t, result.Errors[1], "Invalid quantity")
}

func TestCSVImporter_ImportFromReader_InvalidHeader(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	csvData := `"ID","Symbol","Date"
"1","API","2025-06-22"`

	reader := strings.NewReader(csvData)
	_, err := importer.ImportFromReader(reader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CSV format")
}

func TestCSVImporter_ValidateCSVFormat(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Valid format
	validCSV := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"`
	reader := strings.NewReader(validCSV)
	err := importer.ValidateCSVFormat(reader)
	assert.NoError(t, err)

	// Invalid format
	invalidCSV := `"ID","Symbol","Date"`
	reader = strings.NewReader(invalidCSV)
	err = importer.ValidateCSVFormat(reader)
	assert.Error(t, err)
}

func TestCSVImporter_GenerateImportSummary(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	result := &ImportResult{
		Stats: ImportStats{
			TotalRows:         5,
			SuccessfulImports: 4,
			SkippedRows:       1,
			DefaultPricesUsed: 3,
			TransactionTypes: map[domain.TransactionType]int{
				domain.TransactionBuy:   2,
				domain.TransactionSell:  1,
				domain.TransactionBonus: 1,
			},
		},
		Warnings: []string{
			"Row 2: Using default price",
			"Row 4: Unknown transaction type",
		},
		Errors: []string{
			"Row 3: Invalid date format",
		},
	}

	summary := importer.GenerateImportSummary(result)

	assert.Contains(t, summary, "Total rows processed: 5")
	assert.Contains(t, summary, "Successful imports: 4")
	assert.Contains(t, summary, "Skipped rows: 1")
	assert.Contains(t, summary, "Default prices used: 3")
	assert.Contains(t, summary, "BUY: 2")
	assert.Contains(t, summary, "SELL: 1")
	assert.Contains(t, summary, "BONUS: 1")
	assert.Contains(t, summary, "Warnings: 2")
	assert.Contains(t, summary, "Errors: 1")
}

func TestCSVImporter_RealDataSample(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Test with actual data from your CSV sample
	csvData := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"1","API","2025-06-22","50","-","121.0","ON-CR TD:870496 TX:746047 1301020000003172 SET:1211002025130"
"2","NIBLSF","2025-06-19","182","-","1449.0","CA-Rearrangement          00009000   PUR 09-04-2025 CREDIT"
"9","BPCL","2025-06-11","-","20","0.0","ON-DR TD:263417 TX:431885 1301020000003172 SET:1211002025124"
"24","SCB","2025-04-21","1","-","25.0","CA-Bonus                  00009458   B-6.5%-2023-24 CREDIT"
"145","AHPC","2021-07-16","25","-","125.0","CA-Rights                 00006300     R-50%-076/77 CREDIT"
"71","SNLI","2023-09-15","10","-","10.0","INITIAL PUBLIC OFFERING   00000389         IPO-2080 CREDIT"`

	reader := strings.NewReader(csvData)
	result, err := importer.ImportFromReader(reader)
	require.NoError(t, err)

	// Verify all transactions were imported successfully
	assert.Equal(t, 6, result.Stats.TotalRows)
	assert.Equal(t, 6, result.Stats.SuccessfulImports)
	assert.Equal(t, 0, result.Stats.SkippedRows)

	// Verify transaction types distribution
	assert.Equal(t, 2, result.Stats.TransactionTypes[domain.TransactionBuy])    // API buy + SNLI IPO
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionSell])   // BPCL sell
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionBonus])  // SCB bonus
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionRights]) // AHPC rights
	assert.Equal(t, 1, result.Stats.TransactionTypes[domain.TransactionOther])  // NIBLSF rearrangement

	require.Len(t, result.Transactions, 6)

	// Test specific transactions
	transactions := result.Transactions

	// API buy transaction
	apiTx := transactions[0]
	assert.Equal(t, "API", apiTx.StockSymbol)
	assert.Equal(t, domain.TransactionBuy, apiTx.Type)
	assert.Equal(t, 50, apiTx.Quantity)
	assert.Equal(t, domain.NewMoney(100.0), apiTx.Price) // Default price
	assert.Equal(t, domain.NewMoney(5000.0), apiTx.Cost)
	expectedDate := time.Date(2025, 6, 22, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedDate, apiTx.Date)

	// BPCL sell transaction
	bpclTx := transactions[2] // Index 2 in the result
	assert.Equal(t, "BPCL", bpclTx.StockSymbol)
	assert.Equal(t, domain.TransactionSell, bpclTx.Type)
	assert.Equal(t, 20, bpclTx.Quantity)

	// SCB bonus transaction
	scbTx := transactions[3]
	assert.Equal(t, "SCB", scbTx.StockSymbol)
	assert.Equal(t, domain.TransactionBonus, scbTx.Type)
	assert.Equal(t, 1, scbTx.Quantity)
	assert.Equal(t, domain.Zero(), scbTx.Price) // Bonus shares have zero price

	// AHPC rights transaction
	ahpcTx := transactions[4]
	assert.Equal(t, "AHPC", ahpcTx.StockSymbol)
	assert.Equal(t, domain.TransactionRights, ahpcTx.Type)
	assert.Equal(t, 25, ahpcTx.Quantity)

	// SNLI IPO transaction (should be treated as buy)
	snliTx := transactions[5]
	assert.Equal(t, "SNLI", snliTx.StockSymbol)
	assert.Equal(t, domain.TransactionBuy, snliTx.Type)
	assert.Equal(t, 10, snliTx.Quantity)

	// Should have warnings about default prices and rearrangement
	assert.Greater(t, len(result.Warnings), 0)
}

func TestCSVImporter_MergerTransactions(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Test merger transactions (both credit and debit)
	csvData := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"61","SMPDA","2024-08-12","5","-","5.0","CA-Merger                 00009286 Cr Current Balance"
"62","SABSL","2024-08-12","-","5","0.0","CA-Merger                 00009286 Db Current Balance"`

	reader := strings.NewReader(csvData)
	result, err := importer.ImportFromReader(reader)
	require.NoError(t, err)

	assert.Equal(t, 2, result.Stats.SuccessfulImports)
	assert.Equal(t, 2, result.Stats.TransactionTypes[domain.TransactionMerger])

	// Both should be merger transactions with warnings
	for _, tx := range result.Transactions {
		assert.Equal(t, domain.TransactionMerger, tx.Type)
		assert.Equal(t, 5, tx.Quantity)
	}

	// Should have warnings about merger transactions
	assert.Greater(t, len(result.Warnings), 0)
	foundMergerWarning := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "Merger transaction") {
			foundMergerWarning = true
			break
		}
	}
	assert.True(t, foundMergerWarning)
}

func TestCSVImporter_EdgeCases(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Test edge cases
	csvData := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"abc","API","2025-06-22","50","-","121.0","ON-CR TD:870496 TX:746047 1301020000003172 SET:1211002025130"
"2","","2025-06-21","10","-","10.0","ON-CR TD:870496 TX:746047 1301020000003172 SET:1211002025130"
"3","TEST","2025-06-20","0","-","0.0","ON-CR TD:870496 TX:746047 1301020000003172 SET:1211002025130"`

	reader := strings.NewReader(csvData)
	result, err := importer.ImportFromReader(reader)
	require.NoError(t, err)

	// First row should have warning about invalid S.N but still import
	assert.Equal(t, 1, result.Stats.SuccessfulImports) // Only first row should succeed
	assert.Equal(t, 2, result.Stats.SkippedRows)       // Second and third rows should be skipped

	// Check that warnings include S.N issue
	hasSnWarning := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "Invalid S.N") {
			hasSnWarning = true
			break
		}
	}
	assert.True(t, hasSnWarning)

	// Check errors for empty scrip and zero quantity
	assert.Len(t, result.Errors, 2)
}

func TestCSVImporter_StreamingWithProgress(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))
	importer.SetBatchSize(2) // Small batch size for testing

	csvData := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"1","API","2025-06-22","50","-","121.0","ON-CR TD:870496 TX:746047 1301020000003172 SET:1211002025130"
"2","BPCL","2025-06-11","-","20","0.0","ON-DR TD:263417 TX:431885 1301020000003172 SET:1211002025124"
"3","SCB","2025-04-21","1","-","25.0","CA-Bonus                  00009458   B-6.5%-2023-24 CREDIT"
"4","NABIL","2025-04-20","30","-","30.0","ON-CR TD:761672 TX:528382 1301020000003172 SET:1211002025082"
"5","EBL","2025-04-15","25","-","25.0","ON-CR TD:564968 TX:134973 1301020000003172 SET:1211002025079"`

	// Track progress calls
	progressCalls := make([]int, 0)
	progressCallback := func(processed int, stats ImportStats) {
		progressCalls = append(progressCalls, processed)
	}

	reader := strings.NewReader(csvData)
	result, err := importer.ImportFromReaderWithCallback(reader, progressCallback)
	require.NoError(t, err)

	// Verify all transactions imported
	assert.Equal(t, 5, result.Stats.TotalRows)
	assert.Equal(t, 5, result.Stats.SuccessfulImports)

	// Verify progress was called (should be called at batch intervals and at end)
	assert.Greater(t, len(progressCalls), 0)
	assert.Contains(t, progressCalls, 5) // Final call should have total count
}

func TestCSVImporter_BatchSizeConfiguration(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	// Test default batch size
	assert.Equal(t, 1000, importer.BatchSize)

	// Test setting custom batch size
	importer.SetBatchSize(500)
	assert.Equal(t, 500, importer.BatchSize)

	// Test invalid batch size (should not change)
	importer.SetBatchSize(0)
	assert.Equal(t, 500, importer.BatchSize) // Should remain unchanged

	importer.SetBatchSize(-100)
	assert.Equal(t, 500, importer.BatchSize) // Should remain unchanged
}

func TestCSVImporter_MemoryEfficiencyWithLargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory efficiency test in short mode")
	}

	importer := NewCSVImporter(domain.NewMoney(100.0))
	importer.SetBatchSize(100) // Smaller batch size for memory efficiency

	// Create a large CSV in memory (simulating a large file)
	var csvBuilder strings.Builder
	csvBuilder.WriteString(`"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"` + "\n")

	// Generate 5000 transactions
	for i := 1; i <= 5000; i++ {
		stock := fmt.Sprintf("STOCK%d", (i%50)+1) // 50 different stocks
		date := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i).Format("2006-01-02")
		csvBuilder.WriteString(fmt.Sprintf(`"%d","%s","%s","10","-","%d.0","ON-CR TD:123456 TX:789012 1301020000003172 SET:1211002024015"`+"\n",
			i, stock, date, i*10))
	}

	reader := strings.NewReader(csvBuilder.String())

	// Import with streaming (should use constant memory)
	start := time.Now()
	result, err := importer.ImportFromReader(reader)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 5000, result.Stats.TotalRows)
	assert.Equal(t, 5000, result.Stats.SuccessfulImports)

	// Should be fast (streaming should be efficient)
	assert.Less(t, duration, 2*time.Second, "Streaming import took too long: %v", duration)

	t.Logf("Streamed 5000 transactions in %v", duration)
}

func TestCSVImporter_GetSupportedTransactionTypes(t *testing.T) {
	importer := NewCSVImporter(domain.NewMoney(100.0))

	supportedTypes := importer.GetSupportedTransactionTypes()

	expectedTypes := []domain.TransactionType{
		domain.TransactionBuy,
		domain.TransactionSell,
		domain.TransactionBonus,
		domain.TransactionRights,
		domain.TransactionMerger,
		domain.TransactionOther,
	}

	assert.ElementsMatch(t, expectedTypes, supportedTypes)
}
