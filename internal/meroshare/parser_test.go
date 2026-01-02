package meroshare

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQuantity(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"positive integer", "100", 100},
		{"positive float", "100.5", 100.5},
		{"dash means zero", "-", 0},
		{"empty means zero", "", 0},
		{"whitespace trimmed", "  50  ", 50},
		{"zero", "0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQuantity(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectTransactionType(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want TransactionType
	}{
		{
			name: "bonus",
			desc: "CA-Bonus 00010126 Cr Current Balance",
			want: TypeBonus,
		},
		{
			name: "merger credit",
			desc: "CA-Merger 00010267 Cr Current Balance",
			want: TypeMerger,
		},
		{
			name: "merger debit",
			desc: "CA-Merger 00010267 Db Current Balance",
			want: TypeMerger,
		},
		{
			name: "rights",
			desc: "CA-Rights 00009822 R-27.00%208182 CREDIT",
			want: TypeRights,
		},
		{
			name: "rearrangement",
			desc: "CA-Rearrangement 00009000 PUR 09-04-2025 CREDIT",
			want: TypeRearrangement,
		},
		{
			name: "buy",
			desc: "ON-CR TD:194105 TX:293297 1301020000003172 SET:1211002025185",
			want: TypeBuy,
		},
		{
			name: "sell",
			desc: "ON-DR TD:263417 TX:431885 1301020000003172 SET:1211002025124",
			want: TypeSell,
		},
		{
			name: "ipo",
			desc: "INITIAL PUBLIC OFFERING 00000389 IPO-2080 CREDIT",
			want: TypeIPO,
		},
		{
			name: "demat",
			desc: "Demat 01515373 Close - Cr Confirmed Balance",
			want: TypeDemat,
		},
		{
			name: "unknown type falls back to first word",
			desc: "SOMETHING-NEW 12345 extra data",
			want: TransactionType("SOMETHING-NEW"),
		},
		{
			name: "empty string",
			desc: "",
			want: TransactionType(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTransactionType(tt.desc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseHistoryDescription(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want HistoryDetails
	}{
		{
			name: "bonus with rate",
			desc: "CA-Bonus 00009458 B-6.5%-2023-24 CREDIT",
			want: HistoryDetails{
				Type:           TypeBonus,
				RawDescription: "CA-Bonus 00009458 B-6.5%-2023-24 CREDIT",
				ReferenceID:    "00009458",
				BonusRate:      "B-6.5%-2023-24",
			},
		},
		{
			name: "bonus without rate",
			desc: "CA-Bonus 00010126 Cr Current Balance",
			want: HistoryDetails{
				Type:           TypeBonus,
				RawDescription: "CA-Bonus 00010126 Cr Current Balance",
				ReferenceID:    "00010126",
			},
		},
		{
			name: "rights with rate",
			desc: "CA-Rights 00009822 R-27.00%208182 CREDIT",
			want: HistoryDetails{
				Type:           TypeRights,
				RawDescription: "CA-Rights 00009822 R-27.00%208182 CREDIT",
				ReferenceID:    "00009822",
				RightsRate:     "R-27.00%208182",
			},
		},
		{
			name: "rearrangement with purchase date",
			desc: "CA-Rearrangement 00009000 PUR 09-04-2025 CREDIT",
			want: HistoryDetails{
				Type:           TypeRearrangement,
				RawDescription: "CA-Rearrangement 00009000 PUR 09-04-2025 CREDIT",
				ReferenceID:    "00009000",
				PurchaseDate:   "09-04-2025",
			},
		},
		{
			name: "buy trade",
			desc: "ON-CR TD:194105 TX:293297 1301020000003172 SET:1211002025185",
			want: HistoryDetails{
				Type:           TypeBuy,
				RawDescription: "ON-CR TD:194105 TX:293297 1301020000003172 SET:1211002025185",
				ReferenceID:    "TD:194105",
				TradeID:        "194105",
				TransactionID:  "293297",
				SettlementCode: "1211002025185",
			},
		},
		{
			name: "sell trade",
			desc: "ON-DR TD:263417 TX:431885 1301020000003172 SET:1211002025124",
			want: HistoryDetails{
				Type:           TypeSell,
				RawDescription: "ON-DR TD:263417 TX:431885 1301020000003172 SET:1211002025124",
				ReferenceID:    "TD:263417",
				TradeID:        "263417",
				TransactionID:  "431885",
				SettlementCode: "1211002025124",
			},
		},
		{
			name: "ipo",
			desc: "INITIAL PUBLIC OFFERING 00000389 IPO-2080 CREDIT",
			want: HistoryDetails{
				Type:           TypeIPO,
				RawDescription: "INITIAL PUBLIC OFFERING 00000389 IPO-2080 CREDIT",
				ReferenceID:    "00000389",
			},
		},
		{
			name: "demat",
			desc: "Demat 01515373 Close - Cr Confirmed Balance",
			want: HistoryDetails{
				Type:           TypeDemat,
				RawDescription: "Demat 01515373 Close - Cr Confirmed Balance",
				DematID:        "01515373",
			},
		},
		{
			name: "merger",
			desc: "CA-Merger 00010267 Cr Current Balance",
			want: HistoryDetails{
				Type:           TypeMerger,
				RawDescription: "CA-Merger 00010267 Cr Current Balance",
				ReferenceID:    "00010267",
			},
		},
		{
			name: "empty description",
			desc: "",
			want: HistoryDetails{
				RawDescription: "",
			},
		},
		{
			name: "whitespace only",
			desc: "   ",
			want: HistoryDetails{
				RawDescription: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHistoryDescription(tt.desc)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseTransactions(t *testing.T) {
	t.Run("parses valid csv", func(t *testing.T) {
		csvContent := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"1","MNBBL","2025-12-23","6","-","56.0","CA-Bonus 00010126 Cr Current Balance"
"2","API","2025-08-24","10","-","131.0","ON-CR TD:194105 TX:293297 1301020000003172 SET:1211002025185"
"3","BPCL","2025-06-11","-","20","0.0","ON-DR TD:263417 TX:431885 1301020000003172 SET:1211002025124"
`
		path := createTempCSV(t, csvContent)

		txs, err := ParseTransactions(path)
		require.NoError(t, err)
		require.Len(t, txs, 3)

		// First transaction - bonus
		assert.Equal(t, 1, txs[0].SN)
		assert.Equal(t, "MNBBL", txs[0].Scrip)
		assert.Equal(t, time.Date(2025, 12, 23, 0, 0, 0, 0, time.UTC), txs[0].TransactionDate)
		assert.Equal(t, 6.0, txs[0].CreditQuantity)
		assert.Equal(t, 0.0, txs[0].DebitQuantity)
		assert.Equal(t, 56.0, txs[0].BalanceAfterTransaction)
		assert.Equal(t, TypeBonus, txs[0].HistoryDescription.Type)

		// Second transaction - buy
		assert.Equal(t, "API", txs[1].Scrip)
		assert.Equal(t, TypeBuy, txs[1].HistoryDescription.Type)
		assert.Equal(t, "194105", txs[1].HistoryDescription.TradeID)

		// Third transaction - sell
		assert.Equal(t, "BPCL", txs[2].Scrip)
		assert.Equal(t, TypeSell, txs[2].HistoryDescription.Type)
		assert.Equal(t, 0.0, txs[2].CreditQuantity)
		assert.Equal(t, 20.0, txs[2].DebitQuantity)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := ParseTransactions("/nonexistent/path.csv")
		assert.Error(t, err)
	})

	t.Run("handles empty csv with header only", func(t *testing.T) {
		csvContent := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
`
		path := createTempCSV(t, csvContent)

		txs, err := ParseTransactions(path)
		require.NoError(t, err)
		assert.Empty(t, txs)
	})

	t.Run("handles malformed quantities gracefully", func(t *testing.T) {
		csvContent := `"S.N","Scrip","Transaction Date","Credit Quantity","Debit Quantity","Balance After Transaction","History Description"
"1","TEST","2025-01-01","invalid","-","not-a-number","CA-Bonus 00001 Test"
`
		path := createTempCSV(t, csvContent)

		txs, err := ParseTransactions(path)
		require.NoError(t, err)
		require.Len(t, txs, 1)

		// Should have zero values for unparseable fields
		assert.Equal(t, 0.0, txs[0].CreditQuantity)
		assert.Equal(t, 0.0, txs[0].BalanceAfterTransaction)
	})
}

func TestParseTMSTradeDate(t *testing.T) {
	tests := []struct {
		name    string
		tradeID string
		want    time.Time
	}{
		{
			name:    "parses BS date from trade ID",
			tradeID: "20810910SEQUENCE",
			want:    time.Date(2081, 9, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "parses BS date 2081/04/01",
			tradeID: "20810401SEQUENCE",
			want:    time.Date(2081, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "short trade ID returns zero time",
			tradeID: "208109",
			want:    time.Time{},
		},
		{
			name:    "parses BS date 2080/12/15",
			tradeID: "20801215SEQUENCE",
			want:    time.Date(2080, 12, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTMSTradeDate(tt.tradeID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseTMSTradeBook(t *testing.T) {
	t.Run("parses TMS CSV with BS dates", func(t *testing.T) {
		csvContent := `"CLIENT","CLIENT NAME","SYMBOL","EXCHANGE TRADE ID","TRADE DATE","TRADE TIME","BUY/SELL","TRADE QTY","PRICE(NPR)","Value(NPR)"
"001234","John Doe","NABIL","20810401123456","2081-04-01","14:30:00","Buy","10","500.00","5000.00"
"001234","John Doe","NABIL","20810910789123","2081-09-10","15:00:00","Sell","5","600.00","3000.00"
`
		reader := strings.NewReader(csvContent)

		trades, err := ParseTMSTradeBook(reader)
		require.NoError(t, err)
		require.Len(t, trades, 2)

		assert.Equal(t, "NABIL", trades[0].Symbol)
		assert.Equal(t, "20810401123456", trades[0].TradeID)
		assert.Equal(t, true, trades[0].IsBuy)
		assert.Equal(t, 10.0, trades[0].Quantity)
		assert.Equal(t, 500.0, trades[0].Price)
		assert.Equal(t, 5000.0, trades[0].Value)
		assert.Equal(t, time.Date(2081, 4, 1, 0, 0, 0, 0, time.UTC), trades[0].TradeDate)

		assert.Equal(t, "NABIL", trades[1].Symbol)
		assert.Equal(t, "20810910789123", trades[1].TradeID)
		assert.Equal(t, false, trades[1].IsBuy)
		assert.Equal(t, 5.0, trades[1].Quantity)
		assert.Equal(t, 600.0, trades[1].Price)
		assert.Equal(t, 3000.0, trades[1].Value)
		assert.Equal(t, time.Date(2081, 9, 10, 0, 0, 0, 0, time.UTC), trades[1].TradeDate)
	})

	t.Run("handles empty CSV", func(t *testing.T) {
		csvContent := `"CLIENT","CLIENT NAME","SYMBOL","EXCHANGE TRADE ID","TRADE DATE","TRADE TIME","BUY/SELL","TRADE QTY","PRICE(NPR)","Value(NPR)"
`
		reader := strings.NewReader(csvContent)

		trades, err := ParseTMSTradeBook(reader)
		require.NoError(t, err)
		assert.Empty(t, trades)
	})
}

func TestParseTransactions_RealData(t *testing.T) {
	// Skip if real data file doesn't exist
	realPath := "../../data/transaction.csv"
	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		t.Skip("real transaction.csv not found")
	}

	txs, err := ParseTransactions(realPath)
	require.NoError(t, err)
	assert.NotEmpty(t, txs)

	// Verify all transactions have valid types
	for _, tx := range txs {
		assert.NotEmpty(t, tx.Scrip, "scrip should not be empty")
		assert.NotEmpty(t, tx.HistoryDescription.Type, "type should not be empty")
		assert.NotEmpty(t, tx.HistoryDescription.RawDescription, "raw description should not be preserved")
	}
}

func createTempCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
	return path
}
