package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/voidarchive/ntx/internal/domain/models"
)

// renderStocksPanel renders the right panel with all stocks table
func (a *App) renderStocksPanel(width, height int) string {
	if len(a.quotes) == 0 {
		return a.renderLoadingBox("Loading stocks...", width, height)
	}

	var content strings.Builder

	// Header
	content.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("📋 ALL STOCKS"))
	content.WriteString("\n\n")

	// Table header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	symbolHeader := headerStyle.Copy().Width(8).Render("Symbol")
	ltpHeader := headerStyle.Copy().Width(10).Align(lipgloss.Right).Render("LTP")
	changeHeader := headerStyle.Copy().Width(12).Align(lipgloss.Right).Render("Change")
	volumeHeader := headerStyle.Copy().Width(12).Align(lipgloss.Right).Render("Volume")

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		symbolHeader, ltpHeader, changeHeader, volumeHeader)
	content.WriteString(header + "\n")

	// Separator line
	separator := strings.Repeat("─", width-4)
	content.WriteString(separator + "\n")

	// Table rows (show only as many as fit in height)
	maxRows := height - 8 // Reserve space for header, title, borders
	displayQuotes := a.quotes
	if len(displayQuotes) > maxRows {
		displayQuotes = displayQuotes[:maxRows]
	}

	for _, quote := range displayQuotes {
		content.WriteString(a.renderStockRow(quote) + "\n")
	}

	// Show count info at bottom if truncated
	if len(a.quotes) > maxRows {
		truncateInfo := fmt.Sprintf("Showing %d of %d stocks", maxRows, len(a.quotes))
		content.WriteString("\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render(truncateInfo))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(width).
		Height(height).
		Render(content.String())
}

// renderStockRow renders a single stock row
func (a *App) renderStockRow(quote *models.Quote) string {
	// Calculate change
	change := quote.LTP - quote.PrevClose
	changePct := 0.0
	if quote.PrevClose > 0 {
		changePct = (change / quote.PrevClose) * 100
	}

	// Determine color based on change
	changeColor := lipgloss.Color("196") // Red
	changeSymbol := "▼"
	if change >= 0 {
		changeColor = lipgloss.Color("46") // Green
		changeSymbol = "▲"
	}

	// Format columns
	symbolStyle := lipgloss.NewStyle().
		Width(8).
		Foreground(lipgloss.Color("255")).
		Render(truncateString(quote.Symbol, 7))

	ltpStyle := lipgloss.NewStyle().
		Width(10).
		Align(lipgloss.Right).
		Bold(true).
		Foreground(lipgloss.Color("226")).
		Render(fmt.Sprintf("%.2f", quote.LTP))

	changeStyle := lipgloss.NewStyle().
		Width(12).
		Align(lipgloss.Right).
		Foreground(changeColor).
		Render(fmt.Sprintf("%s %.2f%%", changeSymbol, changePct))

	volumeStyle := lipgloss.NewStyle().
		Width(12).
		Align(lipgloss.Right).
		Foreground(lipgloss.Color("245")).
		Render(formatVolume(quote.Volume))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		symbolStyle, ltpStyle, changeStyle, volumeStyle)
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func formatVolume(volume float64) string {
	if volume >= 1000000 {
		return fmt.Sprintf("%.1fM", volume/1000000)
	} else if volume >= 1000 {
		return fmt.Sprintf("%.1fK", volume/1000)
	}
	return fmt.Sprintf("%.0f", volume)
}
