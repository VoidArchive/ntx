/*
NTX Portfolio Management TUI - Application Model

Bubbletea Model-View-Update pattern ensures predictable state transitions
crucial for financial data integrity - no partial updates or race conditions.

Btop-inspired navigation reduces cognitive load during market analysis sessions
where rapid section switching is common for position monitoring.
*/

package app

import (
	"fmt"
	"ntx/internal/config"
	"ntx/internal/portfolio/services"
	"ntx/internal/ui/components/analysis"
	"ntx/internal/ui/components/dashboard"
	"ntx/internal/ui/components/forms"
	"ntx/internal/ui/components/history"
	"ntx/internal/ui/components/holdings"
	"ntx/internal/ui/components/market"
	"ntx/internal/ui/components/overview"
	"ntx/internal/ui/themes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Section enum enforces navigation consistency and prevents invalid states
type Section int

const (
	SectionDashboard Section = iota // Dashboard provides portfolio command center with overview
	SectionHoldings                 // Primary workflow - where financial decisions are made
	SectionAnalysis                 // TODO: Technical indicators for NEPSE market conditions
	SectionHistory                  // TODO: T+3 settlement tracking for NEPSE transactions
	SectionMarket                   // TODO: Sharesansar integration for price discovery
)

// sectionNames maps enable consistent UI labels and config validation
var sectionNames = map[Section]string{
	SectionDashboard: "Dashboard",
	SectionHoldings:  "Holdings",
	SectionAnalysis:  "Analysis",
	SectionHistory:   "History",
	SectionMarket:    "Market",
}

// Model holds application state with strict separation of concerns
// Ready/quitting flags prevent rendering during state transitions
type Model struct {
	currentSection      Section                    // Holdings default optimizes for primary use case
	ready               bool                       // Terminal resize handling prevents corrupted layouts
	quitting            bool                       // Graceful shutdown preserves data integrity
	themeManager        *themes.ThemeManager       // Live theme switching for extended trading sessions
	config              *config.Config             // Persistent preferences across market sessions
	width               int                        // Terminal width for responsive layout
	height              int                        // Terminal height for responsive layout
	showHelp            bool                       // Help overlay state
	selectedItem        int                        // Currently selected item within sections
	holdingsDisplay     *holdings.HoldingsDisplay  // Holdings component for portfolio management
	overviewDisplay     *overview.OverviewDisplay  // Overview component for portfolio summary
	dashboardDisplay    *dashboard.DashboardDisplay // Dashboard component for portfolio command center
	analysisDisplay     *analysis.AnalysisDisplay  // Analysis component for technical indicators
	historyDisplay      *history.HistoryDisplay    // History component for transaction history
	marketDisplay       *market.MarketDisplay      // Market component for market data
	transactionModal    *forms.Modal               // Modal for transaction entry
	transactionForm     *forms.TransactionForm     // Transaction form component
}

// NewModelWithConfig prioritizes user workflow preferences over defaults
// Configuration cascade prevents frustrating resets during market hours
func NewModelWithConfig(cfg *config.Config) Model {
	// User-specified default section optimizes for individual trading patterns
	var defaultSection Section
	switch cfg.GetDefaultSection() {
	case "dashboard":
		defaultSection = SectionDashboard
	case "holdings":
		defaultSection = SectionHoldings
	case "analysis":
		defaultSection = SectionAnalysis
	case "history":
		defaultSection = SectionHistory
	case "market":
		defaultSection = SectionMarket
	default:
		defaultSection = SectionHoldings // Holdings focus aligns with primary portfolio monitoring task
	}

	// Theme persistence reduces setup friction during daily trading sessions
	themeManager := themes.NewThemeManager()
	if !themeManager.SetThemeByString(cfg.GetTheme()) {
		// Fallback prevents application crash from invalid theme configurations
		themeManager.SetTheme(themes.ThemeTokyoNight)
	}

	// Initialize holdings display with current theme
	holdingsDisplay := holdings.NewHoldingsDisplay(themeManager.GetCurrentTheme())
	holdingsDisplay.UpdateHoldings(holdings.GenerateSampleHoldings()) // TODO: Load from database

	// Initialize overview display with current theme
	overviewDisplay := overview.NewOverviewDisplay(themeManager.GetCurrentTheme())
	// Calculate portfolio totals from holdings
	portfolioTotal := holdingsDisplay.GetPortfolioTotal()
	totalCost := portfolioTotal.MarketValue - portfolioTotal.TotalPL
	overviewDisplay.UpdatePortfolioSummary(
		portfolioTotal.MarketValue,
		totalCost,
		portfolioTotal.TotalPL,
		portfolioTotal.DayPL,
		len(holdingsDisplay.Holdings),
	)

	// Initialize dashboard display with current theme and portfolio data
	dashboardDisplay := dashboard.NewDashboardDisplay(themeManager.GetCurrentTheme())
	dashboardDisplay.UpdatePortfolioMetrics(
		portfolioTotal.MarketValue,
		totalCost,
		portfolioTotal.TotalPL,
		portfolioTotal.DayPL,
		len(holdingsDisplay.Holdings),
	)

	// Initialize analysis display with current theme
	analysisDisplay := analysis.NewAnalysisDisplay(themeManager.GetCurrentTheme())

	// Initialize history display with current theme
	historyDisplay := history.NewHistoryDisplay(themeManager.GetCurrentTheme())

	// Initialize market display with current theme
	marketDisplay := market.NewMarketDisplay(themeManager.GetCurrentTheme())

	return Model{
		currentSection:   defaultSection,
		ready:            false,
		quitting:         false,
		themeManager:     themeManager,
		config:           cfg,
		holdingsDisplay:  holdingsDisplay,
		overviewDisplay:  overviewDisplay,
		dashboardDisplay: dashboardDisplay,
		analysisDisplay:  analysisDisplay,
		historyDisplay:   historyDisplay,
		marketDisplay:    marketDisplay,
	}
}

// Init establishes initial command pipeline for Bubbletea runtime
// No startup commands prevent blocking during market-critical moments
func (m Model) Init() tea.Cmd {
	return nil
}

// Update processes events with financial data consistency guarantees
// State mutations are atomic to prevent portfolio calculation corruption
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Terminal resize without restart preserves portfolio session state
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update holdings display size for responsive layout
		if m.holdingsDisplay != nil {
			m.holdingsDisplay.SetTerminalSize(msg.Width, msg.Height)
		}
		// Update overview display size for responsive layout
		if m.overviewDisplay != nil {
			m.overviewDisplay.SetTerminalSize(msg.Width, msg.Height)
		}
		// Update dashboard display size for responsive layout
		if m.dashboardDisplay != nil {
			m.dashboardDisplay.SetTerminalSize(msg.Width, msg.Height)
		}
		// Update analysis display size for responsive layout
		if m.analysisDisplay != nil {
			m.analysisDisplay.SetTerminalSize(msg.Width, msg.Height)
		}
		// Update history display size for responsive layout
		if m.historyDisplay != nil {
			m.historyDisplay.SetTerminalSize(msg.Width, msg.Height)
		}
		// Update market display size for responsive layout
		if m.marketDisplay != nil {
			m.marketDisplay.SetTerminalSize(msg.Width, msg.Height)
		}

	case tea.KeyMsg:
		// Handle modal events first (when modal is active)
		if m.transactionModal != nil && m.transactionModal.Active {
			var cmd tea.Cmd
			updatedModal, cmd := m.transactionModal.Update(msg)
			if modal, ok := updatedModal.(*forms.Modal); ok {
				m.transactionModal = modal
			}
			return m, cmd
		}

		// Holdings-specific navigation when in holdings section
		if m.currentSection == SectionHoldings && m.holdingsDisplay != nil {
			switch msg.String() {
			case "up", "k":
				m.holdingsDisplay.NavigateUp()
				return m, nil
			case "down", "j":
				m.holdingsDisplay.NavigateDown()
				return m, nil
			case "g":
				m.holdingsDisplay.NavigateTop()
				return m, nil
			case "G":
				m.holdingsDisplay.NavigateBottom()
				return m, nil
			case "s":
				m.holdingsDisplay.CycleSortColumn()
				return m, nil
			case "S":
				m.holdingsDisplay.ToggleSortDirection()
				return m, nil
			case "a":
				// Open add transaction dialog
				m.showTransactionForm()
				return m, nil
			case "d":
				// TODO: Open holding details view
				return m, nil
			case "h":
				// Move left (currently reserved for future use)
				m.holdingsDisplay.NavigateLeft()
				return m, nil
			case "l":
				// Move right (currently reserved for future use)
				m.holdingsDisplay.NavigateRight()
				return m, nil
			case " ":
				// Toggle multi-selection on current row
				m.holdingsDisplay.ToggleSelection()
				return m, nil
			case "enter":
				// Activate current row (show details)
				holding := m.holdingsDisplay.ActivateCurrentRow()
				if holding != nil {
					// TODO: Open holding details view
					// For now, just clear selections as placeholder
					m.holdingsDisplay.ClearSelection()
				}
				return m, nil
			case "escape":
				// Clear multi-selection
				m.holdingsDisplay.ClearSelection()
				return m, nil
			case "r":
				// TODO: Refresh holdings data
				return m, nil
			}
		}

		// Analysis-specific navigation when in analysis section
		if m.currentSection == SectionAnalysis && m.analysisDisplay != nil {
			switch msg.String() {
			case "up", "k":
				m.analysisDisplay.NavigateUp()
				return m, nil
			case "down", "j":
				m.analysisDisplay.NavigateDown()
				return m, nil
			case "left", "h":
				m.analysisDisplay.NavigateLeft()
				return m, nil
			case "right", "l":
				m.analysisDisplay.NavigateRight()
				return m, nil
			case "g":
				m.analysisDisplay.NavigateTop()
				return m, nil
			case "G":
				m.analysisDisplay.NavigateBottom()
				return m, nil
			}
		}

		// Global application shortcuts
		switch msg.String() {
		case "ctrl+c", "q":
			// Clean shutdown prevents data corruption during portfolio updates
			if m.showHelp {
				m.showHelp = false
			} else {
				m.quitting = true
				return m, tea.Quit
			}

		case "?":
			// Toggle help overlay for keybinding reference
			m.showHelp = !m.showHelp

		case "esc":
			// Clear help overlay or selections
			if m.showHelp {
				m.showHelp = false
			} else {
				m.selectedItem = 0
			}

		case "1":
			// Direct section access eliminates navigation latency during price monitoring
			m.currentSection = SectionDashboard
		case "2":
			m.currentSection = SectionHoldings
		case "3":
			m.currentSection = SectionAnalysis
		case "4":
			m.currentSection = SectionHistory
		case "5":
			m.currentSection = SectionMarket

		case "tab":
			// Forward cycling enables rapid workflow transitions during analysis
			m.currentSection = (m.currentSection + 1) % 5
		case "shift+tab":
			// Reverse cycling supports natural navigation patterns from other tools
			if m.currentSection == 0 {
				m.currentSection = 4
			} else {
				m.currentSection = m.currentSection - 1
			}

		case "t":
			// Live theme switching accommodates changing lighting conditions during trading
			m.themeManager.SwitchTheme()
			// Update holdings display theme immediately
			if m.holdingsDisplay != nil {
				m.holdingsDisplay.SetTheme(m.themeManager.GetCurrentTheme())
			}
			// Update overview display theme immediately
			if m.overviewDisplay != nil {
				m.overviewDisplay.SetTheme(m.themeManager.GetCurrentTheme())
			}
			// Update dashboard display theme immediately
			if m.dashboardDisplay != nil {
				m.dashboardDisplay.SetTheme(m.themeManager.GetCurrentTheme())
			}
			// Update analysis display theme immediately
			if m.analysisDisplay != nil {
				m.analysisDisplay.SetTheme(m.themeManager.GetCurrentTheme())
			}
			// Update history display theme immediately
			if m.historyDisplay != nil {
				m.historyDisplay.SetTheme(m.themeManager.GetCurrentTheme())
			}
			// Update market display theme immediately
			if m.marketDisplay != nil {
				m.marketDisplay.SetTheme(m.themeManager.GetCurrentTheme())
			}
			// Immediate persistence prevents theme reset during market volatility
			if m.config != nil {
				m.config.SetTheme(m.themeManager.GetCurrentThemeType())
				// PERF: Async save maintains UI responsiveness during rapid theme changes
				_ = config.Save(m.config)
			}

		// Legacy vim-style navigation for non-holdings sections
		case "h":
			// Move left - for future multi-column layouts
			// Currently no-op but reserved for holdings table navigation
		case "j":
			if m.currentSection != SectionHoldings {
				// Move down in current section
				m.selectedItem++
				// TODO: Add bounds checking when section content is implemented
			}
		case "k":
			if m.currentSection != SectionHoldings {
				// Move up in current section
				if m.selectedItem > 0 {
					m.selectedItem--
				}
			}
		case "l":
			// Move right - for future multi-column layouts
			// Currently no-op but reserved for holdings table navigation
		}
	}

	return m, nil
}

// showTransactionForm initializes and shows the transaction form modal
func (m *Model) showTransactionForm() {
	if m.transactionForm == nil {
		// Use portfolio ID 1 for now (TODO: get from current portfolio context)
		m.transactionForm = forms.NewTransactionForm(1, m.themeManager.GetCurrentTheme())
		
		// Set up callbacks
		m.transactionForm.OnSubmit = func(req services.ExecuteTransactionRequest) error {
			// TODO: Integrate with portfolio service
			// For now, just hide the form to demonstrate it works
			m.hideTransactionForm()
			return nil
		}
		m.transactionForm.OnCancel = func() {
			m.hideTransactionForm()
		}
	}
	
	if m.transactionModal == nil {
		m.transactionModal = forms.NewModal("Add Transaction", m.transactionForm, m.themeManager.GetCurrentTheme())
	}
	
	m.transactionModal.Show()
}

// hideTransactionForm hides the transaction form modal
func (m *Model) hideTransactionForm() {
	if m.transactionModal != nil {
		m.transactionModal.Hide()
	}
}

// View generates immutable UI representation preventing state corruption
// String-based rendering ensures consistent display across terminal types
func (m Model) View() string {
	if m.quitting {
		return "Thanks for using NTX Portfolio Management TUI!\n"
	}

	if !m.ready {
		// Loading state prevents partial render during terminal initialization
		return "Initializing NTX Portfolio Management TUI...\n"
	}

	// Help overlay takes precedence over main interface
	if m.showHelp {
		return m.renderHelpOverlay()
	}

	// Render main interface
	var mainContent string
	
	// All sections now use dedicated components with btop-style borders
	switch m.currentSection {
	case SectionDashboard:
		mainContent = m.renderDashboardSection()
	case SectionHoldings:
		mainContent = m.renderHoldingsSection()
	case SectionAnalysis:
		mainContent = m.renderAnalysisSection()
	case SectionHistory:
		mainContent = m.renderHistorySection()
	case SectionMarket:
		mainContent = m.renderMarketSection()
	default:
		mainContent = m.renderMinimumSizeWarning()
	}
	
	// Overlay modal if active
	if m.transactionModal != nil && m.transactionModal.Active {
		modalView := m.transactionModal.View()
		// Layer modal over main content
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("240")),
		)
	}
	
	return mainContent
}

// renderDashboardSection renders the comprehensive dashboard component
// Provides portfolio command center with overview, market data, and analytics
func (m Model) renderDashboardSection() string {
	if m.dashboardDisplay == nil {
		// Fallback if component is not initialized
		return "Dashboard component not initialized"
	}

	// Minimum size check for dashboard
	if m.width < 60 || m.height < 10 {
		return m.renderMinimumSizeWarning()
	}

	// Update dashboard with current portfolio data
	if m.holdingsDisplay != nil {
		portfolioTotal := m.holdingsDisplay.GetPortfolioTotal()
		totalCost := portfolioTotal.MarketValue - portfolioTotal.TotalPL
		m.dashboardDisplay.UpdatePortfolioMetrics(
			portfolioTotal.MarketValue,
			totalCost,
			portfolioTotal.TotalPL,
			portfolioTotal.DayPL,
			len(m.holdingsDisplay.Holdings),
		)
	}

	// Render complete dashboard
	return m.dashboardDisplay.Render()
}

// renderHoldingsSection renders the btop-style holdings component with overview
// Combines Portfolio Overview widget with holdings table for complete view
func (m Model) renderHoldingsSection() string {
	if m.holdingsDisplay == nil || m.overviewDisplay == nil {
		// Fallback if components are not initialized
		return "Holdings component not initialized"
	}

	// Minimum size check for holdings table
	if m.width < 60 || m.height < 10 {
		return m.renderMinimumSizeWarning()
	}

	// Update overview display with current portfolio data
	portfolioTotal := m.holdingsDisplay.GetPortfolioTotal()
	totalCost := portfolioTotal.MarketValue - portfolioTotal.TotalPL
	m.overviewDisplay.UpdatePortfolioSummary(
		portfolioTotal.MarketValue,
		totalCost,
		portfolioTotal.TotalPL,
		portfolioTotal.DayPL,
		len(m.holdingsDisplay.Holdings),
	)

	// Coordinate widths: Portfolio Overview should match Holdings table width
	// Holdings table width may be adjusted due to minimum column width constraints
	actualTableWidth := m.holdingsDisplay.GetActualTableWidth()
	m.overviewDisplay.SetWidth(actualTableWidth)

	// Render overview widget at top, then holdings table
	overviewWidget := m.overviewDisplay.Render()
	holdingsTable := m.holdingsDisplay.Render()

	return overviewWidget + "\n" + holdingsTable
}

// renderAnalysisSection renders the comprehensive analysis component
// Provides technical indicators, risk metrics, and sector analysis
func (m Model) renderAnalysisSection() string {
	if m.analysisDisplay == nil {
		// Fallback if component is not initialized
		return "Analysis component not initialized"
	}

	// Minimum size check for analysis
	if m.width < 60 || m.height < 10 {
		return m.renderMinimumSizeWarning()
	}

	// Render complete analysis section
	return m.analysisDisplay.Render()
}

// renderHistorySection renders the comprehensive history component
// Provides transaction history, performance tracking, and P/L analysis
func (m Model) renderHistorySection() string {
	if m.historyDisplay == nil {
		// Fallback if component is not initialized
		return "History component not initialized"
	}

	// Minimum size check for history
	if m.width < 60 || m.height < 10 {
		return m.renderMinimumSizeWarning()
	}

	// Render complete history section
	return m.historyDisplay.Render()
}

// renderMarketSection renders the comprehensive market component
// Provides market data, sector performance, and stock information
func (m Model) renderMarketSection() string {
	if m.marketDisplay == nil {
		// Fallback if component is not initialized
		return "Market component not initialized"
	}

	// Minimum size check for market
	if m.width < 60 || m.height < 10 {
		return m.renderMinimumSizeWarning()
	}

	// Render complete market section
	return m.marketDisplay.Render()
}



// renderHelpOverlay displays comprehensive keybinding reference
// Overlay design maintains context while providing complete navigation help
func (m Model) renderHelpOverlay() string {
	theme := m.themeManager.GetCurrentTheme()

	helpTitle := theme.HeaderStyle().Render("🔑 NTX Portfolio Management - Help")

	helpContent := `
📍 SECTION NAVIGATION:
  1-5           Switch to section directly (Dashboard, Holdings, Analysis, History, Market)
  Tab           Next section
  Shift+Tab     Previous section

📍 VIM-STYLE MOVEMENT:
  h/j/k/l       Move left/down/up/right within sections
  g             Go to top of current section
  G             Go to bottom of current section

📍 THEMES & APPEARANCE:
  t             Cycle through themes (Tokyo Night, Rose Pine, Gruvbox, Default)

📍 APPLICATION CONTROLS:
  ?             Toggle this help overlay
  Esc           Close help or clear selections
  q             Quit application
  Ctrl+C        Force quit

📍 FUTURE FEATURES (Coming Soon):
  r             Refresh data
  /             Search/filter
  Enter         Select/drill down
  Space         Multi-select

📍 TERMINAL REQUIREMENTS:
  Minimum (60x10)     Required for basic functionality
  Recommended (120x40) Optimal experience with full features

Press '?' or 'Esc' to close this help.
`

	styledContent := theme.ContentStyle().Render(helpContent)

	// Full-screen overlay with border
	borderStyle := theme.BorderStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary()).
		Padding(1, 2)

	overlay := borderStyle.Render(helpTitle + "\n" + styledContent)

	return overlay
}

// renderMinimumSizeWarning informs users of terminal size requirements
// Clear messaging prevents confusion during terminal setup
func (m Model) renderMinimumSizeWarning() string {
	theme := m.themeManager.GetCurrentTheme()

	warningTitle := theme.ErrorStyle().Render("⚠️  Terminal Too Small")

	warningContent := fmt.Sprintf(`
NTX Portfolio Management requires a minimum terminal size of 60x24.

Current size: %dx%d

Please resize your terminal and the interface will automatically adjust.

Minimum requirements:
• Width: 60 columns (current: %d)
• Height: 24 rows (current: %d)

Recommended size: 120x40 for optimal experience.

Press 'q' to quit.
`,
		m.width, m.height, m.width, m.height)

	return warningTitle + "\n" + theme.ContentStyle().Render(warningContent)
}

