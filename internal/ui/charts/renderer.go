/*
NTX Portfolio Management TUI - ASCII Chart Renderer

Common rendering utilities for ASCII-based charts with terminal compatibility.
Provides consistent character selection, spacing, and layout calculations
across all chart types.

Optimized for financial data visualization with precise alignment
and professional appearance matching the btop-style aesthetic.
*/

package charts

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ASCIIChars defines character sets for different chart elements
// Terminal-safe characters ensure compatibility across all terminal emulators
var ASCIIChars = struct {
	// Sparkline characters (density-based)
	SparkBlocks []string
	
	// Bar chart characters
	HorizontalBar   string
	VerticalBar     string
	BarFill         string
	BarEmpty        string
	
	// Pie chart characters
	PieSlices []string
	
	// Border and connection characters
	Borders   BorderChars
	Corners   CornerChars
}{
	// Sparkline uses block density for smooth curves
	SparkBlocks: []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"},
	
	// Bar chart elements
	HorizontalBar: "█",
	VerticalBar:   "│",
	BarFill:       "█",
	BarEmpty:      "░",
	
	// Pie chart uses different fill patterns for accessibility
	PieSlices: []string{"█", "▓", "▒", "░", "▪", "▫", "●", "○"},
	
	// Border characters matching existing table design
	Borders: BorderChars{
		Horizontal: "─",
		Vertical:   "│",
		Cross:      "┼",
	},
	
	// Corner characters for clean borders
	Corners: CornerChars{
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	},
}

// BorderChars defines border character set
type BorderChars struct {
	Horizontal string
	Vertical   string
	Cross      string
}

// CornerChars defines corner character set
type CornerChars struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
}

// AxisRenderer handles axis drawing and labeling
// Consistent axis formatting across all chart types
type AxisRenderer struct {
	renderer *ChartRenderer
}

// NewAxisRenderer creates axis renderer with shared configuration
func NewAxisRenderer(renderer *ChartRenderer) *AxisRenderer {
	return &AxisRenderer{
		renderer: renderer,
	}
}

// RenderHorizontalAxis creates horizontal axis with optional labels
// Responsive label placement based on available width
func (ar *AxisRenderer) RenderHorizontalAxis(labels []string, width int) string {
	if !ar.renderer.config.ShowAxes {
		return ""
	}
	
	// Create axis line
	axis := strings.Repeat(ASCIIChars.Borders.Horizontal, width)
	styledAxis := ar.renderer.ApplyThemeStyle(axis, StyleAxis)
	
	if !ar.renderer.config.ShowLabels || len(labels) == 0 {
		return styledAxis
	}
	
	// Calculate label positions for even distribution
	labelLine := ar.distributeLabels(labels, width)
	
	return styledAxis + "\n" + labelLine
}

// RenderVerticalAxis creates vertical axis with value labels
// Financial data formatting with proper number representation
func (ar *AxisRenderer) RenderVerticalAxis(height int, minVal, maxVal float64) []string {
	if !ar.renderer.config.ShowAxes {
		return make([]string, height)
	}
	
	lines := make([]string, height)
	
	// Generate axis ticks and labels
	for i := 0; i < height; i++ {
		// Calculate value at this height (inverted for display)
		ratio := float64(height-1-i) / float64(height-1)
		value := minVal + ratio*(maxVal-minVal)
		
		// Format axis tick
		tick := ASCIIChars.Borders.Vertical
		if ar.renderer.config.ShowLabels {
			// Add value label every few lines to prevent clutter
			if i%max(height/5, 1) == 0 {
				label := ar.formatAxisValue(value)
				tick = fmt.Sprintf("%s%s", label, ASCIIChars.Borders.Vertical)
			}
		}
		
		lines[i] = ar.renderer.ApplyThemeStyle(tick, StyleAxis)
	}
	
	return lines
}

// distributeLabels spreads labels evenly across available width
// Prevents label overflow while maintaining readability
func (ar *AxisRenderer) distributeLabels(labels []string, width int) string {
	if len(labels) == 0 {
		return ""
	}
	
	// Calculate spacing between labels
	spacing := width / len(labels)
	if spacing < 1 {
		spacing = 1
	}
	
	var result strings.Builder
	for i, label := range labels {
		// Truncate label if too long
		maxLabelWidth := spacing - 1
		if maxLabelWidth < 1 {
			maxLabelWidth = 1
		}
		
		truncated := ar.renderer.TruncateText(label, maxLabelWidth)
		styledLabel := ar.renderer.ApplyThemeStyle(truncated, StyleAxis)
		
		result.WriteString(styledLabel)
		
		// Add spacing between labels (except last one)
		if i < len(labels)-1 {
			result.WriteString(strings.Repeat(" ", spacing-len(truncated)))
		}
	}
	
	return result.String()
}

// formatAxisValue formats numeric values for axis labels
// Consistent financial formatting matching the application style
func (ar *AxisRenderer) formatAxisValue(value float64) string {
	// Format based on value magnitude for readability
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

// BorderRenderer handles border and frame drawing
// Consistent border styling matching the table design
type BorderRenderer struct {
	renderer *ChartRenderer
}

// NewBorderRenderer creates border renderer with shared configuration
func NewBorderRenderer(renderer *ChartRenderer) *BorderRenderer {
	return &BorderRenderer{
		renderer: renderer,
	}
}

// RenderBorder creates complete border around chart content
// Matches table border design for visual consistency
func (br *BorderRenderer) RenderBorder(content []string, width, height int) []string {
	if len(content) == 0 {
		return br.createEmptyBorder(width, height)
	}
	
	result := make([]string, 0, height+2)
	
	// Top border
	topBorder := br.createTopBorder(width)
	result = append(result, topBorder)
	
	// Content with side borders
	for i, line := range content {
		if i >= height {
			break
		}
		
		// Ensure line fits within border
		paddedLine := br.padLine(line, width-2)
		borderedLine := br.wrapWithSideBorders(paddedLine)
		result = append(result, borderedLine)
	}
	
	// Fill remaining height with empty lines
	for len(result) < height+1 {
		emptyLine := br.wrapWithSideBorders(strings.Repeat(" ", width-2))
		result = append(result, emptyLine)
	}
	
	// Bottom border
	bottomBorder := br.createBottomBorder(width)
	result = append(result, bottomBorder)
	
	return result
}

// createTopBorder generates top border line
func (br *BorderRenderer) createTopBorder(width int) string {
	if width < 2 {
		return ""
	}
	
	border := ASCIIChars.Corners.TopLeft +
		strings.Repeat(ASCIIChars.Borders.Horizontal, width-2) +
		ASCIIChars.Corners.TopRight
	
	return br.renderer.ApplyThemeStyle(border, StyleAxis)
}

// createBottomBorder generates bottom border line
func (br *BorderRenderer) createBottomBorder(width int) string {
	if width < 2 {
		return ""
	}
	
	border := ASCIIChars.Corners.BottomLeft +
		strings.Repeat(ASCIIChars.Borders.Horizontal, width-2) +
		ASCIIChars.Corners.BottomRight
	
	return br.renderer.ApplyThemeStyle(border, StyleAxis)
}

// wrapWithSideBorders adds left and right borders to content line
func (br *BorderRenderer) wrapWithSideBorders(content string) string {
	leftBorder := br.renderer.ApplyThemeStyle(ASCIIChars.Borders.Vertical, StyleAxis)
	rightBorder := br.renderer.ApplyThemeStyle(ASCIIChars.Borders.Vertical, StyleAxis)
	
	return leftBorder + content + rightBorder
}

// padLine ensures line fits exactly within specified width
// Prevents border misalignment from variable content width
func (br *BorderRenderer) padLine(line string, targetWidth int) string {
	// Calculate visual width (handles ANSI sequences)
	visualWidth := br.calculateVisualWidth(line)
	
	if visualWidth >= targetWidth {
		return br.truncateToWidth(line, targetWidth)
	}
	
	// Pad with spaces to reach target width
	padding := targetWidth - visualWidth
	return line + strings.Repeat(" ", padding)
}

// calculateVisualWidth calculates display width ignoring ANSI sequences
// Accurate width calculation prevents border misalignment
func (br *BorderRenderer) calculateVisualWidth(text string) int {
	// Simple approach: count runes and ignore ANSI sequences
	// This is a simplified version - could be enhanced with proper ANSI parsing
	runes := []rune(text)
	count := 0
	inAnsi := false
	
	for _, r := range runes {
		if r == '\033' {
			inAnsi = true
			continue
		}
		if inAnsi && r == 'm' {
			inAnsi = false
			continue
		}
		if !inAnsi {
			count++
		}
	}
	
	return count
}

// truncateToWidth truncates text to exact visual width
// Maintains visual alignment while preventing overflow
func (br *BorderRenderer) truncateToWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	
	runes := []rune(text)
	if len(runes) <= maxWidth {
		return text
	}
	
	if maxWidth == 1 {
		return "…"
	}
	
	return string(runes[:maxWidth-1]) + "…"
}

// createEmptyBorder creates empty bordered area
func (br *BorderRenderer) createEmptyBorder(width, height int) []string {
	result := make([]string, height+2)
	
	// Top border
	result[0] = br.createTopBorder(width)
	
	// Empty content lines
	emptyLine := br.wrapWithSideBorders(strings.Repeat(" ", width-2))
	for i := 1; i <= height; i++ {
		result[i] = emptyLine
	}
	
	// Bottom border
	result[height+1] = br.createBottomBorder(width)
	
	return result
}

// TitleRenderer handles chart title rendering
// Consistent title formatting across all chart types
type TitleRenderer struct {
	renderer *ChartRenderer
}

// NewTitleRenderer creates title renderer with shared configuration
func NewTitleRenderer(renderer *ChartRenderer) *TitleRenderer {
	return &TitleRenderer{
		renderer: renderer,
	}
}

// RenderTitle creates centered title with optional border
// Professional title presentation matching application style
func (tr *TitleRenderer) RenderTitle(title string, width int, withBorder bool) string {
	if title == "" {
		return ""
	}
	
	// Truncate title if too long
	maxTitleWidth := width - 4 // Leave space for padding
	if maxTitleWidth < 1 {
		maxTitleWidth = width
	}
	
	truncated := tr.renderer.TruncateText(title, maxTitleWidth)
	
	// Apply title styling
	styledTitle := tr.renderer.ApplyThemeStyle(truncated, StyleTitle)
	
	// Center title within available width
	titleWidth := utf8.RuneCountInString(truncated)
	if titleWidth >= width {
		return styledTitle
	}
	
	padding := (width - titleWidth) / 2
	leftPad := strings.Repeat(" ", padding)
	rightPad := strings.Repeat(" ", width-titleWidth-padding)
	
	centeredTitle := leftPad + styledTitle + rightPad
	
	if withBorder {
		return tr.addTitleBorder(centeredTitle, width)
	}
	
	return centeredTitle
}

// addTitleBorder adds decorative border around title
func (tr *TitleRenderer) addTitleBorder(title string, width int) string {
	border := strings.Repeat(ASCIIChars.Borders.Horizontal, width)
	styledBorder := tr.renderer.ApplyThemeStyle(border, StyleAxis)
	
	return styledBorder + "\n" + title + "\n" + styledBorder
}

// Helper function for max calculation
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}