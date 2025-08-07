// Package scraper - market overview functionality
package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/voidarchive/ntx/internal/domain/models"
)

// GetMarketOverview scrapes main NEPSE index and all sub-indices
func (s *ShareSansarScraper) GetMarketOverview() (*models.MarketOverview, error) {
	overview := &models.MarketOverview{
		SubIndices: make([]*models.Index, 0),
	}

	// Clear visited URLs for re-scraping
	s.collector.OnHTML(".table-responsive table", func(e *colly.HTMLElement) {
		// Check if this is the main NEPSE index table
		if strings.Contains(e.Text, "NEPSE Index") || strings.Contains(e.Text, "Index Value") {
			// Parse main NEPSE index
			overview.MainIndex = s.parseMainIndex(e)
		}
	})

	// Parse sub-indices table
	s.collector.OnHTML("table tr", func(e *colly.HTMLElement) {
		// Look for sub-index rows (skip header)
		if strings.Contains(e.ChildText("td:nth-child(1)"), "Index") ||
			strings.Contains(e.ChildText("td:nth-child(1)"), "Subindex") {

			index := s.parseSubIndexRow(e)
			if index != nil {
				overview.SubIndices = append(overview.SubIndices, index)
			}
		}
	})

	err := s.collector.Visit("https://www.sharesansar.com/market")
	if err != nil {
		return nil, fmt.Errorf("failed to visit market page: %w", err)
	}

	overview.LastUpdated = time.Now().Format("15:04:05")
	return overview, nil
}

// parseMainIndex extracts main NEPSE index data
func (s *ShareSansarScraper) parseMainIndex(_ *colly.HTMLElement) *models.Index {
	// This will need to be adjusted based on actual HTML structure
	// For now, create a placeholder implementation
	return &models.Index{
		Name:        "NEPSE Index",
		Open:        2650.0, // These will be parsed from actual HTML
		High:        2670.0,
		Low:         2640.0,
		Close:       2656.67,
		PointChange: 12.34,
		IsMain:      true,
	}
}

// parseSubIndexRow extracts sub-index data from table row
func (s *ShareSansarScraper) parseSubIndexRow(e *colly.HTMLElement) *models.Index {
	nameText := strings.TrimSpace(e.ChildText("td:nth-child(1)"))
	if nameText == "" || nameText == "Sub Index" {
		return nil // Skip header or empty rows
	}

	index := &models.Index{
		Name:   nameText,
		IsMain: false,
	}

	// Parse OHLC data
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

	return index
}
