package core

import (
	"testing"
	"time"
)

func TestTransactionCreation(t *testing.T) {
	now := time.Now()
	tx := Transaction{
		ID:       "1",
		Date:     now,
		Symbol:   "NABIL",
		Type:     TransactionTypeBuy,
		Quantity: 100,
		Rate:     500,
		Amount:   50500,
	}

	if tx.Symbol != "NABIL" {
		t.Errorf("Expected symbol NABIL, got %s", tx.Symbol)
	}
	if tx.Type != TransactionTypeBuy {
		t.Errorf("Expected type BUY, got %s", tx.Type)
	}
}
