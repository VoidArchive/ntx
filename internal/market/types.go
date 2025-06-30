package market

import (
	"context"
	"time"

	"ntx/internal/data/models"
)

// ScrapedData represents raw data scraped from NEPSE
type ScrapedData struct {
	Symbol        string       `json:"symbol"`
	LastPrice     models.Money `json:"last_price"`
	ChangeAmount  models.Money `json:"change_amount"`
	ChangePercent float64      `json:"change_percent"`
	Volume        int64        `json:"volume"`
	High          models.Money `json:"high"`
	Low           models.Money `json:"low"`
	Open          models.Money `json:"open"`
	PrevClose     models.Money `json:"prev_close"`
	ScrapedAt     time.Time    `json:"scraped_at"`
}

// MarketSummary represents overall market statistics
type MarketSummary struct {
	TotalTurnover        models.Money `json:"total_turnover"`
	TotalVolume          int64        `json:"total_volume"`
	TotalTransactions    int64        `json:"total_transactions"`
	MarketCapitalization models.Money `json:"market_capitalization"`
	NepseIndex           float64      `json:"nepse_index"`
	SensitiveIndex       float64      `json:"sensitive_index"`
	FloatIndex           float64      `json:"float_index"`
	ScrapedAt            time.Time    `json:"scraped_at"`
}

// Scraper defines the interface for market data scrapers
type Scraper interface {
	// ScrapeAllMarketData retrieves comprehensive market data for all symbols
	ScrapeAllMarketData(ctx context.Context) ([]ScrapedData, error)
	
	// ScrapeSymbol retrieves market data for a specific symbol
	ScrapeSymbol(ctx context.Context, symbol string) (*ScrapedData, error)
	
	// GetHealthStatus checks if the data source is accessible and healthy
	GetHealthStatus(ctx context.Context) error
	
	// Close cleans up scraper resources
	Close() error
}
