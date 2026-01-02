package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	v1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/portfolio"
)

type View int

const (
	HoldingsView View = iota
	TransactionsView
	SummaryView
)

// Messages
type holdingsMsg []*v1.Holding
type transactionsMsg []*v1.Transaction
type summaryMsg *v1.PortfolioSummary
type syncDoneMsg struct{ updated, failed int }
type errMsg error

type Model struct {
	service *portfolio.Service
	ctx     context.Context

	activeView   View
	holdings     []*v1.Holding
	transactions []*v1.Transaction
	summary      *v1.PortfolioSummary

	syncing bool
	spinner spinner.Model
	err     error

	width  int
	height int
}

func New(service *portfolio.Service) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		service:    service,
		ctx:        context.Background(),
		activeView: HoldingsView,
		spinner:    s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadHoldings,
		m.loadTransactions,
		m.loadSummary,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't handle keys while syncing
		if m.syncing {
			return m, nil
		}

		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "1", "h":
			m.activeView = HoldingsView
		case "2", "t":
			m.activeView = TransactionsView
		case "3", "u":
			m.activeView = SummaryView

		case "s":
			m.syncing = true
			return m, tea.Batch(m.spinner.Tick, m.syncPrices)

		case "r":
			return m, tea.Batch(m.loadHoldings, m.loadTransactions, m.loadSummary)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		if m.syncing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case holdingsMsg:
		m.holdings = msg

	case transactionsMsg:
		m.transactions = msg

	case summaryMsg:
		m.summary = msg

	case syncDoneMsg:
		m.syncing = false
		// Auto-reload data after sync
		return m, tea.Batch(m.loadHoldings, m.loadTransactions, m.loadSummary)

	case errMsg:
		m.err = msg
		m.syncing = false
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string
	switch m.activeView {
	case HoldingsView:
		content = m.viewHoldings()
	case TransactionsView:
		content = m.viewTransactions()
	case SummaryView:
		content = m.viewSummary()
	}

	return fmt.Sprintf("%s\n\n%s\n\n%s",
		m.viewTabs(),
		content,
		m.viewStatus(),
	)
}

func (m Model) viewTabs() string {
	tabs := []struct {
		name string
		key  string
		view View
	}{
		{"Holdings", "1", HoldingsView},
		{"Transactions", "2", TransactionsView},
		{"Summary", "3", SummaryView},
	}

	var rendered string
	for _, tab := range tabs {
		style := inactiveTabStyle
		if m.activeView == tab.view {
			style = activeTabStyle
		}
		rendered += style.Render(fmt.Sprintf("[%s] %s", tab.key, tab.name)) + " "
	}

	return rendered
}

func (m Model) viewStatus() string {
	if m.syncing {
		return statusStyle.Render(m.spinner.View() + " Syncing prices...")
	}

	if m.err != nil {
		return lossStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	return helpStyle.Render("s: sync prices  r: refresh  q: quit")
}

// Commands

func (m Model) loadHoldings() tea.Msg {
	holdings, err := m.service.ListHoldings(m.ctx)
	if err != nil {
		return errMsg(err)
	}
	return holdingsMsg(holdings)
}

func (m Model) loadTransactions() tea.Msg {
	txs, _, err := m.service.ListTransactions(m.ctx, "", v1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED, 50, 0)
	if err != nil {
		return errMsg(err)
	}
	return transactionsMsg(txs)
}

func (m Model) loadSummary() tea.Msg {
	summary, err := m.service.Summary(m.ctx)
	if err != nil {
		return errMsg(err)
	}
	return summaryMsg(summary)
}

func (m Model) syncPrices() tea.Msg {
	result, err := m.service.SyncPrices(m.ctx)
	if err != nil {
		return errMsg(err)
	}
	return syncDoneMsg{updated: result.Updated, failed: result.Failed}
}

// Run starts the TUI application.
func Run(service *portfolio.Service) error {
	p := tea.NewProgram(New(service), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
