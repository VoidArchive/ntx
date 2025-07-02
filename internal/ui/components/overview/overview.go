/*
NTX Portfolio Management TUI - Portfolio Overview Component

Single-line overview widget displaying key portfolio metrics in a compact,
professional format similar to btop's CPU/memory summary bars.

Provides essential portfolio metrics at a glance:
- Total portfolio value and cost basis
- Unrealized P/L with percentage change
- Today's change (when available)
- Number of holdings

Designed for FR1.1 implementation as specified in phase3-portfolio-management.md
*/

package overview

import (
	"fmt"
	"strings"
	"time"

	"ntx/internal/ui/themes"

	"github.com/charmbracelet/lipgloss"
)

// OverviewDisplay provides portfolio summary widget
// Compact single-line format maximizes screen real estate for holdings table
type OverviewDisplay struct {
	TotalValue     int64        // Total portfolio market value in paisa
	TotalCost      int64        // Total cost basis in paisa
	UnrealizedPL   int64        // Unrealized P/L in paisa
	DayChange      int64        // Today's change in paisa (optional)
	HoldingsCount  int          // Number of holdings
	LastUpdate     time.Time    // Last data refresh timestamp
	Theme          themes.Theme // Current theme for styling
	TerminalSize   struct {     // Terminal dimensions for responsive layout
		Width  int
		Height int
	}
	OverrideWidth  int          // Override width for alignment with other components (0 = use terminal width)
}

// NewOverviewDisplay creates overview display with default configuration
func NewOverviewDisplay(theme themes.Theme) *OverviewDisplay {
	return &OverviewDisplay{
		TotalValue:    0,
		TotalCost:     0,
		UnrealizedPL:  0,
		DayChange:     0,
		HoldingsCount: 0,
		LastUpdate:    time.Now(),
		Theme:         theme,
		TerminalSize:  struct{ Width, Height int }{Width: 120, Height: 40},
	}
}

// UpdatePortfolioSummary refreshes portfolio metrics
func (od *OverviewDisplay) UpdatePortfolioSummary(totalValue, totalCost, unrealizedPL, dayChange int64, holdingsCount int) {
	od.TotalValue = totalValue
	od.TotalCost = totalCost
	od.UnrealizedPL = unrealizedPL
	od.DayChange = dayChange
	od.HoldingsCount = holdingsCount
	od.LastUpdate = time.Now()
}

// SetTerminalSize updates responsive layout configuration
func (od *OverviewDisplay) SetTerminalSize(width, height int) {
	od.TerminalSize.Width = width
	od.TerminalSize.Height = height
}

// SetTheme updates theme and refreshes styling
func (od *OverviewDisplay) SetTheme(theme themes.Theme) {
	od.Theme = theme
}

// SetWidth sets a specific width for alignment with other components
func (od *OverviewDisplay) SetWidth(width int) {
	od.OverrideWidth = width
}

// Render generates the portfolio overview widget
// Creates btop-style bordered box with key metrics
func (od *OverviewDisplay) Render() string {
	width := od.getOverviewWidth()
	
	// Create bordered box with title
	topBorder := od.renderTopBorder()
	content := od.renderContent(width)
	bottomBorder := od.renderBottomBorder()
	
	return topBorder + "\n" + content + "\n" + bottomBorder
}

// renderTopBorder creates top border with title
func (od *OverviewDisplay) renderTopBorder() string {
	width := od.getOverviewWidth()
	title := "[2]Holdings"
	
	if width < len(title)+10 {
		border := "┌" + strings.Repeat("─", width-2) + "┐"
		return lipgloss.NewStyle().Foreground(od.Theme.Primary()).Render(border)
	}
	
	titleSection := "─" + title + "─"
	remainingWidth := width - len([]rune(titleSection)) - 2
	leftPadding := strings.Repeat("─", remainingWidth)
	
	border := "┌" + titleSection + leftPadding + "┐"
	return lipgloss.NewStyle().Foreground(od.Theme.Primary()).Render(border)
}

// renderBottomBorder creates bottom border
func (od *OverviewDisplay) renderBottomBorder() string {
	width := od.getOverviewWidth()
	border := "└" + strings.Repeat("─", width-2) + "┘"
	return lipgloss.NewStyle().Foreground(od.Theme.Primary()).Render(border)
}

// renderContent creates the main content line with portfolio metrics
func (od *OverviewDisplay) renderContent(width int) string {
	// Calculate percentage change
	var percentChange float64
	if od.TotalCost > 0 {
		percentChange = float64(od.UnrealizedPL) / float64(od.TotalCost) * 100
	}
	
	// Format metric strings
	totalValueStr := FormatCurrency(od.TotalValue)
	percentChangeStr := FormatPercent(percentChange)
	unrealizedPLStr := FormatPL(od.UnrealizedPL)
	
	// Build content string based on terminal width
	var content string
	if width >= 120 {
		// Full format: Total: Rs.2,45,670 (+1.8%) │ Today: +Rs.5,620 │ Unrealized: +Rs.12,340 │ Holdings: 5
		dayChangeStr := ""
		if od.DayChange != 0 {
			dayChangeStr = fmt.Sprintf(" │ Today: %s", FormatPL(od.DayChange))
		}
		content = fmt.Sprintf(" Total: %s (%s)%s │ Unrealized: %s │ Holdings: %d ",
			totalValueStr, percentChangeStr, dayChangeStr, unrealizedPLStr, od.HoldingsCount)
	} else if width >= 100 {
		// Medium format: Total: Rs.2,45,670 (+1.8%) │ P/L: +Rs.12,340 │ Holdings: 5
		content = fmt.Sprintf(" Total: %s (%s) │ P/L: %s │ Holdings: %d ",
			totalValueStr, percentChangeStr, unrealizedPLStr, od.HoldingsCount)
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
	if od.UnrealizedPL > 0 {
		style = lipgloss.NewStyle().Foreground(od.Theme.Success())
	} else if od.UnrealizedPL < 0 {
		style = lipgloss.NewStyle().Foreground(od.Theme.Error())
	} else {
		style = lipgloss.NewStyle().Foreground(od.Theme.Foreground())
	}
	
	styledContent := style.Render(content)
	
	// Pad to exact width
	contentWidth := lipgloss.Width(styledContent)
	availableWidth := width - 2 // Account for borders
	
	if contentWidth < availableWidth {
		padding := strings.Repeat(" ", availableWidth-contentWidth)
		styledContent = styledContent + padding
	} else if contentWidth > availableWidth {
		// Truncate if too long
		styledContent = styledContent[:availableWidth-3] + "..."
	}
	
	// Apply border styling
	borderStyle := lipgloss.NewStyle().Foreground(od.Theme.Primary())
	leftBorder := borderStyle.Render("│")
	rightBorder := borderStyle.Render("│")
	
	return leftBorder + styledContent + rightBorder
}

// getOverviewWidth calculates overview widget width based on terminal size or override
func (od *OverviewDisplay) getOverviewWidth() int {
	// Use override width if set (for alignment with other components)
	if od.OverrideWidth > 0 {
		// Use override width for alignment with other components
		return od.OverrideWidth
	}
	
	// Otherwise use terminal size with minimum constraint
	if od.TerminalSize.Width < 60 {
		return 60
	}
	// Use terminal width when no override is set
	return od.TerminalSize.Width
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