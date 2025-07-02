/*
NTX Portfolio Management TUI - Dashboard Component

Comprehensive dashboard section providing portfolio command center with:
- Portfolio overview metrics (copied from holdings)
- Market summary with NEPSE index and sector performance
- Recent activity tracker
- Top performers and portfolio allocation
- Btop-style professional layout with consistent borders

Designed as Section [1] for high-level portfolio monitoring and quick insights.
*/

package dashboard

import (
	"fmt"
	"strings"
	"time"

	"ntx/internal/ui/charts"
	"ntx/internal/ui/themes"

	"github.com/charmbracelet/lipgloss"
)

// DashboardDisplay provides comprehensive portfolio command center
// Combines overview metrics with market context and activity tracking
type DashboardDisplay struct {
	// Portfolio Metrics (copied from overview)
	TotalValue    int64 // Total portfolio market value in paisa
	TotalCost     int64 // Total cost basis in paisa
	UnrealizedPL  int64 // Unrealized P/L in paisa
	DayChange     int64 // Today's change in paisa
	HoldingsCount int   // Number of holdings

	// Market Data (placeholder for future integration)
	NepseIndex   float64 // NEPSE index value
	NepseChange  float64 // NEPSE daily change percentage
	BankingIndex float64 // Banking sector performance
	HydroIndex   float64 // Hydropower sector performance

	// Recent Activity (placeholder data)
	RecentTrades  []TradeActivity // Recent portfolio transactions
	TopPerformers []Performance   // Best/worst performing holdings

	// Component state
	LastUpdate   time.Time    // Last data refresh timestamp
	Theme        themes.Theme // Current theme for styling
	TerminalSize struct {     // Terminal dimensions for responsive layout
		Width  int
		Height int
	}
}

// TradeActivity represents recent portfolio transaction
type TradeActivity struct {
	Symbol   string    // Stock symbol
	Type     string    // "Buy" or "Sell"
	Quantity int64     // Share quantity
	Price    int64     // Price per share in paisa
	Date     time.Time // Transaction date
}

// Performance represents holding performance metrics
type Performance struct {
	Symbol        string  // Stock symbol
	PercentChange float64 // Performance percentage
	IsTop         bool    // True for top performer, false for worst
}

// NewDashboardDisplay creates dashboard display with default configuration
func NewDashboardDisplay(theme themes.Theme) *DashboardDisplay {
	return &DashboardDisplay{
		TotalValue:    0,
		TotalCost:     0,
		UnrealizedPL:  0,
		DayChange:     0,
		HoldingsCount: 0,

		// Placeholder market data
		NepseIndex:   2089.5,
		NepseChange:  0.8,
		BankingIndex: 1.2,
		HydroIndex:   0.5,

		// Sample recent activity
		RecentTrades:  generateSampleTrades(),
		TopPerformers: generateSamplePerformers(),

		LastUpdate:   time.Now(),
		Theme:        theme,
		TerminalSize: struct{ Width, Height int }{Width: 120, Height: 40},
	}
}

// UpdatePortfolioMetrics updates portfolio summary data
func (dd *DashboardDisplay) UpdatePortfolioMetrics(totalValue, totalCost, unrealizedPL, dayChange int64, holdingsCount int) {
	dd.TotalValue = totalValue
	dd.TotalCost = totalCost
	dd.UnrealizedPL = unrealizedPL
	dd.DayChange = dayChange
	dd.HoldingsCount = holdingsCount
	dd.LastUpdate = time.Now()
}

// SetTerminalSize updates responsive layout configuration
func (dd *DashboardDisplay) SetTerminalSize(width, height int) {
	dd.TerminalSize.Width = width
	dd.TerminalSize.Height = height
}

// SetTheme updates theme and refreshes styling
func (dd *DashboardDisplay) SetTheme(theme themes.Theme) {
	dd.Theme = theme
}

// Render generates the complete dashboard section
func (dd *DashboardDisplay) Render() string {
	width := dd.getDashboardWidth()

	// Main dashboard container with btop-style borders
	topBorder := dd.renderTopBorder()
	portfolioOverview := dd.renderPortfolioOverview(width)
	separator1 := dd.renderSeparator()
	marketSection := dd.renderMarketSection(width)
	separator2 := dd.renderSeparator()
	allocationSection := dd.renderAllocationSection(width)
	bottomBorder := dd.renderBottomBorder()

	return topBorder + "\n" +
		portfolioOverview + "\n" +
		separator1 + "\n" +
		marketSection + "\n" +
		separator2 + "\n" +
		allocationSection + "\n" +
		bottomBorder
}

// renderTopBorder creates top border with dashboard title
func (dd *DashboardDisplay) renderTopBorder() string {
	width := dd.getDashboardWidth()
	title := "[1]Dashboard"

	if width < len(title)+10 {
		border := "┌" + strings.Repeat("─", width-2) + "┐"
		return lipgloss.NewStyle().Foreground(dd.Theme.Primary()).Render(border)
	}

	titleSection := "─" + title + "─"
	remainingWidth := width - len([]rune(titleSection)) - 2
	leftPadding := strings.Repeat("─", remainingWidth)

	border := "┌" + titleSection + leftPadding + "┐"
	return lipgloss.NewStyle().Foreground(dd.Theme.Primary()).Render(border)
}

// renderBottomBorder creates bottom border
func (dd *DashboardDisplay) renderBottomBorder() string {
	width := dd.getDashboardWidth()
	border := "└" + strings.Repeat("─", width-2) + "┘"
	return lipgloss.NewStyle().Foreground(dd.Theme.Primary()).Render(border)
}

// renderSeparator creates horizontal separator between sections
func (dd *DashboardDisplay) renderSeparator() string {
	width := dd.getDashboardWidth()
	leftEdge := lipgloss.NewStyle().Foreground(dd.Theme.Primary()).Render("├")
	rightEdge := lipgloss.NewStyle().Foreground(dd.Theme.Primary()).Render("┤")
	middle := strings.Repeat("─", width-2)
	styledMiddle := lipgloss.NewStyle().Foreground(dd.Theme.Primary()).Render(middle)

	return leftEdge + styledMiddle + rightEdge
}

// renderPortfolioOverview renders portfolio metrics (same as overview component)
func (dd *DashboardDisplay) renderPortfolioOverview(width int) string {
	// Calculate percentage change
	var percentChange float64
	if dd.TotalCost > 0 {
		percentChange = float64(dd.UnrealizedPL) / float64(dd.TotalCost) * 100
	}

	// Format metric strings
	totalValueStr := FormatCurrency(dd.TotalValue)
	percentChangeStr := FormatPercent(percentChange)
	unrealizedPLStr := FormatPL(dd.UnrealizedPL)

	// Build content string based on terminal width
	var content string
	if width >= 120 {
		// Full format: Total: Rs.2,45,670 (+1.8%) │ Today: +Rs.5,620 │ Unrealized: +Rs.12,340 │ Holdings: 5
		dayChangeStr := ""
		if dd.DayChange != 0 {
			dayChangeStr = fmt.Sprintf(" │ Today: %s", FormatPL(dd.DayChange))
		}
		content = fmt.Sprintf(" Total: %s (%s)%s │ Unrealized: %s │ Holdings: %d ",
			totalValueStr, percentChangeStr, dayChangeStr, unrealizedPLStr, dd.HoldingsCount)
	} else if width >= 100 {
		// Medium format: Total: Rs.2,45,670 (+1.8%) │ P/L: +Rs.12,340 │ Holdings: 5
		content = fmt.Sprintf(" Total: %s (%s) │ P/L: %s │ Holdings: %d ",
			totalValueStr, percentChangeStr, unrealizedPLStr, dd.HoldingsCount)
	} else if width >= 80 {
		// Compact format: Total: Rs.2,45,670 (+1.8%) │ P/L: +Rs.12,340
		content = fmt.Sprintf(" Total: %s (%s) │ P/L: %s ",
			totalValueStr, percentChangeStr, unrealizedPLStr)
	} else {
		// Minimal format: Rs.2,45,670 (+1.8%)
		content = fmt.Sprintf(" %s (%s) ",
			totalValueStr, percentChangeStr)
	}

	// Apply styling based on P/L
	var style lipgloss.Style
	if dd.UnrealizedPL > 0 {
		style = lipgloss.NewStyle().Foreground(dd.Theme.Success())
	} else if dd.UnrealizedPL < 0 {
		style = lipgloss.NewStyle().Foreground(dd.Theme.Error())
	} else {
		style = lipgloss.NewStyle().Foreground(dd.Theme.Foreground())
	}

	styledContent := style.Render(content)

	// Pad to exact width and add borders
	return dd.wrapContentWithBorders(styledContent, width)
}

// renderMarketSection renders market overview and recent activity
func (dd *DashboardDisplay) renderMarketSection(width int) string {
	var leftContent, rightContent string

	if width >= 120 {
		// Three-column layout: Market | Top Performers | Recent Activity
		leftWidth := width / 3
		middleWidth := width / 3
		rightWidth := width - leftWidth - middleWidth - 6 // Account for separators

		leftContent = dd.renderMarketOverview(leftWidth - 2)
		middleContent := dd.renderTopPerformers(middleWidth - 2)
		rightContent = dd.renderRecentActivity(rightWidth - 2)

		// Join three sections with separators
		coloredSeparator := lipgloss.NewStyle().Foreground(dd.Theme.Primary()).Render("│")
		fullContent := " " + leftContent + " " + coloredSeparator + " " + middleContent + " " + coloredSeparator + " " + rightContent + " "

		return dd.wrapContentWithBorders(fullContent, width)

	} else if width >= 100 {
		// Two-column layout: Market Overview | Recent Activity
		leftWidth := width / 2
		rightWidth := width - leftWidth - 4 // Account for separators

		leftContent = dd.renderMarketOverview(leftWidth - 2)
		rightContent = dd.renderRecentActivity(rightWidth - 2)

		// Join two sections with separator
		coloredSeparator := lipgloss.NewStyle().Foreground(dd.Theme.Primary()).Render("│")
		fullContent := " " + leftContent + " " + coloredSeparator + " " + rightContent + " "

		return dd.wrapContentWithBorders(fullContent, width)

	} else {
		// Single column: Market Overview only
		content := dd.renderMarketOverview(width - 4)
		return dd.wrapContentWithBorders(" "+content+" ", width)
	}
}

// renderAllocationSection renders portfolio allocation visualization
func (dd *DashboardDisplay) renderAllocationSection(width int) string {
	content := dd.renderPortfolioAllocation(width - 4)
	return dd.wrapContentWithBorders(" "+content+" ", width)
}

// renderMarketOverview renders market statistics with sparklines
func (dd *DashboardDisplay) renderMarketOverview(width int) string {
	nepseChangeStr := FormatPercent(dd.NepseChange)
	bankingChangeStr := FormatPercent(dd.BankingIndex)
	hydroChangeStr := FormatPercent(dd.HydroIndex)

	// Generate sample trend data for sparklines
	nepseData := dd.generateNepseTrendData()
	bankingData := dd.generateSectorTrendData(dd.BankingIndex)
	hydroData := dd.generateSectorTrendData(dd.HydroIndex)

	content := "Market Overview\n"
	
	// Add sparklines for wider terminals
	if width >= 30 {
		// NEPSE with mini sparkline
		nepseSparkline := charts.CreateTrendSparkline(nepseData, 12, dd.Theme)
		content += fmt.Sprintf("NEPSE: %.1f (%s) %s\n", dd.NepseIndex, nepseChangeStr, nepseSparkline.Render())
		
		// Banking sector with sparkline
		bankingSparkline := charts.CreateTrendSparkline(bankingData, 8, dd.Theme)
		content += fmt.Sprintf("Banking: %s %s\n", bankingChangeStr, bankingSparkline.Render())
		
		// Hydro sector with sparkline
		hydroSparkline := charts.CreateTrendSparkline(hydroData, 8, dd.Theme)
		content += fmt.Sprintf("Hydro: %s %s\n", hydroChangeStr, hydroSparkline.Render())
		
		content += "Hotels: -0.2%%"
	} else {
		// Fallback to text-only for narrow terminals
		content += fmt.Sprintf("NEPSE: %.1f (%s)\nBanking: %s\nHydro: %s\nHotels: -0.2%%",
			dd.NepseIndex, nepseChangeStr, bankingChangeStr, hydroChangeStr)
	}

	return dd.truncateToWidth(content, width)
}

// renderTopPerformers renders best/worst performing holdings
func (dd *DashboardDisplay) renderTopPerformers(width int) string {
	content := "Top Performers\n"
	for _, perf := range dd.TopPerformers {
		if perf.IsTop {
			perfStr := FormatPercent(perf.PercentChange)
			content += fmt.Sprintf("%s: %s\n", perf.Symbol, perfStr)
		}
	}

	return dd.truncateToWidth(content, width)
}

// renderRecentActivity renders recent portfolio transactions
func (dd *DashboardDisplay) renderRecentActivity(width int) string {
	content := "Recent Activity\n"
	for i, trade := range dd.RecentTrades {
		if i >= 3 { // Limit to 3 recent trades
			break
		}
		priceStr := FormatCurrencyShort(trade.Price)
		content += fmt.Sprintf("%s %+d @%s\n", trade.Symbol, trade.Quantity, priceStr)
	}

	return dd.truncateToWidth(content, width)
}

// renderPortfolioAllocation renders allocation breakdown with horizontal bar chart
func (dd *DashboardDisplay) renderPortfolioAllocation(width int) string {
	// Sample allocation data (placeholder)
	allocations := []struct {
		Sector  string
		Percent float64
		Value   int64
	}{
		{"Commercial Banking", 65.2, 16017000}, // Rs.1,60,170 in paisa
		{"Hydropower", 25.8, 6348000},          // Rs.63,480 in paisa
		{"Manufacturing", 9.0, 2202000},        // Rs.22,020 in paisa
	}

	// Use enhanced horizontal bar chart for better terminal presentation
	if width >= 60 {
		// Full horizontal bar chart for wider terminals
		chartData := dd.convertAllocationToChartData(allocations)
		
		// Calculate chart dimensions
		chartWidth := width
		chartHeight := len(allocations) + 2 // One bar per sector plus title/spacing
		
		// Create horizontal bar chart with current theme
		barChart := charts.CreateHorizontalBarChart(chartData, chartWidth, chartHeight, dd.Theme)
		
		return barChart.Render()
	} else {
		// Enhanced text format for narrow terminals
		return dd.renderEnhancedAllocationText(allocations, width)
	}
}

// convertAllocationToChartData converts allocation data to chart format
func (dd *DashboardDisplay) convertAllocationToChartData(allocations []struct {
	Sector  string
	Percent float64
	Value   int64
}) charts.ChartData {
	values := make([]float64, len(allocations))
	labels := make([]string, len(allocations))

	for i, alloc := range allocations {
		// Use value in rupees for better chart scaling
		values[i] = float64(alloc.Value) / 100 // Convert paisa to rupees
		labels[i] = alloc.Sector
	}

	return charts.ChartData{
		Values: values,
		Labels: labels,
		Title:  "Portfolio Allocation",
	}
}

// renderEnhancedAllocationText provides enhanced text format with better spacing
func (dd *DashboardDisplay) renderEnhancedAllocationText(allocations []struct {
	Sector  string
	Percent float64
	Value   int64
}, width int) string {
	content := "Portfolio Allocation\n"
	
	// Calculate maximum bar length based on available width
	maxBarLength := max(10, min(25, width-35)) // Reserve space for labels and values
	
	for _, alloc := range allocations {
		// Create proportional bar representation
		barLength := int(alloc.Percent / 100 * float64(maxBarLength))
		if barLength < 1 && alloc.Percent > 0 {
			barLength = 1
		}
		
		// Apply theme colors to bars for better visual distinction
		bar := strings.Repeat("█", barLength)
		coloredBar := dd.Theme.SuccessStyle().Render(bar)
		
		// Format values with proper alignment
		valueStr := FormatCurrency(alloc.Value)
		
		// Create well-aligned line with sectors, bars, percentages, and values
		line := fmt.Sprintf("%-18s %s %5.1f%% %8s", 
			truncateString(alloc.Sector, 18), 
			coloredBar, 
			alloc.Percent, 
			valueStr)
		
		// Truncate if still too long for terminal
		if len([]rune(line)) > width {
			// Calculate safe truncation point accounting for ANSI codes
			visibleLength := lipgloss.Width(line)
			if visibleLength > width {
				line = truncateString(alloc.Sector, width-20) + "... " + valueStr
			}
		}
		content += line + "\n"
	}

	return content
}

// renderAllocationText provides fallback text format for narrow terminals (kept for compatibility)
func (dd *DashboardDisplay) renderAllocationText(allocations []struct {
	Sector  string
	Percent float64
	Value   int64
}, width int) string {
	return dd.renderEnhancedAllocationText(allocations, width)
}

// truncateString safely truncates string to specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateNepseTrendData creates sample NEPSE index trend data
func (dd *DashboardDisplay) generateNepseTrendData() []float64 {
	// Sample 30-day NEPSE trend starting from current index
	baseIndex := dd.NepseIndex
	data := make([]float64, 30)
	
	// Generate realistic trend data with some volatility
	for i := range data {
		volatility := (float64(i%7) - 3) * 0.015 // Weekly volatility pattern
		trend := dd.NepseChange / 100 * float64(i) / 30 // Overall trend
		data[i] = baseIndex * (1 + trend + volatility)
	}
	
	return data
}

// generateSectorTrendData creates sample sector performance trend data
func (dd *DashboardDisplay) generateSectorTrendData(currentChange float64) []float64 {
	// Sample 14-day sector trend
	data := make([]float64, 14)
	
	for i := range data {
		// Generate trend based on current performance
		dailyChange := currentChange / 100 / 14 * float64(i+1)
		volatility := (float64(i%3) - 1) * 0.005 // Minor daily volatility
		data[i] = 1.0 + dailyChange + volatility
	}
	
	return data
}

// Helper functions

// wrapContentWithBorders wraps content with left and right borders
func (dd *DashboardDisplay) wrapContentWithBorders(content string, width int) string {
	// Ensure content fits within borders
	contentWidth := lipgloss.Width(content)
	availableWidth := width - 2 // Account for borders

	if contentWidth < availableWidth {
		padding := strings.Repeat(" ", availableWidth-contentWidth)
		content = content + padding
	} else if contentWidth > availableWidth {
		// Truncate if too long
		content = content[:availableWidth-3] + "..."
	}

	// Apply border styling
	borderStyle := lipgloss.NewStyle().Foreground(dd.Theme.Primary())
	leftBorder := borderStyle.Render("│")
	rightBorder := borderStyle.Render("│")

	return leftBorder + content + rightBorder
}

// truncateToWidth truncates multi-line content to fit width
func (dd *DashboardDisplay) truncateToWidth(content string, width int) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		if len(line) > width {
			line = line[:width-3] + "..."
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// getDashboardWidth calculates dashboard width based on terminal size
func (dd *DashboardDisplay) getDashboardWidth() int {
	if dd.TerminalSize.Width < 60 {
		return 60
	}
	return dd.TerminalSize.Width
}

// generateSampleTrades creates sample trade activity data
func generateSampleTrades() []TradeActivity {
	return []TradeActivity{
		{"NABIL", "Buy", 10, 125000, time.Now().AddDate(0, 0, -1)},
		{"EBL", "Sell", -20, 70000, time.Now().AddDate(0, 0, -2)},
		{"HIDCL", "Buy", 50, 44500, time.Now().AddDate(0, 0, -3)},
	}
}

// generateSamplePerformers creates sample performance data
func generateSamplePerformers() []Performance {
	return []Performance{
		{"HIDCL", 6.0, true},
		{"NABIL", 4.9, true},
		{"EBL", 4.4, true},
	}
}

// FormatCurrency converts paisa to rupees with proper formatting
func FormatCurrency(paisa int64) string {
	rupees := float64(paisa) / 100
	if rupees >= 1000000 {
		return fmt.Sprintf("Rs.%.1fM", rupees/1000000)
	} else if rupees >= 1000 {
		return fmt.Sprintf("Rs.%.1fK", rupees/1000)
	}
	return fmt.Sprintf("Rs.%.0f", rupees)
}

// FormatCurrencyShort converts paisa to short currency format
func FormatCurrencyShort(paisa int64) string {
	rupees := float64(paisa) / 100
	if rupees >= 1000 {
		return fmt.Sprintf("%.0fK", rupees/1000)
	}
	return fmt.Sprintf("%.0f", rupees)
}

// FormatPL formats P/L with appropriate sign
func FormatPL(paisa int64) string {
	if paisa >= 0 {
		return fmt.Sprintf("+%s", FormatCurrency(paisa))
	}
	return FormatCurrency(paisa) // Already includes negative sign
}

// FormatPercent formats percentage with appropriate precision
func FormatPercent(percent float64) string {
	if percent >= 0 {
		return fmt.Sprintf("+%.1f%%", percent)
	}
	return fmt.Sprintf("%.1f%%", percent)
}
