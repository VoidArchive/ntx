package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"ntx/internal/data/models"
	"ntx/internal/market"
	"time"
)

// NepseBotScraper provides fallback data source using nepse.bot API
// Serves as a reliable alternative when direct NEPSE scraping fails
type NepseBotScraper struct {
	client  *http.Client
	config  *NepseBotConfig
	logger  *slog.Logger
	limiter *RateLimiter
}

// NepseBotConfig holds configuration for nepse.bot API integration
type NepseBotConfig struct {
	// Base URL for nepse.bot API (default: https://nepse.bot/api)
	BaseURL string

	// API version (default: v1)
	APIVersion string

	// Request timeout (default: 15 seconds)
	Timeout time.Duration

	// Rate limiting (default: 1 request per 2 seconds)
	RequestDelay time.Duration

	// Maximum retries for failed requests
	MaxRetries int

	// Enable API key authentication (if available)
	APIKey string

	// User agent for API requests
	UserAgent string

	// Enable request/response logging
	LogRequests bool
}

// NepseBotResponse represents the API response structure
type NepseBotResponse struct {
	Success bool                 `json:"success"`
	Data    []NepseBotMarketData `json:"data"`
	Message string               `json:"message"`
	Count   int                  `json:"count"`
	Error   string               `json:"error,omitempty"`
}

// NepseBotMarketData represents market data from nepse.bot API
type NepseBotMarketData struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`     // Company name field
	LastPrice float64 `json:"ltp"`      // Last traded price
	Turnover  float64 `json:"turnover"` // Turnover value

	// Optional fields that might be available in other endpoints
	ChangeAmount  float64 `json:"change,omitempty"`
	ChangePercent float64 `json:"change_percent,omitempty"`
	Volume        int64   `json:"volume,omitempty"`
	High          float64 `json:"high,omitempty"`
	Low           float64 `json:"low,omitempty"`
	Open          float64 `json:"open,omitempty"`
	PrevClose     float64 `json:"prev_close,omitempty"`
	Timestamp     string  `json:"timestamp,omitempty"`
}

// NepseBotIndexData represents index information
type NepseBotIndexData struct {
	ID               int     `json:"id"`
	Index            string  `json:"index"`      // Index name
	Point            float64 `json:"point"`      // Current index value
	Close            float64 `json:"close"`      // Closing value
	High             float64 `json:"high"`       // Daily high
	Low              float64 `json:"low"`        // Daily low
	Difference       float64 `json:"difference"` // Points change
	PreviousClose    float64 `json:"previous_close"`
	PercentChange    float64 `json:"percent_change"`
	FiftyTwoWeekHigh float64 `json:"fifty_two_week_high"`
	FiftyTwoWeekLow  float64 `json:"fifty_two_week_low"`
}

// DefaultNepseBotConfig returns default configuration for nepse.bot API
func DefaultNepseBotConfig() *NepseBotConfig {
	return &NepseBotConfig{
		BaseURL:      "https://data.nepse.bot",
		APIVersion:   "", // No API version path needed
		Timeout:      15 * time.Second,
		RequestDelay: 2 * time.Second,
		MaxRetries:   3,
		UserAgent:    "NTX-Terminal/1.0 (+https://github.com/anish/ntx)",
		LogRequests:  false,
	}
}

// NewNepseBotScraper creates a new nepse.bot API scraper
func NewNepseBotScraper(config *NepseBotConfig, logger *slog.Logger) (*NepseBotScraper, error) {
	if config == nil {
		config = DefaultNepseBotConfig()
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        5,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	// Create rate limiter
	limiter := &RateLimiter{
		delay: config.RequestDelay,
	}

	scraper := &NepseBotScraper{
		client:  client,
		config:  config,
		logger:  logger,
		limiter: limiter,
	}

	logger.Info("NepseBot scraper initialized",
		"base_url", config.BaseURL,
		"api_version", config.APIVersion,
		"request_delay", config.RequestDelay)

	return scraper, nil
}

// ScrapeAllMarketData fetches all market data from nepse.bot API
func (nbs *NepseBotScraper) ScrapeAllMarketData(ctx context.Context) ([]market.ScrapedData, error) {
	nbs.logger.Info("Starting nepse.bot market data fetch")

	// Use top-turnover endpoint which provides comprehensive market data with prices
	url := fmt.Sprintf("%s/top-turnover", nbs.config.BaseURL)

	// Make HTTP request directly since this endpoint doesn't follow the standard API response format
	resp, err := nbs.makeHTTPRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market data from nepse.bot: %w", err)
	}
	defer resp.Body.Close()

	// Parse JSON response directly
	var apiData []NepseBotMarketData
	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return nil, fmt.Errorf("failed to parse market data response: %w", err)
	}

	// Convert API response to internal format
	scrapedData := make([]market.ScrapedData, 0, len(apiData))
	for _, data := range apiData {
		if converted, err := nbs.convertAPIData(data); err == nil {
			scrapedData = append(scrapedData, *converted)
		} else {
			nbs.logger.Warn("Failed to convert API data",
				"symbol", data.Symbol, "error", err)
		}
	}

	nbs.logger.Info("NepseBot market data fetch completed",
		"symbols_found", len(scrapedData))

	return scrapedData, nil
}

// ScrapeSymbol fetches data for a specific symbol from nepse.bot API
func (nbs *NepseBotScraper) ScrapeSymbol(ctx context.Context, symbol string) (*market.ScrapedData, error) {
	nbs.logger.Info("Fetching symbol data from nepse.bot", "symbol", symbol)

	// For now, use the LiveMarket endpoint and filter for the specific symbol
	// TODO: Check if the API has a symbol-specific endpoint
	allData, err := nbs.ScrapeAllMarketData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market data: %w", err)
	}

	// Find the specific symbol
	for _, data := range allData {
		if data.Symbol == symbol {
			return &data, nil
		}
	}

	return nil, fmt.Errorf("no data found for symbol %s", symbol)
}

// ScrapeMarketIndices fetches market indices from nepse.bot API
func (nbs *NepseBotScraper) ScrapeMarketIndices(ctx context.Context) (map[string]float64, error) {
	nbs.logger.Info("Fetching market indices from nepse.bot")

	// Use /index endpoint which provides comprehensive index data
	url := fmt.Sprintf("%s/index", nbs.config.BaseURL)

	// Make HTTP request
	resp, err := nbs.makeHTTPRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch indices: %w", err)
	}
	defer resp.Body.Close()

	// Parse indices response
	var indices []NepseBotIndexData
	if err := json.NewDecoder(resp.Body).Decode(&indices); err != nil {
		return nil, fmt.Errorf("failed to parse indices response: %w", err)
	}

	// Convert to map
	result := make(map[string]float64)
	for _, index := range indices {
		result[index.Index] = index.Point
	}

	nbs.logger.Info("Market indices fetched", "indices_count", len(result))
	return result, nil
}

// GetHealthStatus checks if nepse.bot API is accessible and healthy
func (nbs *NepseBotScraper) GetHealthStatus(ctx context.Context) error {
	nbs.logger.Debug("Checking nepse.bot API health")

	// Try /summary endpoint as health check
	url := fmt.Sprintf("%s/summary", nbs.config.BaseURL)

	resp, err := nbs.makeHTTPRequest(ctx, url)
	if err != nil {
		return fmt.Errorf("nepse.bot API health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("nepse.bot API returned status: %d", resp.StatusCode)
	}

	nbs.logger.Debug("NepseBot API health check passed")
	return nil
}

// Close cleans up scraper resources
func (nbs *NepseBotScraper) Close() error {
	nbs.logger.Info("Closing nepse.bot scraper")
	nbs.client.CloseIdleConnections()
	return nil
}

// Private methods

// makeAPIRequest makes a request to nepse.bot API and parses the standard response format
func (nbs *NepseBotScraper) makeAPIRequest(ctx context.Context, url string) (*NepseBotResponse, error) {
	// Make HTTP request
	resp, err := nbs.makeHTTPRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse JSON response
	var apiResponse NepseBotResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check API response success
	if !apiResponse.Success {
		return nil, fmt.Errorf("API request failed: %s", apiResponse.Error)
	}

	if nbs.config.LogRequests {
		nbs.logger.Debug("API request successful",
			"url", url,
			"data_count", apiResponse.Count)
	}

	return &apiResponse, nil
}

// makeHTTPRequest creates and executes HTTP request with proper headers and rate limiting
func (nbs *NepseBotScraper) makeHTTPRequest(ctx context.Context, url string) (*http.Response, error) {
	// Rate limiting
	if err := nbs.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter cancelled: %w", err)
	}

	// Log request if enabled
	if nbs.config.LogRequests {
		nbs.logger.Debug("Making API request", "url", url)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", nbs.config.UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Add API key if configured
	if nbs.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+nbs.config.APIKey)
	}

	// Execute request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= nbs.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(time.Duration(attempt) * time.Second):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			nbs.logger.Warn("Retrying API request", "attempt", attempt, "url", url)
		}

		resp, lastErr = nbs.client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if lastErr != nil {
			nbs.logger.Warn("API request failed", "attempt", attempt, "error", lastErr)
		} else {
			nbs.logger.Warn("API request returned non-OK status",
				"attempt", attempt, "status", resp.StatusCode)
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("API request failed after %d attempts: %w",
			nbs.config.MaxRetries+1, lastErr)
	}

	return resp, nil
}

// convertAPIData converts nepse.bot API data to internal ScrapedData format
func (nbs *NepseBotScraper) convertAPIData(apiData NepseBotMarketData) (*market.ScrapedData, error) {
	// Parse timestamp if available, otherwise use current time
	var scrapedAt time.Time
	if apiData.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, apiData.Timestamp); err == nil {
			scrapedAt = parsed
		} else {
			// Try alternative timestamp formats
			if parsed, err := time.Parse("2006-01-02 15:04:05", apiData.Timestamp); err == nil {
				scrapedAt = parsed
			} else {
				nbs.logger.Warn("Failed to parse timestamp",
					"timestamp", apiData.Timestamp, "error", err)
				scrapedAt = time.Now()
			}
		}
	} else {
		scrapedAt = time.Now()
	}

	// Convert to internal format with available data
	scrapedData := &market.ScrapedData{
		Symbol:        apiData.Symbol,
		LastPrice:     models.NewMoneyFromRupees(apiData.LastPrice),
		ChangeAmount:  models.NewMoneyFromRupees(apiData.ChangeAmount),
		ChangePercent: apiData.ChangePercent,
		Volume:        apiData.Volume,
		High:          models.NewMoneyFromRupees(apiData.High),
		Low:           models.NewMoneyFromRupees(apiData.Low),
		Open:          models.NewMoneyFromRupees(apiData.Open),
		PrevClose:     models.NewMoneyFromRupees(apiData.PrevClose),
		ScrapedAt:     scrapedAt,
	}

	// If we don't have previous close data, use last price as a fallback for calculations
	if apiData.PrevClose == 0 && apiData.LastPrice > 0 {
		scrapedData.PrevClose = models.NewMoneyFromRupees(apiData.LastPrice)
	}

	return scrapedData, nil
}

// GetDataSourceInfo returns information about the nepse.bot data source
func (nbs *NepseBotScraper) GetDataSourceInfo() map[string]any {
	return map[string]any{
		"source":      "data.nepse.bot",
		"type":        "api",
		"base_url":    nbs.config.BaseURL,
		"api_version": nbs.config.APIVersion,
		"rate_limit":  nbs.config.RequestDelay,
		"reliable":    true,
		"features": []string{
			"market_data",
			"symbol_specific",
			"market_indices",
			"top_performers",
			"market_summary",
		},
	}
}
