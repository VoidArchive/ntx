package meroshare

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/VoidArchive/ntx/core"
)

// ParseCSV reads a Meroshare transaction history CSV and returns a slice of core.Transaction.
func ParseCSV(r io.Reader) ([]core.Transaction, error) {
	reader := csv.NewReader(r)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map headers to indices
	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.TrimSpace(strings.ToLower(h))] = i
	}

	var transactions []core.Transaction
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		tx, err := mapRecordToTransaction(record, colMap)
		if err != nil {
			// In a real app, we might want to collect errors instead of failing immediately
			return nil, fmt.Errorf("failed to map record: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func mapRecordToTransaction(record []string, colMap map[string]int) (core.Transaction, error) {
	var tx core.Transaction

	// Required fields (standard Meroshare headers)
	// Typical headers: S.No, Transaction Date, Symbol, Transaction Type, Units, Rate, Amount

	dateStr := getVal(record, colMap, "transaction date")
	if dateStr == "" {
		dateStr = getVal(record, colMap, "date")
	}

	parsedDate, err := time.Parse("2006-01-02", dateStr) // Assuming YYYY-MM-DD, might need adjustment
	if err != nil {
		// Try another common format if it fails
		parsedDate, err = time.Parse("01/02/2006", dateStr)
		if err != nil {
			return tx, fmt.Errorf("invalid date format: %s", dateStr)
		}
	}
	tx.Date = parsedDate

	tx.Symbol = strings.ToUpper(getVal(record, colMap, "symbol"))

	typeStr := strings.ToUpper(getVal(record, colMap, "transaction type"))
	switch {
	case strings.Contains(typeStr, "BUY"):
		tx.Type = core.TransactionTypeBuy
	case strings.Contains(typeStr, "SELL"):
		tx.Type = core.TransactionTypeSell
	case strings.Contains(typeStr, "BONUS"):
		tx.Type = core.TransactionTypeBonus
	case strings.Contains(typeStr, "RIGHT"):
		tx.Type = core.TransactionTypeRight
	default:
		return tx, fmt.Errorf("unknown transaction type: %s", typeStr)
	}

	qty, _ := strconv.ParseFloat(getVal(record, colMap, "units"), 64)
	if qty == 0 {
		qty, _ = strconv.ParseFloat(getVal(record, colMap, "quantity"), 64)
	}
	tx.Quantity = qty

	rate, _ := strconv.ParseFloat(getVal(record, colMap, "rate"), 64)
	tx.Rate = rate

	amount, _ := strconv.ParseFloat(getVal(record, colMap, "amount"), 64)
	tx.Amount = amount

	return tx, nil
}

func getVal(record []string, colMap map[string]int, key string) string {
	if idx, ok := colMap[key]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}
