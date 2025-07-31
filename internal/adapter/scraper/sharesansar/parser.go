package sharesansar

import (
	"fmt"
	"strings"
	"time"

	"github.com/VoidArchive/ntx/internal/domain/models"
	"github.com/gocolly/colly/v2"
	"github.com/shopspring/decimal"
)

// Parser handles HTML parsing for ShareSansar
type Parser struct{}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{}
}

// ParseQuoteRow extracts quote data from a table row
func (p *Parser) ParseQuoteRow(e *colly.HTMLElement) (*models.Quote, error) {
	cells := e.ChildTexts("td")

	// Skip header rows or empty rows
	if len(cells) < 10 {
		return nil, fmt.Errorf("insufficient columns: got %d, need at least 10", len(cells))
	}

	// Extract symbol - it's in the second column (index 1)
	symbol := strings.TrimSpace(e.ChildText("td:nth-child(2) a"))
	if symbol == "" {
		symbol = strings.TrimSpace(cells[ColSymbol])
	}

	// Skip if no symbol or if it's a header row
	if symbol == "" || symbol == "Symbol" {
		return nil, fmt.Errorf("no symbol found or header row")
	}

	// Parse all numeric fields according to actual column positions
	ltp, err := p.parseDecimal(cells[ColLTP])
	if err != nil {
		return nil, fmt.Errorf("invalid LTP for %s: %w", symbol, err)
	}

	change, _ := p.parseDecimal(cells[ColChange])
	changePct, _ := p.parseDecimal(strings.TrimSuffix(cells[ColChangePct], "%"))
	open, _ := p.parseDecimal(cells[ColOpen])
	high, _ := p.parseDecimal(cells[ColHigh])
	low, _ := p.parseDecimal(cells[ColLow])
	volume := p.parseInt64(cells[ColVolume])

	// Previous close for calculating if market just opened
	previous, _ := p.parseDecimal(cells[ColPrevious])

	// Turnover might be in column 10 if present
	var turnover decimal.Decimal
	if len(cells) > ColTurnover {
		turnover, _ = p.parseDecimal(cells[ColTurnover])
	}

	// Basic validation
	if high.LessThan(low) && !high.IsZero() && !low.IsZero() {
		return nil, fmt.Errorf("high < low for %s: %s < %s", symbol, high, low)
	}

	// If open is 0, market might not have opened yet, use previous close
	if open.IsZero() && !previous.IsZero() {
		open = previous
	}

	quote := &models.Quote{
		Symbol:    symbol,
		Price:     ltp,
		Change:    change,
		ChangePct: changePct,
		High:      high,
		Low:       low,
		Open:      open,
		Volume:    volume,
		Turnover:  turnover,
		Timestamp: time.Now(),
		Stale:     false,
	}

	return quote, nil
}

// ParseCompanyDetails extracts company information
func (p *Parser) ParseCompanyDetails(doc *colly.HTMLElement) map[string]string {
	details := make(map[string]string)

	// Parse each row in company details table
	doc.ForEach(SelectorCompanyTable, func(_ int, e *colly.HTMLElement) {
		key := p.cleanText(e.ChildText("td:nth-child(1)"))
		value := p.cleanText(e.ChildText("td:nth-child(2)"))

		if key != "" && value != "" {
			// Map to normalized field names
			if fieldName, ok := CompanyDetailFields[key]; ok {
				details[fieldName] = value
			} else {
				// Keep unmapped fields with cleaned key
				details[p.normalizeKey(key)] = value
			}
		}
	})

	return details
}

// ParseFloorsheetRow extracts transaction data from floorsheet
func (p *Parser) ParseFloorsheetRow(e *colly.HTMLElement) (*FloorsheetEntry, error) {
	cells := e.ChildTexts("td")

	if len(cells) < 8 {
		return nil, fmt.Errorf("insufficient floorsheet columns")
	}

	// Extract transaction details
	entry := &FloorsheetEntry{
		TransactionNo: p.cleanText(cells[0]),
		Symbol:        p.cleanText(cells[1]),
		BuyerBroker:   p.parseInt(cells[2]),
		SellerBroker:  p.parseInt(cells[3]),
		Quantity:      p.parseInt64(cells[4]),
		Rate:          p.mustParseDecimal(cells[5]),
		Amount:        p.mustParseDecimal(cells[6]),
		Timestamp:     p.parseTime(cells[7]),
	}

	return entry, nil
}

// Helper methods

func (p *Parser) parseDecimal(s string) (decimal.Decimal, error) {
	s = p.cleanNumeric(s)
	if s == "" || s == "-" || s == "N/A" {
		return decimal.Zero, fmt.Errorf("empty or invalid value")
	}
	return decimal.NewFromString(s)
}

func (p *Parser) mustParseDecimal(s string) decimal.Decimal {
	d, _ := p.parseDecimal(s)
	return d
}

func (p *Parser) parseInt64(s string) int64 {
	s = p.cleanNumeric(s)
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

func (p *Parser) parseInt(s string) int {
	return int(p.parseInt64(s))
}

func (p *Parser) cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func (p *Parser) cleanNumeric(s string) string {
	s = p.cleanText(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "Rs.", "")
	s = strings.ReplaceAll(s, "Rs", "")
	s = strings.TrimSpace(s)
	return s
}

func (p *Parser) normalizeKey(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

func (p *Parser) parseTime(s string) time.Time {
	// ShareSansar time format: "HH:MM:SS"
	t, err := time.Parse("15:04:05", p.cleanText(s))
	if err != nil {
		return time.Now()
	}
	// Set to today's date
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
}

// FloorsheetEntry represents a single transaction
type FloorsheetEntry struct {
	TransactionNo string
	Symbol        string
	BuyerBroker   int
	SellerBroker  int
	Quantity      int64
	Rate          decimal.Decimal
	Amount        decimal.Decimal
	Timestamp     time.Time
}
