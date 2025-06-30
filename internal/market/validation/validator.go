package validation

import (
	"fmt"
	"log/slog"
	"math"
	"ntx/internal/data/models"
	"ntx/internal/market"
	"slices"
	"strings"
	"time"
)

// DataValidator provides comprehensive validation for market data integrity
// Ensures data quality, consistency, and business rule compliance
type DataValidator struct {
	logger *slog.Logger
	config *Config
	stats  *ValidationStats
}

// Config holds validation configuration and business rules
type Config struct {
	// Price validation rules
	MinPrice       models.Money // Minimum valid price (default: Rs. 1)
	MaxPrice       models.Money // Maximum valid price (default: Rs. 50,000)
	MaxPriceChange float64      // Maximum % change allowed (default: 50%)

	// Volume validation rules
	MinVolume int64 // Minimum valid volume (default: 0)
	MaxVolume int64 // Maximum valid volume (default: 100M)

	// Timestamp validation
	MaxDataAge      time.Duration // Maximum age for valid data (default: 24 hours)
	FutureThreshold time.Duration // Max time data can be in future (default: 5 minutes)

	// Symbol validation
	ValidSymbolPattern string   // Regex pattern for valid symbols
	KnownSymbols       []string // List of known valid symbols

	// Consistency checks
	EnableCrossValidation  bool // Validate against historical data
	EnableOutlierDetection bool // Detect statistical outliers
	EnableBusinessRules    bool // Apply NEPSE-specific business rules

	// Error handling
	StrictMode             bool // Reject any invalid data vs. warning only
	EnableAutomaticRepairs bool // Attempt to fix minor issues
}

// ValidationStats tracks validation performance and error patterns
type ValidationStats struct {
	TotalValidations  int64 `json:"total_validations"`
	PassedValidations int64 `json:"passed_validations"`
	FailedValidations int64 `json:"failed_validations"`
	RepairedData      int64 `json:"repaired_data"`

	// Error breakdown
	PriceErrors       int64 `json:"price_errors"`
	VolumeErrors      int64 `json:"volume_errors"`
	TimestampErrors   int64 `json:"timestamp_errors"`
	SymbolErrors      int64 `json:"symbol_errors"`
	ConsistencyErrors int64 `json:"consistency_errors"`

	// Performance metrics
	AvgValidationTime time.Duration `json:"avg_validation_time"`
	LastValidation    time.Time     `json:"last_validation"`
}

// ValidationResult represents the outcome of data validation
type ValidationResult struct {
	IsValid     bool                `json:"is_valid"`
	Errors      []ValidationError   `json:"errors"`
	Warnings    []ValidationWarning `json:"warnings"`
	Repairs     []DataRepair        `json:"repairs"`
	ProcessedAt time.Time           `json:"processed_at"`
}

// ValidationError represents a validation failure
type ValidationError struct {
	Type     ErrorType `json:"type"`
	Field    string    `json:"field"`
	Message  string    `json:"message"`
	Value    any       `json:"value"`
	Severity Severity  `json:"severity"`
}

// ValidationWarning represents a non-critical validation issue
type ValidationWarning struct {
	Type    WarningType `json:"type"`
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   any         `json:"value"`
}

// DataRepair represents an automatic data fix
type DataRepair struct {
	Field      string `json:"field"`
	OldValue   any    `json:"old_value"`
	NewValue   any    `json:"new_value"`
	RepairType string `json:"repair_type"`
	Applied    bool   `json:"applied"`
}

// ErrorType categorizes validation errors
type ErrorType string

const (
	ErrorTypeInvalidPrice     ErrorType = "invalid_price"
	ErrorTypeInvalidVolume    ErrorType = "invalid_volume"
	ErrorTypeInvalidTimestamp ErrorType = "invalid_timestamp"
	ErrorTypeInvalidSymbol    ErrorType = "invalid_symbol"
	ErrorTypeInconsistentData ErrorType = "inconsistent_data"
	ErrorTypeBusinessRule     ErrorType = "business_rule"
	ErrorTypeOutlier          ErrorType = "outlier"
)

// WarningType categorizes validation warnings
type WarningType string

const (
	WarningTypeStaleData          WarningType = "stale_data"
	WarningTypeUnusualPattern     WarningType = "unusual_pattern"
	WarningTypeMinorInconsistency WarningType = "minor_inconsistency"
)

// Severity levels for validation issues
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *Config {
	return &Config{
		MinPrice:               models.NewMoneyFromRupees(1.0),
		MaxPrice:               models.NewMoneyFromRupees(50000.0),
		MaxPriceChange:         50.0, // 50% max change
		MinVolume:              0,
		MaxVolume:              100000000, // 100M shares
		MaxDataAge:             24 * time.Hour,
		FutureThreshold:        5 * time.Minute,
		ValidSymbolPattern:     `^[A-Z]{2,10}$`,
		EnableCrossValidation:  true,
		EnableOutlierDetection: true,
		EnableBusinessRules:    true,
		StrictMode:             false,
		EnableAutomaticRepairs: true,
	}
}

// NewDataValidator creates a new data validator
func NewDataValidator(logger *slog.Logger, config *Config, knownSymbols []string) *DataValidator {
	if config == nil {
		config = DefaultValidationConfig()
	}

	config.KnownSymbols = knownSymbols

	validator := &DataValidator{
		logger: logger,
		config: config,
		stats:  &ValidationStats{},
	}

	logger.Info("Data validator initialized",
		"strict_mode", config.StrictMode,
		"auto_repair", config.EnableAutomaticRepairs,
		"known_symbols", len(knownSymbols))

	return validator
}

// ValidateMarketData performs comprehensive validation on market data
func (dv *DataValidator) ValidateMarketData(data *market.ScrapedData, historicalData *market.ScrapedData) *ValidationResult {
	startTime := time.Now()
	defer func() {
		dv.updateStats(time.Since(startTime))
	}()

	result := &ValidationResult{
		IsValid:     true,
		Errors:      make([]ValidationError, 0),
		Warnings:    make([]ValidationWarning, 0),
		Repairs:     make([]DataRepair, 0),
		ProcessedAt: startTime,
	}

	// Basic field validations
	dv.validateSymbol(data, result)
	dv.validatePrices(data, result)
	dv.validateVolume(data, result)
	dv.validateTimestamp(data, result)

	// Cross-validation with historical data
	if dv.config.EnableCrossValidation && historicalData != nil {
		dv.validateConsistency(data, historicalData, result)
	}

	// Outlier detection
	if dv.config.EnableOutlierDetection && historicalData != nil {
		dv.detectOutliers(data, historicalData, result)
	}

	// Business rule validation
	if dv.config.EnableBusinessRules {
		dv.validateBusinessRules(data, result)
	}

	// Apply automatic repairs if enabled
	if dv.config.EnableAutomaticRepairs {
		dv.applyAutomaticRepairs(data, result)
	}

	// Determine final validity
	result.IsValid = len(result.Errors) == 0 || (!dv.config.StrictMode && dv.onlyMinorErrors(result.Errors))

	// Log results
	if !result.IsValid {
		dv.logger.Warn("Data validation failed",
			"symbol", data.Symbol,
			"errors", len(result.Errors),
			"warnings", len(result.Warnings))
	} else if len(result.Warnings) > 0 {
		dv.logger.Debug("Data validation passed with warnings",
			"symbol", data.Symbol,
			"warnings", len(result.Warnings))
	}

	return result
}

// ValidateBatch validates multiple market data entries efficiently
func (dv *DataValidator) ValidateBatch(dataList []market.ScrapedData, historicalMap map[string]*market.ScrapedData) map[string]*ValidationResult {
	results := make(map[string]*ValidationResult)

	dv.logger.Info("Starting batch validation", "count", len(dataList))

	for _, data := range dataList {
		var historical *market.ScrapedData
		if historicalMap != nil {
			historical = historicalMap[data.Symbol]
		}

		results[data.Symbol] = dv.ValidateMarketData(&data, historical)
	}

	// Log batch summary
	valid, invalid, warnings := dv.summarizeBatchResults(results)
	dv.logger.Info("Batch validation completed",
		"valid", valid,
		"invalid", invalid,
		"warnings", warnings)

	return results
}

// GetValidationStats returns current validation statistics
func (dv *DataValidator) GetValidationStats() *ValidationStats {
	// Create a copy to avoid race conditions
	statsCopy := *dv.stats
	return &statsCopy
}

// GetValidationSummary returns a summary of validation health
func (dv *DataValidator) GetValidationSummary() map[string]any {
	successRate := float64(0)
	if dv.stats.TotalValidations > 0 {
		successRate = float64(dv.stats.PassedValidations) / float64(dv.stats.TotalValidations) * 100
	}

	return map[string]any{
		"success_rate":        successRate,
		"total_validations":   dv.stats.TotalValidations,
		"recent_failures":     dv.stats.FailedValidations,
		"data_repairs":        dv.stats.RepairedData,
		"avg_validation_time": dv.stats.AvgValidationTime,
		"last_validation":     dv.stats.LastValidation,
	}
}

// Private validation methods

// validateSymbol checks symbol format and validity
func (dv *DataValidator) validateSymbol(data *market.ScrapedData, result *ValidationResult) {
	symbol := strings.TrimSpace(strings.ToUpper(data.Symbol))

	// Check basic format
	if len(symbol) < 2 || len(symbol) > 10 {
		result.Errors = append(result.Errors, ValidationError{
			Type:     ErrorTypeInvalidSymbol,
			Field:    "symbol",
			Message:  fmt.Sprintf("Symbol length must be 2-10 characters, got %d", len(symbol)),
			Value:    symbol,
			Severity: SeverityCritical,
		})
		return
	}

	// Check against known symbols
	if len(dv.config.KnownSymbols) > 0 {
		isKnown := slices.Contains(dv.config.KnownSymbols, symbol)

		if !isKnown {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    WarningTypeUnusualPattern,
				Field:   "symbol",
				Message: "Symbol not in known symbols list",
				Value:   symbol,
			})
		}
	}

	// Apply repair if needed
	if data.Symbol != symbol && dv.config.EnableAutomaticRepairs {
		result.Repairs = append(result.Repairs, DataRepair{
			Field:      "symbol",
			OldValue:   data.Symbol,
			NewValue:   symbol,
			RepairType: "normalize_case",
			Applied:    false,
		})
	}
}

// validatePrices checks price values and relationships
func (dv *DataValidator) validatePrices(data *market.ScrapedData, result *ValidationResult) {
	// Validate last price
	if data.LastPrice < dv.config.MinPrice || data.LastPrice > dv.config.MaxPrice {
		result.Errors = append(result.Errors, ValidationError{
			Type:  ErrorTypeInvalidPrice,
			Field: "last_price",
			Message: fmt.Sprintf("Price outside valid range [%s, %s]",
				dv.config.MinPrice.FormattedString(),
				dv.config.MaxPrice.FormattedString()),
			Value:    data.LastPrice.FormattedString(),
			Severity: SeverityHigh,
		})
	}

	// Validate price consistency (OHLC relationships)
	if data.High.IsPositive() && data.Low.IsPositive() && data.Open.IsPositive() {
		if data.High < data.Low {
			result.Errors = append(result.Errors, ValidationError{
				Type:     ErrorTypeInconsistentData,
				Field:    "high_low",
				Message:  "High price cannot be less than low price",
				Value:    fmt.Sprintf("High: %s, Low: %s", data.High.FormattedString(), data.Low.FormattedString()),
				Severity: SeverityHigh,
			})
		}

		if data.LastPrice < data.Low || data.LastPrice > data.High {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    WarningTypeMinorInconsistency,
				Field:   "last_price_range",
				Message: "Last price outside day's high-low range",
				Value: fmt.Sprintf("LTP: %s, Range: [%s, %s]",
					data.LastPrice.FormattedString(),
					data.Low.FormattedString(),
					data.High.FormattedString()),
			})
		}
	}
}

// validateVolume checks volume validity
func (dv *DataValidator) validateVolume(data *market.ScrapedData, result *ValidationResult) {
	if data.Volume < dv.config.MinVolume || data.Volume > dv.config.MaxVolume {
		result.Errors = append(result.Errors, ValidationError{
			Type:  ErrorTypeInvalidVolume,
			Field: "volume",
			Message: fmt.Sprintf("Volume outside valid range [%d, %d]",
				dv.config.MinVolume, dv.config.MaxVolume),
			Value:    data.Volume,
			Severity: SeverityMedium,
		})
	}
}

// validateTimestamp checks timestamp validity and freshness
func (dv *DataValidator) validateTimestamp(data *market.ScrapedData, result *ValidationResult) {
	now := time.Now()

	// Check if timestamp is in future
	if data.ScrapedAt.After(now.Add(dv.config.FutureThreshold)) {
		result.Errors = append(result.Errors, ValidationError{
			Type:     ErrorTypeInvalidTimestamp,
			Field:    "scraped_at",
			Message:  "Timestamp is too far in the future",
			Value:    data.ScrapedAt,
			Severity: SeverityMedium,
		})
	}

	// Check if data is stale
	age := now.Sub(data.ScrapedAt)
	if age > dv.config.MaxDataAge {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    WarningTypeStaleData,
			Field:   "scraped_at",
			Message: fmt.Sprintf("Data is stale (age: %v)", age),
			Value:   data.ScrapedAt,
		})
	}
}

// validateConsistency checks consistency with historical data
func (dv *DataValidator) validateConsistency(current, historical *market.ScrapedData, result *ValidationResult) {
	if historical == nil || historical.LastPrice.IsZero() {
		return
	}

	// Calculate price change percentage
	changePercent := math.Abs(float64(current.LastPrice-historical.LastPrice)) / float64(historical.LastPrice) * 100

	if changePercent > dv.config.MaxPriceChange {
		result.Errors = append(result.Errors, ValidationError{
			Type:  ErrorTypeInconsistentData,
			Field: "price_change",
			Message: fmt.Sprintf("Price change too large: %.2f%% (max: %.2f%%)",
				changePercent, dv.config.MaxPriceChange),
			Value:    changePercent,
			Severity: SeverityHigh,
		})
	}
}

// detectOutliers identifies statistical outliers in the data
func (dv *DataValidator) detectOutliers(current, historical *market.ScrapedData, result *ValidationResult) {
	// Simple outlier detection - could be enhanced with more sophisticated algorithms

	// Volume outlier detection
	if historical.Volume > 0 {
		volumeRatio := float64(current.Volume) / float64(historical.Volume)
		if volumeRatio > 10.0 || volumeRatio < 0.1 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    WarningTypeUnusualPattern,
				Field:   "volume",
				Message: fmt.Sprintf("Unusual volume pattern (ratio: %.2fx)", volumeRatio),
				Value:   current.Volume,
			})
		}
	}
}

// validateBusinessRules applies NEPSE-specific business rules
func (dv *DataValidator) validateBusinessRules(data *market.ScrapedData, result *ValidationResult) {
	// NEPSE-specific rules

	// Rule: Trading volume should be positive during market hours
	if data.Volume <= 0 {
		// This could be acceptable for pre/post market or holidays
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    WarningTypeMinorInconsistency,
			Field:   "volume",
			Message: "Zero trading volume",
			Value:   data.Volume,
		})
	}

	// Rule: Change amount should be consistent with price difference
	if data.PrevClose.IsPositive() {
		expectedChange := data.LastPrice.Subtract(data.PrevClose)
		actualChange := data.ChangeAmount

		// Allow small rounding differences
		tolerance := models.NewMoneyFromPaisa(1) // 1 paisa tolerance
		if expectedChange.Subtract(actualChange).Abs() > tolerance {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    WarningTypeMinorInconsistency,
				Field:   "change_amount",
				Message: "Change amount inconsistent with price difference",
				Value: fmt.Sprintf("Expected: %s, Actual: %s",
					expectedChange.FormattedString(),
					actualChange.FormattedString()),
			})
		}
	}
}

// applyAutomaticRepairs attempts to fix minor data issues
func (dv *DataValidator) applyAutomaticRepairs(data *market.ScrapedData, result *ValidationResult) {
	for i := range result.Repairs {
		repair := &result.Repairs[i]

		switch repair.Field {
		case "symbol":
			if repair.RepairType == "normalize_case" {
				data.Symbol = repair.NewValue.(string)
				repair.Applied = true
				dv.stats.RepairedData++
			}
		}
	}
}

// Helper methods

// onlyMinorErrors checks if all errors are non-critical
func (dv *DataValidator) onlyMinorErrors(errors []ValidationError) bool {
	for _, err := range errors {
		if err.Severity == SeverityCritical || err.Severity == SeverityHigh {
			return false
		}
	}
	return true
}

// summarizeBatchResults counts validation outcomes
func (dv *DataValidator) summarizeBatchResults(results map[string]*ValidationResult) (valid, invalid, warnings int) {
	for _, result := range results {
		if result.IsValid {
			valid++
		} else {
			invalid++
		}
		if len(result.Warnings) > 0 {
			warnings++
		}
	}
	return
}

// updateStats tracks validation performance
func (dv *DataValidator) updateStats(duration time.Duration) {
	dv.stats.TotalValidations++
	dv.stats.LastValidation = time.Now()

	// Update average validation time
	if dv.stats.AvgValidationTime == 0 {
		dv.stats.AvgValidationTime = duration
	} else {
		dv.stats.AvgValidationTime = (dv.stats.AvgValidationTime + duration) / 2
	}
}
