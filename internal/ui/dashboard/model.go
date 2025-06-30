package dashboard

import (
	"log/slog"
	"time"

	"ntx/internal/app/services"
	"ntx/internal/ui/common"
	"ntx/internal/ui/portfolio"
	"ntx/internal/ui/watchlist"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the enhanced multi-pane dashboard model
type Model struct {
	// Terminal dimensions
	width  int
	height int
	ready  bool

	// Panes
	panes          map[common.PaneType]common.Pane
	activePaneType common.PaneType

	// Layout management
	layoutManager *common.LayoutManager
	currentLayout string

	// Services integration
	marketService *services.MarketService
	logger        *slog.Logger

	// Update management
	lastUpdate   time.Time
	updateTicker *time.Ticker
	autoUpdate   bool

	// State
	errorMessage string
	isLoading    bool
}

// NewModel creates a new enhanced dashboard model
func NewModel(marketService *services.MarketService, logger *slog.Logger) Model {
	// Create panes
	panes := make(map[common.PaneType]common.Pane)
	panes[common.PaneTypeDashboard] = NewOverviewPane()
	panes[common.PaneTypePortfolio] = portfolio.NewPortfolioPane()
	panes[common.PaneTypeWatchlist] = watchlist.NewWatchlistPane()
	panes[common.PaneTypeMarketStatus] = common.NewMarketStatusPane()

	// Create layout manager with default quadrant layout
	layoutConfig := common.DefaultLayouts["quadrant"]
	layoutManager := common.NewLayoutManager(layoutConfig)

	// Set initial active pane
	activePaneType := common.PaneTypeDashboard
	panes[activePaneType].SetActive(true)

	return Model{
		ready:          false,
		panes:          panes,
		activePaneType: activePaneType,
		layoutManager:  layoutManager,
		currentLayout:  "quadrant",
		marketService:  marketService,
		logger:         logger,
		autoUpdate:     true,
	}
}

// Init initializes the model and starts update routines
func (m Model) Init() tea.Cmd {
	// Initialize all panes
	var cmds []tea.Cmd
	for _, pane := range m.panes {
		cmds = append(cmds, pane.Init())
	}

	// Start auto-update routine
	if m.autoUpdate {
		m.updateTicker = time.NewTicker(5 * time.Second)
		cmds = append(cmds, m.tickCmd())
	}

	// Initial data refresh
	cmds = append(cmds, m.refreshAllData())

	return tea.Batch(cmds...)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update layout manager
		m.layoutManager.SetTerminalSize(msg.Width, msg.Height)

		// Update pane sizes
		dimensions := m.layoutManager.CalculatePaneDimensions()
		for paneType, pane := range m.panes {
			if dim, exists := dimensions[paneType]; exists {
				pane.SetSize(dim.Width, dim.Height)
			}
		}

		return m, nil

	case tea.KeyMsg:
		// Global key handling
		switch msg.String() {
		case "ctrl+c", "q":
			m.cleanup()
			return m, tea.Quit

		case "tab":
			cmds = append(cmds, m.switchToNextPane())

		case "shift+tab":
			cmds = append(cmds, m.switchToPrevPane())

		case "r":
			cmds = append(cmds, m.refreshAllData())

		case "l":
			cmds = append(cmds, m.switchLayout())

		case "1":
			cmds = append(cmds, m.switchToPane(common.PaneTypeDashboard))

		case "2":
			cmds = append(cmds, m.switchToPane(common.PaneTypePortfolio))

		case "3":
			cmds = append(cmds, m.switchToPane(common.PaneTypeWatchlist))

		case "4":
			cmds = append(cmds, m.switchToPane(common.PaneTypeMarketStatus))
		}

		// Pass key to active pane
		if activePane, exists := m.panes[m.activePaneType]; exists {
			updatedPane, cmd := activePane.Update(msg)
			m.panes[m.activePaneType] = updatedPane
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case common.TickMsg:
		// Auto-update tick
		cmds = append(cmds, m.fetchMarketData(), m.tickCmd())

	case common.RefreshMsg:
		// Refresh data
		cmds = append(cmds, m.fetchMarketData())

	case common.LayoutChangeMsg:
		// Change layout
		m.changeLayout(msg.LayoutName)

	case common.PaneFocusMsg:
		// Change pane focus
		cmds = append(cmds, m.switchToPane(msg.PaneType))

	case common.ErrorMsg:
		// Handle errors
		m.errorMessage = msg.Error.Error()
		m.logger.Error("Dashboard error", "component", msg.Component, "error", msg.Error)

	default:
		// Pass message to all panes
		for paneType, pane := range m.panes {
			updatedPane, cmd := pane.Update(msg)
			m.panes[paneType] = updatedPane
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the enhanced dashboard
func (m Model) View() string {
	if !m.ready {
		return m.renderLoading()
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Main content (panes)
	mainContent := m.layoutManager.RenderPanes(m.panes)
	sections = append(sections, mainContent)

	// Footer (help and status)
	sections = append(sections, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
