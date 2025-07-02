/*
NTX Portfolio Management TUI - Pie Charts

ASCII pie charts for sector allocation and portfolio composition visualization.
Circular representation using text characters with legend support for
professional financial data presentation.

Optimized for terminal display with intelligent segment labeling
and theme-aware styling for portfolio analytics.
*/

package charts

import (
	"fmt"
	"math"
	"strings"
)

// PieChart renders ASCII pie charts for portfolio composition
// Circular visualization with legend for sector allocation and holdings distribution
type PieChart struct {
	renderer     *ChartRenderer
	data         ChartData
	config       ChartConfig
	showLegend   bool
	showPercent  bool
	centerRadius int
}

// PieSegment represents a slice of the pie chart
type PieSegment struct {
	Value       float64
	Label       string
	StartAngle  float64
	EndAngle    float64
	Percentage  float64
	Character   string
	Color       StyleType
}

// NewPieChart creates pie chart with configuration
func NewPieChart(config ChartConfig) *PieChart {
	renderer := NewChartRenderer(config, ChartData{})
	
	return &PieChart{
		renderer:     renderer,
		config:       config,
		showLegend:   true,
		showPercent:  true,
		centerRadius: 1, // Minimum radius for readable pie
	}
}

// SetData updates pie chart data with validation
func (pc *PieChart) SetData(data ChartData) error {
	if err := ValidateData(data); err != nil {
		return err
	}
	
	// Calculate total for percentage calculations
	total := 0.0
	for _, value := range data.Values {
		if value < 0 {
			return ValidationError{Message: "pie chart values cannot be negative"}
		}
		total += value
	}
	
	if total == 0 {
		return ValidationError{Message: "pie chart total cannot be zero"}
	}
	
	data.Min = 0
	data.Max = total
	
	pc.data = data
	pc.renderer.data = data
	
	return nil
}

// SetConfig updates rendering configuration
func (pc *PieChart) SetConfig(config ChartConfig) {
	pc.config = config
	pc.renderer.config = config
}

// SetShowLegend controls legend display
func (pc *PieChart) SetShowLegend(show bool) {
	pc.showLegend = show
}

// SetShowPercent controls percentage display in legend
func (pc *PieChart) SetShowPercent(show bool) {
	pc.showPercent = show
}

// GetMinSize returns minimum dimensions for pie chart
func (pc *PieChart) GetMinSize() (width, height int) {
	minDiameter := 15
	legendWidth := 0
	
	if pc.showLegend {
		legendWidth = pc.calculateLegendWidth()
	}
	
	totalWidth := minDiameter + legendWidth + 2 // +2 for spacing
	
	return totalWidth, minDiameter
}

// Render generates ASCII pie chart with legend
func (pc *PieChart) Render() string {
	if len(pc.data.Values) == 0 {
		return pc.renderEmpty()
	}
	
	// Calculate pie segments
	segments := pc.calculateSegments()
	
	// Determine layout: pie + legend side by side
	pieSize := pc.calculatePieSize()
	
	var result strings.Builder
	
	// Add title if present
	if pc.data.Title != "" {
		titleRenderer := NewTitleRenderer(pc.renderer)
		title := titleRenderer.RenderTitle(pc.data.Title, pc.config.Width, false)
		result.WriteString(title)
		result.WriteString("\n")
	}
	
	// Render pie chart and legend
	pieLines := pc.renderPieChart(segments, pieSize)
	legendLines := pc.renderLegend(segments)
	
	// Combine pie and legend side by side
	combinedLines := pc.combinePieAndLegend(pieLines, legendLines)
	
	for _, line := range combinedLines {
		result.WriteString(line)
		result.WriteString("\n")
	}
	
	return strings.TrimSuffix(result.String(), "\n")
}

// calculateSegments converts data into pie segments with angles
func (pc *PieChart) calculateSegments() []PieSegment {
	if len(pc.data.Values) == 0 {
		return []PieSegment{}
	}
	
	// Calculate total value
	total := 0.0
	for _, value := range pc.data.Values {
		total += value
	}
	
	segments := make([]PieSegment, 0, len(pc.data.Values))
	currentAngle := 0.0
	
	for i, value := range pc.data.Values {
		if value <= 0 {
			continue // Skip zero or negative values
		}
		
		// Calculate percentage and angle
		percentage := (value / total) * 100
		angleSize := (value / total) * 360
		
		segment := PieSegment{
			Value:      value,
			Label:      pc.getLabel(i),
			StartAngle: currentAngle,
			EndAngle:   currentAngle + angleSize,
			Percentage: percentage,
			Character:  pc.getSegmentCharacter(i),
			Color:      pc.getSegmentColor(i),
		}
		
		segments = append(segments, segment)
		currentAngle += angleSize
	}
	
	return segments
}

// calculatePieSize determines optimal pie chart dimensions
func (pc *PieChart) calculatePieSize() int {
	availableWidth := pc.config.Width
	availableHeight := pc.config.Height
	
	// Reserve space for legend
	if pc.showLegend {
		legendWidth := pc.calculateLegendWidth()
		availableWidth -= legendWidth + 2 // +2 for spacing
	}
	
	// Reserve space for title
	if pc.data.Title != "" {
		availableHeight -= 2
	}
	
	// Pie chart should be roughly square, so use minimum dimension
	size := min(availableWidth, availableHeight)
	
	// Ensure minimum size
	if size < 10 {
		size = 10
	}
	
	// Ensure odd size for proper centering
	if size%2 == 0 {
		size--
	}
	
	return size
}

// renderPieChart creates the circular pie visualization
func (pc *PieChart) renderPieChart(segments []PieSegment, size int) []string {
	lines := make([]string, size)
	center := size / 2
	radius := float64(center)
	
	for y := 0; y < size; y++ {
		var line strings.Builder
		
		for x := 0; x < size; x++ {
			// Calculate distance from center
			dx := float64(x - center)
			dy := float64(y - center)
			distance := math.Sqrt(dx*dx + dy*dy)
			
			if distance <= radius {
				// Point is inside the circle
				angle := pc.calculateAngle(dx, dy)
				segment := pc.findSegmentForAngle(segments, angle)
				
				if segment != nil {
					char := segment.Character
					styledChar := pc.renderer.ApplyThemeStyle(char, segment.Color)
					line.WriteString(styledChar)
				} else {
					line.WriteString(" ")
				}
			} else {
				// Point is outside the circle
				line.WriteString(" ")
			}
		}
		
		lines[y] = line.String()
	}
	
	return lines
}

// calculateAngle computes angle from center to point
func (pc *PieChart) calculateAngle(dx, dy float64) float64 {
	angle := math.Atan2(dy, dx) * 180 / math.Pi
	
	// Convert to 0-360 range with 0 at top
	angle = angle + 90
	if angle < 0 {
		angle += 360
	}
	
	return angle
}

// findSegmentForAngle returns segment containing the given angle
func (pc *PieChart) findSegmentForAngle(segments []PieSegment, angle float64) *PieSegment {
	for i := range segments {
		segment := &segments[i]
		if angle >= segment.StartAngle && angle < segment.EndAngle {
			return segment
		}
	}
	return nil
}

// renderLegend creates legend with labels and percentages
func (pc *PieChart) renderLegend(segments []PieSegment) []string {
	if !pc.showLegend {
		return []string{}
	}
	
	var lines []string
	
	// Legend header
	header := "Legend:"
	styledHeader := pc.renderer.ApplyThemeStyle(header, StyleTitle)
	lines = append(lines, styledHeader)
	lines = append(lines, "") // Empty line
	
	// Legend entries
	for _, segment := range segments {
		entry := pc.formatLegendEntry(segment)
		lines = append(lines, entry)
	}
	
	return lines
}

// formatLegendEntry creates single legend entry
func (pc *PieChart) formatLegendEntry(segment PieSegment) string {
	var parts []string
	
	// Segment character
	styledChar := pc.renderer.ApplyThemeStyle(segment.Character, segment.Color)
	parts = append(parts, styledChar)
	
	// Label
	parts = append(parts, segment.Label)
	
	// Value and percentage
	if pc.showPercent {
		valueStr := pc.formatValue(segment.Value)
		percentStr := fmt.Sprintf("(%.1f%%)", segment.Percentage)
		parts = append(parts, valueStr, percentStr)
	} else {
		valueStr := pc.formatValue(segment.Value)
		parts = append(parts, valueStr)
	}
	
	return strings.Join(parts, " ")
}

// combinePieAndLegend merges pie chart and legend side by side
func (pc *PieChart) combinePieAndLegend(pieLines, legendLines []string) []string {
	maxLines := max(len(pieLines), len(legendLines))
	result := make([]string, maxLines)
	
	pieWidth := 0
	if len(pieLines) > 0 {
		// Calculate visual width of pie chart
		pieWidth = pc.calculateVisualWidth(pieLines[0])
	}
	
	for i := 0; i < maxLines; i++ {
		var line strings.Builder
		
		// Add pie chart line
		if i < len(pieLines) {
			line.WriteString(pieLines[i])
		} else {
			line.WriteString(strings.Repeat(" ", pieWidth))
		}
		
		// Add spacing
		line.WriteString("  ")
		
		// Add legend line
		if i < len(legendLines) {
			line.WriteString(legendLines[i])
		}
		
		result[i] = line.String()
	}
	
	return result
}

// Helper methods

// getLabel returns label for data point index
func (pc *PieChart) getLabel(index int) string {
	if index < len(pc.data.Labels) {
		return pc.data.Labels[index]
	}
	return fmt.Sprintf("Segment %d", index+1)
}

// getSegmentCharacter returns character for pie segment
func (pc *PieChart) getSegmentCharacter(index int) string {
	charIndex := index % len(ASCIIChars.PieSlices)
	return ASCIIChars.PieSlices[charIndex]
}

// getSegmentColor returns color styling for pie segment
func (pc *PieChart) getSegmentColor(index int) StyleType {
	// Cycle through different color styles
	switch index % 6 {
	case 0:
		return StylePositive
	case 1:
		return StyleNegative
	case 2:
		return StyleNeutral
	case 3:
		return StyleData
	case 4:
		return StyleTitle
	default:
		return StyleAxis
	}
}

// formatValue formats numeric value for display
func (pc *PieChart) formatValue(value float64) string {
	switch {
	case value >= 1000000:
		return fmt.Sprintf("%.1fM", value/1000000)
	case value >= 1000:
		return fmt.Sprintf("%.1fK", value/1000)
	case value >= 1:
		return fmt.Sprintf("%.0f", value)
	default:
		return fmt.Sprintf("%.2f", value)
	}
}

// calculateLegendWidth estimates legend width
func (pc *PieChart) calculateLegendWidth() int {
	if !pc.showLegend {
		return 0
	}
	
	maxWidth := 10 // Minimum legend width
	
	// Check legend header
	headerWidth := len("Legend:")
	if headerWidth > maxWidth {
		maxWidth = headerWidth
	}
	
	// Check legend entries
	for i, value := range pc.data.Values {
		label := pc.getLabel(i)
		valueStr := pc.formatValue(value)
		
		entryWidth := 1 + 1 + len(label) + 1 + len(valueStr) // char + space + label + space + value
		if pc.showPercent {
			entryWidth += 8 // " (XX.X%)"
		}
		
		if entryWidth > maxWidth {
			maxWidth = entryWidth
		}
	}
	
	return maxWidth
}

// calculateVisualWidth calculates display width ignoring ANSI codes
func (pc *PieChart) calculateVisualWidth(text string) int {
	// Simplified version - could be enhanced with proper ANSI parsing
	count := 0
	inAnsi := false
	
	for _, r := range text {
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

// renderEmpty displays empty pie chart state
func (pc *PieChart) renderEmpty() string {
	message := "No data to display"
	
	// Center message in available space
	padding := (pc.config.Width - len(message)) / 2
	if padding < 0 {
		padding = 0
	}
	
	centeredMessage := strings.Repeat(" ", padding) + message
	return pc.renderer.ApplyThemeStyle(centeredMessage, StyleAxis)
}

// Helper functions

// min returns minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CreateSectorPieChart convenience constructor for sector allocation
func CreateSectorPieChart(data ChartData, width, height int, theme Theme) *PieChart {
	config := ChartConfig{
		Width:      width,
		Height:     height,
		Theme:      theme,
		ShowAxes:   false,
		ShowLabels: true,
	}
	
	chart := NewPieChart(config)
	chart.SetData(data)
	chart.SetShowLegend(true)
	chart.SetShowPercent(true)
	
	return chart
}

// CreateSimplePieChart convenience constructor for basic pie chart
func CreateSimplePieChart(values []float64, labels []string, title string, width, height int, theme Theme) *PieChart {
	data := ChartData{
		Values: values,
		Labels: labels,
		Title:  title,
	}
	
	if len(values) > 0 {
		min, max := CalculateDataRange(values)
		data.Min = min
		data.Max = max
	}
	
	config := ChartConfig{
		Width:      width,
		Height:     height,
		Theme:      theme,
		ShowAxes:   false,
		ShowLabels: true,
	}
	
	chart := NewPieChart(config)
	chart.SetData(data)
	
	return chart
}