package meroshare

import (
	"strings"
	"testing"
	"time"
)

func TestParseCSV(t *testing.T) {
	csvData := `S.No,Transaction Date,Symbol,Transaction Type,Units,Rate,Amount
1,2023-10-27,NABIL,Buy,10,500,5050
2,2023-11-01,ADBL,Bonus,5,0,0
3,2023-11-05,NABIL,Sell,5,600,2950`

	transactions, err := ParseCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(transactions) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(transactions))
	}

	// Verify first transaction
	tx1 := transactions[0]
	if tx1.Symbol != "NABIL" {
		t.Errorf("Expected NABIL, got %s", tx1.Symbol)
	}
	expectedDate, _ := time.Parse("2006-01-02", "2023-10-27")
	if !tx1.Date.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, tx1.Date)
	}
	if tx1.Quantity != 10 {
		t.Errorf("Expected quantity 10, got %f", tx1.Quantity)
	}

	// Verify bonus transaction
	tx2 := transactions[1]
	if tx2.Symbol != "ADBL" {
		t.Errorf("Expected ADBL, got %s", tx2.Symbol)
	}
	if tx2.Type != "BONUS" {
		t.Errorf("Expected BONUS, got %s", tx2.Type)
	}
}
