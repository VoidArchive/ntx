package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/voidarchive/ntx/internal/delivery/tui/components"
	"github.com/voidarchive/ntx/internal/domain/models"
)

// renderStocksPanel renders the right panel with an interactive table of stocks.
func (a *App) renderStocksPanel(width, height int) string {
	if len(a.quotes) == 0 {
		return a.renderLoadingBox("Loading stocks...", width, height)
	}

	// Ensure table exists and is sized correctly.
	if a.stockTable.Columns() == nil {
		a.initOrRefreshTable()
	}
	
	// Set proper table width accounting for borders and padding
	tableWidth := width - 6 // account for border (2) + padding (4)
	if tableWidth < 30 {
		tableWidth = 30 // minimum table width
	}
	a.stockTable.SetWidth(tableWidth)

	// Create clean title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("ALL STOCKS")

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
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", a.stockTable.View()))
}

// newStockTable builds a fresh table.Model for the given quotes using our
// components configuration to ensure consistent styling and behavior.
func (a *App) newStockTable(quotes []*models.Quote) table.Model {
	t := components.New(quotes)
	return t
}

// renderStockModal shows a centered overlay with details of the provided quote.
func (a *App) renderStockModal(q *models.Quote, totalWidth, totalHeight int) string {
	// Compose detail body with key fields so users can quickly inspect a symbol.
	body := fmt.Sprintf(
		"%s\n\nLTP: %.2f\nOpen: %.2f  High: %.2f  Low: %.2f\nPrev Close: %.2f\nVolume: %.0f\n\nPress ESC to close",
		lipgloss.NewStyle().Bold(true).Render(q.Symbol),
		q.LTP, q.Open, q.High, q.Low, q.PrevClose, q.Volume,
	)

	modalWidth := minInt(56, totalWidth-8)
	modalHeight := 10

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("105")).
		Background(lipgloss.Color("236")).
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight).
		Render(body)

	// Center the box in the available space
	return lipgloss.Place(totalWidth, totalHeight,
		lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
