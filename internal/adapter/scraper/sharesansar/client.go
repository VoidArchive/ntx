package sharesansar

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/VoidArchive/ntx/internal/domain/interfaces"
	"github.com/VoidArchive/ntx/internal/domain/models"
	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Client implements the MarketFeed interface for ShareSansar
type Client struct {
	collector   *colly.Collector
	parser      *Parser
	rateLimiter *rate.Limiter
	logger      *zap.Logger

	// Health tracking
	mu                sync.RWMutex
	lastSuccess       time.Time
	errorCount        int
	consecutiveErrors int
}

// NewClient creates a new ShareSansar client
func NewClient(logger *zap.Logger) *Client {
	c := colly.NewCollector(
		colly.AllowedDomains("www.sharesansar.com", "sharesansar.com"),
		colly.Async(false), // Synchronous for rate limiting
	)

	// Configure timeouts
	c.SetRequestTimeout(30 * time.Second)

	// Configure rate limiting through colly
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*sharesansar.*",
		Parallelism: 1,
		Delay:       15 * time.Second,
		RandomDelay: 5 * time.Second,
	})

	// Set request headers
	c.OnRequest(func(r *colly.Request) {
		ua := UserAgents[rand.Intn(len(UserAgents))]
		r.Headers.Set("User-Agent", ua)

		for k, v := range DefaultHeaders {
			r.Headers.Set(k, v)
		}

		logger.Debug("requesting",
			zap.String("url", r.URL.String()),
			zap.String("user_agent", ua))
	})

	// Log errors
	c.OnError(func(r *colly.Response, err error) {
		logger.Error("scraper error",
			zap.String("url", r.Request.URL.String()),
			zap.Int("status", r.StatusCode),
			zap.Error(err))
	})

	return &Client{
		collector:   c,
		parser:      NewParser(),
		rateLimiter: rate.NewLimiter(rate.Every(15*time.Second), 1),
		logger:      logger,
	}
}

// GetQuote fetches a single quote
func (c *Client) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	quotes, err := c.GetQuotes(ctx, []string{symbol})
	if err != nil {
		return nil, err
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}

	return quotes[0], nil
}

// GetQuotes fetches multiple quotes (filters if symbols provided)
func (c *Client) GetQuotes(ctx context.Context, symbols []string) ([]*models.Quote, error) {
	// Rate limit
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	var quotes []*models.Quote
	var mu sync.Mutex
	var scrapeErr error

	// Create symbol map for filtering
	symbolFilter := make(map[string]bool)
	for _, s := range symbols {
		symbolFilter[s] = true
	}

	// Setup scraping callback
	c.collector.OnHTML(SelectorLiveTable, func(e *colly.HTMLElement) {
		quote, err := c.parser.ParseQuoteRow(e)
		if err != nil {
			// Skip invalid rows silently (could be headers, etc)
			return
		}

		// Filter by requested symbols if specified
		if len(symbols) > 0 && !symbolFilter[quote.Symbol] {
			return
		}

		mu.Lock()
		quotes = append(quotes, quote)
		mu.Unlock()
	})

	// Visit the page
	if err := c.collector.Visit(URLLiveTrading); err != nil {
		c.recordError()
		return nil, fmt.Errorf("failed to visit ShareSansar: %w", err)
	}

	// Wait for async operations
	c.collector.Wait()

	if scrapeErr != nil {
		c.recordError()
		return nil, scrapeErr
	}

	// Record success
	c.recordSuccess()

	c.logger.Info("scraped quotes",
		zap.Int("requested", len(symbols)),
		zap.Int("found", len(quotes)))

	return quotes, nil
}

// GetCandles fetches historical data (not implemented yet)
func (c *Client) GetCandles(ctx context.Context, symbol string, from, to time.Time) ([]*models.Candle, error) {
	// ShareSansar historical data requires different endpoint
	// TODO: Implement when we find the historical data URL
	return nil, fmt.Errorf("historical data not implemented")
}

// Subscribe creates a channel for real-time updates
func (c *Client) Subscribe(ctx context.Context, symbols []string) (<-chan *models.Quote, error) {
	ch := make(chan *models.Quote, len(symbols))

	// ShareSansar doesn't have WebSocket, so we poll
	go func() {
		defer close(ch)

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				quotes, err := c.GetQuotes(ctx, symbols)
				if err != nil {
					c.logger.Error("subscription fetch failed", zap.Error(err))
					continue
				}

				for _, quote := range quotes {
					select {
					case ch <- quote:
					case <-ctx.Done():
						return
					default:
						// Channel full, skip
						c.logger.Warn("subscription channel full, dropping quote",
							zap.String("symbol", quote.Symbol))
					}
				}
			}
		}
	}()

	return ch, nil
}

// Health returns the current health status
func (c *Client) Health() interfaces.HealthStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := "healthy"
	if c.consecutiveErrors > 3 {
		status = "degraded"
	}
	if c.consecutiveErrors > 10 {
		status = "unhealthy"
	}

	latency := time.Duration(0)
	if !c.lastSuccess.IsZero() {
		// Estimate based on typical response time
		latency = 2 * time.Second
	}

	return interfaces.HealthStatus{
		Status:      status,
		LastSuccess: c.lastSuccess,
		ErrorCount:  c.errorCount,
		Latency:     latency,
	}
}

// GetCompanyDetails fetches detailed company information
func (c *Client) GetCompanyDetails(ctx context.Context, symbol string) (map[string]string, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	details := make(map[string]string)
	var scrapeErr error

	c.collector.OnHTML("body", func(e *colly.HTMLElement) {
		details = c.parser.ParseCompanyDetails(e)
	})

	url := fmt.Sprintf(URLCompanyDetail, symbol)
	if err := c.collector.Visit(url); err != nil {
		c.recordError()
		return nil, fmt.Errorf("failed to fetch company details: %w", err)
	}

	c.collector.Wait()

	if scrapeErr != nil {
		return nil, scrapeErr
	}

	c.recordSuccess()
	return details, nil
}

// Helper methods

func (c *Client) recordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastSuccess = time.Now()
	c.consecutiveErrors = 0
}

func (c *Client) recordError() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.errorCount++
	c.consecutiveErrors++
}

// ResetCollector creates a fresh collector (useful for cleaning up state)
func (c *Client) ResetCollector() {
	c.collector = colly.NewCollector(
		colly.AllowedDomains("www.sharesansar.com", "sharesansar.com"),
		colly.Async(false),
	)

	// Reapply all configurations
	c.collector.SetRequestTimeout(30 * time.Second)
	c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*sharesansar.*",
		Parallelism: 1,
		Delay:       15 * time.Second,
		RandomDelay: 5 * time.Second,
	})
}
