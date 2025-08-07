// Package scraper scrapes nepse data from sharesansar live trading view
package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/voidarchive/ntx/internal/domain/models"
)

// ShareSansarScraper implements stock data scraping from ShareSansar
type ShareSansarScraper struct {
	collector *colly.Collector
}

// NewShareSansarScraper creates a new ShareSansar scraper
func NewShareSansarScraper() *ShareSansarScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("www.sharesansar.com"),
		colly.AllowURLRevisit(),
	)

	c.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"

	// Be nice to the server
	c.OnRequest(func(r *colly.Request) {
		time.Sleep(500 * time.Millisecond)
	})

	return &ShareSansarScraper{
		collector: c,
	}
}

// GetQuote scrapes a single stock quote (inefficient - prefer GetAllQuotes)
func (s *ShareSansarScraper) GetQuote(symbol string) (*models.Quote, error) {
	// For single quotes, we'll scrape all and filter
	// This is more efficient than making separate requests
	allQuotes, err := s.GetAllQuotes()
	if err != nil {
		return nil, err
	}

	for _, quote := range allQuotes {
		if quote.Symbol == symbol {
			return quote, nil
		}
	}

	return nil, fmt.Errorf("symbol %s not found", symbol)
}

// GetAllQuotes scrapes all available stock quotes
func (s *ShareSansarScraper) GetAllQuotes() ([]*models.Quote, error) {
	var quotes []*models.Quote

	s.collector = s.collector.Clone()

	// Clear visited URLs to allow re-scraping
	s.collector.OnHTML("table tr", func(e *colly.HTMLElement) {
		if e.ChildText("td:nth-child(2)") == "" {
			return
		}

		cellSymbol := strings.TrimSpace(e.ChildText("td:nth-child(2)"))
		if cellSymbol == "" {
			return
		}

		quote := &models.Quote{Symbol: cellSymbol}

		if ltp, err := parseFloat(e.ChildText("td:nth-child(3)")); err == nil {
			quote.LTP = ltp
		}
		if open, err := parseFloat(e.ChildText("td:nth-child(6)")); err == nil {
			quote.Open = open
		}
		if high, err := parseFloat(e.ChildText("td:nth-child(7)")); err == nil {
			quote.High = high
		}
		if low, err := parseFloat(e.ChildText("td:nth-child(8)")); err == nil {
			quote.Low = low
		}
		if volume, err := parseFloat(e.ChildText("td:nth-child(9)")); err == nil {
			quote.Volume = volume
		}
		if prevClose, err := parseFloat(e.ChildText("td:nth-child(10)")); err == nil {
			quote.PrevClose = prevClose
		}

		quotes = append(quotes, quote)
	})

	err := s.collector.Visit("https://www.sharesansar.com/live-trading")
	if err != nil {
		return nil, fmt.Errorf("failed to visit page: %w", err)
	}

	return quotes, nil
}

func parseFloat(s string) (float64, error) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	if cleaned == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseFloat(cleaned, 64)
}
