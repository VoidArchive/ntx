// Package components
package components

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/voidarchive/ntx/internal/domain/models"
)

var (
	gruvBg     = lipgloss.Color("#282828")
	gruvOrange = lipgloss.Color("#d79921")
	gruvBlue   = lipgloss.Color("#458588")
	greenColor = lipgloss.Color("46")  // Green for positive
	redColor   = lipgloss.Color("196") // Red for negative
	header     = lipgloss.NewStyle().Background(gruvBg).Foreground(gruvOrange).Bold(true)
	focusedRow = lipgloss.NewStyle().Foreground(gruvBlue)
)

func QuoteRow(q *models.Quote) table.Row {
	changePercent := q.PercentageChange()

	// Format change with color
	changeColor := redColor
	if q.IsPositive() {
		changeColor = greenColor
	}

	changeStr := lipgloss.NewStyle().
		Foreground(changeColor).
		Render(fmt.Sprintf("%+.2f%%", changePercent))

	return table.Row{
		q.Symbol,
		fmt.Sprintf("%.2f", q.LTP),
		changeStr,
	}
}

func New(quotes []*models.Quote) table.Model {
	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].Symbol < quotes[j].Symbol
	})

	// Calculate dynamic column widths based on available width
	// Assume minimum table width of 32 chars for proper display
	symbolWidth := 8
	ltpWidth := 10
	changeWidth := 10

	cols := []table.Column{
		{Title: "Symbol", Width: symbolWidth},
		{Title: "LTP", Width: ltpWidth},
		{Title: "Change", Width: changeWidth},
	}
	rows := make([]table.Row, len(quotes))
	for i, q := range quotes {
		rows[i] = QuoteRow(q)
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	styles := table.DefaultStyles()
	styles.Header = header
	styles.Selected = focusedRow
	t.SetStyles(styles)

	return t
}
