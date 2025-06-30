package scraper

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"ntx/internal/market"
	"ntx/internal/market/parser"
)

// NepseScraper handles web scraping from NEPSE website with rate limiting
// Implements respectful scraping practices with proper delays and error handling
type NepseScraper struct {
	client    *http.Client
	config    *Config
	logger    *slog.Logger
	limiter   *RateLimiter
	parser    *parser.NepseParser
	mu        sync.RWMutex
	lastReq   time.Time
	userAgent string
}

// Config holds scraper configuration
type Config struct {
	// Base URL for NEPSE data (default: https://www.nepalstock.com)
	BaseURL string

	// Request delay between calls for respectful scraping (default: 3 seconds)
	RequestDelay time.Duration

	// HTTP client timeout (default: 10 seconds)
	Timeout time.Duration

	// Maximum retries for failed requests (default: 3)
	MaxRetries int

	// User agent string for identification
	UserAgent string

	// Enable request logging for debugging
	LogRequests bool
}

// RateLimiter implements respectful request rate limiting
type RateLimiter struct {
	lastRequest time.Time
	delay       time.Duration
	mu          sync.Mutex
}

// DefaultConfig returns default scraper configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:      "https://www.nepalstock.com",
		RequestDelay: 3 * time.Second,
		Timeout:      10 * time.Second,
		MaxRetries:   3,
		UserAgent:    "NTX-Terminal/1.0 (+https://github.com/anish/ntx)",
		LogRequests:  false,
	}
}

// NewNepseScraper creates a new NEPSE scraper with rate limiting
func NewNepseScraper(config *Config, logger *slog.Logger) (*NepseScraper, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create HTTP client with timeout and SSL configuration
	client := &http.Client{
		Timeout: config.Timeout,
		// INFO: Secure HTTPS transport with proper TLS configuration
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
			TLSHandshakeTimeout: 10 * time.Second,
			// Secure TLS configuration - never skip certificate verification
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,           // Always verify certificates for security
				MinVersion:         tls.VersionTLS12, // Enforce minimum TLS 1.2
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				},
			},
		},
	}

	limiter := &RateLimiter{
		delay: config.RequestDelay,
	}

	scraper := &NepseScraper{
		client:    client,
		config:    config,
		logger:    logger,
		limiter:   limiter,
		parser:    parser.NewNepseParser(logger),
		userAgent: config.UserAgent,
	}

	logger.Info("NEPSE scraper initialized",
		"base_url", config.BaseURL,
		"request_delay", config.RequestDelay,
		"timeout", config.Timeout,
		"max_retries", config.MaxRetries)

	return scraper, nil
}

// ScrapeAllMarketData scrapes comprehensive market data from NEPSE
func (ns *NepseScraper) ScrapeAllMarketData(ctx context.Context) ([]market.ScrapedData, error) {
	ns.logger.Info("Starting comprehensive market data scrape")

	// Construct market data URL
	marketDataURL := fmt.Sprintf("%s/market", ns.config.BaseURL)

	// Make HTTP request
	resp, err := ns.makeRequest(ctx, marketDataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market data: %w", err)
	}
	defer resp.Body.Close()

	// Parse HTML content
	parseConfig := &parser.ParseConfig{
		SkipInvalidSymbols: true,
		MaxSymbols:         0, // No limit
		DebugMode:          ns.config.LogRequests,
	}

	scrapedData, err := ns.parser.ParseMarketData(resp.Body, parseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse market data: %w", err)
	}

	ns.logger.Info("Market data scrape completed", "symbols_found", len(scrapedData))
	return scrapedData, nil
}

// ScrapeSymbol scrapes data for a specific symbol
func (ns *NepseScraper) ScrapeSymbol(ctx context.Context, symbol string) (*market.ScrapedData, error) {
	ns.logger.Info("Scraping specific symbol", "symbol", symbol)

	// Construct symbol-specific URL
	symbolURL := fmt.Sprintf("%s/company/%s", ns.config.BaseURL, symbol)

	// Make HTTP request
	resp, err := ns.makeRequest(ctx, symbolURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch symbol data: %w", err)
	}
	defer resp.Body.Close()

	// Parse HTML content for specific symbol
	scrapedData, err := ns.parser.ParseSymbolDetails(resp.Body, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to parse symbol data: %w", err)
	}

	return scrapedData, nil
}

// ScrapeMarketSummary scrapes overall market statistics
func (ns *NepseScraper) ScrapeMarketSummary(ctx context.Context) (*market.MarketSummary, error) {
	ns.logger.Info("Scraping market summary statistics")

	// Construct market summary URL
	summaryURL := fmt.Sprintf("%s/market-summary", ns.config.BaseURL)

	// Make HTTP request
	resp, err := ns.makeRequest(ctx, summaryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market summary: %w", err)
	}
	defer resp.Body.Close()

	// Parse HTML content for market summary
	summary, err := ns.parser.ParseMarketSummary(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse market summary: %w", err)
	}

	return summary, nil
}

// GetHealthStatus checks if NEPSE website is accessible
func (ns *NepseScraper) GetHealthStatus(ctx context.Context) error {
	ns.logger.Debug("Checking NEPSE website health")

	// Create a simple health check request
	req, err := http.NewRequestWithContext(ctx, "GET", ns.config.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Set proper headers
	req.Header.Set("User-Agent", ns.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// Enforce rate limiting
	if err := ns.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter cancelled: %w", err)
	}

	resp, err := ns.client.Do(req)
	if err != nil {
		return fmt.Errorf("NEPSE website health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NEPSE website returned status: %d", resp.StatusCode)
	}

	ns.logger.Debug("NEPSE website health check passed")
	return nil
}

// Close cleans up scraper resources
func (ns *NepseScraper) Close() error {
	ns.logger.Info("Closing NEPSE scraper")

	// Close HTTP client connections
	ns.client.CloseIdleConnections()

	return nil
}

// RateLimiter methods

// Wait enforces rate limiting between requests
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if !rl.lastRequest.IsZero() {
		elapsed := time.Since(rl.lastRequest)
		if elapsed < rl.delay {
			waitTime := rl.delay - elapsed

			select {
			case <-time.After(waitTime):
				// Wait completed
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	rl.lastRequest = time.Now()
	return nil
}

// SetDelay updates the rate limit delay
func (rl *RateLimiter) SetDelay(delay time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.delay = delay
}

// GetDelay returns the current rate limit delay
func (rl *RateLimiter) GetDelay() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.delay
}

// Helper methods for making HTTP requests (to be used in actual scraping implementation)

// makeRequest creates and executes an HTTP request with proper headers and rate limiting
func (ns *NepseScraper) makeRequest(ctx context.Context, url string) (*http.Response, error) {
	// Log request if enabled
	if ns.config.LogRequests {
		ns.logger.Debug("Making HTTP request", "url", url)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set proper headers for web scraping
	req.Header.Set("User-Agent", ns.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Enforce rate limiting
	if err := ns.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter cancelled: %w", err)
	}

	// Execute request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= ns.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(time.Duration(attempt) * time.Second):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			ns.logger.Warn("Retrying request", "attempt", attempt, "url", url)
		}

		resp, lastErr = ns.client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			// Success
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if lastErr != nil {
			ns.logger.Warn("Request failed", "attempt", attempt, "error", lastErr)
		} else {
			ns.logger.Warn("Request returned non-OK status",
				"attempt", attempt, "status", resp.StatusCode)
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w",
			ns.config.MaxRetries+1, lastErr)
	}

	// Log successful request
	if ns.config.LogRequests {
		ns.logger.Debug("Request completed successfully",
			"url", url, "status", resp.StatusCode)
	}

	return resp, nil
}
