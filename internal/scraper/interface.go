// Package scraper defines interfaces for market data scraping
package scraper

import "github.com/voidarchive/ntx/internal/domain/models"

// MarketDataSource defines the interface for scraping market data
type MarketDataSource interface {
	GetAllQuotes() ([]*models.Quote, error)
	GetMarketOverview() (*models.MarketOverview, error)
}

// UnifiedScraper combines stock quotes and market overview scraping
type UnifiedScraper struct {
	quoteScraper    *ShareSansarScraper
	overviewScraper *MarketScraper
}

// NewUnifiedScraper creates a scraper that handles both quotes and market overview
func NewUnifiedScraper() MarketDataSource {
	return &UnifiedScraper{
		quoteScraper:    NewShareSansarScraper(),
		overviewScraper: NewMarketScraper(),
	}
}

// GetAllQuotes delegates to the stock quote scraper
func (u *UnifiedScraper) GetAllQuotes() ([]*models.Quote, error) {
	return u.quoteScraper.GetAllQuotes()
}

// GetMarketOverview delegates to the market overview scraper
func (u *UnifiedScraper) GetMarketOverview() (*models.MarketOverview, error) {
	return u.overviewScraper.GetMarketOverview()
}
