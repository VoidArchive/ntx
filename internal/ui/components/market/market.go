/*
NTX Portfolio Management TUI - Market Section Component

Market Data & Sector Information section providing comprehensive market context
with NEPSE index, sector performance, news, and stock screening capabilities.

Features:
- Real-time NEPSE index and daily changes
- Sector performance tracking (Banking, Hydro, Manufacturing, Hotels)
- Market news and announcements
- Stock screener and watchlist management
- Btop-style borders matching other components

Designed as Section [5] for market context and stock discovery.
*/

package market

import (
	"fmt"
	"strings"
	"time"

	"ntx/internal/ui/themes"

	"github.com/charmbracelet/lipgloss"
)

// MarketDisplay provides comprehensive market data and sector information
type MarketDisplay struct {
	// NEPSE Index Data
	NepseIndex      float64 // Current NEPSE index value
	NepseChange     float64 // Daily change in points
	NepsePercent    float64 // Daily percentage change
	NepseVolume     int64   // Trading volume
	NepseTurnover   int64   // Trading turnover in paisa
	
	// Sector Performance
	SectorData      []SectorInfo // Sector performance data
	
	// Market News
	MarketNews      []NewsItem // Recent market announcements
	
	// Top Gainers/Losers
	TopGainers      []StockInfo // Best performing stocks
	TopLosers       []StockInfo // Worst performing stocks
	
	// Market Summary
	TradingStatus   string    // Market status (Open/Closed)
	LastUpdate      time.Time // Last market data update
	
	// Component state
	Theme           themes.Theme // Current theme for styling
	TerminalSize    struct {     // Terminal dimensions for responsive layout
		Width  int
		Height int
	}
	SelectedSection int // Currently selected market section
}

// SectorInfo represents sector performance data
type SectorInfo struct {
	Name        string  // Sector name
	Index       float64 // Sector index value
	Change      float64 // Daily change
	Percent     float64 // Percentage change
	Volume      int64   // Sector trading volume
	TopStock    string  // Best performing stock in sector
}

// NewsItem represents market news or announcement
type NewsItem struct {
	Headline    string    // News headline
	Source      string    // News source
	Time        time.Time // Publication time
	Category    string    // News category
	Priority    string    // News priority (High/Medium/Low)
}

// StockInfo represents individual stock performance
type StockInfo struct {
	Symbol      string  // Stock symbol
	Price       int64   // Current price in paisa
	Change      int64   // Price change in paisa
	Percent     float64 // Percentage change
	Volume      int64   // Trading volume
	Turnover    int64   // Trading turnover in paisa
}

// NewMarketDisplay creates market display with default configuration
func NewMarketDisplay(theme themes.Theme) *MarketDisplay {
	return &MarketDisplay{
		// Sample NEPSE data
		NepseIndex:    2089.5,
		NepseChange:   16.8,
		NepsePercent:  0.81,
		NepseVolume:   2500000,
		NepseTurnover: 125000000000, // Rs.12.5 crore in paisa
		
		// Sample sector data
		SectorData:   generateSampleSectorData(),
		
		// Sample market news
		MarketNews:   generateSampleNews(),
		
		// Sample gainers/losers
		TopGainers:   generateSampleGainers(),
		TopLosers:    generateSampleLosers(),
		
		// Market status
		TradingStatus: "Closed", // NEPSE trades Sun-Thu 11:00-15:00
		LastUpdate:    time.Now(),
		
		Theme:           theme,
		TerminalSize:    struct{ Width, Height int }{Width: 120, Height: 40},
		SelectedSection: 0,
	}
}

// SetTerminalSize updates responsive layout configuration
func (md *MarketDisplay) SetTerminalSize(width, height int) {
	md.TerminalSize.Width = width
	md.TerminalSize.Height = height
}

// SetTheme updates theme and refreshes styling
func (md *MarketDisplay) SetTheme(theme themes.Theme) {
	md.Theme = theme
}

// NavigateSection moves between market subsections
func (md *MarketDisplay) NavigateSection(direction int) {
	md.SelectedSection = (md.SelectedSection + direction) % 4
	if md.SelectedSection < 0 {
		md.SelectedSection = 3
	}
}

// Render generates the complete market section with btop-style borders
func (md *MarketDisplay) Render() string {
	width := md.getMarketWidth()
	
	// Main market container with btop-style borders
	topBorder := md.renderTopBorder()
	nepseSection := md.renderNepseOverview(width)
	separator1 := md.renderSeparator()
	sectorSection := md.renderSectorPerformance(width)
	separator2 := md.renderSeparator()
	activitySection := md.renderMarketActivity(width)
	bottomBorder := md.renderBottomBorder()
	
	return topBorder + "\n" +
		nepseSection + "\n" +
		separator1 + "\n" +
		sectorSection + "\n" +
		separator2 + "\n" +
		activitySection + "\n" +
		bottomBorder
}

// renderTopBorder creates top border with market title
func (md *MarketDisplay) renderTopBorder() string {
	width := md.getMarketWidth()
	title := "[5]Market"
	
	if width < len(title)+10 {
		border := "┌" + strings.Repeat("─", width-2) + "┐"
		return lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render(border)
	}
	
	titleSection := "─" + title + "─"
	remainingWidth := width - len([]rune(titleSection)) - 2
	leftPadding := strings.Repeat("─", remainingWidth)
	
	border := "┌" + titleSection + leftPadding + "┐"
	return lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render(border)
}

// renderBottomBorder creates bottom border
func (md *MarketDisplay) renderBottomBorder() string {
	width := md.getMarketWidth()
	border := "└" + strings.Repeat("─", width-2) + "┘"
	return lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render(border)
}

// renderSeparator creates horizontal separator between sections
func (md *MarketDisplay) renderSeparator() string {
	width := md.getMarketWidth()
	leftEdge := lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render("├")
	rightEdge := lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render("┤")
	middle := strings.Repeat("─", width-2)
	styledMiddle := lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render(middle)
	
	return leftEdge + styledMiddle + rightEdge
}

// renderNepseOverview renders NEPSE index and market status
func (md *MarketDisplay) renderNepseOverview(width int) string {
	var content string
	
	if width >= 120 {
		// Three-column layout for wide terminals
		leftContent := md.renderNepseIndex()
		middleContent := md.renderMarketStatus()
		rightContent := md.renderTradingStats()
		
		coloredSeparator := lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render("│")
		fullContent := " " + leftContent + " " + coloredSeparator + " " + middleContent + " " + coloredSeparator + " " + rightContent + " "
		
		return md.wrapContentWithBorders(fullContent, width)
	} else if width >= 80 {
		// Two-column layout for medium terminals
		leftContent := md.renderNepseIndex()
		rightContent := md.renderMarketStatus()
		
		coloredSeparator := lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render("│")
		fullContent := " " + leftContent + " " + coloredSeparator + " " + rightContent + " "
		
		return md.wrapContentWithBorders(fullContent, width)
	} else {
		// Single column for narrow terminals
		content = md.renderNepseIndex()
		return md.wrapContentWithBorders(" "+content+" ", width)
	}
}

// renderNepseIndex renders NEPSE index information
func (md *MarketDisplay) renderNepseIndex() string {
	content := "NEPSE Index\n"
	content += fmt.Sprintf("%.1f", md.NepseIndex)
	
	// Add change indicators
	changeStr := FormatChangeWithSign(md.NepseChange)
	percentStr := FormatPercent(md.NepsePercent)
	content += fmt.Sprintf("\n%s (%s)", changeStr, percentStr)
	
	// Add trend indicator
	if md.NepsePercent > 0 {
		content += " ↗"
	} else if md.NepsePercent < 0 {
		content += " ↘"
	} else {
		content += " →"
	}
	
	return content
}

// renderMarketStatus renders current market status
func (md *MarketDisplay) renderMarketStatus() string {
	content := "Market Status\n"
	content += md.TradingStatus
	
	// Add market hours info
	if md.TradingStatus == "Open" {
		content += "\nTrading until 15:00"
	} else {
		content += "\nNext: Sun 11:00"
	}
	
	content += fmt.Sprintf("\nLast Update: %s", md.LastUpdate.Format("15:04"))
	
	return content
}

// renderTradingStats renders trading volume and turnover
func (md *MarketDisplay) renderTradingStats() string {
	content := "Trading Stats\n"
	content += fmt.Sprintf("Volume: %.1fM", float64(md.NepseVolume)/1000000)
	content += fmt.Sprintf("\nTurnover: %s", FormatCurrency(md.NepseTurnover))
	content += "\nTransactions: 12.5K"
	
	return content
}

// renderSectorPerformance renders sector performance data
func (md *MarketDisplay) renderSectorPerformance(width int) string {
	content := "Sector Performance\n"
	
	// Table header
	if width >= 100 {
		content += "Sector           Index    Change    %      Top Stock\n"
		content += strings.Repeat("─", width-4) + "\n"
	} else {
		content += "Sector       Index  Change   %\n"
		content += strings.Repeat("─", width-4) + "\n"
	}
	
	// Sector rows
	for _, sector := range md.SectorData {
		changeStr := FormatChangeWithSign(sector.Change)
		percentStr := FormatPercent(sector.Percent)
		
		if width >= 100 {
			// Full format with top stock
			content += fmt.Sprintf("%-15s %7.1f %8s %6s  %s\n",
				sector.Name, sector.Index, changeStr, percentStr, sector.TopStock)
		} else {
			// Compact format
			content += fmt.Sprintf("%-11s %6.1f %7s %5s\n",
				sector.Name, sector.Index, changeStr, percentStr)
		}
	}
	
	return md.wrapContentWithBorders(" "+content+" ", width)
}

// renderMarketActivity renders gainers, losers, and news
func (md *MarketDisplay) renderMarketActivity(width int) string {
	var content string
	
	if width >= 120 {
		// Three-column layout: Gainers | Losers | News
		leftContent := md.renderTopGainers()
		middleContent := md.renderTopLosers()
		rightContent := md.renderMarketNews()
		
		coloredSeparator := lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render("│")
		fullContent := " " + leftContent + " " + coloredSeparator + " " + middleContent + " " + coloredSeparator + " " + rightContent + " "
		
		return md.wrapContentWithBorders(fullContent, width)
	} else if width >= 80 {
		// Two-column layout: Gainers | Losers
		leftContent := md.renderTopGainers()
		rightContent := md.renderTopLosers()
		
		coloredSeparator := lipgloss.NewStyle().Foreground(md.Theme.Primary()).Render("│")
		fullContent := " " + leftContent + " " + coloredSeparator + " " + rightContent + " "
		
		return md.wrapContentWithBorders(fullContent, width)
	} else {
		// Single column: Gainers only
		content = md.renderTopGainers()
		return md.wrapContentWithBorders(" "+content+" ", width)
	}
}

// renderTopGainers renders best performing stocks
func (md *MarketDisplay) renderTopGainers() string {
	content := "Top Gainers\n"
	
	for i, stock := range md.TopGainers {
		if i >= 4 { // Limit to 4 stocks
			break
		}
		
		priceStr := FormatCurrencyShort(stock.Price)
		percentStr := FormatPercent(stock.Percent)
		content += fmt.Sprintf("%s: %s (%s)\n", stock.Symbol, priceStr, percentStr)
	}
	
	return content
}

// renderTopLosers renders worst performing stocks
func (md *MarketDisplay) renderTopLosers() string {
	content := "Top Losers\n"
	
	for i, stock := range md.TopLosers {
		if i >= 4 { // Limit to 4 stocks
			break
		}
		
		priceStr := FormatCurrencyShort(stock.Price)
		percentStr := FormatPercent(stock.Percent)
		content += fmt.Sprintf("%s: %s (%s)\n", stock.Symbol, priceStr, percentStr)
	}
	
	return content
}

// renderMarketNews renders recent market news
func (md *MarketDisplay) renderMarketNews() string {
	content := "Market News\n"
	
	for i, news := range md.MarketNews {
		if i >= 3 { // Limit to 3 news items
			break
		}
		
		timeStr := news.Time.Format("15:04")
		content += fmt.Sprintf("[%s] %s\n", timeStr, news.Headline)
	}
	
	return content
}

// wrapContentWithBorders wraps content with left and right borders
func (md *MarketDisplay) wrapContentWithBorders(content string, width int) string {
	contentWidth := lipgloss.Width(content)
	availableWidth := width - 2
	
	if contentWidth < availableWidth {
		padding := strings.Repeat(" ", availableWidth-contentWidth)
		content = content + padding
	} else if contentWidth > availableWidth {
		content = content[:availableWidth-3] + "..."
	}
	
	borderStyle := lipgloss.NewStyle().Foreground(md.Theme.Primary())
	leftBorder := borderStyle.Render("│")
	rightBorder := borderStyle.Render("│")
	
	return leftBorder + content + rightBorder
}

// getMarketWidth calculates market width based on terminal size
func (md *MarketDisplay) getMarketWidth() int {
	if md.TerminalSize.Width < 60 {
		return 60
	}
	return md.TerminalSize.Width
}

// Helper functions for sample data generation

// generateSampleSectorData creates sample sector performance data
func generateSampleSectorData() []SectorInfo {
	return []SectorInfo{
		{"Commercial Banking", 1856.7, 22.4, 1.22, 1200000, "NABIL"},
		{"Hydropower", 2145.3, 15.8, 0.74, 800000, "CHCL"},
		{"Manufacturing", 1789.2, -8.5, -0.47, 450000, "UNL"},
		{"Hotels & Tourism", 2234.1, -12.3, -0.55, 320000, "OHL"},
	}
}

// generateSampleNews creates sample market news
func generateSampleNews() []NewsItem {
	now := time.Now()
	return []NewsItem{
		{"NEPSE Index Gains 16.8 Points", "NEPSE", now.Add(-1*time.Hour), "Market", "Medium"},
		{"Banking Sector Leads Gains", "ShareSansar", now.Add(-2*time.Hour), "Sector", "Medium"},
		{"New IPO Application Opens", "MeroShare", now.Add(-3*time.Hour), "IPO", "High"},
	}
}

// generateSampleGainers creates sample top gainers data
func generateSampleGainers() []StockInfo {
	return []StockInfo{
		{"CHCL", 55000, 4500, 8.9, 25000, 137500000},
		{"UNL", 42000, 3200, 8.2, 18000, 75600000},
		{"OHL", 38500, 2800, 7.8, 12000, 46200000},
		{"API", 67000, 4600, 7.4, 8500, 56950000},
	}
}

// generateSampleLosers creates sample top losers data
func generateSampleLosers() []StockInfo {
	return []StockInfo{
		{"GBIME", 31500, -2800, -8.2, 22000, 69300000},
		{"RMDC", 125000, -9500, -7.1, 5200, 65000000},
		{"SRBL", 28700, -2100, -6.8, 18500, 53095000},
		{"MLBL", 45600, -3200, -6.6, 15000, 68400000},
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

// FormatChangeWithSign formats change with appropriate sign
func FormatChangeWithSign(change float64) string {
	if change >= 0 {
		return fmt.Sprintf("+%.1f", change)
	}
	return fmt.Sprintf("%.1f", change)
}

// FormatPercent formats percentage with appropriate precision
func FormatPercent(percent float64) string {
	if percent >= 0 {
		return fmt.Sprintf("+%.1f%%", percent)
	}
	return fmt.Sprintf("%.1f%%", percent)
}