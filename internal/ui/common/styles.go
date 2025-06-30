package common

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Color scheme for NEPSE Power Terminal
var (
	// Primary brand colors
	Primary   = lipgloss.Color("#7D56F4") // Purple
	Secondary = lipgloss.Color("#874BFD") // Light purple
	Accent    = lipgloss.Color("#F25D94") // Pink

	// Market colors
	GainColor    = lipgloss.Color("#00FF87") // Bright green
	LossColor    = lipgloss.Color("#FF5757") // Bright red
	NeutralColor = lipgloss.Color("#8E8E93") // Gray

	// UI colors
	Background = lipgloss.Color("#1E1E2E") // Dark background
	Surface    = lipgloss.Color("#313244") // Card/pane background
	Text       = lipgloss.Color("#CDD6F4") // Primary text
	Muted      = lipgloss.Color("#6C7086") // Secondary text
	Border     = lipgloss.Color("#45475A") // Border color

	// Status colors
	Success = lipgloss.Color("#A6E3A1") // Green
	Warning = lipgloss.Color("#F9E2AF") // Yellow
	Error   = lipgloss.Color("#F38BA8") // Red
	Info    = lipgloss.Color("#89B4FA") // Blue
)

// Base styles for consistent theming
var (
	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(Primary).
			Padding(0, 1).
			Bold(true)

	SubHeaderStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(Secondary).
			Padding(0, 1)

	// Panel styles
	PanelStyle = lipgloss.NewStyle().
			Background(Surface).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2)

	ActivePanelStyle = lipgloss.NewStyle().
				Background(Surface).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(1, 2)

	// Content styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(Text).
			Bold(true).
			Margin(0, 0, 1, 0)

	ContentStyle = lipgloss.NewStyle().
			Foreground(Text)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Market data styles
	GainStyle = lipgloss.NewStyle().
			Foreground(GainColor).
			Bold(true)

	LossStyle = lipgloss.NewStyle().
			Foreground(LossColor).
			Bold(true)

	NeutralStyle = lipgloss.NewStyle().
			Foreground(NeutralColor)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Info)

	// Navigation styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Margin(1, 0)

	KeyStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(Text).
				Background(Surface).
				Bold(true).
				Padding(0, 1)

	TableCellStyle = lipgloss.NewStyle().
			Foreground(Text).
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(Background).
				Background(Primary).
				Padding(0, 1)
)

// FormatMoney formats money with appropriate color coding
func FormatMoney(amount float64, showSign bool) string {
	var style lipgloss.Style
	var prefix string

	if amount > 0 {
		style = GainStyle
		if showSign {
			prefix = "+"
		}
	} else if amount < 0 {
		style = LossStyle
		// Negative sign already included in amount
	} else {
		style = NeutralStyle
	}

	return style.Render(prefix + formatCurrency(amount))
}

// FormatPercentage formats percentage with color coding
func FormatPercentage(percent float64) string {
	var style lipgloss.Style
	var prefix string

	if percent > 0 {
		style = GainStyle
		prefix = "+"
	} else if percent < 0 {
		style = LossStyle
	} else {
		style = NeutralStyle
	}

	return style.Render(prefix + formatPercent(percent))
}

// Helper functions for number formatting
func formatCurrency(amount float64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.2fM", amount/1000000)
	} else if amount >= 1000 {
		return fmt.Sprintf("%.2fK", amount/1000)
	}
	return fmt.Sprintf("%.2f", amount)
}

func formatPercent(percent float64) string {
	return fmt.Sprintf("%.2f%%", percent)
}

// Layout helper functions
func CreateRow(items ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, items...)
}

func CreateColumn(items ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

// Responsive width calculation
func CalculatePaneWidth(totalWidth, paneCount int) int {
	border := 2 // Border width per pane
	margin := 1 // Margin between panes

	usableWidth := totalWidth - (paneCount * border) - ((paneCount - 1) * margin)
	return usableWidth / paneCount
}

// Market status indicators
func MarketStatusIndicator(isOpen bool) string {
	if isOpen {
		return SuccessStyle.Render("● OPEN")
	}
	return ErrorStyle.Render("● CLOSED")
}

// Data freshness indicator
func DataFreshnessIndicator(age time.Duration) string {
	if age < time.Minute {
		return SuccessStyle.Render("● LIVE")
	} else if age < 5*time.Minute {
		return WarningStyle.Render("● RECENT")
	} else {
		return ErrorStyle.Render("● STALE")
	}
}
