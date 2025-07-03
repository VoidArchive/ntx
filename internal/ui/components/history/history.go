/*
NTX Portfolio Management TUI - History Section Component

Transaction History & Performance Tracking section providing comprehensive
transaction log and historical performance analysis with btop-style borders.

Features:
- Complete transaction history with filtering
- Performance timeline and metrics
- Realized P/L tracking and tax implications
- T+3 settlement tracking for NEPSE
- Btop-style borders matching other components

Designed as Section [4] for transaction history and performance analysis.
*/

package history

import (
	"fmt"
	"strings"
	"time"

	"ntx/internal/ui/themes"

	"github.com/charmbracelet/lipgloss"
)

// HistoryDisplay provides comprehensive transaction history and performance tracking
type HistoryDisplay struct {
	// Transaction History
	RecentTransactions []TransactionRecord // Recent transaction history
	TotalTransactions  int                 // Total transaction count
	
	// Performance History
	PerformanceHistory []PerformanceRecord // Historical performance data
	
	// Realized P/L
	RealizedPL         int64   // Total realized P/L in paisa
	TaxableGains       int64   // Taxable gains in paisa
	TaxLosses          int64   // Tax losses in paisa
	
	// Settlement Tracking
	PendingSettlements []Settlement // T+3 pending settlements
	
	// Filtering and Display
	FilterSymbol       string // Symbol filter
	FilterType         string // Transaction type filter
	FilterPeriod       string // Time period filter
	
	// Component state
	LastUpdate         time.Time    // Last data refresh timestamp
	Theme              themes.Theme // Current theme for styling
	TerminalSize       struct {     // Terminal dimensions for responsive layout
		Width  int
		Height int
	}
	SelectedRow        int // Currently selected transaction row
}

// TransactionRecord represents a single transaction in history
type TransactionRecord struct {
	ID          int       // Transaction ID
	Symbol      string    // Stock symbol
	Type        string    // Buy/Sell
	Quantity    int64     // Share quantity
	Price       int64     // Price per share in paisa
	TotalValue  int64     // Total transaction value in paisa
	Date        time.Time // Transaction date
	Status      string    // Settlement status
	RealizedPL  int64     // Realized P/L for sells (in paisa)
}

// PerformanceRecord represents portfolio performance at a point in time
type PerformanceRecord struct {
	Date           time.Time // Performance date
	PortfolioValue int64     // Portfolio value in paisa
	DailyChange    int64     // Daily change in paisa
	DailyPercent   float64   // Daily percentage change
}

// Settlement represents T+3 settlement information
type Settlement struct {
	TransactionID int       // Related transaction ID
	Symbol        string    // Stock symbol
	Quantity      int64     // Share quantity
	Amount        int64     // Settlement amount in paisa
	SettleDate    time.Time // Expected settlement date
	Status        string    // Settlement status
}

// NewHistoryDisplay creates history display with default configuration
func NewHistoryDisplay(theme themes.Theme) *HistoryDisplay {
	return &HistoryDisplay{
		// Sample transaction data
		RecentTransactions: generateSampleTransactions(),
		TotalTransactions:  25,
		
		// Sample performance data
		PerformanceHistory: generateSamplePerformance(),
		
		// Sample P/L data
		RealizedPL:   1250000, // Rs.12,500 in paisa
		TaxableGains: 980000,  // Rs.9,800 in paisa
		TaxLosses:    -230000, // Rs.-2,300 in paisa
		
		// Sample settlement data
		PendingSettlements: generateSampleSettlements(),
		
		// Default filters
		FilterSymbol: "",
		FilterType:   "All",
		FilterPeriod: "30d",
		
		LastUpdate:   time.Now(),
		Theme:        theme,
		TerminalSize: struct{ Width, Height int }{Width: 120, Height: 40},
		SelectedRow:  0,
	}
}

// SetTerminalSize updates responsive layout configuration
func (hd *HistoryDisplay) SetTerminalSize(width, height int) {
	hd.TerminalSize.Width = width
	hd.TerminalSize.Height = height
}

// SetTheme updates theme and refreshes styling
func (hd *HistoryDisplay) SetTheme(theme themes.Theme) {
	hd.Theme = theme
}

// NavigateUp moves selection up in transaction list
func (hd *HistoryDisplay) NavigateUp() {
	if hd.SelectedRow > 0 {
		hd.SelectedRow--
	}
}

// NavigateDown moves selection down in transaction list
func (hd *HistoryDisplay) NavigateDown() {
	if hd.SelectedRow < len(hd.RecentTransactions)-1 {
		hd.SelectedRow++
	}
}

// Render generates the complete history section with btop-style borders
func (hd *HistoryDisplay) Render() string {
	width := hd.getHistoryWidth()
	
	// Main history container with btop-style borders
	topBorder := hd.renderTopBorder()
	summarySection := hd.renderSummarySection(width)
	separator1 := hd.renderSeparator()
	transactionSection := hd.renderTransactionHistory(width)
	separator2 := hd.renderSeparator()
	performanceSection := hd.renderPerformanceSection(width)
	bottomBorder := hd.renderBottomBorder()
	
	return topBorder + "\n" +
		summarySection + "\n" +
		separator1 + "\n" +
		transactionSection + "\n" +
		separator2 + "\n" +
		performanceSection + "\n" +
		bottomBorder
}

// renderTopBorder creates top border with history title
func (hd *HistoryDisplay) renderTopBorder() string {
	width := hd.getHistoryWidth()
	title := "[4]History"
	
	if width < len(title)+10 {
		border := "┌" + strings.Repeat("─", width-2) + "┐"
		return lipgloss.NewStyle().Foreground(hd.Theme.Primary()).Render(border)
	}
	
	titleSection := "─" + title + "─"
	remainingWidth := width - len([]rune(titleSection)) - 2
	leftPadding := strings.Repeat("─", remainingWidth)
	
	border := "┌" + titleSection + leftPadding + "┐"
	return lipgloss.NewStyle().Foreground(hd.Theme.Primary()).Render(border)
}

// renderBottomBorder creates bottom border
func (hd *HistoryDisplay) renderBottomBorder() string {
	width := hd.getHistoryWidth()
	border := "└" + strings.Repeat("─", width-2) + "┘"
	return lipgloss.NewStyle().Foreground(hd.Theme.Primary()).Render(border)
}

// renderSeparator creates horizontal separator between sections
func (hd *HistoryDisplay) renderSeparator() string {
	width := hd.getHistoryWidth()
	leftEdge := lipgloss.NewStyle().Foreground(hd.Theme.Primary()).Render("├")
	rightEdge := lipgloss.NewStyle().Foreground(hd.Theme.Primary()).Render("┤")
	middle := strings.Repeat("─", width-2)
	styledMiddle := lipgloss.NewStyle().Foreground(hd.Theme.Primary()).Render(middle)
	
	return leftEdge + styledMiddle + rightEdge
}

// renderSummarySection renders P/L summary and filters
func (hd *HistoryDisplay) renderSummarySection(width int) string {
	var content string
	
	if width >= 100 {
		// Two-column layout for wide terminals
		leftContent := hd.renderPLSummary()
		rightContent := hd.renderFilterInfo()
		
		coloredSeparator := lipgloss.NewStyle().Foreground(hd.Theme.Primary()).Render("│")
		fullContent := " " + leftContent + " " + coloredSeparator + " " + rightContent + " "
		
		return hd.wrapContentWithBorders(fullContent, width)
	} else {
		// Single column for narrow terminals
		content = hd.renderPLSummary()
		return hd.wrapContentWithBorders(" "+content+" ", width)
	}
}

// renderPLSummary renders realized P/L summary
func (hd *HistoryDisplay) renderPLSummary() string {
	content := "Realized P/L Summary\n"
	content += fmt.Sprintf("Total Realized: %s\n", FormatCurrency(hd.RealizedPL))
	content += fmt.Sprintf("Taxable Gains: %s\n", FormatCurrency(hd.TaxableGains))
	content += fmt.Sprintf("Tax Losses: %s\n", FormatCurrency(hd.TaxLosses))
	content += fmt.Sprintf("Transactions: %d", hd.TotalTransactions)
	
	return content
}

// renderFilterInfo renders current filter information
func (hd *HistoryDisplay) renderFilterInfo() string {
	content := "Filter Settings\n"
	content += fmt.Sprintf("Symbol: %s\n", hd.FilterSymbol)
	content += fmt.Sprintf("Type: %s\n", hd.FilterType)
	content += fmt.Sprintf("Period: %s\n", hd.FilterPeriod)
	content += fmt.Sprintf("Showing: %d records", len(hd.RecentTransactions))
	
	return content
}

// renderTransactionHistory renders transaction history table
func (hd *HistoryDisplay) renderTransactionHistory(width int) string {
	content := "Recent Transactions\n"
	
	// Table header
	if width >= 100 {
		content += "Date       Symbol   Type  Qty    Price     Total       P/L\n"
		content += strings.Repeat("─", width-4) + "\n"
	} else {
		content += "Date     Symbol Type  Qty   Price    Total\n"
		content += strings.Repeat("─", width-4) + "\n"
	}
	
	// Transaction rows
	for i, tx := range hd.RecentTransactions {
		if i >= 8 { // Limit to 8 transactions for display
			break
		}
		
		// Selection indicator
		indicator := " "
		if i == hd.SelectedRow {
			indicator = "►"
		}
		
		dateStr := tx.Date.Format("01-02")
		priceStr := FormatCurrencyShort(tx.Price)
		totalStr := FormatCurrencyShort(tx.TotalValue)
		
		if width >= 100 {
			// Full format with P/L
			plStr := ""
			if tx.RealizedPL != 0 {
				plStr = FormatPL(tx.RealizedPL)
			}
			
			content += fmt.Sprintf("%s%-8s %-8s %-4s %4d %8s %10s %8s\n",
				indicator, dateStr, tx.Symbol, tx.Type, tx.Quantity, priceStr, totalStr, plStr)
		} else {
			// Compact format
			content += fmt.Sprintf("%s%-7s %-6s %-4s %3d %7s %9s\n",
				indicator, dateStr, tx.Symbol, tx.Type, tx.Quantity, priceStr, totalStr)
		}
	}
	
	return hd.wrapContentWithBorders(" "+content+" ", width)
}

// renderPerformanceSection renders performance history
func (hd *HistoryDisplay) renderPerformanceSection(width int) string {
	content := "Performance Timeline\n"
	
	// Show recent performance points
	for i, perf := range hd.PerformanceHistory {
		if i >= 5 { // Limit to 5 recent points
			break
		}
		
		dateStr := perf.Date.Format("01-02")
		valueStr := FormatCurrencyShort(perf.PortfolioValue)
		changeStr := FormatPL(perf.DailyChange)
		percentStr := FormatPercent(perf.DailyPercent)
		
		content += fmt.Sprintf("%s: %s (%s, %s)\n",
			dateStr, valueStr, changeStr, percentStr)
	}
	
	return hd.wrapContentWithBorders(" "+content+" ", width)
}

// wrapContentWithBorders wraps content with left and right borders
func (hd *HistoryDisplay) wrapContentWithBorders(content string, width int) string {
	contentWidth := lipgloss.Width(content)
	availableWidth := width - 2
	
	if contentWidth < availableWidth {
		padding := strings.Repeat(" ", availableWidth-contentWidth)
		content = content + padding
	} else if contentWidth > availableWidth {
		content = content[:availableWidth-3] + "..."
	}
	
	borderStyle := lipgloss.NewStyle().Foreground(hd.Theme.Primary())
	leftBorder := borderStyle.Render("│")
	rightBorder := borderStyle.Render("│")
	
	return leftBorder + content + rightBorder
}

// getHistoryWidth calculates history width based on terminal size
func (hd *HistoryDisplay) getHistoryWidth() int {
	if hd.TerminalSize.Width < 60 {
		return 60
	}
	return hd.TerminalSize.Width
}

// Helper functions for sample data generation

// generateSampleTransactions creates sample transaction data
func generateSampleTransactions() []TransactionRecord {
	now := time.Now()
	return []TransactionRecord{
		{1, "NABIL", "Buy", 10, 125000, 1250000, now.AddDate(0, 0, -1), "Settled", 0},
		{2, "EBL", "Sell", -20, 70000, -1400000, now.AddDate(0, 0, -2), "Settled", 40000},
		{3, "HIDCL", "Buy", 50, 44500, 2225000, now.AddDate(0, 0, -3), "Settled", 0},
		{4, "KTM", "Buy", 25, 89000, 2225000, now.AddDate(0, 0, -5), "Settled", 0},
		{5, "ADBL", "Sell", -15, 55000, -825000, now.AddDate(0, 0, -7), "Settled", 15000},
	}
}

// generateSamplePerformance creates sample performance data
func generateSamplePerformance() []PerformanceRecord {
	now := time.Now()
	return []PerformanceRecord{
		{now.AddDate(0, 0, -1), 24567000, 562000, 2.4},
		{now.AddDate(0, 0, -2), 24005000, -125000, -0.5},
		{now.AddDate(0, 0, -3), 24130000, 380000, 1.6},
		{now.AddDate(0, 0, -4), 23750000, -95000, -0.4},
		{now.AddDate(0, 0, -5), 23845000, 210000, 0.9},
	}
}

// generateSampleSettlements creates sample settlement data
func generateSampleSettlements() []Settlement {
	now := time.Now()
	return []Settlement{
		{1, "NABIL", 10, 1250000, now.AddDate(0, 0, 2), "Pending"},
		{3, "HIDCL", 50, 2225000, now.AddDate(0, 0, 1), "Pending"},
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
	return FormatCurrency(paisa)
}

// FormatPercent formats percentage with appropriate precision
func FormatPercent(percent float64) string {
	if percent >= 0 {
		return fmt.Sprintf("+%.1f%%", percent)
	}
	return fmt.Sprintf("%.1f%%", percent)
}