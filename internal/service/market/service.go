// Package market
package market

import (
	"context"

	"github.com/voidarchive/ntx/internal/domain/models"
	"github.com/voidarchive/ntx/internal/scraper"
)

// Service is what the TUI depends on.
type Service interface {
	GetLiveQuotes(ctx context.Context) ([]*models.Quote, error)
	GetMarketOverview(ctx context.Context) (*models.MarketOverview, error)
}

// ----- concrete implementation -----

type marketService struct {
	src scraper.MarketDataSource // unified interface
}

// New wires any MarketDataSource.
func New(src scraper.MarketDataSource) Service {
	return &marketService{src: src}
}

func (s *marketService) GetLiveQuotes(ctx context.Context) ([]*models.Quote, error) {
	// add ctx timeouts, logging, caching, etc. here if needed
	return s.src.GetAllQuotes()
}

func (s *marketService) GetMarketOverview(ctx context.Context) (*models.MarketOverview, error) {
	// add ctx timeouts, logging, caching, etc. here if needed
	return s.src.GetMarketOverview()
}

// NewWithShareSansar Helper for production wiring
func NewWithShareSansar() Service {
	return New(scraper.NewUnifiedScraper())
}
