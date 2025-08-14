package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/voidarchive/ntx/internal/domain/models"
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

	// Create clean panel with title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("NEPSE INDEX")
		
	combinedContent := mainPanel + "\n\n" + subPanel
	
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true).
		Padding(1).
		Width(width).
		Height(height).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", combinedContent))
}

// renderMainIndex renders the main NEPSE index overview
func (a *App) renderMainIndex(width, height int) string {
	if a.overview.MainIndex == nil {
		return a.renderErrorBox("Main index not available", width, height)
	}

	idx := a.overview.MainIndex

	// Determine color based on point change
	changeColor := lipgloss.Color("196") // Red
	if idx.PointChange >= 0 {
		changeColor = lipgloss.Color("46") // Green
	}

	// Build content - no duplicate title
	var content strings.Builder

	// Calculate percentage change
	percentChange := 0.0
	if idx.Open > 0 {
		percentChange = ((idx.Close - idx.Open) / idx.Open) * 100
	}

	// Main index display with better formatting
	mainIndexStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226"))
	
	content.WriteString(mainIndexStyle.Render(fmt.Sprintf("%.2f", idx.Close)))
	content.WriteString("  ")
	
	changeStyle := lipgloss.NewStyle().Foreground(changeColor)
	content.WriteString(changeStyle.Render(fmt.Sprintf("%.2f (%.2f%%)", idx.PointChange, percentChange)))
	content.WriteString("\n\n")

	// OHLC data in compact format
	content.WriteString(fmt.Sprintf("Open: %.2f  |  High: %.2f\n", idx.Open, idx.High))
	content.WriteString(fmt.Sprintf("Low: %.2f   |  Time: %s", idx.Low, a.overview.LastUpdated))

	return content.String()
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
		Render("SUB INDICES"))
	content.WriteString("\n\n")

	// Filter and create table-like format for relevant sub-indices only
	var relevantIndices []*models.Index
	for _, idx := range a.overview.SubIndices {
		// Only include proper sector indices, skip Sensitive, Float, etc.
		name := strings.ToLower(idx.Name)
		if strings.Contains(name, "sensitive") || 
		   strings.Contains(name, "float") ||
		   strings.Contains(name, "trading") ||
		   strings.Contains(name, "others") {
			continue // Skip these indices
		}
		relevantIndices = append(relevantIndices, idx)
	}
	
	for _, idx := range relevantIndices {
		// Determine color based on point change
		changeColor := lipgloss.Color("196") // Red
		if idx.PointChange >= 0 {
			changeColor = lipgloss.Color("46") // Green
		}

		// Format name (clean up index names)
		name := strings.ReplaceAll(idx.Name, "Index", "")
		name = strings.ReplaceAll(name, "SubIndex", "")
		name = strings.TrimSpace(name)
		
		if len(name) > 16 {
			name = name[:13] + "..."
		}

		// Build row with better spacing
		nameStyle := lipgloss.NewStyle().Width(18).Render(name)
		valueStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226")).
			Width(12).
			Align(lipgloss.Right).
			Render(fmt.Sprintf("%.2f", idx.Close))

		changeStyle := lipgloss.NewStyle().
			Foreground(changeColor).
			Width(8).
			Align(lipgloss.Right).
			Render(fmt.Sprintf("%.1f", idx.PointChange))

		row := lipgloss.JoinHorizontal(lipgloss.Top, nameStyle, valueStyle, changeStyle)
		content.WriteString(row + "\n")
	}

	return content.String()
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

// The following helpers are declared in app.go but referenced here via methods
// on App; keeping them in app.go maintains separation of concerns:
// - initOrRefreshTable()
// - syncTableSize()
// - findQuoteBySymbol()
