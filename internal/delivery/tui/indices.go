package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderIndicesPanel renders the left panel with NEPSE overview and sub-indices
func (a *App) renderIndicesPanel(width, height int) string {
	if a.loading {
		return a.renderLoadingBox("Loading market data...", width, height)
	}

	if a.err != nil {
		return a.renderErrorBox(fmt.Sprintf("Error: %v", a.err), width, height)
	}

	if a.overview == nil {
		return a.renderErrorBox("No market data available", width, height)
	}

	// Split height: 1/3 for main index, 2/3 for sub-indices
	mainHeight := height / 3
	subHeight := height - mainHeight - 1 // -1 for separator

	// Render main NEPSE overview
	mainPanel := a.renderMainIndex(width, mainHeight)

	// Render sub-indices
	subPanel := a.renderSubIndices(width, subHeight)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		mainPanel,
		subPanel,
	)
}

// renderMainIndex renders the main NEPSE index overview
func (a *App) renderMainIndex(width, height int) string {
	if a.overview.MainIndex == nil {
		return a.renderErrorBox("Main index not available", width, height)
	}

	idx := a.overview.MainIndex

	// Determine color based on point change
	changeColor := lipgloss.Color("196") // Red
	changeSymbol := "▼"
	if idx.PointChange >= 0 {
		changeColor = lipgloss.Color("46") // Green
		changeSymbol = "▲"
	}

	// Build content
	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("📈 NEPSE INDEX"))
	content.WriteString("\n\n")

	// Index value and change
	indexValue := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")).
		Render(fmt.Sprintf("%.2f", idx.Close))

	pointChange := lipgloss.NewStyle().
		Foreground(changeColor).
		Render(fmt.Sprintf("%s %.2f", changeSymbol, idx.PointChange))

	content.WriteString(fmt.Sprintf("%s %s\n\n", indexValue, pointChange))

	// OHLC data
	content.WriteString(fmt.Sprintf("Open:  %.2f\n", idx.Open))
	content.WriteString(fmt.Sprintf("High:  %.2f\n", idx.High))
	content.WriteString(fmt.Sprintf("Low:   %.2f\n", idx.Low))
	content.WriteString(fmt.Sprintf("Time:  %s", a.overview.LastUpdated))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(width).
		Height(height).
		Render(content.String())
}

// renderSubIndices renders all sub-indices in a scrollable list
func (a *App) renderSubIndices(width, height int) string {
	if len(a.overview.SubIndices) == 0 {
		return a.renderErrorBox("No sub-indices available", width, height)
	}

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33")).
		Render("📊 SUB INDICES"))
	content.WriteString("\n\n")

	// Create table-like format for sub-indices
	for _, idx := range a.overview.SubIndices {
		// Determine color based on point change
		changeColor := lipgloss.Color("196") // Red
		changeSymbol := "▼"
		if idx.PointChange >= 0 {
			changeColor = lipgloss.Color("46") // Green
			changeSymbol = "▲"
		}

		// Format name (truncate if too long)
		name := idx.Name
		if len(name) > 15 {
			name = name[:12] + "..."
		}

		// Build row
		nameStyle := lipgloss.NewStyle().Width(18).Render(name)
		valueStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226")).
			Width(10).
			Align(lipgloss.Right).
			Render(fmt.Sprintf("%.2f", idx.Close))

		changeStyle := lipgloss.NewStyle().
			Foreground(changeColor).
			Width(10).
			Align(lipgloss.Right).
			Render(fmt.Sprintf("%s%.1f", changeSymbol, idx.PointChange))

		row := lipgloss.JoinHorizontal(lipgloss.Top, nameStyle, valueStyle, changeStyle)
		content.WriteString(row + "\n")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(width).
		Height(height).
		Render(content.String())
}

// Helper methods for loading and error states
func (a *App) renderLoadingBox(message string, width, height int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("214")).
		Render(message)
}

func (a *App) renderErrorBox(message string, width, height int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("196")).
		Render(message)
}
