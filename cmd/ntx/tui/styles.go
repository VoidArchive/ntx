package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorGreen  = lipgloss.Color("2")
	colorRed    = lipgloss.Color("1")
	colorDim    = lipgloss.Color("8")
	colorAccent = lipgloss.Color("6")
	colorWhite  = lipgloss.Color("15")

	// Text styles
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(colorDim)
	profitStyle = lipgloss.NewStyle().Foreground(colorGreen)
	lossStyle   = lipgloss.NewStyle().Foreground(colorRed)
	dimStyle    = lipgloss.NewStyle().Foreground(colorDim)

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorAccent).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Padding(0, 2)

	// Help style
	helpStyle = lipgloss.NewStyle().Foreground(colorDim)

	// Status bar
	statusStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)
)
