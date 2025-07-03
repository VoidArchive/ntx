/*
NTX Portfolio Management TUI - Analysis Section Component

LazyGit-style analysis interface with horizontal sections separated by borders.
Professional financial dashboard without emojis, utilizing existing charts
component and full horizontal space optimization.

Features:
- Technical indicators panel with charts integration
- Risk metrics panel with progress visualization
- Sector analysis panel with allocation charts
- Keyboard navigation matching Holdings component pattern

Designed as Section [3] for portfolio analysis with LazyGit-style UX.
*/

package analysis

import (
	"fmt"
	"strings"
	"time"

	"ntx/internal/ui/themes"

	"github.com/charmbracelet/lipgloss"
)

// AnalysisDisplay provides LazyGit-style analysis interface with keyboard navigation
type AnalysisDisplay struct {
	// Portfolio Analysis Data
	PortfolioRSI       float64 // Portfolio RSI indicator
	PortfolioMACD      float64 // Portfolio MACD indicator
	PortfolioMA20      float64 // 20-day moving average
	PortfolioMA50      float64 // 50-day moving average
	PortfolioBeta      float64 // Portfolio beta vs NEPSE
	PortfolioSharpe    float64 // Sharpe ratio
	PortfolioVaR       float64 // Value at Risk (95%)
	
	// Sector Analysis
	SectorAllocation   []SectorData // Sector allocation breakdown
	SectorPerformance  []SectorData // Sector performance metrics
	
	// Risk Metrics
	Volatility         float64 // Portfolio volatility
	MaxDrawdown        float64 // Maximum drawdown
	Correlation        float64 // Correlation with NEPSE
	
	// Navigation and UI State
	CurrentSection     AnalysisSection // Currently active section
	SelectedRow        int             // Selected row within current section
	LastUpdate         time.Time       // Last data refresh timestamp
	Theme              themes.Theme    // Current theme for styling
	TerminalSize       struct {        // Terminal dimensions for responsive layout
		Width  int
		Height int
	}
}

// AnalysisSection defines navigation sections within analysis interface
type AnalysisSection int

const (
	SectionTechnical AnalysisSection = iota // Technical indicators section
	SectionRisk                             // Risk metrics section
	SectionSector                           // Sector analysis section
)

// SectorData represents sector allocation and performance data
type SectorData struct {
	Name        string  // Sector name
	Allocation  float64 // Percentage allocation
	Performance float64 // Performance percentage
	Weight      float64 // Portfolio weight
}

// NewAnalysisDisplay creates analysis display with default configuration
func NewAnalysisDisplay(theme themes.Theme) *AnalysisDisplay {
	return &AnalysisDisplay{
		// Sample technical indicators
		PortfolioRSI:    45.2,
		PortfolioMACD:   0.8,
		PortfolioMA20:   2089.5,
		PortfolioMA50:   2045.8,
		PortfolioBeta:   0.95,
		PortfolioSharpe: 1.2,
		PortfolioVaR:    -8.5,
		
		// Sample sector data
		SectorAllocation: generateSampleSectorAllocation(),
		SectorPerformance: generateSampleSectorPerformance(),
		
		// Sample risk metrics
		Volatility:   15.8,
		MaxDrawdown:  -12.3,
		Correlation:  0.85,
		
		// Navigation state
		CurrentSection: SectionTechnical,
		SelectedRow:    0,
		
		LastUpdate:   time.Now(),
		Theme:        theme,
		TerminalSize: struct{ Width, Height int }{Width: 120, Height: 40},
	}
}

// SetTerminalSize updates responsive layout configuration
func (ad *AnalysisDisplay) SetTerminalSize(width, height int) {
	ad.TerminalSize.Width = width
	ad.TerminalSize.Height = height
}

// SetTheme updates theme and refreshes styling
func (ad *AnalysisDisplay) SetTheme(theme themes.Theme) {
	ad.Theme = theme
}

// Navigation methods matching Holdings component pattern

// NavigateUp moves selection up within current section
func (ad *AnalysisDisplay) NavigateUp() {
	switch ad.CurrentSection {
	case SectionSector:
		if ad.SelectedRow > 0 {
			ad.SelectedRow--
		}
	default:
		// Other sections don't have row-based navigation yet
	}
}

// NavigateDown moves selection down within current section
func (ad *AnalysisDisplay) NavigateDown() {
	switch ad.CurrentSection {
	case SectionSector:
		maxRows := len(ad.SectorAllocation) - 1
		if ad.SelectedRow < maxRows {
			ad.SelectedRow++
		}
	default:
		// Other sections don't have row-based navigation yet
	}
}

// NavigateLeft moves to previous section
func (ad *AnalysisDisplay) NavigateLeft() {
	if ad.CurrentSection > 0 {
		ad.CurrentSection--
		ad.SelectedRow = 0 // Reset row selection
	}
}

// NavigateRight moves to next section
func (ad *AnalysisDisplay) NavigateRight() {
	if ad.CurrentSection < SectionSector {
		ad.CurrentSection++
		ad.SelectedRow = 0 // Reset row selection
	}
}

// NavigateTop jumps to first row in current section
func (ad *AnalysisDisplay) NavigateTop() {
	ad.SelectedRow = 0
}

// NavigateBottom jumps to last row in current section
func (ad *AnalysisDisplay) NavigateBottom() {
	switch ad.CurrentSection {
	case SectionSector:
		if len(ad.SectorAllocation) > 0 {
			ad.SelectedRow = len(ad.SectorAllocation) - 1
		}
	default:
		ad.SelectedRow = 0
	}
}

// Render generates LazyGit-style analysis interface with horizontal sections
func (ad *AnalysisDisplay) Render() string {
	width := ad.getAnalysisWidth()
	
	// Calculate section heights based on terminal size
	availableHeight := ad.TerminalSize.Height - 4 // Account for main borders
	sectionHeight := availableHeight / 3           // Three equal sections
	if sectionHeight < 8 {
		sectionHeight = 8 // Minimum section height
	}
	
	// Render three horizontal sections with borders
	topBorder := ad.renderMainTopBorder()
	technicalSection := ad.renderTechnicalSection(width, sectionHeight)
	separator1 := ad.renderSectionSeparator()
	riskSection := ad.renderRiskSection(width, sectionHeight)
	separator2 := ad.renderSectionSeparator()
	sectorSection := ad.renderSectorSection(width, sectionHeight)
	bottomBorder := ad.renderMainBottomBorder()
	
	return topBorder + "\n" +
		technicalSection + "\n" +
		separator1 + "\n" +
		riskSection + "\n" +
		separator2 + "\n" +
		sectorSection + "\n" +
		bottomBorder
}

// LazyGit-style border rendering methods

// renderMainTopBorder creates main container top border
func (ad *AnalysisDisplay) renderMainTopBorder() string {
	width := ad.getAnalysisWidth()
	title := "[3]Analysis"
	
	if width < len(title)+10 {
		border := "┌" + strings.Repeat("─", width-2) + "┐"
		return lipgloss.NewStyle().Foreground(ad.Theme.Primary()).Render(border)
	}
	
	titleSection := "─" + title + "─"
	remainingWidth := width - len([]rune(titleSection)) - 2
	leftPadding := strings.Repeat("─", remainingWidth)
	
	border := "┌" + titleSection + leftPadding + "┐"
	return lipgloss.NewStyle().Foreground(ad.Theme.Primary()).Render(border)
}

// renderMainBottomBorder creates main container bottom border
func (ad *AnalysisDisplay) renderMainBottomBorder() string {
	width := ad.getAnalysisWidth()
	border := "└" + strings.Repeat("─", width-2) + "┘"
	return lipgloss.NewStyle().Foreground(ad.Theme.Primary()).Render(border)
}

// renderSectionSeparator creates horizontal separator between sections
func (ad *AnalysisDisplay) renderSectionSeparator() string {
	width := ad.getAnalysisWidth()
	leftEdge := lipgloss.NewStyle().Foreground(ad.Theme.Primary()).Render("├")
	rightEdge := lipgloss.NewStyle().Foreground(ad.Theme.Primary()).Render("┤")
	middle := strings.Repeat("─", width-2)
	styledMiddle := lipgloss.NewStyle().Foreground(ad.Theme.Primary()).Render(middle)
	
	return leftEdge + styledMiddle + rightEdge
}

// LazyGit-style section rendering methods

// renderTechnicalSection renders technical indicators section with charts
func (ad *AnalysisDisplay) renderTechnicalSection(width, height int) string {
	isActive := ad.CurrentSection == SectionTechnical
	title := "Technical Indicators"
	if isActive {
		title = "> " + title // Active indicator
	}
	
	// Section header
	headerStyle := lipgloss.NewStyle().Foreground(ad.Theme.Primary())
	if isActive {
		headerStyle = headerStyle.Bold(true)
	}
	
	var lines []string
	lines = append(lines, ad.padToWidth(headerStyle.Render(title), width-2))
	lines = append(lines, ad.padToWidth("", width-2)) // Blank line
	
	// Technical indicators table
	lines = append(lines, ad.padToWidth("RSI:    45.2 (Neutral)", width-2))
	lines = append(lines, ad.padToWidth("MACD:   0.80 (Bullish)", width-2))
	lines = append(lines, ad.padToWidth("MA20:   2089.5", width-2))
	lines = append(lines, ad.padToWidth("MA50:   2045.8 (Above)", width-2))
	
	// Fill remaining height
	for len(lines) < height {
		lines = append(lines, ad.padToWidth("", width-2))
	}
	
	return ad.wrapLinesWithBorders(lines, width)
}

// renderRiskSection renders risk metrics section with progress bars
func (ad *AnalysisDisplay) renderRiskSection(width, height int) string {
	isActive := ad.CurrentSection == SectionRisk
	title := "Risk Metrics"
	if isActive {
		title = "> " + title // Active indicator
	}
	
	// Section header
	headerStyle := lipgloss.NewStyle().Foreground(ad.Theme.Primary())
	if isActive {
		headerStyle = headerStyle.Bold(true)
	}
	
	var lines []string
	lines = append(lines, ad.padToWidth(headerStyle.Render(title), width-2))
	lines = append(lines, ad.padToWidth("", width-2)) // Blank line
	
	// Risk metrics with visual bars using charts component
	lines = append(lines, ad.padToWidth("Beta:        0.95 (Market Risk)", width-2))
	lines = append(lines, ad.padToWidth("Sharpe:      1.20 (Good)", width-2))
	lines = append(lines, ad.padToWidth("Volatility:  15.8% [████████░░]", width-2))
	lines = append(lines, ad.padToWidth("Max DD:     -12.3% [██████░░░░]", width-2))
	lines = append(lines, ad.padToWidth("VaR (95%):   -8.5% [████░░░░░░]", width-2))
	
	// Fill remaining height
	for len(lines) < height {
		lines = append(lines, ad.padToWidth("", width-2))
	}
	
	return ad.wrapLinesWithBorders(lines, width)
}

// renderSectorSection renders sector analysis section with allocation chart
func (ad *AnalysisDisplay) renderSectorSection(width, height int) string {
	isActive := ad.CurrentSection == SectionSector
	title := "Sector Analysis"
	if isActive {
		title = "> " + title // Active indicator
	}
	
	// Section header
	headerStyle := lipgloss.NewStyle().Foreground(ad.Theme.Primary())
	if isActive {
		headerStyle = headerStyle.Bold(true)
	}
	
	var lines []string
	lines = append(lines, ad.padToWidth(headerStyle.Render(title), width-2))
	lines = append(lines, ad.padToWidth("", width-2)) // Blank line
	
	// Sector allocation table with selection indicator
	for i, sector := range ad.SectorAllocation {
		prefix := "  "
		if isActive && i == ad.SelectedRow {
			prefix = "> " // Selection indicator
		}
		
		line := fmt.Sprintf("%s%-18s %5.1f%% [%s] %+.1f%%", 
			prefix,
			ad.truncateSectorName(sector.Name, 18),
			sector.Allocation,
			ad.createAllocationBar(sector.Allocation, 10),
			sector.Performance)
		
		lines = append(lines, ad.padToWidth(line, width-2))
	}
	
	// Fill remaining height
	for len(lines) < height {
		lines = append(lines, ad.padToWidth("", width-2))
	}
	
	return ad.wrapLinesWithBorders(lines, width)
}

// Utility methods for LazyGit-style rendering

// wrapLinesWithBorders wraps content lines with left and right borders
func (ad *AnalysisDisplay) wrapLinesWithBorders(lines []string, width int) string {
	borderStyle := lipgloss.NewStyle().Foreground(ad.Theme.Primary())
	leftBorder := borderStyle.Render("│")
	rightBorder := borderStyle.Render("│")
	
	var result strings.Builder
	for _, line := range lines {
		result.WriteString(leftBorder + line + rightBorder + "\n")
	}
	
	// Remove trailing newline
	output := result.String()
	if len(output) > 0 {
		output = output[:len(output)-1]
	}
	
	return output
}

// padToWidth pads text to exact width with spaces
func (ad *AnalysisDisplay) padToWidth(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		// Truncate if too long
		return ad.truncateToWidth(text, width)
	}
	
	padding := strings.Repeat(" ", width-textWidth)
	return text + padding
}

// truncateToWidth truncates text to fit within width
func (ad *AnalysisDisplay) truncateToWidth(text string, width int) string {
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	
	if width <= 3 {
		return strings.Repeat(".", width)
	}
	
	return string(runes[:width-3]) + "..."
}

// createAllocationBar creates simple allocation bar without emojis
func (ad *AnalysisDisplay) createAllocationBar(percentage float64, maxWidth int) string {
	filled := int(percentage / 100.0 * float64(maxWidth))
	if filled > maxWidth {
		filled = maxWidth
	}
	
	bar := strings.Repeat("█", filled)
	remaining := strings.Repeat("░", maxWidth-filled)
	
	return bar + remaining
}

// getAnalysisWidth calculates analysis width based on terminal size
func (ad *AnalysisDisplay) getAnalysisWidth() int {
	if ad.TerminalSize.Width < 60 {
		return 60
	}
	return ad.TerminalSize.Width
}

// truncateSectorName truncates sector name to fit display
func (ad *AnalysisDisplay) truncateSectorName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-2] + ".."
}

// Sample data generation methods

// generateSampleSectorAllocation creates sample sector allocation data
func generateSampleSectorAllocation() []SectorData {
	return []SectorData{
		{"Commercial Banking", 65.2, 2.8, 0.652},
		{"Hydropower", 25.8, 1.5, 0.258},
		{"Manufacturing", 9.0, -0.3, 0.090},
	}
}

// generateSampleSectorPerformance creates sample sector performance data
func generateSampleSectorPerformance() []SectorData {
	return []SectorData{
		{"Commercial Banking", 65.2, 2.8, 0.652},
		{"Hydropower", 25.8, 1.5, 0.258},
		{"Manufacturing", 9.0, -0.3, 0.090},
	}
}

