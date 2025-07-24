// Package importer handles CSV transaction file import and parsing
package importer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/VoidArchive/ntx/internal/domain"
)

// CSVImporter handles MeroShare transaction CSV file imports
type CSVImporter struct {
	DefaultPrice domain.Money // Default price for transactions missing price data
	BatchSize    int          // Number of transactions to process in each batch
}

// ImportResult contains the results of a CSV import operation
type ImportResult struct {
	Transactions []domain.Transaction
	Warnings     []string
	Errors       []string
	Stats        ImportStats
}

// ImportStats provides statistics about the import operation
type ImportStats struct {
	TotalRows         int
	SuccessfulImports int
	SkippedRows       int
	DefaultPricesUsed int
	TransactionTypes  map[domain.TransactionType]int
}

// CSVRecord represents a single row from the MeroShare CSV
type CSVRecord struct {
	SN                      string
	Scrip                   string
	TransactionDate         string
	CreditQuantity          string
	DebitQuantity           string
	BalanceAfterTransaction string
	HistoryDescription      string
}

// NewCSVImporter creates a new CSV importer with default price and batch size
func NewCSVImporter(defaultPrice domain.Money) *CSVImporter {
	return &CSVImporter{
		DefaultPrice: defaultPrice,
		BatchSize:    1000, // Process 1000 transactions at a time
	}
}

// ImportFromReader imports transactions from a CSV reader using streaming approach
func (c *CSVImporter) ImportFromReader(reader io.Reader) (*ImportResult, error) {
	// Use buffered reader for efficient streaming
	bufferedReader := bufio.NewReader(reader)
	csvReader := csv.NewReader(bufferedReader)
	csvReader.TrimLeadingSpace = true

	result := &ImportResult{
		Transactions: make([]domain.Transaction, 0, c.BatchSize), // Pre-allocate with batch size
		Warnings:     make([]string, 0),
		Errors:       make([]string, 0),
		Stats: ImportStats{
			TransactionTypes: make(map[domain.TransactionType]int),
		},
	}

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header format
	if err := c.validateHeader(header); err != nil {
		return nil, fmt.Errorf("invalid CSV format: %w", err)
	}

	// Process records in streaming fashion
	rowNum := 2 // Start from 2 (header is row 1)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Failed to parse CSV: %v", rowNum, err))
			rowNum++
			continue
		}

		result.Stats.TotalRows++

		// Parse the CSV record
		csvRecord := c.parseCSVRecord(record)

		// Convert to domain transaction
		transaction, warnings, parseErr := c.convertToTransaction(csvRecord, rowNum)
		if parseErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", rowNum, parseErr))
			result.Stats.SkippedRows++
		} else {
			result.Transactions = append(result.Transactions, *transaction)
			result.Stats.SuccessfulImports++
			result.Stats.TransactionTypes[transaction.Type]++

			// Track default price usage
			if transaction.Price.Equal(c.DefaultPrice) {
				result.Stats.DefaultPricesUsed++
			}
		}

		// Add warnings
		for _, warning := range warnings {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Row %d: %s", rowNum, warning))
		}

		rowNum++

		// Optional: Yield control periodically for large files (every 1000 rows)
		if rowNum%c.BatchSize == 0 {
			// Could add progress callback here if needed
			// progressCallback(rowNum, result.Stats)
		}
	}

	return result, nil
}

// ImportFromReaderWithCallback imports transactions with progress callback for large files
func (c *CSVImporter) ImportFromReaderWithCallback(reader io.Reader, progressCallback func(processed int, stats ImportStats)) (*ImportResult, error) {
	// Use buffered reader for efficient streaming
	bufferedReader := bufio.NewReader(reader)
	csvReader := csv.NewReader(bufferedReader)
	csvReader.TrimLeadingSpace = true

	result := &ImportResult{
		Transactions: make([]domain.Transaction, 0, c.BatchSize),
		Warnings:     make([]string, 0),
		Errors:       make([]string, 0),
		Stats: ImportStats{
			TransactionTypes: make(map[domain.TransactionType]int),
		},
	}

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header format
	if err := c.validateHeader(header); err != nil {
		return nil, fmt.Errorf("invalid CSV format: %w", err)
	}

	// Process records in streaming fashion with progress reporting
	rowNum := 2 // Start from 2 (header is row 1)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Failed to parse CSV: %v", rowNum, err))
			rowNum++
			continue
		}

		result.Stats.TotalRows++

		// Parse the CSV record
		csvRecord := c.parseCSVRecord(record)

		// Convert to domain transaction
		transaction, warnings, parseErr := c.convertToTransaction(csvRecord, rowNum)
		if parseErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", rowNum, parseErr))
			result.Stats.SkippedRows++
		} else {
			result.Transactions = append(result.Transactions, *transaction)
			result.Stats.SuccessfulImports++
			result.Stats.TransactionTypes[transaction.Type]++

			// Track default price usage
			if transaction.Price.Equal(c.DefaultPrice) {
				result.Stats.DefaultPricesUsed++
			}
		}

		// Add warnings
		for _, warning := range warnings {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Row %d: %s", rowNum, warning))
		}

		// Report progress periodically based on processed rows
		if result.Stats.TotalRows%c.BatchSize == 0 && progressCallback != nil {
			progressCallback(result.Stats.TotalRows, result.Stats)
		}

		rowNum++
	}

	// Final progress callback with total processed rows
	if progressCallback != nil {
		progressCallback(result.Stats.TotalRows, result.Stats)
	}

	return result, nil
}

// SetBatchSize allows customizing the batch size for large file processing
func (c *CSVImporter) SetBatchSize(batchSize int) {
	if batchSize > 0 {
		c.BatchSize = batchSize
	}
}

// validateHeader checks if the CSV header matches expected MeroShare format
func (c *CSVImporter) validateHeader(header []string) error {
	expectedHeaders := []string{
		"S.N", "Scrip", "Transaction Date", "Credit Quantity",
		"Debit Quantity", "Balance After Transaction", "History Description",
	}

	if len(header) != len(expectedHeaders) {
		return fmt.Errorf("expected %d columns, got %d", len(expectedHeaders), len(header))
	}

	for i, expected := range expectedHeaders {
		if strings.TrimSpace(header[i]) != expected {
			return fmt.Errorf("column %d: expected '%s', got '%s'", i+1, expected, header[i])
		}
	}

	return nil
}

// parseCSVRecord converts a CSV record array to a structured CSVRecord
func (c *CSVImporter) parseCSVRecord(record []string) CSVRecord {
	return CSVRecord{
		SN:                      strings.TrimSpace(record[0]),
		Scrip:                   strings.TrimSpace(record[1]),
		TransactionDate:         strings.TrimSpace(record[2]),
		CreditQuantity:          strings.TrimSpace(record[3]),
		DebitQuantity:           strings.TrimSpace(record[4]),
		BalanceAfterTransaction: strings.TrimSpace(record[5]),
		HistoryDescription:      strings.TrimSpace(record[6]),
	}
}

// convertToTransaction converts a CSV record to a domain Transaction
func (c *CSVImporter) convertToTransaction(record CSVRecord, rowNum int) (*domain.Transaction, []string, error) {
	var warnings []string

	// Parse transaction date
	date, err := c.parseTransactionDate(record.TransactionDate)
	if err != nil {
		return nil, warnings, fmt.Errorf("invalid transaction date '%s': %w", record.TransactionDate, err)
	}

	// Validate scrip symbol
	if record.Scrip == "" {
		return nil, warnings, fmt.Errorf("empty scrip symbol")
	}

	// Determine transaction type and quantity
	txType, quantity, typeWarnings := c.determineTransactionType(record)
	warnings = append(warnings, typeWarnings...)

	if txType == "" {
		return nil, warnings, fmt.Errorf("could not determine transaction type from description: %s", record.HistoryDescription)
	}

	if quantity <= 0 {
		return nil, warnings, fmt.Errorf("Invalid quantity: %d", quantity)
	}

	// Determine price based on transaction type
	var price domain.Money
	if txType == domain.TransactionBonus || txType == domain.TransactionMerger {
		// NOTE: Bonus shares and mergers have zero cost basis
		price = domain.Zero()
	} else {
		// NOTE: Buy/sell/rights use default price (user can edit later)
		price = c.DefaultPrice
	}
	cost := price.Multiply(quantity)

	// Add warning about default price usage for transactions that need pricing
	if txType == domain.TransactionBuy || txType == domain.TransactionRights {
		warnings = append(warnings, fmt.Sprintf("Using default price %s (please verify and update)", price.String()))
	}

	// Generate transaction ID from S.N
	id, err := strconv.Atoi(record.SN)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Invalid S.N '%s', using row number as ID", record.SN))
		id = rowNum
	}

	transaction := &domain.Transaction{
		ID:          id,
		StockSymbol: record.Scrip,
		Date:        date,
		Type:        txType,
		Quantity:    quantity,
		Price:       price,
		Cost:        cost,
		Description: record.HistoryDescription,
		Note:        "", // User can add notes later
	}

	return transaction, warnings, nil
}

// parseTransactionDate parses the transaction date from YYYY-MM-DD format
func (c *CSVImporter) parseTransactionDate(dateStr string) (time.Time, error) {
	// MeroShare uses YYYY-MM-DD format
	return time.Parse("2006-01-02", dateStr)
}

// determineTransactionType analyzes the CSV record to determine transaction type and quantity
func (c *CSVImporter) determineTransactionType(record CSVRecord) (domain.TransactionType, int, []string) {
	var warnings []string
	description := strings.ToUpper(record.HistoryDescription)

	// Determine transaction type based on history description patterns
	var txType domain.TransactionType
	var quantityStr string

	switch {
	case strings.Contains(description, "ON-CR"):
		txType = domain.TransactionBuy
		quantityStr = record.CreditQuantity

	case strings.Contains(description, "ON-DR"):
		txType = domain.TransactionSell
		quantityStr = record.DebitQuantity

	case strings.Contains(description, "CA-BONUS"):
		txType = domain.TransactionBonus
		quantityStr = record.CreditQuantity

	case strings.Contains(description, "CA-RIGHTS"):
		txType = domain.TransactionRights
		quantityStr = record.CreditQuantity

	case strings.Contains(description, "CA-MERGER"):
		// For mergers, check if it's credit (receiving) or debit (giving up)
		if record.CreditQuantity != "-" && record.CreditQuantity != "" {
			txType = domain.TransactionMerger
			quantityStr = record.CreditQuantity
			warnings = append(warnings, "Merger transaction detected - please verify details")
		} else if record.DebitQuantity != "-" && record.DebitQuantity != "" {
			txType = domain.TransactionMerger
			quantityStr = record.DebitQuantity
			warnings = append(warnings, "Merger transaction (shares given up) - please verify details")
		}

	case strings.Contains(description, "CA-REARRANGEMENT"):
		txType = domain.TransactionOther
		quantityStr = record.CreditQuantity
		warnings = append(warnings, "Corporate rearrangement - may need manual review")

	case strings.Contains(description, "INITIAL PUBLIC OFFERING") || strings.Contains(description, "IPO"):
		txType = domain.TransactionBuy
		quantityStr = record.CreditQuantity

	default:
		// Try to infer from credit/debit quantities
		if record.CreditQuantity != "-" && record.CreditQuantity != "" {
			txType = domain.TransactionOther
			quantityStr = record.CreditQuantity
			warnings = append(warnings, fmt.Sprintf("Unknown transaction type, inferred as OTHER: %s", record.HistoryDescription))
		} else if record.DebitQuantity != "-" && record.DebitQuantity != "" {
			txType = domain.TransactionOther
			quantityStr = record.DebitQuantity
			warnings = append(warnings, fmt.Sprintf("Unknown transaction type, inferred as OTHER: %s", record.HistoryDescription))
		}
	}

	// Parse quantity
	quantity := 0
	if quantityStr != "" && quantityStr != "-" {
		var err error
		quantity, err = strconv.Atoi(quantityStr)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Invalid quantity '%s', defaulting to 0", quantityStr))
			quantity = 0
		}
	}

	return txType, quantity, warnings
}

// GetSupportedTransactionTypes returns the transaction types that the importer can handle
func (c *CSVImporter) GetSupportedTransactionTypes() []domain.TransactionType {
	return []domain.TransactionType{
		domain.TransactionBuy,
		domain.TransactionSell,
		domain.TransactionBonus,
		domain.TransactionRights,
		domain.TransactionMerger,
		domain.TransactionOther,
	}
}

// ValidateCSVFormat performs a quick validation of CSV format without full import
func (c *CSVImporter) ValidateCSVFormat(reader io.Reader) error {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	// Read and validate header
	header, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	return c.validateHeader(header)
}

// GenerateImportSummary creates a human-readable summary of the import results
func (c *CSVImporter) GenerateImportSummary(result *ImportResult) string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("CSV Import Summary:\n"))
	summary.WriteString(fmt.Sprintf("- Total rows processed: %d\n", result.Stats.TotalRows))
	summary.WriteString(fmt.Sprintf("- Successful imports: %d\n", result.Stats.SuccessfulImports))
	summary.WriteString(fmt.Sprintf("- Skipped rows: %d\n", result.Stats.SkippedRows))
	summary.WriteString(fmt.Sprintf("- Default prices used: %d\n", result.Stats.DefaultPricesUsed))

	if len(result.Stats.TransactionTypes) > 0 {
		summary.WriteString("\nTransaction types found:\n")
		for txType, count := range result.Stats.TransactionTypes {
			summary.WriteString(fmt.Sprintf("- %s: %d\n", txType, count))
		}
	}

	if len(result.Warnings) > 0 {
		summary.WriteString(fmt.Sprintf("\nWarnings: %d\n", len(result.Warnings)))
		for i, warning := range result.Warnings {
			if i < 5 { // Show first 5 warnings
				summary.WriteString(fmt.Sprintf("- %s\n", warning))
			} else if i == 5 {
				summary.WriteString(fmt.Sprintf("- ... and %d more warnings\n", len(result.Warnings)-5))
				break
			}
		}
	}

	if len(result.Errors) > 0 {
		summary.WriteString(fmt.Sprintf("\nErrors: %d\n", len(result.Errors)))
		for i, error := range result.Errors {
			if i < 3 { // Show first 3 errors
				summary.WriteString(fmt.Sprintf("- %s\n", error))
			} else if i == 3 {
				summary.WriteString(fmt.Sprintf("- ... and %d more errors\n", len(result.Errors)-3))
				break
			}
		}
	}

	return summary.String()
}
