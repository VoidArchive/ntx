package realtime

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"
)

// UpdateManager handles real-time market data updates with intelligent scheduling
// Adapts update frequency based on market hours, volatility, and data staleness
type UpdateManager struct {
	logger *slog.Logger
	config *Config
	mu     sync.RWMutex

	// Update scheduling
	activeTicker  *time.Ticker
	passiveTicker *time.Ticker
	stopChan      chan struct{}

	// Market state tracking
	currentState MarketState
	lastUpdate   time.Time
	updateCount  int64

	// Callback handlers
	updateHandler UpdateHandler
	stateHandler  StateChangeHandler

	// Performance tracking
	stats *UpdateStats
}

// Config holds real-time update configuration
type Config struct {
	// Active trading hours update interval (default: 15 seconds)
	ActiveUpdateInterval time.Duration

	// Non-trading hours update interval (default: 5 minutes)
	PassiveUpdateInterval time.Duration

	// Pre-market update interval (default: 1 minute)
	PreMarketUpdateInterval time.Duration

	// Post-market update interval (default: 2 minutes)
	PostMarketUpdateInterval time.Duration

	// Market hours (NPT timezone)
	MarketOpenHour  int // 11 AM
	MarketCloseHour int // 3 PM

	// Trading days (0=Sunday, 1=Monday, ..., 6=Saturday)
	TradingDays []int // [0,1,2,3,4] for Sun-Thu

	// Staleness threshold - trigger immediate update if data is older
	StalenessThreshold time.Duration

	// Maximum update failures before backing off
	MaxFailures int

	// Backoff multiplier for failed updates
	BackoffMultiplier float64

	// Enable adaptive updates based on volatility
	EnableAdaptiveUpdates bool

	// Enable pre/post market updates
	EnableExtendedHours bool
}

// MarketState represents current market operational state
type MarketState int

const (
	MarketClosed MarketState = iota
	PreMarket
	MarketOpen
	PostMarket
	MarketHoliday
)

func (ms MarketState) String() string {
	switch ms {
	case MarketClosed:
		return "closed"
	case PreMarket:
		return "pre_market"
	case MarketOpen:
		return "open"
	case PostMarket:
		return "post_market"
	case MarketHoliday:
		return "holiday"
	default:
		return "unknown"
	}
}

// UpdateHandler is called when market data should be updated
type UpdateHandler func(ctx context.Context, state MarketState) error

// StateChangeHandler is called when market state changes
type StateChangeHandler func(oldState, newState MarketState, timestamp time.Time)

// UpdateStats tracks update performance and reliability
type UpdateStats struct {
	mu sync.RWMutex

	// Update counters by state
	ActiveUpdates  int64 `json:"active_updates"`
	PassiveUpdates int64 `json:"passive_updates"`
	FailedUpdates  int64 `json:"failed_updates"`

	// Timing statistics
	LastUpdateTime    time.Time     `json:"last_update_time"`
	AvgUpdateDuration time.Duration `json:"avg_update_duration"`

	// Market state tracking
	StateChanges int64             `json:"state_changes"`
	StateHistory []StateTransition `json:"state_history"`
	CurrentState MarketState       `json:"current_state"`

	// Performance metrics
	UpdateSuccessRate   float64 `json:"update_success_rate"`
	ConsecutiveFailures int     `json:"consecutive_failures"`

	// Operational metrics
	StartTime   time.Time `json:"start_time"`
	UptimeHours float64   `json:"uptime_hours"`
}

// StateTransition records market state changes
type StateTransition struct {
	From      MarketState   `json:"from"`
	To        MarketState   `json:"to"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

// DefaultUpdateConfig returns default real-time update configuration
func DefaultUpdateConfig() *Config {
	return &Config{
		ActiveUpdateInterval:     15 * time.Second,
		PassiveUpdateInterval:    5 * time.Minute,
		PreMarketUpdateInterval:  1 * time.Minute,
		PostMarketUpdateInterval: 2 * time.Minute,
		MarketOpenHour:           11,
		MarketCloseHour:          15,
		TradingDays:              []int{0, 1, 2, 3, 4}, // Sun-Thu
		StalenessThreshold:       2 * time.Minute,
		MaxFailures:              5,
		BackoffMultiplier:        2.0,
		EnableAdaptiveUpdates:    true,
		EnableExtendedHours:      true,
	}
}

// NewUpdateManager creates a new real-time update manager
func NewUpdateManager(logger *slog.Logger, config *Config) *UpdateManager {
	if config == nil {
		config = DefaultUpdateConfig()
	}

	um := &UpdateManager{
		logger:   logger,
		config:   config,
		stopChan: make(chan struct{}),
		stats: &UpdateStats{
			StartTime:    time.Now(),
			StateHistory: make([]StateTransition, 0, 100),
			CurrentState: MarketClosed,
		},
	}

	// Initialize current market state
	um.currentState = um.determineMarketState(time.Now())
	um.stats.CurrentState = um.currentState

	logger.Info("Real-time update manager initialized",
		"current_state", um.currentState.String(),
		"active_interval", config.ActiveUpdateInterval,
		"passive_interval", config.PassiveUpdateInterval)

	return um
}

// Start begins real-time update scheduling based on market state
func (um *UpdateManager) Start(ctx context.Context, updateHandler UpdateHandler, stateHandler StateChangeHandler) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	um.updateHandler = updateHandler
	um.stateHandler = stateHandler

	um.logger.Info("Starting real-time update manager")

	// Start main update loop
	go um.updateLoop(ctx)

	// Start market state monitoring
	go um.stateMonitorLoop(ctx)

	return nil
}

// Stop stops the update manager and all running routines
func (um *UpdateManager) Stop() error {
	um.mu.Lock()
	defer um.mu.Unlock()

	um.logger.Info("Stopping real-time update manager")

	if um.activeTicker != nil {
		um.activeTicker.Stop()
	}
	if um.passiveTicker != nil {
		um.passiveTicker.Stop()
	}

	close(um.stopChan)
	return nil
}

// TriggerUpdate forces an immediate update regardless of schedule
func (um *UpdateManager) TriggerUpdate(ctx context.Context, reason string) error {
	um.logger.Info("Triggering immediate update", "reason", reason)

	if um.updateHandler == nil {
		return fmt.Errorf("no update handler configured")
	}

	return um.performUpdate(ctx, um.GetCurrentState())
}

// GetCurrentState returns the current market state
func (um *UpdateManager) GetCurrentState() MarketState {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return um.currentState
}

// GetStats returns current update statistics
func (um *UpdateManager) GetStats() *UpdateStats {
	um.stats.mu.RLock()
	defer um.stats.mu.RUnlock()

	// Create a copy without the mutex to avoid race conditions
	statsCopy := UpdateStats{
		ActiveUpdates:       um.stats.ActiveUpdates,
		PassiveUpdates:      um.stats.PassiveUpdates,
		FailedUpdates:       um.stats.FailedUpdates,
		LastUpdateTime:      um.stats.LastUpdateTime,
		AvgUpdateDuration:   um.stats.AvgUpdateDuration,
		StateChanges:        um.stats.StateChanges,
		StateHistory:        slices.Clone(um.stats.StateHistory),
		CurrentState:        um.stats.CurrentState,
		UpdateSuccessRate:   um.stats.UpdateSuccessRate,
		ConsecutiveFailures: um.stats.ConsecutiveFailures,
		StartTime:           um.stats.StartTime,
		UptimeHours:         time.Since(um.stats.StartTime).Hours(),
	}

	// Calculate success rate
	totalUpdates := um.stats.ActiveUpdates + um.stats.PassiveUpdates
	if totalUpdates > 0 {
		successUpdates := totalUpdates - um.stats.FailedUpdates
		statsCopy.UpdateSuccessRate = float64(successUpdates) / float64(totalUpdates) * 100.0
	}

	return &statsCopy
}

// SetStalenessThreshold updates the staleness threshold for triggering updates
func (um *UpdateManager) SetStalenessThreshold(threshold time.Duration) {
	um.mu.Lock()
	defer um.mu.Unlock()

	um.config.StalenessThreshold = threshold
	um.logger.Info("Staleness threshold updated", "threshold", threshold)
}

// IsDataStale checks if data is stale and needs immediate update
func (um *UpdateManager) IsDataStale() bool {
	um.mu.RLock()
	defer um.mu.RUnlock()

	if um.lastUpdate.IsZero() {
		return true
	}

	return time.Since(um.lastUpdate) > um.config.StalenessThreshold
}

// Private methods

// updateLoop handles scheduled updates based on market state
func (um *UpdateManager) updateLoop(ctx context.Context) {
	// Initial update
	if err := um.performUpdate(ctx, um.GetCurrentState()); err != nil {
		um.logger.Error("Initial update failed", "error", err)
	}

	for {
		interval := um.getUpdateInterval(um.GetCurrentState())
		ticker := time.NewTicker(interval)

		select {
		case <-ticker.C:
			state := um.GetCurrentState()
			if err := um.performUpdate(ctx, state); err != nil {
				um.logger.Error("Scheduled update failed",
					"state", state.String(), "error", err)
			}

		case <-um.stopChan:
			ticker.Stop()
			return

		case <-ctx.Done():
			ticker.Stop()
			return
		}

		ticker.Stop()
	}
}

// stateMonitorLoop monitors market state changes and adjusts update strategy
func (um *UpdateManager) stateMonitorLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check state every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			um.checkAndUpdateMarketState(time.Now())

		case <-um.stopChan:
			return

		case <-ctx.Done():
			return
		}
	}
}

// checkAndUpdateMarketState checks current time and updates market state if changed
func (um *UpdateManager) checkAndUpdateMarketState(now time.Time) {
	newState := um.determineMarketState(now)

	um.mu.Lock()
	oldState := um.currentState
	um.mu.Unlock()

	if newState != oldState {
		um.recordStateChange(oldState, newState, now)

		// Notify state change handler
		if um.stateHandler != nil {
			um.stateHandler(oldState, newState, now)
		}

		um.logger.Info("Market state changed",
			"from", oldState.String(),
			"to", newState.String())
	}
}

// determineMarketState calculates current market state based on time and configuration
func (um *UpdateManager) determineMarketState(now time.Time) MarketState {
	// Check if it's a trading day
	weekday := int(now.Weekday())
	isTradingDay := slices.Contains(um.config.TradingDays, weekday)

	if !isTradingDay {
		return MarketClosed
	}

	hour := now.Hour()

	// Market hours: 11:00-15:00 NPT
	if hour >= um.config.MarketOpenHour && hour < um.config.MarketCloseHour {
		return MarketOpen
	}

	// Extended hours (if enabled)
	if um.config.EnableExtendedHours {
		// Pre-market: 9:00-11:00
		if hour >= 9 && hour < um.config.MarketOpenHour {
			return PreMarket
		}

		// Post-market: 15:00-17:00
		if hour >= um.config.MarketCloseHour && hour < 17 {
			return PostMarket
		}
	}

	return MarketClosed
}

// getUpdateInterval returns appropriate update interval for current market state
func (um *UpdateManager) getUpdateInterval(state MarketState) time.Duration {
	switch state {
	case MarketOpen:
		return um.config.ActiveUpdateInterval
	case PreMarket:
		return um.config.PreMarketUpdateInterval
	case PostMarket:
		return um.config.PostMarketUpdateInterval
	default:
		return um.config.PassiveUpdateInterval
	}
}

// performUpdate executes the update handler and tracks performance
func (um *UpdateManager) performUpdate(ctx context.Context, state MarketState) error {
	startTime := time.Now()

	// Execute update handler
	err := um.updateHandler(ctx, state)

	// Update statistics
	duration := time.Since(startTime)
	um.recordUpdate(state, duration, err)

	if err != nil {
		um.logger.Warn("Update failed",
			"state", state.String(),
			"duration", duration,
			"error", err)
		return err
	}

	um.logger.Debug("Update completed",
		"state", state.String(),
		"duration", duration)

	return nil
}

// recordStateChange logs and tracks market state transitions
func (um *UpdateManager) recordStateChange(oldState, newState MarketState, timestamp time.Time) {
	um.mu.Lock()
	defer um.mu.Unlock()

	// Calculate duration in previous state
	var duration time.Duration
	if !um.lastUpdate.IsZero() {
		duration = timestamp.Sub(um.lastUpdate)
	}

	// Record transition
	transition := StateTransition{
		From:      oldState,
		To:        newState,
		Timestamp: timestamp,
		Duration:  duration,
	}

	// Update stats
	um.stats.mu.Lock()
	um.stats.StateChanges++
	um.stats.CurrentState = newState

	// Keep limited history
	if len(um.stats.StateHistory) >= 100 {
		um.stats.StateHistory = um.stats.StateHistory[1:]
	}
	um.stats.StateHistory = append(um.stats.StateHistory, transition)
	um.stats.mu.Unlock()

	// Update current state
	um.currentState = newState
}

// recordUpdate tracks update performance statistics
func (um *UpdateManager) recordUpdate(state MarketState, duration time.Duration, err error) {
	um.mu.Lock()
	um.lastUpdate = time.Now()
	um.updateCount++
	um.mu.Unlock()

	um.stats.mu.Lock()
	defer um.stats.mu.Unlock()

	// Update timing statistics
	um.stats.LastUpdateTime = time.Now()
	if um.stats.AvgUpdateDuration == 0 {
		um.stats.AvgUpdateDuration = duration
	} else {
		um.stats.AvgUpdateDuration = (um.stats.AvgUpdateDuration + duration) / 2
	}

	// Update counters based on state and result
	if err != nil {
		um.stats.FailedUpdates++
		um.stats.ConsecutiveFailures++
	} else {
		um.stats.ConsecutiveFailures = 0

		switch state {
		case MarketOpen:
			um.stats.ActiveUpdates++
		default:
			um.stats.PassiveUpdates++
		}
	}
}
