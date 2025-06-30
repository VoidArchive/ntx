package parser

import (
	"fmt"
	"io"
	"log/slog"
	"ntx/internal/data/models"
	"ntx/internal/market"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NepseParser handles parsing HTML content from NEPSE website
// Extracts market data from various NEPSE pages with robust error handling
type NepseParser struct {
	logger *slog.Logger
}

// ParseConfig holds configuration for parsing operations
type ParseConfig struct {
	// Skip symbols with invalid data
	SkipInvalidSymbols bool

	// Maximum symbols to parse (0 = no limit)
	MaxSymbols int

	// Enable debug logging for parser
	DebugMode bool
}

// NewNepseParser creates a new NEPSE HTML parser
func NewNepseParser(logger *slog.Logger) *NepseParser {
	return &NepseParser{
		logger: logger,
	}
}

// ParseMarketData parses the main market data page from NEPSE
// Extracts live trading data for all listed symbols
func (np *NepseParser) ParseMarketData(htmlContent io.Reader, config *ParseConfig) ([]market.ScrapedData, error) {
	if config == nil {
		config = &ParseConfig{
			SkipInvalidSymbols: true,
			MaxSymbols:         0,
			DebugMode:          false,
		}
	}

	doc, err := goquery.NewDocumentFromReader(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var results []market.ScrapedData
	parseTime := time.Now()

	// INFO: NEPSE website structure - look for market data table
	// The exact selectors will need to be updated based on current NEPSE website structure
	tableSelector := "table.table-striped, table.market-data, .table-responsive table"

	doc.Find(tableSelector).Each(func(i int, table *goquery.Selection) {
		if config.DebugMode {
			np.logger.Debug("Found potential market data table", "index", i)
		}

		// Look for table rows with market data
		table.Find("tbody tr, tr").Each(func(j int, row *goquery.Selection) {
			// Check if we've reached the symbol limit
			if config.MaxSymbols > 0 && len(results) >= config.MaxSymbols {
				return
			}

			if data, err := np.parseTableRow(row, parseTime, config.DebugMode); err != nil {
				if config.DebugMode {
					np.logger.Debug("Failed to parse table row", "row", j, "error", fmt.Errorf("row %d parsing failed: %w", j, err))
				}
				if !config.SkipInvalidSymbols {
					np.logger.Warn("Invalid row data", "row", j, "error", fmt.Errorf("row %d validation failed: %w", j, err))
				}
			} else if data != nil {
				results = append(results, *data)
			}
		})
	})

	if len(results) == 0 {
		// Try alternative parsing strategies
		if altResults, err := np.parseAlternativeStructure(doc, config, parseTime); err == nil && len(altResults) > 0 {
			results = altResults
		}
	}

	np.logger.Info("Parsed market data",
		"symbols_found", len(results),
		"parse_time", time.Since(parseTime))

	return results, nil
}

// ParseMarketSummary parses market summary statistics from NEPSE
func (np *NepseParser) ParseMarketSummary(htmlContent io.Reader) (*market.MarketSummary, error) {
	doc, err := goquery.NewDocumentFromReader(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	summary := &market.MarketSummary{
		ScrapedAt: time.Now(),
	}

	// INFO: Parse market indices - NEPSE Index, Sensitive Index, Float Index
	np.parseIndexValues(doc, summary)

	// INFO: Parse market statistics - turnover, volume, transactions
	np.parseMarketStats(doc, summary)

	np.logger.Info("Parsed market summary",
		"nepse_index", summary.NepseIndex,
		"total_turnover", summary.TotalTurnover.FormattedString(),
		"total_volume", summary.TotalVolume)

	return summary, nil
}

// ParseSymbolDetails parses detailed information for a specific symbol
func (np *NepseParser) ParseSymbolDetails(htmlContent io.Reader, symbol string) (*market.ScrapedData, error) {
	doc, err := goquery.NewDocumentFromReader(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	parseTime := time.Now()

	// Look for symbol-specific data in various possible locations
	selectors := []string{
		fmt.Sprintf("tr:contains('%s')", symbol),
		fmt.Sprintf(".symbol-row[data-symbol='%s']", symbol),
		fmt.Sprintf("#symbol-%s", strings.ToLower(symbol)),
	}

	for _, selector := range selectors {
		if row := doc.Find(selector).First(); row.Length() > 0 {
			if data, err := np.parseTableRow(row, parseTime, true); err == nil && data != nil {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("symbol data not found: %s", symbol)
}

// Private parsing methods

// parseTableRow parses a single table row containing market data
func (np *NepseParser) parseTableRow(row *goquery.Selection, parseTime time.Time, debug bool) (*market.ScrapedData, error) {
	cells := row.Find("td, th")
	if cells.Length() < 5 {
		return nil, fmt.Errorf("insufficient columns in row: %d", cells.Length())
	}

	// Extract cell data - this mapping may need adjustment based on actual NEPSE HTML structure
	var cellData []string
	cells.Each(func(i int, cell *goquery.Selection) {
		cellData = append(cellData, strings.TrimSpace(cell.Text()))
	})

	if debug {
		np.logger.Debug("Parsing row data", "cells", cellData)
	}

	// Skip header rows or invalid data
	if len(cellData) == 0 || cellData[0] == "" || strings.Contains(cellData[0], "Symbol") {
		return nil, fmt.Errorf("header or empty row")
	}

	// Basic validation - symbol should be alphanumeric
	symbol := strings.ToUpper(strings.TrimSpace(cellData[0]))
	if len(symbol) < 2 || len(symbol) > 10 {
		return nil, fmt.Errorf("invalid symbol format: %s", symbol)
	}

	data := &market.ScrapedData{
		Symbol:    symbol,
		ScrapedAt: parseTime,
	}

	// Parse common market data fields
	// INFO: Column mapping may vary based on NEPSE table structure
	// Common columns: Symbol, LTP, Change, %Change, Volume, High, Low, Open, PrevClose
	if err := np.parseMarketDataFields(data, cellData); err != nil {
		return nil, fmt.Errorf("failed to parse market data for %s: %w", symbol, err)
	}

	return data, nil
}

// parseMarketDataFields parses individual market data fields from cell data
func (np *NepseParser) parseMarketDataFields(data *market.ScrapedData, cells []string) error {
	// INFO: This is a flexible parser that handles various NEPSE table formats
	// Column order may vary, so we use positional parsing with fallbacks

	// Assuming common NEPSE table structure:
	// [0] Symbol, [1] LTP, [2] Change, [3] %Change, [4] Volume, [5] High, [6] Low, [7] Open, [8] PrevClose

	if len(cells) >= 2 {
		// Parse Last Trading Price (LTP)
		if ltp, err := np.parseMoneyValue(cells[1]); err == nil {
			data.LastPrice = ltp
		}
	}

	if len(cells) >= 3 {
		// Parse Change Amount
		if change, err := np.parseMoneyValue(cells[2]); err == nil {
			data.ChangeAmount = change
		}
	}

	if len(cells) >= 4 {
		// Parse Change Percentage
		if changePct, err := np.parsePercentageValue(cells[3]); err == nil {
			data.ChangePercent = changePct
		}
	}

	if len(cells) >= 5 {
		// Parse Volume
		if volume, err := np.parseVolumeValue(cells[4]); err == nil {
			data.Volume = volume
		}
	}

	if len(cells) >= 6 {
		// Parse High
		if high, err := np.parseMoneyValue(cells[5]); err == nil {
			data.High = high
		}
	}

	if len(cells) >= 7 {
		// Parse Low
		if low, err := np.parseMoneyValue(cells[6]); err == nil {
			data.Low = low
		}
	}

	if len(cells) >= 8 {
		// Parse Open
		if open, err := np.parseMoneyValue(cells[7]); err == nil {
			data.Open = open
		}
	}

	if len(cells) >= 9 {
		// Parse Previous Close
		if prevClose, err := np.parseMoneyValue(cells[8]); err == nil {
			data.PrevClose = prevClose
		}
	}

	return nil
}

// parseAlternativeStructure tries alternative parsing when main table parsing fails
func (np *NepseParser) parseAlternativeStructure(doc *goquery.Document, config *ParseConfig, parseTime time.Time) ([]market.ScrapedData, error) {
	var results []market.ScrapedData

	// Try parsing div-based layouts
	doc.Find(".stock-item, .market-item, .symbol-data").Each(func(i int, item *goquery.Selection) {
		if config.MaxSymbols > 0 && len(results) >= config.MaxSymbols {
			return
		}

		if data := np.parseDivBasedData(item, parseTime); data != nil {
			results = append(results, *data)
		}
	})

	return results, nil
}

// parseDivBasedData parses market data from div-based layouts
func (np *NepseParser) parseDivBasedData(item *goquery.Selection, parseTime time.Time) *market.ScrapedData {
	// Extract symbol
	symbol := strings.TrimSpace(item.Find(".symbol, .stock-symbol").First().Text())
	if symbol == "" {
		return nil
	}

	data := &market.ScrapedData{
		Symbol:    strings.ToUpper(symbol),
		ScrapedAt: parseTime,
	}

	// Parse price fields
	if price := item.Find(".price, .ltp").First().Text(); price != "" {
		if ltp, err := np.parseMoneyValue(price); err == nil {
			data.LastPrice = ltp
		}
	}

	if change := item.Find(".change, .change-amount").First().Text(); change != "" {
		if changeAmt, err := np.parseMoneyValue(change); err == nil {
			data.ChangeAmount = changeAmt
		}
	}

	return data
}

// parseIndexValues extracts market index values from the document
func (np *NepseParser) parseIndexValues(doc *goquery.Document, summary *market.MarketSummary) {
	// INFO: Look for NEPSE Index, Sensitive Index, Float Index
	indexSelectors := map[string]*float64{
		"nepse":     &summary.NepseIndex,
		"sensitive": &summary.SensitiveIndex,
		"float":     &summary.FloatIndex,
	}

	for indexName, field := range indexSelectors {
		selectors := []string{
			fmt.Sprintf(".%s-index", indexName),
			fmt.Sprintf("#%s-index", indexName),
			fmt.Sprintf("*:contains('%s Index')", cases.Title(language.English).String(indexName)),
		}

		for _, selector := range selectors {
			if element := doc.Find(selector).First(); element.Length() > 0 {
				if value, err := np.parseFloatValue(element.Text()); err == nil {
					*field = value
					break
				}
			}
		}
	}
}

// parseMarketStats extracts market statistics from the document
func (np *NepseParser) parseMarketStats(doc *goquery.Document, summary *market.MarketSummary) {
	// Parse turnover
	if turnover := doc.Find(".turnover, #turnover, *:contains('Turnover')").First(); turnover.Length() > 0 {
		if value, err := np.parseMoneyValue(turnover.Text()); err == nil {
			summary.TotalTurnover = value
		}
	}

	// Parse volume
	if volume := doc.Find(".volume, #volume, *:contains('Volume')").First(); volume.Length() > 0 {
		if value, err := np.parseVolumeValue(volume.Text()); err == nil {
			summary.TotalVolume = value
		}
	}

	// Parse transactions
	if transactions := doc.Find(".transactions, #transactions, *:contains('Transactions')").First(); transactions.Length() > 0 {
		if value, err := np.parseVolumeValue(transactions.Text()); err == nil {
			summary.TotalTransactions = value
		}
	}
}

// Helper methods for parsing specific data types

// parseMoneyValue parses a money value from string
func (np *NepseParser) parseMoneyValue(text string) (models.Money, error) {
	// Clean the text
	cleaned := strings.TrimSpace(text)
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "Rs", "")
	cleaned = strings.ReplaceAll(cleaned, "Rs.", "")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" || cleaned == "-" {
		return models.Money(0), nil
	}

	// Handle negative values
	negative := strings.HasPrefix(cleaned, "-")
	if negative {
		cleaned = strings.TrimPrefix(cleaned, "-")
	}

	value, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return models.Money(0), fmt.Errorf("invalid money value: %s", text)
	}

	money := models.NewMoneyFromRupees(value)
	if negative {
		money = -money
	}

	return money, nil
}

// parsePercentageValue parses a percentage value from string
func (np *NepseParser) parsePercentageValue(text string) (float64, error) {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.ReplaceAll(cleaned, "%", "")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" || cleaned == "-" {
		return 0, nil
	}

	return strconv.ParseFloat(cleaned, 64)
}

// parseVolumeValue parses a volume value from string
func (np *NepseParser) parseVolumeValue(text string) (int64, error) {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" || cleaned == "-" {
		return 0, nil
	}

	return strconv.ParseInt(cleaned, 10, 64)
}

// parseFloatValue parses a float value from string
func (np *NepseParser) parseFloatValue(text string) (float64, error) {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" || cleaned == "-" {
		return 0, nil
	}

	return strconv.ParseFloat(cleaned, 64)
}
