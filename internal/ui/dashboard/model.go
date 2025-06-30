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

	// Vim navigation state
	lastKey      string
	searchMode   bool
	commandMode  bool
	searchQuery  string
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
		// Handle search and command modes first
		if m.searchMode {
			return m.handleSearchMode(msg)
		}
		if m.commandMode {
			return m.handleCommandMode(msg)
		}

		// Global key handling
		switch msg.String() {
		case "ctrl+c", "q":
			m.cleanup()
			return m, tea.Quit

		// Vim navigation
		case "h", "left":
			cmds = append(cmds, m.handleVimLeft())
		case "j", "down":
			cmds = append(cmds, m.handleVimDown())
		case "k", "up":
			cmds = append(cmds, m.handleVimUp())
		case "l", "right":
			cmds = append(cmds, m.handleVimRight())

		// Vim motions
		case "g":
			if m.lastKey == "g" {
				cmds = append(cmds, m.handleVimTop())
				m.lastKey = ""
			} else {
				m.lastKey = "g"
			}
		case "G":
			cmds = append(cmds, m.handleVimBottom())

		// Search mode
		case "/":
			m.enterSearchMode()

		// Command mode
		case ":":
			m.enterCommandMode()

		// Pane resizing with Shift+hjkl
		case "shift+h":
			cmds = append(cmds, m.resizePaneLeft())
		case "shift+j":
			cmds = append(cmds, m.resizePaneDown())
		case "shift+k":
			cmds = append(cmds, m.resizePaneUp())
		case "shift+l":
			cmds = append(cmds, m.resizePaneRight())

		// Layout switching with F-keys
		case "f1":
			cmds = append(cmds, m.switchToLayoutByName("market_focus"))
		case "f2":
			cmds = append(cmds, m.switchToLayoutByName("portfolio_focus"))
		case "f3":
			cmds = append(cmds, m.switchToLayoutByName("analysis_focus"))

		// Existing btop-style navigation (maintained for compatibility)
		case "tab":
			cmds = append(cmds, m.switchToNextPane())

		case "shift+tab":
			cmds = append(cmds, m.switchToPrevPane())

		case "r":
			cmds = append(cmds, m.refreshAllData())

		case "1":
			cmds = append(cmds, m.switchToPane(common.PaneTypeDashboard))

		case "2":
			cmds = append(cmds, m.switchToPane(common.PaneTypePortfolio))

		case "3":
			cmds = append(cmds, m.switchToPane(common.PaneTypeWatchlist))

		case "4":
			cmds = append(cmds, m.switchToPane(common.PaneTypeMarketStatus))

		default:
			// Reset last key for multi-key sequences
			m.lastKey = ""
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

// View renders the enhanced dashboard with responsive design
func (m Model) View() string {
	// Check minimum terminal size (60x24 threshold)
	if m.width < 60 || m.height < 24 {
		return m.renderTerminalTooSmall()
	}

	if !m.ready {
		return m.renderLoading()
	}

	// Adapt layout for narrow terminals
	if m.width < 100 {
		return m.renderNarrowTerminal()
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

// Vim navigation helper methods
func (m Model) handleVimLeft() tea.Cmd {
	// Move to previous pane (left)
	return m.switchToPrevPane()
}

func (m Model) handleVimRight() tea.Cmd {
	// Move to next pane (right)
	return m.switchToNextPane()
}

func (m Model) handleVimUp() tea.Cmd {
	// Pass up navigation to active pane
	if activePane, exists := m.panes[m.activePaneType]; exists {
		updatedPane, cmd := activePane.Update(tea.KeyMsg{Type: tea.KeyUp})
		m.panes[m.activePaneType] = updatedPane
		return cmd
	}
	return nil
}

func (m Model) handleVimDown() tea.Cmd {
	// Pass down navigation to active pane
	if activePane, exists := m.panes[m.activePaneType]; exists {
		updatedPane, cmd := activePane.Update(tea.KeyMsg{Type: tea.KeyDown})
		m.panes[m.activePaneType] = updatedPane
		return cmd
	}
	return nil
}

func (m Model) handleVimTop() tea.Cmd {
	// Go to top of active pane
	if activePane, exists := m.panes[m.activePaneType]; exists {
		updatedPane, cmd := activePane.Update(tea.KeyMsg{Type: tea.KeyHome})
		m.panes[m.activePaneType] = updatedPane
		return cmd
	}
	return nil
}

func (m Model) handleVimBottom() tea.Cmd {
	// Go to bottom of active pane
	if activePane, exists := m.panes[m.activePaneType]; exists {
		updatedPane, cmd := activePane.Update(tea.KeyMsg{Type: tea.KeyEnd})
		m.panes[m.activePaneType] = updatedPane
		return cmd
	}
	return nil
}

// Search and command mode handlers
func (m *Model) enterSearchMode() {
	m.searchMode = true
	m.searchQuery = ""
}

func (m *Model) enterCommandMode() {
	m.commandMode = true
}

func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searchMode = false
		m.searchQuery = ""
	case tea.KeyEnter:
		// TODO: Implement search functionality
		m.searchMode = false
	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.searchQuery += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.commandMode = false
	case tea.KeyEnter:
		// TODO: Implement command execution
		m.commandMode = false
	}
	return m, nil
}

// Pane resizing methods
func (m Model) resizePaneLeft() tea.Cmd {
	// INFO: Pane resizing adjusts layout ratios within the current layout type
	// For now, we provide placeholder functionality that will be enhanced later
	m.logger.Debug("Pane resize left requested")
	return nil
}

func (m Model) resizePaneRight() tea.Cmd {
	// INFO: Pane resizing adjusts layout ratios within the current layout type
	m.logger.Debug("Pane resize right requested")
	return nil
}

func (m Model) resizePaneUp() tea.Cmd {
	// INFO: Pane resizing adjusts layout ratios within the current layout type
	m.logger.Debug("Pane resize up requested")
	return nil
}

func (m Model) resizePaneDown() tea.Cmd {
	// INFO: Pane resizing adjusts layout ratios within the current layout type
	m.logger.Debug("Pane resize down requested")
	return nil
}

// Layout switching
func (m Model) switchToLayoutByName(layoutName string) tea.Cmd {
	if layoutConfig, exists := common.DefaultLayouts[layoutName]; exists {
		m.currentLayout = layoutName
		m.layoutManager = common.NewLayoutManager(layoutConfig)
		m.layoutManager.SetTerminalSize(m.width, m.height)
		
		// Update pane sizes
		dimensions := m.layoutManager.CalculatePaneDimensions()
		for paneType, pane := range m.panes {
			if dim, exists := dimensions[paneType]; exists {
				pane.SetSize(dim.Width, dim.Height)
			}
		}
	}
	return nil
}
