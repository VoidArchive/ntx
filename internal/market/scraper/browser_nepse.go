package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"ntx/internal/data/models"
	"ntx/internal/market"

	"github.com/chromedp/chromedp"
)

// BrowserNepseScraper implements NEPSE data scraping using ChromeDP for JavaScript-rendered content
type BrowserNepseScraper struct {
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *slog.Logger
	fallback market.Scraper // HTTP fallback scraper
	isActive bool
}

// NewBrowserNepseScraper creates a new browser-based NEPSE scraper with fallback
func NewBrowserNepseScraper(logger *slog.Logger, fallback market.Scraper) *BrowserNepseScraper {
	return &BrowserNepseScraper{
		logger:   logger,
		fallback: fallback,
		isActive: false,
	}
}

// StartPersistentBrowser initializes and starts the persistent browser process
func (bs *BrowserNepseScraper) StartPersistentBrowser() error {
	// Create context with Chrome allocator
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("NTX-NEPSE-Terminal/1.0 (Portfolio Management Tool)"),
	)

	bs.cancel = cancel
	bs.ctx, _ = chromedp.NewContext(allocCtx)

	// Test browser startup with about:blank
	err := chromedp.Run(bs.ctx, chromedp.Navigate("about:blank"))
	if err != nil {
		bs.logger.Error("Failed to start browser", "error", err)
		return fmt.Errorf("browser startup failed: %w", err)
	}

	bs.isActive = true
	bs.logger.Info("Browser scraper started successfully")
	return nil
}

// StopBrowser shuts down the persistent browser process
func (bs *BrowserNepseScraper) StopBrowser() {
	if bs.cancel != nil {
		bs.cancel()
		bs.isActive = false
		bs.logger.Info("Browser scraper stopped")
	}
}

// ScrapeAllMarketData implements the market.Scraper interface
func (bs *BrowserNepseScraper) ScrapeAllMarketData(ctx context.Context) ([]market.ScrapedData, error) {
	if !bs.isActive {
		bs.logger.Warn("Browser not active, using fallback scraper")
		return bs.fallback.ScrapeAllMarketData(ctx)
	}

	var allData []market.ScrapedData
	scrapeCtx, cancel := context.WithTimeout(bs.ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(scrapeCtx,
		chromedp.Navigate("https://nepalstock.com/market-data"),
		chromedp.WaitVisible(`table.market-data-table`, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // Allow JavaScript to load data

		// Extract market data rows
		chromedp.ActionFunc(func(ctx context.Context) error {
			var rows []string
			err := chromedp.Evaluate(`
				Array.from(document.querySelectorAll('table.market-data-table tbody tr'))
					.map(row => Array.from(row.cells).map(cell => cell.textContent.trim()).join('|'))
					.join('\n')
			`, &rows).Do(ctx)
			
			if err != nil {
				return err
			}

			// Parse the extracted data
			for _, row := range strings.Split(strings.Join(rows, ""), "\n") {
				if row == "" {
					continue
				}
				
				cells := strings.Split(row, "|")
				if len(cells) >= 8 {
					data := market.ScrapedData{
						Symbol:        cells[0],
						LastPrice:     models.Money(parseFloatSafe(cells[1]) * 100), // Convert to paisa
						ChangeAmount:  models.Money(parseFloatSafe(cells[2]) * 100),
						ChangePercent: parseFloatSafe(cells[3]),
						Volume:        int64(parseIntSafe(cells[4])),
						High:          models.Money(parseFloatSafe(cells[5]) * 100),
						Low:           models.Money(parseFloatSafe(cells[6]) * 100),
						Open:          models.Money(parseFloatSafe(cells[7]) * 100),
						ScrapedAt:     time.Now(),
					}
					if len(cells) >= 9 {
						data.PrevClose = models.Money(parseFloatSafe(cells[8]) * 100)
					}
					allData = append(allData, data)
				}
			}
			
			return nil
		}),
	)

	if err != nil {
		bs.logger.Error("Browser market data scraping failed, falling back to HTTP", "error", err)
		return bs.fallback.ScrapeAllMarketData(ctx)
	}

	bs.logger.Debug("Market data fetched via browser", "symbols", len(allData))
	return allData, nil
}

// ScrapeSymbol implements the market.Scraper interface for a specific symbol
func (bs *BrowserNepseScraper) ScrapeSymbol(ctx context.Context, symbol string) (*market.ScrapedData, error) {
	if !bs.isActive {
		bs.logger.Warn("Browser not active, using fallback scraper")
		return bs.fallback.ScrapeSymbol(ctx, symbol)
	}

	data := &market.ScrapedData{Symbol: symbol, ScrapedAt: time.Now()}
	scrapeCtx, cancel := context.WithTimeout(bs.ctx, 10*time.Second)
	defer cancel()

	var ltp, change, volume, high, low, open, prevClose string
	url := fmt.Sprintf("https://nepalstock.com/company/detail/%s", symbol)
	
	err := chromedp.Run(scrapeCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`.company-detail`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),

		// Extract symbol data
		chromedp.Text(`.ltp-value`, &ltp, chromedp.ByQuery),
		chromedp.Text(`.change-value`, &change, chromedp.ByQuery),
		chromedp.Text(`.volume-value`, &volume, chromedp.ByQuery),
		chromedp.Text(`.high-value`, &high, chromedp.ByQuery),
		chromedp.Text(`.low-value`, &low, chromedp.ByQuery),
		chromedp.Text(`.open-value`, &open, chromedp.ByQuery),
		chromedp.Text(`.prev-close-value`, &prevClose, chromedp.ByQuery),
	)

	if err != nil {
		bs.logger.Error("Browser symbol scraping failed, falling back to HTTP", 
			"symbol", symbol, "error", err)
		return bs.fallback.ScrapeSymbol(ctx, symbol)
	}

	// Parse scraped values and convert to proper types
	data.LastPrice = models.Money(parseFloatSafe(ltp) * 100)      // Convert to paisa
	data.ChangeAmount = models.Money(parseFloatSafe(change) * 100)
	data.Volume = int64(parseIntSafe(volume))
	data.High = models.Money(parseFloatSafe(high) * 100)
	data.Low = models.Money(parseFloatSafe(low) * 100)
	data.Open = models.Money(parseFloatSafe(open) * 100)
	data.PrevClose = models.Money(parseFloatSafe(prevClose) * 100)

	// Calculate percentage change
	if data.PrevClose > 0 {
		data.ChangePercent = float64(data.ChangeAmount) / float64(data.PrevClose) * 100
	}

	bs.logger.Debug("Symbol data fetched via browser", 
		"symbol", symbol, "ltp", data.LastPrice)

	return data, nil
}

// GetHealthStatus implements the market.Scraper interface
func (bs *BrowserNepseScraper) GetHealthStatus(ctx context.Context) error {
	if !bs.isActive {
		return fmt.Errorf("browser not active")
	}

	healthCtx, cancel := context.WithTimeout(bs.ctx, 5*time.Second)
	defer cancel()

	var title string
	err := chromedp.Run(healthCtx,
		chromedp.Navigate("about:blank"),
		chromedp.Title(&title),
	)

	if err != nil {
		bs.logger.Error("Browser health check failed", "error", err)
		// Attempt restart
		if restartErr := bs.RestartBrowser(); restartErr != nil {
			return fmt.Errorf("health check failed and restart failed: %w", restartErr)
		}
	}

	return nil
}

// Close implements the market.Scraper interface
func (bs *BrowserNepseScraper) Close() error {
	bs.StopBrowser()
	return nil
}

// RestartBrowser restarts the browser process if it crashes
func (bs *BrowserNepseScraper) RestartBrowser() error {
	bs.logger.Info("Restarting browser process")
	
	if bs.isActive {
		bs.StopBrowser()
	}
	
	return bs.StartPersistentBrowser()
}

// Helper functions
func (bs *BrowserNepseScraper) isMarketOpen(status string) bool {
	// INFO: NEPSE trading hours are 11:00-15:00 NPT, Sunday-Thursday
	now := time.Now()
	
	// Check if it's a trading day (Sunday = 0, Thursday = 4 in Nepal)
	weekday := now.Weekday()
	if weekday == time.Friday || weekday == time.Saturday {
		return false
	}
	
	// Check trading hours (11:00-15:00 NPT)
	hour := now.Hour()
	if hour < 11 || hour >= 15 {
		return false
	}
	
	// Also check status text for additional confirmation
	statusLower := strings.ToLower(strings.TrimSpace(status))
	return strings.Contains(statusLower, "open") || strings.Contains(statusLower, "trading")
}

func parseFloatSafe(s string) float64 {
	cleaned := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	if val, err := strconv.ParseFloat(cleaned, 64); err == nil {
		return val
	}
	return 0.0
}

func parseIntSafe(s string) int {
	cleaned := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	if val, err := strconv.Atoi(cleaned); err == nil {
		return val
	}
	return 0
}