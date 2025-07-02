/*
NTX Portfolio Management TUI - Charts Interface

Chart system provides ASCII-based visualization for portfolio analytics,
technical indicators, and performance metrics integrated with the existing
theme system and responsive layout architecture.

Design prioritizes terminal compatibility and professional appearance
matching the btop-style aesthetic throughout the application.
*/

package charts

import (
	"github.com/charmbracelet/lipgloss"
)

// ChartData represents common data structure for all chart types
// Unified interface enables polymorphic chart rendering and theme consistency
type ChartData struct {
	Values []float64 // Core data points for rendering
	Labels []string  // Optional labels for data points
	Title  string    // Chart title for professional presentation
	Min    float64   // Data range minimum for scaling calculations
	Max    float64   // Data range maximum for scaling calculations
}

// ChartConfig defines rendering parameters for responsive design
// Centralizes layout calculations for consistent appearance across chart types
type ChartConfig struct {
	Width   int  // Target width in characters (terminal-aware)
	Height  int  // Target height in characters (terminal-aware)
	Theme   Theme // Theme integration for consistent styling
	ShowAxes bool // Axis visibility for dense layouts
	ShowLabels bool // Label visibility for compact displays
}

// Theme interface matches existing theme system for consistent integration
// Prevents theme dependency injection issues and styling inconsistencies
type Theme interface {
	Primary() lipgloss.Color
	Success() lipgloss.Color
	Warning() lipgloss.Color
	Error() lipgloss.Color
	Foreground() lipgloss.Color
	Muted() lipgloss.Color
	Background() lipgloss.Color
}

// Chart interface enables polymorphic chart rendering
// Common interface supports theme switching and responsive layout updates
type Chart interface {
	Render() string                    // Generate ASCII chart with theme styling
	SetData(data ChartData) error     // Update chart data with validation
	SetConfig(config ChartConfig)     // Apply rendering configuration
	GetMinSize() (width, height int)  // Return minimum viable dimensions
}

// ChartType enum prevents invalid chart type references
// Type safety ensures correct chart instantiation patterns
type ChartType string

const (
	ChartTypeSparkline ChartType = "sparkline"
	ChartTypeBar       ChartType = "bar"
	ChartTypePie       ChartType = "pie"
)

// ChartRenderer provides common ASCII rendering utilities
// Shared functionality eliminates code duplication across chart implementations
type ChartRenderer struct {
	config ChartConfig
	data   ChartData
}

// NewChartRenderer creates renderer with configuration
func NewChartRenderer(config ChartConfig, data ChartData) *ChartRenderer {
	return &ChartRenderer{
		config: config,
		data:   data,
	}
}

// ScaleValue maps data value to display coordinates
// Linear scaling ensures accurate visual representation of financial data
func (cr *ChartRenderer) ScaleValue(value float64, targetRange int) int {
	if cr.data.Max == cr.data.Min {
		return targetRange / 2 // Center value when no variance
	}
	
	ratio := (value - cr.data.Min) / (cr.data.Max - cr.data.Min)
	scaled := int(ratio * float64(targetRange-1))
	
	// Clamp to valid range to prevent rendering errors
	if scaled < 0 {
		return 0
	}
	if scaled >= targetRange {
		return targetRange - 1
	}
	
	return scaled
}

// TruncateText truncates text to fit within specified width
// Unicode-aware truncation prevents display corruption in terminal
func (cr *ChartRenderer) TruncateText(text string, maxWidth int) string {
	runes := []rune(text)
	if len(runes) <= maxWidth {
		return text
	}
	
	if maxWidth <= 1 {
		return "…"
	}
	
	return string(runes[:maxWidth-1]) + "…"
}

// ApplyThemeStyle applies consistent theme styling to chart elements
// Centralized styling ensures visual consistency across all chart types
func (cr *ChartRenderer) ApplyThemeStyle(text string, styleType StyleType) string {
	var style lipgloss.Style
	
	switch styleType {
	case StyleTitle:
		style = lipgloss.NewStyle().
			Foreground(cr.config.Theme.Primary()).
			Bold(true)
	case StyleAxis:
		style = lipgloss.NewStyle().
			Foreground(cr.config.Theme.Muted())
	case StyleData:
		style = lipgloss.NewStyle().
			Foreground(cr.config.Theme.Foreground())
	case StylePositive:
		style = lipgloss.NewStyle().
			Foreground(cr.config.Theme.Success())
	case StyleNegative:
		style = lipgloss.NewStyle().
			Foreground(cr.config.Theme.Error())
	case StyleNeutral:
		style = lipgloss.NewStyle().
			Foreground(cr.config.Theme.Warning())
	default:
		style = lipgloss.NewStyle().
			Foreground(cr.config.Theme.Foreground())
	}
	
	return style.Render(text)
}

// StyleType defines styling categories for consistent theming
type StyleType int

const (
	StyleTitle StyleType = iota
	StyleAxis
	StyleData
	StylePositive // For gains/profits
	StyleNegative // For losses
	StyleNeutral  // For neutral values
)

// ValidationError represents chart data validation failures
type ValidationError struct {
	Message string
}

func (ve ValidationError) Error() string {
	return ve.Message
}

// ValidateData performs common data validation across all chart types
// Prevents rendering errors from invalid or malformed data inputs
func ValidateData(data ChartData) error {
	if len(data.Values) == 0 {
		return ValidationError{Message: "chart data cannot be empty"}
	}
	
	if len(data.Labels) > 0 && len(data.Labels) != len(data.Values) {
		return ValidationError{Message: "labels length must match values length"}
	}
	
	// Check for valid numeric values
	for i, val := range data.Values {
		if val != val { // NaN check
			return ValidationError{Message: "invalid numeric value at index " + string(rune(i))}
		}
	}
	
	return nil
}

// CalculateDataRange computes min/max values from data slice
// Accurate range calculation ensures proper chart scaling
func CalculateDataRange(values []float64) (min, max float64) {
	if len(values) == 0 {
		return 0, 0
	}
	
	min = values[0]
	max = values[0]
	
	for _, val := range values[1:] {
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}
	
	return min, max
}