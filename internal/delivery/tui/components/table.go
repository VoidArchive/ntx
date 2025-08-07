// Package components
package components

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/voidarchive/ntx/internal/domain/models"
)

var (
	gruvBg     = lipgloss.Color("#282828")
	gruvOrange = lipgloss.Color("#d79921")
	gruvBlue   = lipgloss.Color("#458588")
	header     = lipgloss.NewStyle().Background(gruvBg).Foreground(gruvOrange).Bold(true)
	focusedRow = lipgloss.NewStyle().Foreground(gruvBlue)
)

func QuoteRow(q *models.Quote) table.Row {
	return table.Row{
		q.Symbol,
		fmt.Sprintf("%.2f", q.LTP),
		fmt.Sprintf("%.2f", q.High),
		fmt.Sprintf("%.2f", q.Low),
		humanize.Comma(int64(q.Volume)),
	}
}

func New(quotes []*models.Quote) table.Model {
	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].Symbol < quotes[j].Symbol
	})

	cols := []table.Column{
		{Title: "Symbol", Width: 8},
		{Title: "LTP", Width: 10},
		{Title: "High", Width: 10},
		{Title: "Low", Width: 10},
		{Title: "Vol", Width: 12},
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
