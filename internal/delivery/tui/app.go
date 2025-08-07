// Package tui
package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/voidarchive/ntx/internal/delivery/tui/components"
	"github.com/voidarchive/ntx/internal/service/market"
)

const refreshInterval = 30 * time.Second

type refreshMsg struct{}

type quoteMsg struct {
	quotes []*table.Row
	err    error
}

type model struct {
	svc    market.Service
	table  table.Model
	err    error
	width  int
	height int
}

func New(svc market.Service) tea.Model {
	t := components.New(nil)
	return &model{svc: svc, table: t}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.fetchCmd(),
		tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }),
	)
}

func (m *model) fetchCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		qs, err := m.svc.GetLiveQuotes(ctx)
		if err != nil {
			return quoteMsg{err: err}
		}
		rows := make([]*table.Row, len(qs))
		for i, q := range qs {
			r := components.QuoteRow(q)
			rows[i] = &r
		}
		return quoteMsg{quotes: rows}
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case quoteMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.table.SetRows(deref(msg.quotes))
		m.resizeCols(m.table.Width())
		return m, nil
	case refreshMsg:
		return m, m.fetchCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			return m, m.fetchCmd()
		}
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 2)
		m.resizeCols(msg.Width)
		return m, nil
	}
	var cmd tea.Cmd
	if m.table.Columns() != nil {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

func (m *model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\n%s", m.err, m.table.View())
	}
	if len(m.table.Rows()) == 0 {
		return "Loading quotes..."
	}
	return m.table.View()
}

func deref[T any](ptrs []*T) []T {
	out := make([]T, len(ptrs))
	for i, p := range ptrs {
		out[i] = *p
	}
	return out
}

func (m *model) resizeCols(total int) {
	padding := 4 // gutters
	avail := total - padding
	widths := []int{8, 10, 10, 10, 12} // base widths

	sum := 0
	for _, w := range widths {
		sum += w
	}
	extra := avail - sum
	if extra > 0 {
		share := extra / len(widths)
		for i := range widths {
			widths[i] += share
		}
		widths[len(widths)-1] += extra % len(widths) // remainder
	}

	cols := m.table.Columns()
	for i, w := range widths {
		cols[i].Width = w
	}
	m.table.SetColumns(cols)
}
