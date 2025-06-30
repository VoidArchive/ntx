package common

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Custom messages for the TUI application

// MarketDataMsg carries real-time market data updates
type MarketDataMsg struct {
	Symbol    string
	LastPrice float64
	Change    float64
	Volume    int64
	Timestamp time.Time
}

// MarketStatusMsg carries market status updates
type MarketStatusMsg struct {
	IsOpen        bool
	NextOpen      time.Time
	NextClose     time.Time
	ActiveSymbols int
	DataAge       time.Duration
}

// PortfolioUpdateMsg carries portfolio data updates
type PortfolioUpdateMsg struct {
	TotalValue float64
	DayChange  float64
	TotalGain  float64
	Holdings   []PortfolioHolding
	LastUpdate time.Time
}

// PortfolioHolding represents a single holding in the portfolio
type PortfolioHolding struct {
	Symbol       string
	Quantity     int64
	AvgPrice     float64
	CurrentPrice float64
	Value        float64
	Gain         float64
	GainPercent  float64
}

// WatchlistUpdateMsg carries watchlist updates
type WatchlistUpdateMsg struct {
	Symbols    []string
	Prices     map[string]float64
	Changes    map[string]float64
	LastUpdate time.Time
}

// SystemHealthMsg carries system health status
type SystemHealthMsg struct {
	Status         string
	ComponentsUp   int
	ComponentsDown int
	CacheHitRate   float64
	LastUpdate     time.Time
}

// ErrorMsg carries error information
type ErrorMsg struct {
	Component string
	Error     error
	Timestamp time.Time
}

// RefreshMsg triggers a refresh of all data
type RefreshMsg struct {
	Force bool
}

// LayoutChangeMsg triggers a layout change
type LayoutChangeMsg struct {
	LayoutName string
}

// PaneFocusMsg changes focus to a specific pane
type PaneFocusMsg struct {
	PaneType PaneType
}

// TickMsg represents a ticker event for auto-updates
type TickMsg time.Time

// Message creation helpers

// NewMarketDataMsg creates a new market data message
func NewMarketDataMsg(symbol string, price, change float64, volume int64) tea.Msg {
	return MarketDataMsg{
		Symbol:    symbol,
		LastPrice: price,
		Change:    change,
		Volume:    volume,
		Timestamp: time.Now(),
	}
}

// NewMarketStatusMsg creates a new market status message
func NewMarketStatusMsg(isOpen bool, nextOpen, nextClose time.Time, activeSymbols int, dataAge time.Duration) tea.Msg {
	return MarketStatusMsg{
		IsOpen:        isOpen,
		NextOpen:      nextOpen,
		NextClose:     nextClose,
		ActiveSymbols: activeSymbols,
		DataAge:       dataAge,
	}
}

// NewPortfolioUpdateMsg creates a new portfolio update message
func NewPortfolioUpdateMsg(totalValue, dayChange, totalGain float64, holdings []PortfolioHolding) tea.Msg {
	return PortfolioUpdateMsg{
		TotalValue: totalValue,
		DayChange:  dayChange,
		TotalGain:  totalGain,
		Holdings:   holdings,
		LastUpdate: time.Now(),
	}
}

// NewWatchlistUpdateMsg creates a new watchlist update message
func NewWatchlistUpdateMsg(symbols []string, prices, changes map[string]float64) tea.Msg {
	return WatchlistUpdateMsg{
		Symbols:    symbols,
		Prices:     prices,
		Changes:    changes,
		LastUpdate: time.Now(),
	}
}

// NewSystemHealthMsg creates a new system health message
func NewSystemHealthMsg(status string, up, down int, cacheHitRate float64) tea.Msg {
	return SystemHealthMsg{
		Status:         status,
		ComponentsUp:   up,
		ComponentsDown: down,
		CacheHitRate:   cacheHitRate,
		LastUpdate:     time.Now(),
	}
}

// NewErrorMsg creates a new error message
func NewErrorMsg(component string, err error) tea.Msg {
	return ErrorMsg{
		Component: component,
		Error:     err,
		Timestamp: time.Now(),
	}
}

// NewRefreshMsg creates a new refresh message
func NewRefreshMsg(force bool) tea.Cmd {
	return func() tea.Msg {
		return RefreshMsg{Force: force}
	}
}

// NewLayoutChangeMsg creates a new layout change message
func NewLayoutChangeMsg(layoutName string) tea.Cmd {
	return func() tea.Msg {
		return LayoutChangeMsg{LayoutName: layoutName}
	}
}

// NewPaneFocusMsg creates a new pane focus message
func NewPaneFocusMsg(paneType PaneType) tea.Cmd {
	return func() tea.Msg {
		return PaneFocusMsg{PaneType: paneType}
	}
}

// NewTickMsg creates a new tick message
func NewTickMsg() tea.Msg {
	return TickMsg(time.Now())
}
