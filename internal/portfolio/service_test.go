package portfolio

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/voidarchive/ntx/internal/database"
)

func TestImportCSV(t *testing.T) {
	db, err := database.OpenTestDB()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	err = database.AutoMigrate(db)
	require.NoError(t, err)

	service := NewService(db)

	// Buy 10, Buy 5, Sell 3 = 12 shares
	csvData := `S.N.,Scrip,Transaction Date,Credit Quantity,Debit Quantity,Balance After Transaction,History Description
1,NABIL,2026-01-01,10,-,10.0,ON-CR 001234 TD:ABC123 SET:NPL001
2,NABIL,2026-01-02,5,-,15.0,ON-CR 001235 TD:DEF456 SET:NPL002
3,NABIL,2026-01-03,-,3,12.0,ON-DR 001236 TD:GHI789 SET:NPL003`

	result, err := service.ImportCSV(ctx, []byte(csvData))
	require.NoError(t, err)
	require.Equal(t, 3, result.Imported)
	require.Equal(t, 0, result.Skipped)

	// Verify holdings
	holdings, err := service.ListHoldings(ctx)
	require.NoError(t, err)
	require.Len(t, holdings, 1)
	require.Equal(t, "NABIL", holdings[0].Stock.Symbol)
	require.Equal(t, 12.0, holdings[0].Quantity) // 10 + 5 - 3 = 12
}

func TestImportCSV_DuplicateDetection(t *testing.T) {
	db, err := database.OpenTestDB()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	err = database.AutoMigrate(db)
	require.NoError(t, err)

	service := NewService(db)

	csvData := `S.N.,Scrip,Transaction Date,Credit Quantity,Debit Quantity,Balance After Transaction,History Description
1,NABIL,2026-01-01,10,-,10.0,ON-CR 001234 TD:ABC123 SET:NPL001`

	// First import
	result, err := service.ImportCSV(ctx, []byte(csvData))
	require.NoError(t, err)
	require.Equal(t, 1, result.Imported)

	// Second import - should skip duplicate
	result, err = service.ImportCSV(ctx, []byte(csvData))
	require.NoError(t, err)
	require.Equal(t, 0, result.Imported)
	require.Equal(t, 1, result.Skipped)

	// Holdings should still be 10, not 20
	holdings, err := service.ListHoldings(ctx)
	require.NoError(t, err)
	require.Len(t, holdings, 1)
	require.Equal(t, 10.0, holdings[0].Quantity)
}

func TestImportCSV_SellReducesHoldings(t *testing.T) {
	db, err := database.OpenTestDB()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	err = database.AutoMigrate(db)
	require.NoError(t, err)

	service := NewService(db)

	csvData := `S.N.,Scrip,Transaction Date,Credit Quantity,Debit Quantity,Balance After Transaction,History Description
1,NABIL,2026-01-01,100,-,100.0,ON-CR 001234 TD:ABC123 SET:NPL001
2,NABIL,2026-01-02,-,40,60.0,ON-DR 001235 TD:DEF456 SET:NPL002`

	result, err := service.ImportCSV(ctx, []byte(csvData))
	require.NoError(t, err)
	require.Equal(t, 2, result.Imported)

	holdings, err := service.ListHoldings(ctx)
	require.NoError(t, err)
	require.Len(t, holdings, 1)
	require.Equal(t, 60.0, holdings[0].Quantity) // 100 - 40 = 60
}

func TestImportCSV_SoldOutRemovesHolding(t *testing.T) {
	db, err := database.OpenTestDB()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	err = database.AutoMigrate(db)
	require.NoError(t, err)

	service := NewService(db)

	csvData := `S.N.,Scrip,Transaction Date,Credit Quantity,Debit Quantity,Balance After Transaction,History Description
1,NABIL,2026-01-01,50,-,50.0,ON-CR 001234 TD:ABC123 SET:NPL001
2,NABIL,2026-01-02,-,50,0.0,ON-DR 001235 TD:DEF456 SET:NPL002`

	result, err := service.ImportCSV(ctx, []byte(csvData))
	require.NoError(t, err)
	require.Equal(t, 2, result.Imported)

	// Holding should be removed when quantity = 0
	holdings, err := service.ListHoldings(ctx)
	require.NoError(t, err)
	require.Len(t, holdings, 0)
}

func TestImportCSV_MultipleSymbols(t *testing.T) {
	db, err := database.OpenTestDB()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	err = database.AutoMigrate(db)
	require.NoError(t, err)

	service := NewService(db)

	csvData := `S.N.,Scrip,Transaction Date,Credit Quantity,Debit Quantity,Balance After Transaction,History Description
1,NABIL,2026-01-01,10,-,10.0,ON-CR 001234 TD:ABC123 SET:NPL001
2,EBL,2026-01-01,20,-,20.0,ON-CR 001235 TD:DEF456 SET:NPL002
3,NABIL,2026-01-02,5,-,15.0,ON-CR 001236 TD:GHI789 SET:NPL003`

	result, err := service.ImportCSV(ctx, []byte(csvData))
	require.NoError(t, err)
	require.Equal(t, 3, result.Imported)

	holdings, err := service.ListHoldings(ctx)
	require.NoError(t, err)
	require.Len(t, holdings, 2)

	// Check each holding (sorted by symbol)
	holdingsMap := make(map[string]float64)
	for _, h := range holdings {
		holdingsMap[h.Stock.Symbol] = h.Quantity
	}
	require.Equal(t, 20.0, holdingsMap["EBL"])
	require.Equal(t, 15.0, holdingsMap["NABIL"]) // 10 + 5
}

func TestDetectTransactionType(t *testing.T) {
	tests := []struct {
		name     string
		desc     string
		credit   float64
		debit    float64
		expected int64
	}{
		{"buy", "ON-CR 001234", 10, 0, 1},
		{"sell", "ON-DR 001234", 0, 10, 2},
		{"ipo", "INITIAL PUBLIC OFFERING 001234", 10, 0, 5},
		{"bonus", "CA-BONUS 10%", 10, 0, 3},
		{"rights", "CA-RIGHTS 1:4", 10, 0, 4},
		{"merger in", "CA-MERGER", 10, 0, 6},
		{"merger out", "CA-MERGER", 0, 10, 7},
		{"demat", "Demat 12345", 10, 0, 8},
		{"rearrangement", "CA-REARRANGEMENT 00009000 PUR 09-04-2025", 10, 0, 9},
		{"unknown", "UNKNOWN TYPE", 10, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectTransactionType(tt.desc, tt.credit, tt.debit)
			require.Equal(t, tt.expected, result)
		})
	}
}
