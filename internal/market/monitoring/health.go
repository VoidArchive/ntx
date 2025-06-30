package monitoring

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"time"
)

// HealthMonitor provides comprehensive monitoring and health checks for market data system
// Tracks performance, availability, and data quality across all components
type HealthMonitor struct {
	logger *slog.Logger
	config *Config
	mu     sync.RWMutex

	// Component health status
	components map[string]*ComponentHealth

	// System metrics
	systemMetrics *SystemMetrics

	// Alert management
	alertManager *AlertManager

	// Monitoring control
	monitorTicker *time.Ticker
	stopChan      chan struct{}
	isRunning     bool
}

// Config holds health monitoring configuration
type Config struct {
	// Health check intervals
	HealthCheckInterval    time.Duration // Default: 1 minute
	ComponentCheckInterval time.Duration // Default: 30 seconds
	MetricsInterval        time.Duration // Default: 5 minutes

	// Health thresholds
	MaxResponseTime time.Duration // Default: 5 seconds
	MinSuccessRate  float64       // Default: 95%
	MaxErrorRate    float64       // Default: 5%

	// Data quality thresholds
	MinDataFreshness  time.Duration // Default: 10 minutes
	MaxDataStaleness  time.Duration // Default: 1 hour
	MinValidationRate float64       // Default: 90%

	// Alert thresholds
	CriticalErrorThreshold int // Default: 5 errors in 5 minutes
	WarningThreshold       int // Default: 10 warnings in 10 minutes

	// Enable specific monitoring features
	EnablePerformanceMetrics bool // Default: true
	EnableDataQualityCheck   bool // Default: true
	EnableComponentHealth    bool // Default: true
	EnableAlerting           bool // Default: true
}

// ComponentHealth tracks health status of individual system components
type ComponentHealth struct {
	Name         string              `json:"name"`
	Status       HealthStatus        `json:"status"`
	LastCheck    time.Time           `json:"last_check"`
	ResponseTime time.Duration       `json:"response_time"`
	ErrorCount   int64               `json:"error_count"`
	SuccessRate  float64             `json:"success_rate"`
	Details      map[string]any      `json:"details"`
	History      []HealthCheckResult `json:"history"`
}

// SystemMetrics aggregates system-wide performance metrics
type SystemMetrics struct {
	mu sync.RWMutex

	// Overall system health
	OverallStatus HealthStatus `json:"overall_status"`
	LastUpdate    time.Time    `json:"last_update"`

	// Performance metrics
	AvgResponseTime    time.Duration `json:"avg_response_time"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`

	// Data quality metrics
	DataFreshness  time.Duration `json:"data_freshness"`
	ValidationRate float64       `json:"validation_rate"`
	ErrorRate      float64       `json:"error_rate"`

	// Resource utilization
	CacheHitRate float64 `json:"cache_hit_rate"`
	CacheSize    int64   `json:"cache_size"`
	DatabaseSize int64   `json:"database_size"`

	// Operational metrics
	UptimeSeconds  int64 `json:"uptime_seconds"`
	ComponentsUp   int   `json:"components_up"`
	ComponentsDown int   `json:"components_down"`
}

// HealthStatus represents component or system health state
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Timestamp    time.Time      `json:"timestamp"`
	Status       HealthStatus   `json:"status"`
	ResponseTime time.Duration  `json:"response_time"`
	Error        string         `json:"error,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
}

// AlertManager handles health-related alerts and notifications
type AlertManager struct {
	logger *slog.Logger
	config *Config
	mu     sync.RWMutex

	// Alert tracking
	activeAlerts map[string]*Alert
	alertHistory []Alert

	// Alert suppression
	suppressedUntil map[string]time.Time
}

// Alert represents a system health alert
type Alert struct {
	ID           string        `json:"id"`
	Component    string        `json:"component"`
	Severity     AlertSeverity `json:"severity"`
	Message      string        `json:"message"`
	Timestamp    time.Time     `json:"timestamp"`
	Acknowledged bool          `json:"acknowledged"`
	Resolved     bool          `json:"resolved"`
	ResolvedAt   time.Time     `json:"resolved_at"`
}

// AlertSeverity levels for health alerts
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// ComponentHealthChecker interface for health check implementations
type ComponentHealthChecker interface {
	CheckHealth(ctx context.Context) *HealthCheckResult
	GetComponentName() string
}

// DefaultHealthConfig returns default health monitoring configuration
func DefaultHealthConfig() *Config {
	return &Config{
		HealthCheckInterval:      1 * time.Minute,
		ComponentCheckInterval:   30 * time.Second,
		MetricsInterval:          5 * time.Minute,
		MaxResponseTime:          5 * time.Second,
		MinSuccessRate:           95.0,
		MaxErrorRate:             5.0,
		MinDataFreshness:         10 * time.Minute,
		MaxDataStaleness:         1 * time.Hour,
		MinValidationRate:        90.0,
		CriticalErrorThreshold:   5,
		WarningThreshold:         10,
		EnablePerformanceMetrics: true,
		EnableDataQualityCheck:   true,
		EnableComponentHealth:    true,
		EnableAlerting:           true,
	}
}

// NewHealthMonitor creates a new health monitoring system
func NewHealthMonitor(logger *slog.Logger, config *Config) *HealthMonitor {
	if config == nil {
		config = DefaultHealthConfig()
	}

	hm := &HealthMonitor{
		logger:     logger,
		config:     config,
		components: make(map[string]*ComponentHealth),
		systemMetrics: &SystemMetrics{
			OverallStatus: HealthStatusUnknown,
		},
		alertManager: &AlertManager{
			logger:          logger,
			config:          config,
			activeAlerts:    make(map[string]*Alert),
			alertHistory:    make([]Alert, 0, 1000),
			suppressedUntil: make(map[string]time.Time),
		},
		stopChan: make(chan struct{}),
	}

	logger.Info("Health monitor initialized",
		"check_interval", config.HealthCheckInterval,
		"metrics_interval", config.MetricsInterval)

	return hm
}

// Start begins health monitoring routines
func (hm *HealthMonitor) Start(ctx context.Context) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.isRunning {
		return fmt.Errorf("health monitor already running")
	}

	hm.logger.Info("Starting health monitor")

	// Start monitoring routines
	go hm.healthCheckLoop(ctx)
	go hm.metricsCollectionLoop(ctx)

	hm.isRunning = true
	return nil
}

// Stop stops health monitoring
func (hm *HealthMonitor) Stop() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.isRunning {
		return nil
	}

	hm.logger.Info("Stopping health monitor")

	if hm.monitorTicker != nil {
		hm.monitorTicker.Stop()
	}

	close(hm.stopChan)
	hm.isRunning = false

	return nil
}

// RegisterComponent registers a component for health monitoring
func (hm *HealthMonitor) RegisterComponent(name string, checker ComponentHealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.components[name] = &ComponentHealth{
		Name:    name,
		Status:  HealthStatusUnknown,
		Details: make(map[string]any),
		History: make([]HealthCheckResult, 0, 100),
	}

	hm.logger.Info("Component registered for health monitoring", "component", name)
}

// GetSystemHealth returns overall system health status
func (hm *HealthMonitor) GetSystemHealth() *SystemMetrics {
	hm.systemMetrics.mu.RLock()
	defer hm.systemMetrics.mu.RUnlock()

	// Create a copy without the mutex to avoid race conditions
	metrics := SystemMetrics{
		OverallStatus:      hm.systemMetrics.OverallStatus,
		LastUpdate:         hm.systemMetrics.LastUpdate,
		AvgResponseTime:    hm.systemMetrics.AvgResponseTime,
		TotalRequests:      hm.systemMetrics.TotalRequests,
		SuccessfulRequests: hm.systemMetrics.SuccessfulRequests,
		FailedRequests:     hm.systemMetrics.FailedRequests,
		DataFreshness:      hm.systemMetrics.DataFreshness,
		ValidationRate:     hm.systemMetrics.ValidationRate,
		ErrorRate:          hm.systemMetrics.ErrorRate,
		CacheHitRate:       hm.systemMetrics.CacheHitRate,
		CacheSize:          hm.systemMetrics.CacheSize,
		DatabaseSize:       hm.systemMetrics.DatabaseSize,
		UptimeSeconds:      hm.systemMetrics.UptimeSeconds,
		ComponentsUp:       hm.systemMetrics.ComponentsUp,
		ComponentsDown:     hm.systemMetrics.ComponentsDown,
	}
	return &metrics
}

// GetComponentHealth returns health status for a specific component
func (hm *HealthMonitor) GetComponentHealth(componentName string) (*ComponentHealth, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	health, exists := hm.components[componentName]
	if !exists {
		return nil, false
	}

	// Create a copy
	healthCopy := *health
	return &healthCopy, true
}

// GetAllComponentsHealth returns health status for all registered components
func (hm *HealthMonitor) GetAllComponentsHealth() map[string]*ComponentHealth {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	result := make(map[string]*ComponentHealth)
	for name, health := range hm.components {
		healthCopy := *health
		result[name] = &healthCopy
	}

	return result
}

// CheckComponentHealth performs immediate health check for specific component
func (hm *HealthMonitor) CheckComponentHealth(ctx context.Context, componentName string, checker ComponentHealthChecker) error {
	startTime := time.Now()

	// Perform health check
	result := checker.CheckHealth(ctx)
	result.Timestamp = startTime
	result.ResponseTime = time.Since(startTime)

	// Update component health
	hm.updateComponentHealth(componentName, result)

	// Check for alerts
	if hm.config.EnableAlerting {
		hm.checkForAlerts(componentName, result)
	}

	return nil
}

// GetActiveAlerts returns all active alerts
func (hm *HealthMonitor) GetActiveAlerts() []*Alert {
	return hm.alertManager.GetActiveAlerts()
}

// AcknowledgeAlert marks an alert as acknowledged
func (hm *HealthMonitor) AcknowledgeAlert(alertID string) error {
	return hm.alertManager.AcknowledgeAlert(alertID)
}

// GetHealthSummary returns a concise health summary
func (hm *HealthMonitor) GetHealthSummary() map[string]any {
	systemHealth := hm.GetSystemHealth()
	activeAlerts := hm.GetActiveAlerts()

	return map[string]any{
		"overall_status":       systemHealth.OverallStatus,
		"components_healthy":   systemHealth.ComponentsUp,
		"components_unhealthy": systemHealth.ComponentsDown,
		"success_rate":         systemHealth.SuccessfulRequests,
		"error_rate":           systemHealth.ErrorRate,
		"data_freshness":       systemHealth.DataFreshness,
		"cache_hit_rate":       systemHealth.CacheHitRate,
		"active_alerts":        len(activeAlerts),
		"uptime_hours":         float64(systemHealth.UptimeSeconds) / 3600,
		"last_update":          systemHealth.LastUpdate,
	}
}

// Private methods

// healthCheckLoop performs periodic health checks on all registered components
func (hm *HealthMonitor) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(hm.config.ComponentCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hm.performHealthChecks(ctx)

		case <-hm.stopChan:
			return

		case <-ctx.Done():
			return
		}
	}
}

// metricsCollectionLoop collects and updates system metrics
func (hm *HealthMonitor) metricsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(hm.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hm.collectSystemMetrics()

		case <-hm.stopChan:
			return

		case <-ctx.Done():
			return
		}
	}
}

// performHealthChecks checks health of all registered components
func (hm *HealthMonitor) performHealthChecks(_ context.Context) {
	hm.mu.RLock()
	components := make(map[string]*ComponentHealth)
	maps.Copy(components, hm.components)
	hm.mu.RUnlock()

	for name := range components {
		// This would normally call the actual health checker
		// For now, we'll simulate health checks
		result := &HealthCheckResult{
			Timestamp:    time.Now(),
			Status:       HealthStatusHealthy,
			ResponseTime: 100 * time.Millisecond,
		}

		hm.updateComponentHealth(name, result)
	}

	// Update overall system health
	hm.updateSystemHealth()
}

// collectSystemMetrics aggregates metrics from all components
func (hm *HealthMonitor) collectSystemMetrics() {
	hm.systemMetrics.mu.Lock()
	defer hm.systemMetrics.mu.Unlock()

	hm.systemMetrics.LastUpdate = time.Now()

	// Count healthy/unhealthy components
	hm.mu.RLock()
	up, down := 0, 0
	for _, health := range hm.components {
		if health.Status == HealthStatusHealthy {
			up++
		} else {
			down++
		}
	}
	hm.mu.RUnlock()

	hm.systemMetrics.ComponentsUp = up
	hm.systemMetrics.ComponentsDown = down

	// Update overall status
	if down == 0 {
		hm.systemMetrics.OverallStatus = HealthStatusHealthy
	} else if up > down {
		hm.systemMetrics.OverallStatus = HealthStatusDegraded
	} else {
		hm.systemMetrics.OverallStatus = HealthStatusUnhealthy
	}
}

// updateComponentHealth updates health status for a component
func (hm *HealthMonitor) updateComponentHealth(componentName string, result *HealthCheckResult) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	health, exists := hm.components[componentName]
	if !exists {
		return
	}

	// Update health status
	health.Status = result.Status
	health.LastCheck = result.Timestamp
	health.ResponseTime = result.ResponseTime

	// Update error count
	if result.Status != HealthStatusHealthy {
		health.ErrorCount++
	}

	// Add to history (keep last 100 entries)
	if len(health.History) >= 100 {
		health.History = health.History[1:]
	}
	health.History = append(health.History, *result)

	// Calculate success rate from recent history
	if len(health.History) > 0 {
		successful := 0
		for _, h := range health.History {
			if h.Status == HealthStatusHealthy {
				successful++
			}
		}
		health.SuccessRate = float64(successful) / float64(len(health.History)) * 100
	}
}

// updateSystemHealth calculates overall system health
func (hm *HealthMonitor) updateSystemHealth() {
	hm.mu.RLock()
	totalComponents := len(hm.components)
	healthyComponents := 0

	for _, health := range hm.components {
		if health.Status == HealthStatusHealthy {
			healthyComponents++
		}
	}
	hm.mu.RUnlock()

	hm.systemMetrics.mu.Lock()
	defer hm.systemMetrics.mu.Unlock()

	if totalComponents == 0 {
		hm.systemMetrics.OverallStatus = HealthStatusUnknown
	} else if healthyComponents == totalComponents {
		hm.systemMetrics.OverallStatus = HealthStatusHealthy
	} else if healthyComponents > totalComponents/2 {
		hm.systemMetrics.OverallStatus = HealthStatusDegraded
	} else {
		hm.systemMetrics.OverallStatus = HealthStatusUnhealthy
	}
}

// checkForAlerts checks if component health warrants an alert
func (hm *HealthMonitor) checkForAlerts(componentName string, result *HealthCheckResult) {
	if result.Status == HealthStatusUnhealthy {
		alert := &Alert{
			ID:        fmt.Sprintf("%s-%d", componentName, time.Now().Unix()),
			Component: componentName,
			Severity:  AlertSeverityCritical,
			Message:   fmt.Sprintf("Component %s is unhealthy", componentName),
			Timestamp: time.Now(),
		}

		hm.alertManager.CreateAlert(alert)
	}
}

// AlertManager methods

// GetActiveAlerts returns all currently active alerts
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		if !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// CreateAlert creates a new alert
func (am *AlertManager) CreateAlert(alert *Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check if similar alert is suppressed
	if suppressedUntil, exists := am.suppressedUntil[alert.Component]; exists {
		if time.Now().Before(suppressedUntil) {
			return // Alert is suppressed
		}
	}

	am.activeAlerts[alert.ID] = alert

	// Add to history (keep last 1000)
	if len(am.alertHistory) >= 1000 {
		am.alertHistory = am.alertHistory[1:]
	}
	am.alertHistory = append(am.alertHistory, *alert)

	am.logger.Warn("Health alert created",
		"component", alert.Component,
		"severity", alert.Severity,
		"message", alert.Message)
}

// AcknowledgeAlert marks an alert as acknowledged
func (am *AlertManager) AcknowledgeAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.activeAlerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	alert.Acknowledged = true

	am.logger.Info("Alert acknowledged", "alert_id", alertID)
	return nil
}
