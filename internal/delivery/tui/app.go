// Package tui provides terminal user interface
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/voidarchive/ntx/internal/domain/models"
	"github.com/voidarchive/ntx/internal/service/market"
)

// App is the main TUI application
type App struct {
	marketService market.Service
	width         int
	height        int

	// Data
	overview *models.MarketOverview
	quotes   []*models.Quote

	// UI State
	loading bool
	err     error

	// Interactive stocks table and modal state
	stockTable    table.Model
	showModal     bool
	selectedQuote *models.Quote
}

// NewApp creates a new TUI application
func NewApp(marketService market.Service) *App {
	return &App{
		marketService: marketService,
		loading:       true,
	}
}

// Init starts the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		loadMarketOverview(a.marketService),
		loadQuotes(a.marketService),
		tick(), // Start automatic refresh timer
	)
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Keep interactive table sized with viewport
		if a.stockTable.Columns() != nil {
			a.syncTableSize()
		}
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if a.showModal {
				a.showModal = false
				return a, nil
			}
			return a, tea.Quit
		case "r", "F5":
			a.loading = true
			a.err = nil
			return a, tea.Batch(
				loadMarketOverview(a.marketService),
				loadQuotes(a.marketService),
			)
		case "enter", " ":
			if a.stockTable.Rows() != nil && len(a.stockTable.Rows()) > 0 {
				row := a.stockTable.SelectedRow()
				if len(row) > 0 {
					a.selectedQuote = a.findQuoteBySymbol(row[0])
					if a.selectedQuote != nil {
						a.showModal = true
					}
				}
			}
			return a, nil
		case "esc":
			if a.showModal {
				a.showModal = false
				return a, nil
			}
			return a, nil
		}

		// Forward unhandled keys to table
		if a.stockTable.Columns() != nil {
			var cmd tea.Cmd
			a.stockTable, cmd = a.stockTable.Update(msg)
			return a, cmd
		}

	case overviewMsg:
		a.overview = msg.overview
		a.loading = false
		return a, nil

	case quotesMsg:
		a.quotes = msg.quotes
		a.initOrRefreshTable()
		return a, nil

	case errorMsg:
		a.err = msg.err
		a.loading = false
		return a, nil

	case tickMsg:
		// Auto-refresh every 30 seconds
		return a, tea.Batch(
			loadMarketOverview(a.marketService),
			loadQuotes(a.marketService),
			tick(), // Schedule next tick
		)

	case tea.MouseMsg:
		if a.stockTable.Columns() != nil {
			var cmd tea.Cmd
			a.stockTable, cmd = a.stockTable.Update(msg)
			return a, cmd
		}
		return a, nil
	}

	return a, nil
}

// View renders the dashboard
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Initializing..."
	}

	if a.showModal && a.selectedQuote != nil {
		return a.renderStockModal(a.selectedQuote, a.width, a.height) + a.renderStatusBar()
	}

	// Calculate layout dimensions - exact 50/50 split
	leftWidth := a.width / 2               // 50% for indices  
	rightWidth := a.width - leftWidth - 1  // 50% for stocks (minus separator)

	// Build left panel (indices)
	leftPanel := a.renderIndicesPanel(leftWidth, a.height-4)

	// Build right panel (stocks)
	rightPanel := a.renderStocksPanel(rightWidth, a.height-4)

	// Combine panels side by side with better styling
	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	
	// Add main container style for better appearance
	container := lipgloss.NewStyle().
		Padding(0, 1).
		Render(layout)
		
	return container + a.renderStatusBar()
}

// renderStatusBar renders the bottom status bar
func (a *App) renderStatusBar() string {
	leftStatus := "r: Refresh • q: Quit"

	var rightStatus string
	if a.overview != nil {
		rightStatus = fmt.Sprintf("Updated: %s", a.overview.LastUpdated)
	}

	// Calculate spacing
	totalWidth := a.width
	statusWidth := totalWidth - len(leftStatus) - len(rightStatus)
	spacing := ""
	if statusWidth > 0 {
		spacing = strings.Repeat(" ", statusWidth)
	}

	statusLine := leftStatus + spacing + rightStatus

	return "\n" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(totalWidth).
		Render(statusLine)
}

// Messages for async operations
type overviewMsg struct {
	overview *models.MarketOverview
}

type quotesMsg struct {
	quotes []*models.Quote
}

type errorMsg struct {
	err error
}

type tickMsg time.Time

// Commands for loading data
func loadMarketOverview(service market.Service) tea.Cmd {
	return func() tea.Msg {
		overview, err := service.GetMarketOverview(context.TODO())
		if err != nil {
			return errorMsg{err}
		}
		return overviewMsg{overview}
	}
}

func loadQuotes(service market.Service) tea.Cmd {
	return func() tea.Msg {
		quotes, err := service.GetLiveQuotes(context.TODO())
		if err != nil {
			return errorMsg{err}
		}
		return quotesMsg{quotes}
	}
}

// tick returns a command that sends a tickMsg after 30 seconds
func tick() tea.Cmd {
	return tea.Tick(time.Second*30, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// initOrRefreshTable constructs or updates the interactive stock table when
// new quotes arrive. Keeping this in the root model centralizes selection state.
func (a *App) initOrRefreshTable() {
	if len(a.quotes) == 0 {
		return
	}
	a.stockTable = a.newStockTable(a.quotes)
	a.syncTableSize()
}

// syncTableSize aligns the table height with the right panel's available size
// to avoid overflow and ensure a stable layout during terminal resizes.
func (a *App) syncTableSize() {
	if a.height == 0 || a.width == 0 {
		return
	}
	panelHeight := a.height - 4 // matches renderStocksPanel(..., a.height-4)
	tableHeight := panelHeight - 6
	if tableHeight < 3 {
		tableHeight = 3
	}
	a.stockTable.SetHeight(tableHeight)
}

// findQuoteBySymbol returns the matching quote for the provided symbol.
func (a *App) findQuoteBySymbol(symbol string) *models.Quote {
	for _, q := range a.quotes {
		if q.Symbol == symbol {
			return q
		}
	}
	return nil
}
