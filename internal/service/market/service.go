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
}

// ----- concrete implementation -----

type quoteSource interface { // tiny, unexported
	GetAllQuotes() ([]*models.Quote, error)
}

type marketService struct {
	src quoteSource // can be scraper, cache, mock…
}

// New wires any quoteSource (start with the scraper).
func New(src quoteSource) Service {
	return &marketService{src: src}
}

func (s *marketService) GetLiveQuotes(ctx context.Context) ([]*models.Quote, error) {
	// add ctx timeouts, logging, caching, etc. here if needed
	return s.src.GetAllQuotes()
}

// NewWithShareSansar Helper for production wiring
func NewWithShareSansar() Service {
	return New(scraper.NewShareSansarScraper())
}
