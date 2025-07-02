/*
NTX Portfolio Management TUI - Sparkline Charts

Mini trend visualization for individual holdings and portfolio metrics.
Compact ASCII sparklines provide at-a-glance trend analysis within
table cells and dashboard widgets.

Optimized for financial time series data with intelligent scaling
and theme-aware coloring for gains/losses visualization.
*/

package charts

import (
	"strings"
)

// Sparkline renders compact trend charts for financial data
// Single-line visualization perfect for table integration and dashboard widgets
type Sparkline struct {
	renderer *ChartRenderer
	data     ChartData
	config   ChartConfig
}

// NewSparkline creates sparkline chart with configuration
func NewSparkline(config ChartConfig) *Sparkline {
	renderer := NewChartRenderer(config, ChartData{})
	
	return &Sparkline{
		renderer: renderer,
		config:   config,
	}
}

// SetData updates sparkline data with validation
// Recalculates data range for accurate visual scaling
func (s *Sparkline) SetData(data ChartData) error {
	if err := ValidateData(data); err != nil {
		return err
	}
	
	// Calculate data range for scaling
	if len(data.Values) > 0 {
		min, max := CalculateDataRange(data.Values)
		data.Min = min
		data.Max = max
	}
	
	s.data = data
	s.renderer.data = data
	
	return nil
}

// SetConfig updates rendering configuration
func (s *Sparkline) SetConfig(config ChartConfig) {
	s.config = config
	s.renderer.config = config
}

// GetMinSize returns minimum dimensions for sparkline
// Compact design requires minimal space while remaining readable
func (s *Sparkline) GetMinSize() (width, height int) {
	return 10, 1 // Minimum 10 characters wide, 1 line tall
}

// Render generates ASCII sparkline with theme coloring
// Single line visualization with density-based character selection
func (s *Sparkline) Render() string {
	if len(s.data.Values) == 0 {
		return s.renderEmpty()
	}
	
	// Handle single value case
	if len(s.data.Values) == 1 {
		return s.renderSingleValue()
	}
	
	// Calculate sparkline width (reserve space for title if present)
	sparklineWidth := s.config.Width
	if s.data.Title != "" && s.config.ShowLabels {
		titleLength := len([]rune(s.data.Title))
		if titleLength+2 < s.config.Width {
			sparklineWidth = s.config.Width - titleLength - 2
		}
	}
	
	// Ensure minimum width
	if sparklineWidth < 3 {
		sparklineWidth = 3
	}
	
	// Generate sparkline characters
	sparkline := s.generateSparkline(sparklineWidth)
	
	// Add title if configured
	if s.data.Title != "" && s.config.ShowLabels {
		title := s.renderer.TruncateText(s.data.Title, 15)
		styledTitle := s.renderer.ApplyThemeStyle(title, StyleTitle)
		return styledTitle + ": " + sparkline
	}
	
	return sparkline
}

// generateSparkline creates character-based trend visualization
// Maps data values to block density characters for smooth representation
func (s *Sparkline) generateSparkline(width int) string {
	if len(s.data.Values) == 0 {
		return strings.Repeat(" ", width)
	}
	
	var result strings.Builder
	
	// Handle case where we have fewer data points than width
	if len(s.data.Values) <= width {
		// Direct mapping: one character per data point
		for i, value := range s.data.Values {
			if i >= width {
				break
			}
			
			char := s.valueToSparkChar(value)
			styledChar := s.applyValueStyling(char, value)
			result.WriteString(styledChar)
		}
		
		// Pad remaining width with spaces
		remaining := width - len(s.data.Values)
		if remaining > 0 {
			result.WriteString(strings.Repeat(" ", remaining))
		}
	} else {
		// Downsample data to fit width
		for i := 0; i < width; i++ {
			// Calculate which data point(s) this character represents
			dataIndex := float64(i) * float64(len(s.data.Values)-1) / float64(width-1)
			
			// Use nearest neighbor for simplicity
			nearestIndex := int(dataIndex + 0.5)
			if nearestIndex >= len(s.data.Values) {
				nearestIndex = len(s.data.Values) - 1
			}
			
			value := s.data.Values[nearestIndex]
			char := s.valueToSparkChar(value)
			styledChar := s.applyValueStyling(char, value)
			result.WriteString(styledChar)
		}
	}
	
	return result.String()
}

// valueToSparkChar maps numeric value to sparkline character
// Uses data range normalization for consistent visual scaling
func (s *Sparkline) valueToSparkChar(value float64) string {
	if s.data.Max == s.data.Min {
		// No variation - use middle character
		return ASCIIChars.SparkBlocks[len(ASCIIChars.SparkBlocks)/2]
	}
	
	// Normalize value to 0-1 range
	normalized := (value - s.data.Min) / (s.data.Max - s.data.Min)
	
	// Map to character index (0 = lowest, max index = highest)
	charIndex := int(normalized * float64(len(ASCIIChars.SparkBlocks)-1))
	
	// Clamp to valid range
	if charIndex < 0 {
		charIndex = 0
	}
	if charIndex >= len(ASCIIChars.SparkBlocks) {
		charIndex = len(ASCIIChars.SparkBlocks) - 1
	}
	
	return ASCIIChars.SparkBlocks[charIndex]
}

// applyValueStyling applies theme colors based on value sentiment
// Financial data coloring: green for gains, red for losses
func (s *Sparkline) applyValueStyling(char string, value float64) string {
	// Determine value sentiment for coloring
	if s.isPositiveValue(value) {
		return s.renderer.ApplyThemeStyle(char, StylePositive)
	} else if s.isNegativeValue(value) {
		return s.renderer.ApplyThemeStyle(char, StyleNegative)
	}
	
	return s.renderer.ApplyThemeStyle(char, StyleData)
}

// isPositiveValue determines if value represents a gain
// Contextual evaluation based on data range and typical financial metrics
func (s *Sparkline) isPositiveValue(value float64) bool {
	// If we have both positive and negative values, zero is the threshold
	if s.data.Min < 0 && s.data.Max > 0 {
		return value > 0
	}
	
	// If all positive, compare to average
	if s.data.Min >= 0 {
		avg := (s.data.Min + s.data.Max) / 2
		return value > avg
	}
	
	// If all negative, less negative is "positive"
	if s.data.Max <= 0 {
		avg := (s.data.Min + s.data.Max) / 2
		return value > avg
	}
	
	return false
}

// isNegativeValue determines if value represents a loss
func (s *Sparkline) isNegativeValue(value float64) bool {
	// If we have both positive and negative values, zero is the threshold
	if s.data.Min < 0 && s.data.Max > 0 {
		return value < 0
	}
	
	// If all positive, compare to average
	if s.data.Min >= 0 {
		avg := (s.data.Min + s.data.Max) / 2
		return value < avg
	}
	
	// If all negative, more negative is "negative"
	if s.data.Max <= 0 {
		avg := (s.data.Min + s.data.Max) / 2
		return value < avg
	}
	
	return false
}

// renderEmpty displays empty sparkline state
func (s *Sparkline) renderEmpty() string {
	emptyChar := strings.Repeat("─", s.config.Width)
	return s.renderer.ApplyThemeStyle(emptyChar, StyleAxis)
}

// renderSingleValue displays single data point
func (s *Sparkline) renderSingleValue() string {
	value := s.data.Values[0]
	char := s.valueToSparkChar(value)
	
	// Fill width with same character
	singleLine := strings.Repeat(char, s.config.Width)
	return s.applyValueStyling(singleLine, value)
}

// CreateTrendSparkline creates sparkline optimized for financial trend data
// Convenience constructor for common portfolio trend visualization
func CreateTrendSparkline(values []float64, width int, theme Theme) *Sparkline {
	config := ChartConfig{
		Width:      width,
		Height:     1,
		Theme:      theme,
		ShowAxes:   false,
		ShowLabels: false,
	}
	
	sparkline := NewSparkline(config)
	
	data := ChartData{
		Values: values,
		Title:  "",
	}
	
	if len(values) > 0 {
		min, max := CalculateDataRange(values)
		data.Min = min
		data.Max = max
	}
	
	sparkline.SetData(data)
	
	return sparkline
}

// CreateLabeledSparkline creates sparkline with title label
// Used in dashboard widgets and detailed views
func CreateLabeledSparkline(title string, values []float64, width int, theme Theme) *Sparkline {
	config := ChartConfig{
		Width:      width,
		Height:     1,
		Theme:      theme,
		ShowAxes:   false,
		ShowLabels: true,
	}
	
	sparkline := NewSparkline(config)
	
	data := ChartData{
		Values: values,
		Title:  title,
	}
	
	if len(values) > 0 {
		min, max := CalculateDataRange(values)
		data.Min = min
		data.Max = max
	}
	
	sparkline.SetData(data)
	
	return sparkline
}

// GetTrendDirection analyzes sparkline data for overall trend
// Returns trend classification for automated decision support
type TrendDirection int

const (
	TrendFlat TrendDirection = iota
	TrendUp
	TrendDown
	TrendVolatile
)

// AnalyzeTrend performs trend analysis on sparkline data
// Statistical analysis for portfolio decision support
func (s *Sparkline) AnalyzeTrend() TrendDirection {
	if len(s.data.Values) < 2 {
		return TrendFlat
	}
	
	// Calculate linear trend using first and last points
	first := s.data.Values[0]
	last := s.data.Values[len(s.data.Values)-1]
	
	// Calculate percentage change
	var pctChange float64
	if first != 0 {
		pctChange = (last - first) / first
	}
	
	// Analyze volatility
	volatility := s.calculateVolatility()
	
	// Trend classification thresholds
	const (
		significantChange = 0.05  // 5% change threshold
		highVolatility    = 0.10  // 10% volatility threshold
	)
	
	// High volatility indicates volatile trend regardless of direction
	if volatility > highVolatility {
		return TrendVolatile
	}
	
	// Directional trend based on change magnitude
	if pctChange > significantChange {
		return TrendUp
	} else if pctChange < -significantChange {
		return TrendDown
	}
	
	return TrendFlat
}

// calculateVolatility computes data volatility for trend analysis
func (s *Sparkline) calculateVolatility() float64 {
	if len(s.data.Values) < 2 {
		return 0
	}
	
	// Calculate mean
	var sum float64
	for _, val := range s.data.Values {
		sum += val
	}
	mean := sum / float64(len(s.data.Values))
	
	// Calculate variance
	var variance float64
	for _, val := range s.data.Values {
		diff := val - mean
		variance += diff * diff
	}
	variance /= float64(len(s.data.Values))
	
	// Return standard deviation as volatility measure
	if variance == 0 {
		return 0
	}
	
	// Simple square root approximation for integer math
	return s.approxSqrt(variance)
}

// approxSqrt provides simple square root approximation
// Avoids math package dependency for basic volatility calculation
func (s *Sparkline) approxSqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	
	// Newton's method approximation
	guess := x / 2
	for i := 0; i < 10; i++ {
		guess = (guess + x/guess) / 2
	}
	
	return guess
}