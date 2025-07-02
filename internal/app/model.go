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
	"strings"
	"ntx/internal/config"
	"ntx/internal/ui/components/dashboard"
	"ntx/internal/ui/components/holdings"
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
	currentSection   Section                    // Holdings default optimizes for primary use case
	ready            bool                       // Terminal resize handling prevents corrupted layouts
	quitting         bool                       // Graceful shutdown preserves data integrity
	themeManager     *themes.ThemeManager       // Live theme switching for extended trading sessions
	config           *config.Config             // Persistent preferences across market sessions
	width            int                        // Terminal width for responsive layout
	height           int                        // Terminal height for responsive layout
	showHelp         bool                       // Help overlay state
	selectedItem     int                        // Currently selected item within sections
	holdingsDisplay  *holdings.HoldingsDisplay  // Holdings component for portfolio management
	overviewDisplay  *overview.OverviewDisplay  // Overview component for portfolio summary
	dashboardDisplay *dashboard.DashboardDisplay // Dashboard component for portfolio command center
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

	return Model{
		currentSection:   defaultSection,
		ready:            false,
		quitting:         false,
		themeManager:     themeManager,
		config:           cfg,
		holdingsDisplay:  holdingsDisplay,
		overviewDisplay:  overviewDisplay,
		dashboardDisplay: dashboardDisplay,
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

	case tea.KeyMsg:
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
				// TODO: Open add transaction dialog
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

	// Dashboard and Holdings sections use dedicated components, others use legacy rendering
	if m.currentSection == SectionDashboard {
		return m.renderDashboardSection()
	}
	
	if m.currentSection == SectionHoldings {
		return m.renderHoldingsSection()
	}

	// Structured layout ensures consistent information hierarchy for financial data
	return m.renderMainInterface()
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

// renderMainInterface implements btop-style section rendering for non-dashboard/holdings sections
// All sections now use consistent btop-style borders and layout
func (m Model) renderMainInterface() string {
	// Minimum size handling prevents corrupted layouts
	if m.width < 60 || m.height < 10 {
		return m.renderMinimumSizeWarning()
	}

	// Render section with btop-style borders
	return m.renderStandardSection()
}

// renderStandardSection renders Analysis, History, and Market sections with btop-style borders
// Provides consistent UI across all sections with professional appearance
func (m Model) renderStandardSection() string {
	width := m.width

	var sectionIcon string
	var sectionName string
	var sectionNumber string
	var description string

	switch m.currentSection {
	case SectionAnalysis:
		sectionIcon = "📈"
		sectionName = "Analysis"
		sectionNumber = "[3]"
		description = `Portfolio Analysis & Technical Indicators

• Technical Indicators: RSI, MACD, Moving Averages
• Risk Metrics: Beta, VaR, Sharpe Ratio, Portfolio Correlation
• Performance Analysis: Risk-adjusted returns, drawdown analysis
• Sector Analysis: Industry allocation and performance comparison

This section will provide comprehensive portfolio analytics
for informed investment decisions on NEPSE.`

	case SectionHistory:
		sectionIcon = "📋"
		sectionName = "History"
		sectionNumber = "[4]"
		description = `Transaction History & Performance Tracking

• Complete Transaction Log: All buy/sell orders with timestamps
• Performance Timeline: Portfolio value changes over time
• Realized P/L Tracking: Completed trades and tax implications
• Settlement Tracking: T+3 NEPSE settlement status

This section will show detailed transaction history
and historical performance metrics.`

	case SectionMarket:
		sectionIcon = "🌐"
		sectionName = "Market"
		sectionNumber = "[5]"
		description = `Market Data & Sector Information

• NEPSE Index: Real-time market index and daily changes
• Sector Performance: Banking, Hydro, Manufacturing, Hotels
• Market News: Latest NEPSE announcements and market updates
• Stock Screener: Find stocks by criteria and watchlist management

This section will provide market context for
portfolio decisions and stock discovery.`

	default:
		sectionIcon = "❓"
		sectionName = "Unknown"
		sectionNumber = "[?]"
		description = "This section is not recognized."
	}

	// Create btop-style bordered section
	title := sectionNumber + sectionName
	
	// Top border with integrated title
	topBorder := m.renderSectionTopBorder(title, width)
	
	// Content area with proper padding and styling
	content := m.renderSectionContent(sectionIcon, sectionName, description, width)
	
	// Bottom border
	bottomBorder := m.renderSectionBottomBorder(width)

	return topBorder + "\n" + content + "\n" + bottomBorder
}

// renderSectionTopBorder creates top border with integrated title (btop-style)
func (m Model) renderSectionTopBorder(title string, width int) string {
	theme := m.themeManager.GetCurrentTheme()
	
	if width < len(title)+10 {
		border := "┌" + strings.Repeat("─", width-2) + "┐"
		return lipgloss.NewStyle().Foreground(theme.Primary()).Render(border)
	}
	
	titleSection := "─" + title + "─"
	remainingWidth := width - len([]rune(titleSection)) - 2
	leftPadding := strings.Repeat("─", remainingWidth)
	
	border := "┌" + titleSection + leftPadding + "┐"
	return lipgloss.NewStyle().Foreground(theme.Primary()).Render(border)
}

// renderSectionBottomBorder creates bottom border
func (m Model) renderSectionBottomBorder(width int) string {
	theme := m.themeManager.GetCurrentTheme()
	border := "└" + strings.Repeat("─", width-2) + "┘"
	return lipgloss.NewStyle().Foreground(theme.Primary()).Render(border)
}

// renderSectionContent renders section content with proper styling and padding
func (m Model) renderSectionContent(sectionIcon, sectionName, description string, width int) string {
	theme := m.themeManager.GetCurrentTheme()
	
	// Format the content
	headerText := sectionIcon + " " + sectionName + " Section"
	styledHeader := theme.HighlightStyle().Render(headerText)
	
	// Style description with theme
	styledDescription := theme.ContentStyle().Render(description)
	
	// Combine header and description
	fullContent := styledHeader + "\n\n" + styledDescription
	
	// Split content into lines and apply borders
	lines := strings.Split(fullContent, "\n")
	var borderedLines []string
	
	for _, line := range lines {
		// Calculate visual width for Unicode characters
		visualWidth := lipgloss.Width(line)
		contentWidth := width - 2 // Account for borders
		
		if visualWidth < contentWidth {
			padding := strings.Repeat(" ", contentWidth-visualWidth)
			line = line + padding
		} else if visualWidth > contentWidth {
			// Truncate if too long
			line = line[:contentWidth-3] + "..."
		}
		
		// Add borders
		borderStyle := lipgloss.NewStyle().Foreground(theme.Primary())
		leftBorder := borderStyle.Render("│")
		rightBorder := borderStyle.Render("│")
		
		borderedLines = append(borderedLines, leftBorder+line+rightBorder)
	}
	
	// Add some padding rows if needed
	minHeight := 15 // Minimum content height
	for len(borderedLines) < minHeight {
		emptyLine := strings.Repeat(" ", width-2)
		borderStyle := lipgloss.NewStyle().Foreground(theme.Primary())
		leftBorder := borderStyle.Render("│")
		rightBorder := borderStyle.Render("│")
		borderedLines = append(borderedLines, leftBorder+emptyLine+rightBorder)
	}
	
	return strings.Join(borderedLines, "\n")
}

// renderSectionHeader maintains visual continuity across theme changes
// Consistent header reduces disorientation during rapid section switching
func (m Model) renderSectionHeader() string {
	theme := m.themeManager.GetCurrentTheme()
	sectionName := sectionNames[m.currentSection]

	// Responsive header content adapts to terminal width
	var headerText string
	if m.width >= 100 {
		headerText = "NTX Portfolio Management - " + sectionName + " Section"
	} else if m.width >= 80 {
		headerText = "NTX - " + sectionName + " Section"
	} else {
		headerText = sectionName
	}

	// Theme-aware styling maintains readability across lighting conditions
	header := theme.HeaderStyle().Render(headerText)

	// Responsive width styling adapts to available terminal space
	headerWidth := m.width
	if headerWidth == 0 {
		headerWidth = 78 // Fallback for initialization
	}

	// Visual separation prevents information bleeding between interface zones
	borderStyle := theme.BorderStyle().
		BorderTop(true).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		Width(headerWidth - 2)

	return borderStyle.Render(header)
}

// renderStatusBar provides persistent navigation context without cluttering main content
// Essential shortcuts remain visible during intensive portfolio analysis sessions
func (m Model) renderStatusBar() string {
	theme := m.themeManager.GetCurrentTheme()
	currentSectionName := sectionNames[m.currentSection]
	currentThemeName := m.themeManager.GetCurrentThemeName()

	// Responsive status content adapts to terminal width
	var statusContent string
	if m.width >= 120 {
		// Wide layout - full navigation hints
		statusContent = "[1]Dashboard [2]Holdings [3]Analysis [4]History [5]Market | " +
			"hjkl: Move | ?: Help | t: Theme | Current: " + currentSectionName + " (" + currentThemeName + ") | q: Quit"
	} else if m.width >= 80 {
		// Medium layout - condensed hints
		statusContent = "[1-5]: Sections | hjkl: Move | ?: Help | t: Theme | " + currentSectionName + " | q: Quit"
	} else {
		// Narrow layout - minimal hints
		statusContent = "1-5: Sections | ?: Help | " + currentSectionName + " | q: Quit"
	}

	// Responsive width styling adapts to available terminal space
	statusWidth := m.width
	if statusWidth == 0 {
		statusWidth = 78 // Fallback for initialization
	}

	styledStatusBar := theme.StatusBarStyle().
		Width(statusWidth - 2). // Account for padding
		Render(statusContent)

	// Theme-consistent separator maintains visual coherence across color schemes
	separator := theme.BorderStyle().
		BorderTop(true).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		Width(statusWidth - 2).
		Render("")

	return separator + "\n" + styledStatusBar
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

