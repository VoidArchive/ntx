package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"ntx/internal/data/models"
	"ntx/internal/database"
	"ntx/internal/market"
	"ntx/internal/market/cache"
	"ntx/internal/market/monitoring"
	"ntx/internal/market/realtime"
	"ntx/internal/market/scraper"
	"ntx/internal/market/validation"
	"slices"
	"sync"
	"time"
)

// MarketService handles NEPSE market data collection and management.
// It integrates with the database using SQLC for type-safe queries and provides real-time price updates.
type MarketService struct {
	dbManager       *database.Manager
	scraper         market.Scraper
	fallbackScraper market.Scraper
	logger          *slog.Logger
	config          *MarketConfig
	stopChan        chan struct{}

	cacheManager  *cache.CacheManager
	updateManager *realtime.UpdateManager
	validator     *validation.DataValidator
	healthMonitor *monitoring.HealthMonitor

	// Synchronization for concurrent operations
	mu         sync.RWMutex
	lastUpdate time.Time

	// Channel-based update coordination to prevent race conditions
	updateChan chan updateRequest
	updateWg   sync.WaitGroup
}

// updateRequest represents a queued market data update
type updateRequest struct {
	ctx   context.Context
	state realtime.MarketState
}

// MarketConfig holds market service configuration.
type MarketConfig struct {
	UpdateInterval         time.Duration
	TradingStartHour       int
	TradingEndHour         int
	TradingDays            []int
	MaxDataAge             time.Duration
	BatchSize              int
	MaxRetries             int
	RetryInterval          time.Duration
	RequestDelay           time.Duration
	EnableFallbackScraper  bool
	EnableDataValidation   bool
	EnableHealthMonitoring bool
	EnableAdvancedCaching  bool
	KnownSymbols           []string
}

// MarketData represents real-time market data for a symbol.
type MarketData struct {
	Symbol        string            `json:"symbol"`
	LastPrice     models.Money      `json:"last_price"`
	ChangeAmount  models.Money      `json:"change_amount"`
	ChangePercent models.Percentage `json:"change_percent"`
	Volume        int64             `json:"volume"`
	Timestamp     time.Time         `json:"timestamp"`
	High          models.Money      `json:"high,omitempty"`
	Low           models.Money      `json:"low,omitempty"`
	Open          models.Money      `json:"open,omitempty"`
	PrevClose     models.Money      `json:"prev_close,omitempty"`
}

// MarketStatus represents current market status.
type MarketStatus struct {
	IsOpen        bool          `json:"is_open"`
	NextOpen      time.Time     `json:"next_open"`
	NextClose     time.Time     `json:"next_close"`
	LastUpdate    time.Time     `json:"last_update"`
	DataAge       time.Duration `json:"data_age"`
	ActiveSymbols int           `json:"active_symbols"`
}

// DefaultMarketConfig returns the default market configuration.
func DefaultMarketConfig() *MarketConfig {
	return &MarketConfig{
		UpdateInterval:   5 * time.Minute,
		TradingStartHour: 11,
		TradingEndHour:   15,
		TradingDays:      []int{0, 1, 2, 3, 4}, // Sunday-Thursday
		MaxDataAge:       30 * time.Minute,
		BatchSize:        100,
		MaxRetries:       3,
		RetryInterval:    30 * time.Second,
		RequestDelay:     3 * time.Second,

		EnableFallbackScraper:  true,
		EnableDataValidation:   true,
		EnableHealthMonitoring: true,
		EnableAdvancedCaching:  true,
		KnownSymbols: []string{
			"NABIL", "EBL", "ADBL", "KTM", "HIDCL", "GBIME", "BOKL", "NICA",
			"SANIMA", "MBL", "CZBIL", "SBI", "PRVU", "PCBL", "LBL", "SCB",
		},
	}
}

// NewMarketService creates a new market service.
func NewMarketService(dbManager *database.Manager, logger *slog.Logger, config *MarketConfig) (*MarketService, error) {
	if config == nil {
		config = DefaultMarketConfig()
	}

	nepseScraper, err := scraper.NewNepseScraper(scraper.DefaultConfig(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create NEPSE scraper: %w", err)
	}

	ms := &MarketService{
		dbManager:  dbManager,
		scraper:    nepseScraper,
		logger:     logger,
		config:     config,
		stopChan:   make(chan struct{}),
		updateChan: make(chan updateRequest, 100), // Buffered channel for update requests
	}

	if err := ms.initializeSubsystems(config, logger); err != nil {
		return nil, err
	}

	logger.Info("Enhanced market service initialized",
		"fallback_scraper", config.EnableFallbackScraper,
		"data_validation", config.EnableDataValidation,
		"health_monitoring", config.EnableHealthMonitoring,
		"advanced_caching", config.EnableAdvancedCaching)

	return ms, nil
}

func (ms *MarketService) initializeSubsystems(config *MarketConfig, logger *slog.Logger) error {
	if config.EnableFallbackScraper {
		fallbackScraper, err := scraper.NewNepseBotScraper(scraper.DefaultNepseBotConfig(), logger)
		if err != nil {
			logger.Warn("Failed to initialize fallback scraper", "error", err)
		} else {
			ms.fallbackScraper = fallbackScraper
		}
	}

	if config.EnableAdvancedCaching {
		ms.cacheManager = cache.NewCacheManager(logger, cache.DefaultCacheConfig())
	}

	if config.EnableDataValidation {
		ms.validator = validation.NewDataValidator(logger, validation.DefaultValidationConfig(), config.KnownSymbols)
	}

	updateConfig := realtime.DefaultUpdateConfig()
	updateConfig.MarketOpenHour = config.TradingStartHour
	updateConfig.MarketCloseHour = config.TradingEndHour
	updateConfig.TradingDays = config.TradingDays
	ms.updateManager = realtime.NewUpdateManager(logger, updateConfig)

	if config.EnableHealthMonitoring {
		ms.healthMonitor = monitoring.NewHealthMonitor(logger, monitoring.DefaultHealthConfig())
	}

	return nil
}

// Start begins automatic market data collection.
func (ms *MarketService) Start(ctx context.Context) error {
	ms.logger.Info("Starting enhanced market service")

	if ms.healthMonitor != nil {
		if err := ms.healthMonitor.Start(ctx); err != nil {
			ms.logger.Error("Failed to start health monitor", "error", err)
		}
		ms.registerHealthCheckers()
	}

	if ms.updateManager != nil {
		// Thread-safe update handler using channels
		updateHandler := func(ctx context.Context, state realtime.MarketState) error {
			select {
			case ms.updateChan <- updateRequest{ctx: ctx, state: state}:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Channel full, log warning but don't block
				ms.logger.Warn("Update channel full, dropping request", "state", state.String())
				return nil
			}
		}

		stateHandler := func(oldState, newState realtime.MarketState, timestamp time.Time) {
			ms.logger.Info("Market state changed", "from", oldState.String(), "to", newState.String())
		}

		if err := ms.updateManager.Start(ctx, updateHandler, stateHandler); err != nil {
			ms.logger.Error("Failed to start update manager", "error", err)
		}
	}

	// Start update worker goroutine
	ms.updateWg.Add(1)
	go ms.updateWorker()

	go ms.loadCachedData(context.Background())

	ms.logger.Info("Enhanced market service started successfully")
	return nil
}

// Stop stops the market service.
func (ms *MarketService) Stop() error {
	ms.logger.Info("Stopping enhanced market service")

	// Signal all workers to stop
	close(ms.stopChan)

	// Close update channel to stop worker
	close(ms.updateChan)

	// Wait for update worker to finish
	ms.updateWg.Wait()

	if ms.updateManager != nil {
		if err := ms.updateManager.Stop(); err != nil {
			ms.logger.Error("Failed to stop update manager", "error", err)
		}
	}

	if ms.healthMonitor != nil {
		if err := ms.healthMonitor.Stop(); err != nil {
			ms.logger.Error("Failed to stop health monitor", "error", err)
		}
	}

	if ms.cacheManager != nil {
		if err := ms.cacheManager.Close(); err != nil {
			ms.logger.Error("Failed to close cache manager", "error", err)
		}
	}

	return nil
}

// GetMarketData retrieves market data for a specific symbol.
func (ms *MarketService) GetMarketData(ctx context.Context, symbol string) (*MarketData, error) {
	if ms.cacheManager != nil {
		if data, exists := ms.cacheManager.Get(symbol, ms.isMarketOpen(time.Now())); exists {
			return ms.convertScrapedToMarketData(*data), nil
		}
	}

	dbData, err := ms.dbManager.Queries().GetMarketData(ctx, symbol)
	if err != nil {
		return ms.fetchAndCacheSymbol(ctx, symbol)
	}

	return ms.convertDBToMarketData(dbData), nil
}

// GetMultipleMarketData retrieves market data for multiple symbols efficiently.
func (ms *MarketService) GetMultipleMarketData(ctx context.Context, symbols []string) (map[string]*MarketData, error) {
	if len(symbols) == 0 {
		return make(map[string]*MarketData), nil
	}

	results := make(map[string]*MarketData)
	var symbolsToFetch []string

	if ms.cacheManager != nil {
		for _, symbol := range symbols {
			if data, exists := ms.cacheManager.Get(symbol, ms.isMarketOpen(time.Now())); exists {
				results[symbol] = ms.convertScrapedToMarketData(*data)
			} else {
				symbolsToFetch = append(symbolsToFetch, symbol)
			}
		}
	} else {
		symbolsToFetch = symbols
	}

	if len(symbolsToFetch) > 0 {
		dbResults, err := ms.dbManager.Queries().GetMarketDataBatch(ctx, symbolsToFetch)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch market data batch: %w", err)
		}

		for _, dbData := range dbResults {
			results[dbData.Symbol] = ms.convertDBToMarketData(dbData)
		}
	}

	return results, nil
}

// GetMarketStatus returns current market status and statistics.
func (ms *MarketService) GetMarketStatus(ctx context.Context) (*MarketStatus, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	now := time.Now()
	status := &MarketStatus{
		IsOpen:     ms.isMarketOpen(now),
		LastUpdate: ms.lastUpdate,
	}

	if ms.cacheManager != nil {
		stats := ms.cacheManager.GetStats()
		status.ActiveSymbols = stats.HotEntries + stats.WarmEntries
	}

	if !ms.lastUpdate.IsZero() {
		status.DataAge = now.Sub(ms.lastUpdate)
	}

	status.NextOpen, status.NextClose = ms.calculateNextTradingTimes(now)

	return status, nil
}

// RefreshAllData forces a refresh of all market data.
func (ms *MarketService) RefreshAllData(ctx context.Context) error {
	ms.logger.Info("Starting manual refresh of all market data")
	return ms.performEnhancedDataUpdate(ctx, ms.updateManager.GetCurrentState())
}

// RefreshSymbol forces a refresh of specific symbol data.
func (ms *MarketService) RefreshSymbol(ctx context.Context, symbol string) error {
	ms.logger.Info("Refreshing specific symbol", "symbol", symbol)
	_, err := ms.fetchAndCacheSymbol(ctx, symbol)
	return err
}

// updateWorker processes market data updates sequentially to prevent race conditions
func (ms *MarketService) updateWorker() {
	defer ms.updateWg.Done()

	for {
		select {
		case request, ok := <-ms.updateChan:
			if !ok {
				// Channel closed, worker should exit
				ms.logger.Info("Update worker stopping")
				return
			}

			// Process update with error handling
			if err := ms.performEnhancedDataUpdate(request.ctx, request.state); err != nil {
				ms.logger.Error("Update processing failed", "error", err, "state", request.state.String())
			}

		case <-ms.stopChan:
			// Service stopping
			ms.logger.Info("Update worker received stop signal")
			return
		}
	}
}

func (ms *MarketService) loadCachedData(ctx context.Context) error {
	if ms.cacheManager == nil {
		return nil
	}

	ms.logger.Info("Loading cached market data from database")
	allData, err := ms.dbManager.Queries().GetAllMarketData(ctx)
	if err != nil {
		return fmt.Errorf("failed to load cached data: %w", err)
	}

	cacheData := make(map[string]*market.ScrapedData)
	for _, dbData := range allData {
		cacheData[dbData.Symbol] = ms.convertDBToScrapedData(dbData)
	}
	ms.cacheManager.SetBatch(cacheData, ms.isMarketOpen(time.Now()))

	ms.mu.Lock()
	if len(allData) > 0 {
		ms.lastUpdate = allData[0].Timestamp.Time
	}
	ms.mu.Unlock()

	ms.logger.Info("Loaded cached market data", "symbols_count", len(allData))
	return nil
}

func (ms *MarketService) performEnhancedDataUpdate(ctx context.Context, state realtime.MarketState) error {
	startTime := time.Now()
	ms.logger.Info("Starting enhanced data update", "market_state", state.String())

	scrapedData, err := ms.scrapeDataWithFallback(ctx)
	if err != nil {
		return err
	}

	if len(scrapedData) == 0 {
		ms.logger.Warn("No data scraped from any source")
		return nil
	}

	if ms.validator != nil {
		scrapedData = ms.validateData(scrapedData, state)
	}

	if err := ms.batchUpdateDatabase(ctx, scrapedData); err != nil {
		ms.logger.Error("Failed to update database", "error", err)
	}

	if ms.cacheManager != nil {
		cacheData := make(map[string]*market.ScrapedData)
		for i := range scrapedData {
			cacheData[scrapedData[i].Symbol] = &scrapedData[i]
		}
		ms.cacheManager.SetBatch(cacheData, state == realtime.MarketOpen)
	}

	ms.mu.Lock()
	ms.lastUpdate = startTime
	ms.mu.Unlock()

	ms.logger.Info("Enhanced data update completed",
		"symbols_updated", len(scrapedData),
		"duration", time.Since(startTime))

	return nil
}

func (ms *MarketService) scrapeDataWithFallback(ctx context.Context) ([]market.ScrapedData, error) {
	scrapedData, err := ms.scraper.ScrapeAllMarketData(ctx)
	if err != nil {
		ms.logger.Warn("Primary scraper failed, trying fallback", "error", err)
		if ms.fallbackScraper != nil {
			return ms.fallbackScraper.ScrapeAllMarketData(ctx)
		}
		return nil, fmt.Errorf("primary scraper failed and no fallback available: %w", err)
	}
	return scrapedData, nil
}

func (ms *MarketService) validateData(scrapedData []market.ScrapedData, state realtime.MarketState) []market.ScrapedData {
	validatedData := make([]market.ScrapedData, 0, len(scrapedData))
	for _, data := range scrapedData {
		var historical *market.ScrapedData
		if ms.cacheManager != nil {
			if cached, exists := ms.cacheManager.Get(data.Symbol, state == realtime.MarketOpen); exists {
				historical = cached
			}
		}

		result := ms.validator.ValidateMarketData(&data, historical)
		if result.IsValid {
			validatedData = append(validatedData, data)
		} else {
			ms.logger.Warn("Data validation failed", "symbol", data.Symbol, "errors", len(result.Errors))
		}
	}
	return validatedData
}

func (ms *MarketService) fetchAndCacheSymbol(ctx context.Context, symbol string) (*MarketData, error) {
	ms.logger.Info("Fetching symbol data", "symbol", symbol)

	scrapedData, err := ms.scraper.ScrapeSymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape symbol %s: %w", symbol, err)
	}

	if err := ms.upsertSymbolData(ctx, *scrapedData); err != nil {
		ms.logger.Warn("Failed to update database for symbol", "symbol", symbol, "error", err)
	}

	if ms.cacheManager != nil {
		ms.cacheManager.Set(symbol, scrapedData, ms.isMarketOpen(time.Now()))
	}

	return ms.convertScrapedToMarketData(*scrapedData), nil
}

func (ms *MarketService) batchUpdateDatabase(ctx context.Context, scrapedData []market.ScrapedData) error {
	batchSize := ms.config.BatchSize
	for i := 0; i < len(scrapedData); i += batchSize {
		end := min(i+batchSize, len(scrapedData))
		batch := scrapedData[i:end]

		for _, data := range batch {
			if err := ms.upsertSymbolData(ctx, data); err != nil {
				ms.logger.Error("Failed to upsert symbol data", "symbol", data.Symbol, "error", err)
			}
		}
	}
	return nil
}

func (ms *MarketService) upsertSymbolData(ctx context.Context, data market.ScrapedData) error {
	params := database.UpsertMarketDataParams{
		Symbol:       data.Symbol,
		LastPrice:    data.LastPrice,
		ChangeAmount: data.ChangeAmount,
		Timestamp:    sql.NullTime{Time: data.ScrapedAt, Valid: true},
	}

	if data.ChangePercent != 0 {
		params.ChangePercent = sql.NullInt64{Int64: int64(data.ChangePercent * 100), Valid: true}
	}

	if data.Volume > 0 {
		params.Volume = sql.NullInt64{Int64: data.Volume, Valid: true}
	}

	_, err := ms.dbManager.Queries().UpsertMarketData(ctx, params)
	return err
}

func (ms *MarketService) convertDBToMarketData(dbData database.MarketDatum) *MarketData {
	return &MarketData{
		Symbol:        dbData.Symbol,
		LastPrice:     dbData.LastPrice,
		ChangeAmount:  dbData.ChangeAmount,
		ChangePercent: models.Percentage(dbData.ChangePercent.Int64),
		Volume:        dbData.Volume.Int64,
		Timestamp:     dbData.Timestamp.Time,
	}
}

func (ms *MarketService) convertDBToScrapedData(dbData database.MarketDatum) *market.ScrapedData {
	return &market.ScrapedData{
		Symbol:        dbData.Symbol,
		LastPrice:     dbData.LastPrice,
		ChangeAmount:  dbData.ChangeAmount,
		ChangePercent: float64(dbData.ChangePercent.Int64) / 100,
		Volume:        dbData.Volume.Int64,
		ScrapedAt:     dbData.Timestamp.Time,
	}
}

func (ms *MarketService) convertScrapedToMarketData(data market.ScrapedData) *MarketData {
	return &MarketData{
		Symbol:        data.Symbol,
		LastPrice:     data.LastPrice,
		ChangeAmount:  data.ChangeAmount,
		ChangePercent: models.NewPercentageFromFloat(data.ChangePercent),
		Volume:        data.Volume,
		Timestamp:     data.ScrapedAt,
		High:          data.High,
		Low:           data.Low,
		Open:          data.Open,
		PrevClose:     data.PrevClose,
	}
}

func (ms *MarketService) isMarketOpen(now time.Time) bool {
	weekday := int(now.Weekday())
	if slices.Contains(ms.config.TradingDays, weekday) {
		hour := now.Hour()
		return hour >= ms.config.TradingStartHour && hour < ms.config.TradingEndHour
	}
	return false
}

func (ms *MarketService) calculateNextTradingTimes(now time.Time) (time.Time, time.Time) {
	if ms.isMarketOpen(now) {
		nextClose := time.Date(now.Year(), now.Month(), now.Day(), ms.config.TradingEndHour, 0, 0, 0, now.Location())
		nextOpen := ms.findNextTradingDay(now.AddDate(0, 0, 1))
		return nextOpen, nextClose
	}

	nextOpen := ms.findNextTradingDay(now)
	nextClose := time.Date(nextOpen.Year(), nextOpen.Month(), nextOpen.Day(), ms.config.TradingEndHour, 0, 0, 0, nextOpen.Location())
	return nextOpen, nextClose
}

func (ms *MarketService) findNextTradingDay(from time.Time) time.Time {
	nextDay := from
	if from.Hour() >= ms.config.TradingEndHour {
		nextDay = nextDay.AddDate(0, 0, 1)
	}

	for {
		isTradingDay := slices.Contains(ms.config.TradingDays, int(nextDay.Weekday()))
		if isTradingDay {
			return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), ms.config.TradingStartHour, 0, 0, 0, nextDay.Location())
		}
		nextDay = nextDay.AddDate(0, 0, 1)
	}
}

func (ms *MarketService) registerHealthCheckers() {
	if ms.healthMonitor == nil {
		return
	}

	ms.healthMonitor.RegisterComponent("nepse_scraper", &ScraperHealthChecker{scraper: ms.scraper, name: "nepse_scraper"})
	if ms.fallbackScraper != nil {
		ms.healthMonitor.RegisterComponent("nepsebot_scraper", &ScraperHealthChecker{scraper: ms.fallbackScraper, name: "nepsebot_scraper"})
	}
	ms.healthMonitor.RegisterComponent("database", &DatabaseHealthChecker{dbManager: ms.dbManager, name: "database"})
}

// Health checker implementations
type ScraperHealthChecker struct {
	scraper market.Scraper
	name    string
}

func (shc *ScraperHealthChecker) CheckHealth(ctx context.Context) *monitoring.HealthCheckResult {
	startTime := time.Now()
	if err := shc.scraper.GetHealthStatus(ctx); err != nil {
		return &monitoring.HealthCheckResult{Status: monitoring.HealthStatusUnhealthy, Error: err.Error(), ResponseTime: time.Since(startTime)}
	}
	return &monitoring.HealthCheckResult{Status: monitoring.HealthStatusHealthy, ResponseTime: time.Since(startTime)}
}

func (shc *ScraperHealthChecker) GetComponentName() string {
	return shc.name
}

type DatabaseHealthChecker struct {
	dbManager *database.Manager
	name      string
}

func (dhc *DatabaseHealthChecker) CheckHealth(ctx context.Context) *monitoring.HealthCheckResult {
	startTime := time.Now()
	if err := dhc.dbManager.Ping(ctx); err != nil {
		return &monitoring.HealthCheckResult{Status: monitoring.HealthStatusUnhealthy, Error: err.Error(), ResponseTime: time.Since(startTime)}
	}
	return &monitoring.HealthCheckResult{Status: monitoring.HealthStatusHealthy, ResponseTime: time.Since(startTime)}
}

func (dhc *DatabaseHealthChecker) GetComponentName() string {
	return dhc.name
}
