// Package tui provides terminal user interface
package tui

import (
	"context"
	"fmt"
	"strings"

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
	)
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "r", "F5":
			a.loading = true
			a.err = nil
			return a, tea.Batch(
				loadMarketOverview(a.marketService),
				loadQuotes(a.marketService),
			)
		}

	case overviewMsg:
		a.overview = msg.overview
		a.loading = false
		return a, nil

	case quotesMsg:
		a.quotes = msg.quotes
		return a, nil

	case errorMsg:
		a.err = msg.err
		a.loading = false
		return a, nil
	}

	return a, nil
}

// View renders the dashboard
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Initializing..."
	}

	// Calculate layout dimensions
	leftWidth := a.width * 50 / 100       // 50% for indices
	rightWidth := a.width - leftWidth - 3 // Rest for stocks table (minus borders)

	// Build left panel (indices)
	leftPanel := a.renderIndicesPanel(leftWidth, a.height-4)

	// Build right panel (stocks)
	rightPanel := a.renderStocksPanel(rightWidth, a.height-4)

	// Combine panels side by side
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		rightPanel,
	) + a.renderStatusBar()
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
