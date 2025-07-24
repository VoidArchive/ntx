package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#3C82F6") // Blue
	ColorSecondary = lipgloss.Color("#6B7280") // Gray
	ColorAccent    = lipgloss.Color("#8B5CF6") // Purple

	// Status colors
	ColorSuccess = lipgloss.Color("#10B981") // Green
	ColorDanger  = lipgloss.Color("#EF4444") // Red
	ColorWarning = lipgloss.Color("#F59E0B") // Yellow
	ColorInfo    = lipgloss.Color("#3B82F6") // Blue

	// Neutral colors
	ColorBorder      = lipgloss.Color("#374151") // Dark gray
	ColorBorderLight = lipgloss.Color("#6B7280") // Light gray
	ColorBackground  = lipgloss.Color("#111827") // Very dark gray
	ColorForeground  = lipgloss.Color("#F9FAFB") // Off white
	ColorMuted       = lipgloss.Color("#9CA3AF") // Medium gray
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Background(ColorBackground)

	FocusedStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Background(lipgloss.Color("#1F2937")).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)

// Border styles
var (
	RoundedBorder = lipgloss.RoundedBorder()
	ThickBorder   = lipgloss.ThickBorder()
	NormalBorder  = lipgloss.NormalBorder()

	PanelStyle = lipgloss.NewStyle().
			Border(RoundedBorder).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	HeaderStyle = lipgloss.NewStyle().
			Border(lipgloss.Border{
				Top:    "─",
				Bottom: "─",
				Left:   "│",
				Right:  "│",
			}).
			BorderForeground(ColorBorderLight).
			Bold(true).
			Align(lipgloss.Center).
			Padding(0, 1)

	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				MarginBottom(1)
)

// Table styles
var (
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorForeground).
				Background(ColorBorder).
				Padding(0, 1).
				Align(lipgloss.Center)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Align(lipgloss.Left)

	TableRowEvenStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1F2937"))

	TableRowOddStyle = lipgloss.NewStyle().
				Background(ColorBackground)

	TableSelectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#374151")).
				Foreground(ColorAccent).
				Bold(true)
)

// Financial styles
var (
	ProfitStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	LossStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	NeutralStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	MoneyStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Bold(true)

	PercentageStyle = lipgloss.NewStyle().
				Bold(true)
)

// Button styles
var (
	ActiveButtonStyle = lipgloss.NewStyle().
				Foreground(ColorForeground).
				Background(ColorPrimary).
				Padding(0, 3).
				MarginRight(1).
				Bold(true)

	InactiveButtonStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Background(ColorBorder).
				Padding(0, 3).
				MarginRight(1)

	DangerButtonStyle = lipgloss.NewStyle().
				Foreground(ColorForeground).
				Background(ColorDanger).
				Padding(0, 3).
				MarginRight(1).
				Bold(true)
)

// Input styles
var (
	InputStyle = lipgloss.NewStyle().
			Border(NormalBorder).
			BorderForeground(ColorBorderLight).
			Padding(0, 1).
			Width(20)

	FocusedInputStyle = lipgloss.NewStyle().
				Border(NormalBorder).
				BorderForeground(ColorPrimary).
				Padding(0, 1).
				Width(20)
)

// Status styles
var (
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Bold(true)
)

// Navigation styles
var (
	NavigationStyle = lipgloss.NewStyle().
			Border(lipgloss.Border{
				Top:    "─",
				Bottom: "─",
			}).
			BorderForeground(ColorBorderLight).
			Padding(0, 2).
			MarginBottom(1)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Background(ColorPrimary).
			Padding(0, 2).
			MarginRight(1).
			Bold(true)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 2).
				MarginRight(1)
)

// Progress bar styles
var (
	ProgressBarStyle = lipgloss.NewStyle().
				Border(NormalBorder).
				BorderForeground(ColorBorderLight).
				Padding(0, 1).
				Width(40)

	ProgressFillStyle = lipgloss.NewStyle().
				Background(ColorSuccess)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Background(ColorBorder)
)

// Dialog styles
var (
	DialogStyle = lipgloss.NewStyle().
			Border(RoundedBorder).
			BorderForeground(ColorBorderLight).
			Padding(1, 2).
			Background(ColorBackground).
			Align(lipgloss.Center)

	DialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				Align(lipgloss.Center).
				MarginBottom(1)

	DialogContentStyle = lipgloss.NewStyle().
				Foreground(ColorForeground).
				MarginBottom(1)
)

// Help styles
var (
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	KeybindStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)

// Utility functions
func StyleForMoney(isPositive bool, isZero bool) lipgloss.Style {
	if isZero {
		return NeutralStyle
	}
	if isPositive {
		return ProfitStyle
	}
	return LossStyle
}

func StyleForPercentage(percentage float64) lipgloss.Style {
	style := PercentageStyle
	if percentage > 0 {
		return style.Foreground(ColorSuccess)
	} else if percentage < 0 {
		return style.Foreground(ColorDanger)
	}
	return style.Foreground(ColorMuted)
}

func StyleForTableRow(index int, isSelected bool) lipgloss.Style {
	if isSelected {
		return TableSelectedRowStyle
	}
	if index%2 == 0 {
		return TableRowEvenStyle
	}
	return TableRowOddStyle
}

func WithWidth(style lipgloss.Style, width int) lipgloss.Style {
	return style.Width(width)
}

func WithHeight(style lipgloss.Style, height int) lipgloss.Style {
	return style.Height(height)
}