// Package scraper - market overview functionality
package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/voidarchive/ntx/internal/domain/models"
)

// MarketScraper handles scraping market overview data from ShareSansar
type MarketScraper struct {
	collector *colly.Collector
}

// NewMarketScraper creates a new market overview scraper
func NewMarketScraper() *MarketScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("www.sharesansar.com"),
	)

	c.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"

	// Be nice to the server
	c.OnRequest(func(r *colly.Request) {
		time.Sleep(500 * time.Millisecond)
	})

	return &MarketScraper{
		collector: c,
	}
}

// GetMarketOverview scrapes main NEPSE index and all sub-indices
func (s *MarketScraper) GetMarketOverview() (*models.MarketOverview, error) {
	overview := &models.MarketOverview{
		SubIndices: make([]*models.Index, 0),
	}

	// Parse all table rows - both main index and sub-indices
	s.collector.OnHTML("table tr", func(e *colly.HTMLElement) {
		nameText := strings.TrimSpace(e.ChildText("td:nth-child(1)"))
		if nameText == "" {
			return // Skip empty rows
		}

		// Parse the row into an index
		index := s.parseIndexRow(e)
		if index == nil {
			return
		}

		// Determine if this is main NEPSE index or sub-index
		if nameText == "NEPSE Index" {
			index.IsMain = true
			overview.MainIndex = index
		} else if strings.Contains(nameText, "Index") || strings.Contains(nameText, "SubIndex") {
			index.IsMain = false
			overview.SubIndices = append(overview.SubIndices, index)
		}
	})

	err := s.collector.Visit("https://www.sharesansar.com/market")
	if err != nil {
		return nil, fmt.Errorf("failed to visit market page: %w", err)
	}

	overview.LastUpdated = time.Now().Format("15:04:05")
	return overview, nil
}

// parseIndexRow extracts index data from table row (works for both main and sub-indices)
func (s *MarketScraper) parseIndexRow(e *colly.HTMLElement) *models.Index {
	nameText := strings.TrimSpace(e.ChildText("td:nth-child(1)"))
	if nameText == "" {
		return nil // Skip empty rows
	}

	index := &models.Index{
		Name: nameText,
	}

	// Parse OHLC data based on the HTML structure you provided
	if open, err := parseFloat(e.ChildText("td:nth-child(2)")); err == nil {
		index.Open = open
	}
	if high, err := parseFloat(e.ChildText("td:nth-child(3)")); err == nil {
		index.High = high
	}
	if low, err := parseFloat(e.ChildText("td:nth-child(4)")); err == nil {
		index.Low = low
	}
	if close, err := parseFloat(e.ChildText("td:nth-child(5)")); err == nil {
		index.Close = close
	}
	if pointChange, err := parseFloat(e.ChildText("td:nth-child(6)")); err == nil {
		index.PointChange = pointChange
	}

	// Only return if we have the essential data
	if index.Close == 0 {
		return nil
	}

	return index
}
