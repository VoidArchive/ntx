package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Transaction represents a single transaction from the CSV
type Transaction struct {
	ID               int       `json:"id"`
	Scrip            string    `json:"scrip"`
	Date             time.Time `json:"date"`
	Quantity         int       `json:"quantity"`
	Price            float64   `json:"price"`
	TransactionType  string    `json:"transaction_type"`
	Description      string    `json:"description"`
	BalanceAfter     float64   `json:"balance_after"`
}

// ParseCSV parses the Meroshare CSV file and returns transactions
func ParseCSV(filename string) ([]Transaction, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Skip header row
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file has no data rows")
	}

	var transactions []Transaction
	for i, record := range records[1:] { // Skip header
		if len(record) < 7 {
			continue // Skip malformed rows
		}

		transaction, err := parseTransaction(record, i+1)
		if err != nil {
			// Log error but continue processing other rows
			fmt.Printf("Warning: Failed to parse row %d: %v\n", i+2, err)
			continue
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// parseTransaction converts a CSV record to a Transaction struct
func parseTransaction(record []string, rowNum int) (Transaction, error) {
	var t Transaction
	
	// Parse serial number (not used as ID)
	
	// Parse scrip
	t.Scrip = strings.TrimSpace(record[1])
	if t.Scrip == "" {
		return t, fmt.Errorf("scrip is required")
	}

	// Parse date
	dateStr := strings.TrimSpace(record[2])
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return t, fmt.Errorf("invalid date format: %s", dateStr)
	}
	t.Date = date

	// Parse quantities (credit and debit)
	creditStr := strings.TrimSpace(record[3])
	debitStr := strings.TrimSpace(record[4])
	
	if creditStr != "" && creditStr != "-" {
		credit, err := strconv.Atoi(creditStr)
		if err != nil {
			return t, fmt.Errorf("invalid credit quantity: %s", creditStr)
		}
		t.Quantity = credit // Positive for buy/credit
	} else if debitStr != "" && debitStr != "-" {
		debit, err := strconv.Atoi(debitStr)
		if err != nil {
			return t, fmt.Errorf("invalid debit quantity: %s", debitStr)
		}
		t.Quantity = -debit // Negative for sell/debit
	} else {
		return t, fmt.Errorf("no valid quantity found")
	}

	// Parse balance after transaction
	balanceStr := strings.TrimSpace(record[5])
	if balanceStr != "" {
		balance, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return t, fmt.Errorf("invalid balance: %s", balanceStr)
		}
		t.BalanceAfter = balance
	}

	// Parse description
	t.Description = strings.TrimSpace(record[6])

	// Determine transaction type based on description
	t.TransactionType = determineTransactionType(t.Description)

	// Price is not available in CSV - will be entered later
	t.Price = 0.0

	return t, nil
}

// determineTransactionType analyzes the description to determine transaction type
func determineTransactionType(description string) string {
	desc := strings.ToUpper(description)
	
	// Check for IPO
	if strings.Contains(desc, "INITIAL PUBLIC OFFERING") || 
	   strings.Contains(desc, "IPO") {
		return "IPO"
	}
	
	// Check for Bonus
	if strings.Contains(desc, "CA-BONUS") || 
	   strings.Contains(desc, "BONUS") {
		return "BONUS"
	}
	
	// Check for Rights
	if strings.Contains(desc, "CA-RIGHTS") || 
	   strings.Contains(desc, "RIGHTS") {
		return "RIGHTS"
	}
	
	// Check for Merger
	if strings.Contains(desc, "CA-MERGER") || 
	   strings.Contains(desc, "MERGER") {
		return "MERGER"
	}
	
	// Check for Rearrangement (mutual fund)
	if strings.Contains(desc, "CA-REARRANGEMENT") || 
	   strings.Contains(desc, "REARRANGEMENT") {
		return "REARRANGEMENT"
	}
	
	// Check for regular trading
	if strings.Contains(desc, "ON-CR") || 
	   strings.Contains(desc, "ON-DR") {
		return "REGULAR"
	}
	
	// Default to REGULAR for unknown types
	return "REGULAR"
}

// IsBuy returns true if the transaction is a buy (positive quantity)
func (t Transaction) IsBuy() bool {
	return t.Quantity > 0
}

// IsSell returns true if the transaction is a sell (negative quantity)
func (t Transaction) IsSell() bool {
	return t.Quantity < 0
}

// AbsQuantity returns the absolute value of quantity
func (t Transaction) AbsQuantity() int {
	if t.Quantity < 0 {
		return -t.Quantity
	}
	return t.Quantity
}

// NeedsPrice returns true if the transaction needs a price to be entered
func (t Transaction) NeedsPrice() bool {
	// IPO, Bonus, Rights, Merger, and Rearrangement don't need market prices
	switch t.TransactionType {
	case "BONUS", "RIGHTS", "MERGER", "REARRANGEMENT":
		return false
	case "IPO":
		return t.Price == 0.0 // IPO needs price if not already set
	default:
		return t.Price == 0.0 // Regular transactions need price
	}
}