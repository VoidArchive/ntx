package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"ntx/internal/money"
	"os"
	"strconv"
	"strings"
	"time"
)

// ParseCSV parses a Meroshare transaction history CSV file and returns
// a slice of Transaction structs. It handles various transaction types
// including regular trades, IPOs, bonus shares, rights issues, and corporate actions.
//
// The CSV file is expected to have the following columns:
// S.N, Scrip, Transaction Date, Credit Quantity, Debit Quantity, Balance After Transaction, History Description
//
// Returns an error if the file cannot be read or parsed.
func ParseCSV(ctx context.Context, filename string) ([]Transaction, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file %q: %w", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, ErrEmptyFile
	}

	// Validate header row
	header := records[0]
	if len(header) < CSVMinFields {
		return nil, fmt.Errorf("invalid CSV header: expected %d fields, got %d", CSVMinFields, len(header))
	}

	// Skip header row
	if len(records) < 2 {
		return nil, ErrNoDataRows
	}

	var transactions []Transaction
	for i, record := range records[1:] { // Skip header
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if len(record) < CSVMinFields {
			slog.Warn("Skipping malformed CSV row",
				"row", i+2,
				"expected_fields", CSVMinFields,
				"actual_fields", len(record))
			continue
		}

		transaction, err := parseTransaction(record)
		if err != nil {
			// Log error but continue processing other rows
			slog.Warn("Failed to parse CSV row",
				"row", i+2,
				"error", err,
				"record", record)
			continue
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// parseTransaction converts a single CSV record to a Transaction struct.
// It handles quantity parsing (credit/debit), date parsing, and transaction
// type detection based on the description field.
func parseTransaction(record []string) (Transaction, error) {
	var t Transaction

	// Validate record length
	if len(record) < CSVMinFields {
		return t, fmt.Errorf("record has %d fields, expected %d", len(record), CSVMinFields)
	}

	// Parse serial number (not used as ID)

	// Parse scrip
	t.Scrip = strings.TrimSpace(record[CSVFieldScrip])
	if t.Scrip == "" {
		return t, fmt.Errorf("%w: scrip", ErrRequiredField)
	}

	// Parse date
	dateStr := strings.TrimSpace(record[CSVFieldDate])
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return t, fmt.Errorf("invalid date format %q: %w", dateStr, err)
	}
	t.Date = date

	// Parse quantities (credit and debit)
	creditStr := strings.TrimSpace(record[CSVFieldCredit])
	debitStr := strings.TrimSpace(record[CSVFieldDebit])

	if creditStr != "" && creditStr != "-" {
		credit, err := strconv.Atoi(creditStr)
		if err != nil {
			return t, fmt.Errorf("invalid credit quantity %q: %w", creditStr, err)
		}
		t.Quantity = credit // Positive for buy/credit
	} else if debitStr != "" && debitStr != "-" {
		debit, err := strconv.Atoi(debitStr)
		if err != nil {
			return t, fmt.Errorf("invalid debit quantity %q: %w", debitStr, err)
		}
		t.Quantity = -debit // Negative for sell/debit
	} else {
		return t, ErrInvalidQuantity
	}

	// Parse balance after transaction (share quantity)
	balanceStr := strings.TrimSpace(record[CSVFieldBalance])
	if balanceStr != "" {
		// Parse as float first to handle decimal points in CSV, then convert to int
		balanceFloat, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return t, fmt.Errorf("invalid balance %q: %w", balanceStr, err)
		}
		t.BalanceAfter = int(balanceFloat) // Convert to whole shares
	}

	// Parse description
	t.Description = strings.TrimSpace(record[CSVFieldDescription])

	// Determine transaction type based on description
	t.TransactionType = determineTransactionType(t.Description)

	// Price is not available in CSV - will be entered later
	t.Price = money.Money(0)

	return t, nil
}

// determineTransactionType analyzes the transaction description to categorize
// the transaction type. It recognizes IPOs, bonus shares, rights issues,
// mergers, rearrangements, and regular trading based on keywords in the description.
func determineTransactionType(description string) TransactionType {
	desc := strings.ToUpper(description)

	// Check for IPO
	if strings.Contains(desc, "INITIAL PUBLIC OFFERING") ||
		strings.Contains(desc, "IPO") {
		return TransactionTypeIPO
	}

	// Check for Bonus
	if strings.Contains(desc, "CA-BONUS") ||
		strings.Contains(desc, "BONUS") {
		return TransactionTypeBonus
	}

	// Check for Rights
	if strings.Contains(desc, "CA-RIGHTS") ||
		strings.Contains(desc, "RIGHTS") {
		return TransactionTypeRights
	}

	// Check for Merger
	if strings.Contains(desc, "CA-MERGER") ||
		strings.Contains(desc, "MERGER") {
		return TransactionTypeMerger
	}

	// Check for Rearrangement (mutual fund)
	if strings.Contains(desc, "CA-REARRANGEMENT") ||
		strings.Contains(desc, "REARRANGEMENT") {
		return TransactionTypeRearrangement
	}

	// Check for regular trading
	if strings.Contains(desc, "ON-CR") ||
		strings.Contains(desc, "ON-DR") {
		return TransactionTypeRegular
	}

	// Default to REGULAR for unknown types
	return TransactionTypeRegular
}
