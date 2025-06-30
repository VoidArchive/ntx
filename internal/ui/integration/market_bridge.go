package integration

import (
	"context"
	"log/slog"
	"ntx/internal/app/services"
	"ntx/internal/ui/common"
	"slices"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// MarketBridge connects the TUI dashboard to the MarketService for real-time updates
type MarketBridge struct {
	marketService    *services.MarketService
	portfolioService *services.PortfolioService
	logger           *slog.Logger
	program          *tea.Program

	// Update management
	updateInterval time.Duration
	stopChan       chan struct{}
	mu             sync.RWMutex
	isRunning      bool

	// Data cache for delta detection
	lastMarketStatus *services.MarketStatus
	lastPrices       map[string]float64

	// Configuration
	enableDeltaUpdates bool
	batchSize          int
	watchedSymbols     []string
}

// NewMarketBridge creates a new market data bridge
func NewMarketBridge(marketService *services.MarketService, portfolioService *services.PortfolioService, logger *slog.Logger) *MarketBridge {
	return &MarketBridge{
		marketService:      marketService,
		portfolioService:   portfolioService,
		logger:             logger,
		updateInterval:     5 * time.Second,
		stopChan:           make(chan struct{}),
		lastPrices:         make(map[string]float64),
		enableDeltaUpdates: true,
		batchSize:          50,
		watchedSymbols:     []string{"NABIL", "EBL", "KTM", "HIDCL", "ADBL", "GBIME"},
	}
}

// SetProgram sets the Bubbletea program for sending updates
func (mb *MarketBridge) SetProgram(program *tea.Program) {
	mb.program = program
}

// Start begins the market data bridge
func (mb *MarketBridge) Start(ctx context.Context) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.isRunning {
		return nil
	}

	mb.logger.Info("Starting market data bridge")
	mb.isRunning = true

	// Start update routines
	go mb.updateLoop(ctx)

	return nil
}

// Stop stops the market data bridge
func (mb *MarketBridge) Stop() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if !mb.isRunning {
		return nil
	}

	mb.logger.Info("Stopping market data bridge")
	close(mb.stopChan)
	mb.isRunning = false

	return nil
}

// updateLoop continuously updates market data
func (mb *MarketBridge) updateLoop(ctx context.Context) {
	ticker := time.NewTicker(mb.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mb.fetchAndSendData(ctx)

		case <-mb.stopChan:
			return

		case <-ctx.Done():
			return
		}
	}
}

// fetchAndSendData fetches and sends all data types
func (mb *MarketBridge) fetchAndSendData(ctx context.Context) {
	// Send market status
	status, err := mb.marketService.GetMarketStatus(ctx)
	if err == nil {
		msg := common.NewMarketStatusMsg(
			status.IsOpen,
			status.NextOpen,
			status.NextClose,
			status.ActiveSymbols,
			status.DataAge,
		)
		mb.sendMessage(msg)
	}

	// Send real portfolio data
	mb.sendRealPortfolioData(ctx)

	// Send simulated market data
	mb.sendSimulatedMarketData()
}

// sendRealPortfolioData sends real portfolio data using the portfolio service
func (mb *MarketBridge) sendRealPortfolioData(ctx context.Context) {
	if mb.portfolioService == nil {
		mb.logger.Warn("Portfolio service not available")
		return
	}

	// Get real portfolio data
	portfolioData, err := mb.portfolioService.GetPortfolioData(ctx)
	if err != nil {
		mb.logger.Error("Failed to get portfolio data", "error", err)
		return
	}

	// Convert portfolio data to UI format
	holdings := make([]common.PortfolioHolding, 0, len(portfolioData.Holdings))
	for _, holding := range portfolioData.Holdings {
		uiHolding := common.PortfolioHolding{
			Symbol:       holding.Symbol,
			Quantity:     holding.Quantity.Int64(),
			AvgPrice:     holding.AvgCost.Rupees(),
			CurrentPrice: holding.CurrentPrice.Rupees(),
			Value:        holding.CurrentValue.Rupees(),
			Gain:         holding.UnrealizedGain.Rupees(),
			GainPercent:  holding.GainPercent.Float(),
		}
		holdings = append(holdings, uiHolding)
	}

	// Send portfolio update message with real data
	msg := common.NewPortfolioUpdateMsg(
		portfolioData.TotalValue.Rupees(),
		portfolioData.DayChange.Rupees(),
		portfolioData.TotalGain.Rupees(),
		holdings,
	)

	mb.sendMessage(msg)

	mb.logger.Debug("Sent real portfolio data update",
		"total_value", portfolioData.TotalValue.FormattedString(),
		"day_change", portfolioData.DayChange.FormattedString(),
		"holdings_count", len(holdings))
}

// sendSimulatedMarketData sends simulated market data
func (mb *MarketBridge) sendSimulatedMarketData() {
	symbols := []string{"NABIL", "EBL", "KTM", "HIDCL"}
	prices := map[string]float64{
		"NABIL": 1250.0,
		"EBL":   780.0,
		"KTM":   520.0,
		"HIDCL": 350.0,
	}
	changes := map[string]float64{
		"NABIL": 25.0,
		"EBL":   -15.0,
		"KTM":   10.0,
		"HIDCL": 5.0,
	}

	msg := common.NewWatchlistUpdateMsg(symbols, prices, changes)
	mb.sendMessage(msg)
}

// sendMessage sends a message to the TUI program
func (mb *MarketBridge) sendMessage(msg tea.Msg) {
	if mb.program == nil {
		return
	}
	mb.program.Send(msg)
}

// TriggerRefresh forces an immediate refresh
func (mb *MarketBridge) TriggerRefresh() {
	if !mb.isRunning {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		mb.fetchAndSendData(ctx)
	}()
}

// AddWatchedSymbol adds a symbol to watch
func (mb *MarketBridge) AddWatchedSymbol(symbol string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if slices.Contains(mb.watchedSymbols, symbol) {
		return
	}

	mb.watchedSymbols = append(mb.watchedSymbols, symbol)
}

// IsRunning returns if the bridge is running
func (mb *MarketBridge) IsRunning() bool {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.isRunning
}
