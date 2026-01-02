// Package meroshare parses Meroshare transaction history CSV exports.
package meroshare

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type Transaction struct {
	SN                      int
	Scrip                   string
	TransactionDate         time.Time
	CreditQuantity          float64
	DebitQuantity           float64
	BalanceAfterTransaction float64
	HistoryDescription      HistoryDetails
}

// TransactionType is a string to allow future types without code changes.
// Meroshare may add new corporate action types; using string lets us
// capture them without failing.
type TransactionType string

// Known types for convenience. The parser will still work with unknown types.
const (
	TypeBonus         TransactionType = "CA-Bonus"
	TypeMerger        TransactionType = "CA-Merger"
	TypeRights        TransactionType = "CA-Rights"
	TypeRearrangement TransactionType = "CA-Rearrangement"
	TypeBuy           TransactionType = "ON-CR"
	TypeSell          TransactionType = "ON-DR"
	TypeIPO           TransactionType = "IPO"
	TypeDemat         TransactionType = "Demat"
)

// HistoryDetails holds parsed fields from the description column.
// RawDescription is always preserved for debugging or handling unknown formats.
type HistoryDetails struct {
	Type           TransactionType
	RawDescription string
	ReferenceID    string

	TradeID        string
	TransactionID  string
	SettlementCode string

	BonusRate    string
	RightsRate   string
	PurchaseDate string
	DematID      string
}

// ParseTransactions reads a Meroshare CSV export and returns parsed transactions.
// Malformed rows are skipped rather than failing the entire parse, since partial
// data is often more useful than none when analyzing transaction history.
func ParseTransactions(filepath string) ([]Transaction, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Pre-allocate to avoid repeated slice growth.
	transactions := make([]Transaction, 0, len(records)-1)

	// Skip header row (i=0).
	for i := 1; i < len(records); i++ {
		rec := records[i]

		// Parse errors are ignored per-field; zero values are acceptable
		// since missing data in exports is common and shouldn't break parsing.
		sn, _ := strconv.Atoi(strings.TrimSpace(rec[0]))
		scrip := strings.TrimSpace(rec[1])
		date, _ := time.Parse("2006-01-02", strings.TrimSpace(rec[2]))
		creditQty := ParseQuantity(rec[3])
		debitQty := ParseQuantity(rec[4])
		balance, _ := strconv.ParseFloat(strings.TrimSpace(rec[5]), 64)
		history := parseHistoryDescription(rec[6])

		transactions = append(transactions, Transaction{
			SN:                      sn,
			Scrip:                   scrip,
			TransactionDate:         date,
			CreditQuantity:          creditQty,
			DebitQuantity:           debitQty,
			BalanceAfterTransaction: balance,
			HistoryDescription:      history,
		})
	}
	return transactions, nil
}

// ParseQuantity handles Meroshare's quantity format where "-" means zero.
func ParseQuantity(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "-" || s == "" {
		return 0
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func parseHistoryDescription(desc string) HistoryDetails {
	desc = strings.TrimSpace(desc)
	details := HistoryDetails{RawDescription: desc}

	if desc == "" {
		return details
	}

	details.Type = detectTransactionType(desc)
	parts := strings.Fields(desc)

	// Reference ID position varies by type due to inconsistent Meroshare formatting.
	switch details.Type {
	case TypeIPO:
		// IPO has 3 words before the ID: "INITIAL PUBLIC OFFERING 00000389 ..."
		if len(parts) > 3 {
			details.ReferenceID = parts[3]
		}
	case TypeDemat:
		// Demat entries have the demat number as second field.
		if len(parts) > 1 {
			details.DematID = parts[1]
		}
	default:
		// All other types: "TYPE REFID ..."
		if len(parts) > 1 {
			details.ReferenceID = parts[1]
		}
	}

	parseTradeFields(&details, parts)
	parseRateFields(&details, desc)
	parsePurchaseDate(&details, desc)

	return details
}

// detectTransactionType identifies the type from description prefix.
// Uses prefix matching rather than exact match so new CA-* or ON-* types
// from Meroshare are captured automatically.
func detectTransactionType(desc string) TransactionType {
	// Multi-word type must be checked first.
	if strings.HasPrefix(desc, "INITIAL PUBLIC OFFERING") {
		return TypeIPO
	}

	// CA- prefix indicates corporate actions (bonus, merger, rights, etc).
	if strings.HasPrefix(desc, "CA-") {
		parts := strings.Fields(desc)
		if len(parts) > 0 {
			return TransactionType(parts[0])
		}
	}

	// ON- prefix indicates online trades (buy/sell).
	if strings.HasPrefix(desc, "ON-") {
		parts := strings.Fields(desc)
		if len(parts) > 0 {
			return TransactionType(parts[0])
		}
	}

	if strings.HasPrefix(desc, "Demat") {
		return TypeDemat
	}

	// Fallback: use first word so unknown types still get captured.
	parts := strings.Fields(desc)
	if len(parts) > 0 {
		return TransactionType(parts[0])
	}

	return ""
}

// parseTradeFields extracts trade identifiers used for reconciliation
// with broker statements.
func parseTradeFields(details *HistoryDetails, parts []string) {
	for _, part := range parts {
		switch {
		case strings.HasPrefix(part, "TD:"):
			details.TradeID = strings.TrimPrefix(part, "TD:")
		case strings.HasPrefix(part, "TX:"):
			details.TransactionID = strings.TrimPrefix(part, "TX:")
		case strings.HasPrefix(part, "SET:"):
			details.SettlementCode = strings.TrimPrefix(part, "SET:")
		}
	}
}

// parseRateFields extracts percentage rates from bonus/rights descriptions.
// Format: "B-6.5%-2023-24" for bonus, "R-27.00%" for rights.
func parseRateFields(details *HistoryDetails, desc string) {
	for part := range strings.FieldsSeq(desc) {
		switch {
		case strings.HasPrefix(part, "B-") && strings.Contains(part, "%"):
			details.BonusRate = part
		case strings.HasPrefix(part, "R-") && strings.Contains(part, "%"):
			details.RightsRate = part
		}
	}
}

// parsePurchaseDate extracts the original purchase date from rearrangement
// entries, needed for calculating holding period.
func parsePurchaseDate(details *HistoryDetails, desc string) {
	_, after, found := strings.Cut(desc, "PUR ")
	if found && len(after) >= 10 {
		details.PurchaseDate = after[:10]
	}
}

// TMSTrade represents a row from TMS Trade Book export.
type TMSTrade struct {
	Symbol    string
	TradeID   string
	TradeDate time.Time
	IsBuy     bool
	Quantity  float64
	Price     float64 // per share in NPR
	Value     float64 // total value in NPR
}

// ParseTMSTradeBook parses TMS Trade Book CSV export.
// CSV format: CLIENT, CLIENT NAME, SYMBOL, EXCHANGE TRADE ID, TRADE DATE, TRADE TIME, BUY/SELL, TRADE QTY, PRICE(NPR), Value(NPR)
func ParseTMSTradeBook(r io.Reader) ([]TMSTrade, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return []TMSTrade{}, nil
	}

	trades := make([]TMSTrade, 0, len(records)-1)

	for i := 1; i < len(records); i++ {
		rec := records[i]
		if len(rec) < 10 {
			continue
		}

		symbol := strings.TrimSpace(rec[2])
		if symbol == "" {
			continue
		}

		tradeID := strings.TrimSpace(rec[3])
		// Date is embedded in trade ID: YYYYMMDD followed by sequence
		tradeDate := parseTMSTradeDate(tradeID)

		buySell := strings.TrimSpace(rec[6])
		isBuy := strings.EqualFold(buySell, "Buy")

		quantity, _ := strconv.ParseFloat(strings.TrimSpace(rec[7]), 64)
		price, _ := strconv.ParseFloat(strings.TrimSpace(rec[8]), 64)
		value, _ := strconv.ParseFloat(strings.TrimSpace(rec[9]), 64)

		trades = append(trades, TMSTrade{
			Symbol:    symbol,
			TradeID:   tradeID,
			TradeDate: tradeDate,
			IsBuy:     isBuy,
			Quantity:  quantity,
			Price:     price,
			Value:     value,
		})
	}

	return trades, nil
}

// parseTMSTradeDate extracts date from TMS trade ID (format: YYYYMMDD...)
// TMS uses Bikram Sambat (BS) dates embedded in the trade ID
// These are kept in BS format to match Meroshare exports which also use BS
func parseTMSTradeDate(tradeID string) time.Time {
	if len(tradeID) < 8 {
		return time.Time{}
	}
	dateStr := tradeID[:8]
	t, _ := time.Parse("20060102", dateStr)
	return t
}

// WAC represents a row from Meroshare's WACC Report CSV export.
type WAC struct {
	Symbol       string
	Quantity     float64
	Rate         float64 // Average cost per share
	TotalCost    float64
	LastModified time.Time
}

// ParseWACCReport parses Meroshare's WACC Report CSV from a reader.
// CSV format: S.N, Demat, Scrip Name, WACC Calculated Quantity, WACC Rate, Total Cost of Capital, Last Modification Date
func ParseWACCReport(r io.Reader) ([]WAC, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return []WAC{}, nil
	}

	wacs := make([]WAC, 0, len(records)-1)

	for i := 1; i < len(records); i++ {
		rec := records[i]
		if len(rec) < 7 {
			continue
		}

		symbol := strings.TrimSpace(rec[2])
		if symbol == "" {
			continue
		}

		quantity, _ := strconv.ParseFloat(strings.TrimSpace(rec[3]), 64)
		rate, _ := strconv.ParseFloat(strings.TrimSpace(rec[4]), 64)
		totalCost, _ := strconv.ParseFloat(strings.TrimSpace(rec[5]), 64)
		lastMod, _ := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(rec[6]))

		wacs = append(wacs, WAC{
			Symbol:       symbol,
			Quantity:     quantity,
			Rate:         rate,
			TotalCost:    totalCost,
			LastModified: lastMod,
		})
	}

	return wacs, nil
}
