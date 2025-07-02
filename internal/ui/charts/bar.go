/*
NTX Portfolio Management TUI - Bar Charts

Enhanced horizontal and vertical bar charts for portfolio analytics,
sector allocation, and performance comparisons. ASCII-based rendering
with responsive design and professional financial data visualization.

Supports both horizontal and vertical orientations with intelligent
scaling, value labels, and theme-aware styling for gains/losses.
*/

package charts

import (
	"fmt"
	"strings"
)

// BarChart renders horizontal or vertical bar charts
// Flexible orientation and scaling for various financial visualizations
type BarChart struct {
	renderer    *ChartRenderer
	data        ChartData
	config      ChartConfig
	orientation BarOrientation
	showValues  bool
}

// BarOrientation defines chart orientation
type BarOrientation int

const (
	OrientationHorizontal BarOrientation = iota
	OrientationVertical
)

// NewBarChart creates bar chart with specified orientation
func NewBarChart(config ChartConfig, orientation BarOrientation) *BarChart {
	renderer := NewChartRenderer(config, ChartData{})
	
	return &BarChart{
		renderer:    renderer,
		config:      config,
		orientation: orientation,
		showValues:  true, // Default to showing values
	}
}

// SetData updates bar chart data with validation
func (bc *BarChart) SetData(data ChartData) error {
	if err := ValidateData(data); err != nil {
		return err
	}
	
	// Calculate data range for scaling
	if len(data.Values) > 0 {
		min, max := CalculateDataRange(data.Values)
		data.Min = min
		data.Max = max
	}
	
	bc.data = data
	bc.renderer.data = data
	
	return nil
}

// SetConfig updates rendering configuration
func (bc *BarChart) SetConfig(config ChartConfig) {
	bc.config = config
	bc.renderer.config = config
}

// SetShowValues controls value label display
func (bc *BarChart) SetShowValues(show bool) {
	bc.showValues = show
}

// GetMinSize returns minimum dimensions for bar chart
func (bc *BarChart) GetMinSize() (width, height int) {
	if bc.orientation == OrientationHorizontal {
		// Horizontal: need width for bars, height for each data point
		return 20, max(len(bc.data.Values), 5)
	} else {
		// Vertical: need width for each data point, height for bars
		return max(len(bc.data.Values)*3, 15), 10
	}
}

// Render generates ASCII bar chart with theme styling
func (bc *BarChart) Render() string {
	if len(bc.data.Values) == 0 {
		return bc.renderEmpty()
	}
	
	// Add title if present
	var result strings.Builder
	if bc.data.Title != "" {
		titleRenderer := NewTitleRenderer(bc.renderer)
		title := titleRenderer.RenderTitle(bc.data.Title, bc.config.Width, false)
		result.WriteString(title)
		result.WriteString("\n")
	}
	
	// Render chart based on orientation
	var chartLines []string
	if bc.orientation == OrientationHorizontal {
		chartLines = bc.renderHorizontalBars()
	} else {
		chartLines = bc.renderVerticalBars()
	}
	
	// Add borders if configured
	if bc.config.ShowAxes {
		borderRenderer := NewBorderRenderer(bc.renderer)
		chartLines = borderRenderer.RenderBorder(chartLines, bc.config.Width, len(chartLines))
	}
	
	for _, line := range chartLines {
		result.WriteString(line)
		result.WriteString("\n")
	}
	
	return strings.TrimSuffix(result.String(), "\n")
}

// renderHorizontalBars creates horizontal bar chart
// Each bar represents one data point with optional labels
func (bc *BarChart) renderHorizontalBars() []string {
	if len(bc.data.Values) == 0 {
		return []string{}
	}
	
	// Calculate dimensions
	availableWidth := bc.config.Width
	if bc.config.ShowAxes {
		availableWidth -= 2 // Account for borders
	}
	
	// Reserve space for labels and values
	labelWidth := bc.calculateMaxLabelWidth()
	valueWidth := bc.calculateMaxValueWidth()
	
	// Calculate bar area width
	barAreaWidth := availableWidth - labelWidth - valueWidth - 2 // -2 for spacing
	if barAreaWidth < 5 {
		barAreaWidth = 5
	}
	
	var lines []string
	
	for i, value := range bc.data.Values {
		line := bc.renderHorizontalBar(i, value, labelWidth, barAreaWidth, valueWidth)
		lines = append(lines, line)
	}
	
	return lines
}

// renderHorizontalBar creates single horizontal bar
func (bc *BarChart) renderHorizontalBar(index int, value float64, labelWidth, barWidth, valueWidth int) string {
	var parts []string
	
	// Label (left-aligned)
	label := bc.getLabel(index)
	if labelWidth > 0 {
		paddedLabel := bc.padToWidth(label, labelWidth, false)
		styledLabel := bc.renderer.ApplyThemeStyle(paddedLabel, StyleData)
		parts = append(parts, styledLabel)
	}
	
	// Bar visualization
	bar := bc.createHorizontalBar(value, barWidth)
	parts = append(parts, bar)
	
	// Value (right-aligned)
	if bc.showValues && valueWidth > 0 {
		valueStr := bc.formatValue(value)
		paddedValue := bc.padToWidth(valueStr, valueWidth, true)
		styledValue := bc.applyValueStyling(paddedValue, value)
		parts = append(parts, styledValue)
	}
	
	return strings.Join(parts, " ")
}

// createHorizontalBar generates bar visual representation
func (bc *BarChart) createHorizontalBar(value float64, width int) string {
	if width <= 0 {
		return ""
	}
	
	// Handle negative values by showing direction
	isNegative := value < 0
	absValue := value
	if isNegative {
		absValue = -value
	}
	
	// Scale bar length
	var barLength int
	if bc.data.Max == bc.data.Min {
		barLength = width / 2
	} else {
		// Scale based on absolute maximum for consistent scaling
		maxAbs := max(int(bc.data.Max), int(-bc.data.Min))
		if maxAbs > 0 {
			barLength = int(absValue * float64(width) / float64(maxAbs))
		}
	}
	
	// Clamp bar length
	if barLength > width {
		barLength = width
	}
	if barLength < 0 {
		barLength = 0
	}
	
	// Create bar visualization
	var bar string
	if isNegative {
		// Negative bar: empty space then filled portion
		emptySpace := width - barLength
		bar = strings.Repeat(" ", emptySpace) + strings.Repeat(ASCIIChars.BarFill, barLength)
		return bc.renderer.ApplyThemeStyle(bar, StyleNegative)
	} else {
		// Positive bar: filled portion then empty space
		emptySpace := width - barLength
		bar = strings.Repeat(ASCIIChars.BarFill, barLength) + strings.Repeat(" ", emptySpace)
		return bc.renderer.ApplyThemeStyle(bar, StylePositive)
	}
}

// renderVerticalBars creates vertical bar chart
func (bc *BarChart) renderVerticalBars() []string {
	if len(bc.data.Values) == 0 {
		return []string{}
	}
	
	// Calculate dimensions
	availableWidth := bc.config.Width
	if bc.config.ShowAxes {
		availableWidth -= 2 // Account for borders
	}
	
	chartHeight := bc.config.Height
	if bc.data.Title != "" {
		chartHeight -= 1 // Reserve space for title
	}
	if bc.showValues {
		chartHeight -= 1 // Reserve space for values
	}
	if bc.config.ShowLabels && len(bc.data.Labels) > 0 {
		chartHeight -= 1 // Reserve space for labels
	}
	
	// Calculate bar width and spacing
	barCount := len(bc.data.Values)
	totalSpacing := barCount - 1
	barWidth := max((availableWidth-totalSpacing)/barCount, 1)
	
	var lines []string
	
	// Render bars from top to bottom
	for row := chartHeight - 1; row >= 0; row-- {
		line := bc.renderVerticalBarRow(row, chartHeight, barWidth)
		lines = append(lines, line)
	}
	
	// Add value labels if enabled
	if bc.showValues {
		valueLine := bc.renderVerticalValueLabels(barWidth)
		lines = append(lines, valueLine)
	}
	
	// Add category labels if available
	if bc.config.ShowLabels && len(bc.data.Labels) > 0 {
		labelLine := bc.renderVerticalCategoryLabels(barWidth)
		lines = append(lines, labelLine)
	}
	
	return lines
}

// renderVerticalBarRow creates single row of vertical bars
func (bc *BarChart) renderVerticalBarRow(row, totalHeight, barWidth int) string {
	var parts []string
	
	for _, value := range bc.data.Values {
		// Calculate if this row should show bar fill
		barHeight := bc.calculateVerticalBarHeight(value, totalHeight)
		
		var barSegment string
		if row < barHeight {
			// Fill bar segment
			barSegment = strings.Repeat(ASCIIChars.BarFill, barWidth)
			barSegment = bc.applyValueStyling(barSegment, value)
		} else {
			// Empty bar segment
			barSegment = strings.Repeat(" ", barWidth)
		}
		
		parts = append(parts, barSegment)
	}
	
	return strings.Join(parts, " ")
}

// calculateVerticalBarHeight determines bar height for value
func (bc *BarChart) calculateVerticalBarHeight(value float64, maxHeight int) int {
	if bc.data.Max == bc.data.Min {
		return maxHeight / 2
	}
	
	// Handle negative values
	if value < 0 {
		return 0
	}
	
	// Scale to height
	ratio := value / bc.data.Max
	height := int(ratio * float64(maxHeight))
	
	// Clamp to valid range
	if height < 0 {
		height = 0
	}
	if height > maxHeight {
		height = maxHeight
	}
	
	return height
}

// renderVerticalValueLabels creates value labels for vertical bars
func (bc *BarChart) renderVerticalValueLabels(barWidth int) string {
	var parts []string
	
	for _, value := range bc.data.Values {
		valueStr := bc.formatValue(value)
		
		// Truncate value to fit bar width
		if len(valueStr) > barWidth {
			if barWidth > 1 {
				valueStr = valueStr[:barWidth-1] + "…"
			} else {
				valueStr = "…"
			}
		}
		
		// Pad value to bar width
		paddedValue := bc.padToWidth(valueStr, barWidth, true)
		styledValue := bc.applyValueStyling(paddedValue, value)
		
		parts = append(parts, styledValue)
	}
	
	return strings.Join(parts, " ")
}

// renderVerticalCategoryLabels creates category labels for vertical bars
func (bc *BarChart) renderVerticalCategoryLabels(barWidth int) string {
	var parts []string
	
	for i := range bc.data.Values {
		label := bc.getLabel(i)
		
		// Truncate label to fit bar width
		if len(label) > barWidth {
			if barWidth > 1 {
				label = label[:barWidth-1] + "…"
			} else {
				label = "…"
			}
		}
		
		// Pad label to bar width
		paddedLabel := bc.padToWidth(label, barWidth, false)
		styledLabel := bc.renderer.ApplyThemeStyle(paddedLabel, StyleAxis)
		
		parts = append(parts, styledLabel)
	}
	
	return strings.Join(parts, " ")
}

// Helper methods

// getLabel returns label for data point index
func (bc *BarChart) getLabel(index int) string {
	if index < len(bc.data.Labels) {
		return bc.data.Labels[index]
	}
	return fmt.Sprintf("Item %d", index+1)
}

// formatValue formats numeric value for display
func (bc *BarChart) formatValue(value float64) string {
	abs := value
	if abs < 0 {
		abs = -abs
	}
	
	switch {
	case abs >= 1000000:
		return fmt.Sprintf("%.1fM", value/1000000)
	case abs >= 1000:
		return fmt.Sprintf("%.1fK", value/1000)
	case abs >= 1:
		return fmt.Sprintf("%.0f", value)
	default:
		return fmt.Sprintf("%.2f", value)
	}
}

// applyValueStyling applies theme colors based on value
func (bc *BarChart) applyValueStyling(text string, value float64) string {
	if value > 0 {
		return bc.renderer.ApplyThemeStyle(text, StylePositive)
	} else if value < 0 {
		return bc.renderer.ApplyThemeStyle(text, StyleNegative)
	}
	return bc.renderer.ApplyThemeStyle(text, StyleNeutral)
}

// calculateMaxLabelWidth finds maximum label width
func (bc *BarChart) calculateMaxLabelWidth() int {
	if !bc.config.ShowLabels {
		return 0
	}
	
	maxWidth := 0
	for i := range bc.data.Values {
		label := bc.getLabel(i)
		if len(label) > maxWidth {
			maxWidth = len(label)
		}
	}
	
	// Limit label width to reasonable maximum
	const maxLabelWidth = 15
	if maxWidth > maxLabelWidth {
		maxWidth = maxLabelWidth
	}
	
	return maxWidth
}

// calculateMaxValueWidth finds maximum value width
func (bc *BarChart) calculateMaxValueWidth() int {
	if !bc.showValues {
		return 0
	}
	
	maxWidth := 0
	for _, value := range bc.data.Values {
		valueStr := bc.formatValue(value)
		if len(valueStr) > maxWidth {
			maxWidth = len(valueStr)
		}
	}
	
	return maxWidth
}

// padToWidth pads string to specified width
func (bc *BarChart) padToWidth(text string, width int, rightAlign bool) string {
	if len(text) >= width {
		return text
	}
	
	padding := width - len(text)
	if rightAlign {
		return strings.Repeat(" ", padding) + text
	}
	return text + strings.Repeat(" ", padding)
}

// renderEmpty displays empty chart state
func (bc *BarChart) renderEmpty() string {
	message := "No data to display"
	
	// Center message in available space
	padding := (bc.config.Width - len(message)) / 2
	if padding < 0 {
		padding = 0
	}
	
	centeredMessage := strings.Repeat(" ", padding) + message
	return bc.renderer.ApplyThemeStyle(centeredMessage, StyleAxis)
}

// CreateHorizontalBarChart convenience constructor for horizontal bars
func CreateHorizontalBarChart(data ChartData, width, height int, theme Theme) *BarChart {
	config := ChartConfig{
		Width:      width,
		Height:     height,
		Theme:      theme,
		ShowAxes:   true,
		ShowLabels: true,
	}
	
	chart := NewBarChart(config, OrientationHorizontal)
	chart.SetData(data)
	
	return chart
}

// CreateVerticalBarChart convenience constructor for vertical bars
func CreateVerticalBarChart(data ChartData, width, height int, theme Theme) *BarChart {
	config := ChartConfig{
		Width:      width,
		Height:     height,
		Theme:      theme,
		ShowAxes:   true,
		ShowLabels: true,
	}
	
	chart := NewBarChart(config, OrientationVertical)
	chart.SetData(data)
	
	return chart
}