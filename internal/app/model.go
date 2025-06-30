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
	"ntx/internal/ui/themes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Section enum enforces navigation consistency and prevents invalid states
type Section int

const (
	SectionOverview Section = iota // Default focus optimizes for most common portfolio monitoring task
	SectionHoldings                // Primary workflow - where financial decisions are made
	SectionAnalysis                // TODO: Technical indicators for NEPSE market conditions
	SectionHistory                 // TODO: T+3 settlement tracking for NEPSE transactions
	SectionMarket                  // TODO: Sharesansar integration for price discovery
)

// sectionNames maps enable consistent UI labels and config validation
var sectionNames = map[Section]string{
	SectionOverview: "Overview",
	SectionHoldings: "Holdings",
	SectionAnalysis: "Analysis",
	SectionHistory:  "History",
	SectionMarket:   "Market",
}

// Model holds application state with strict separation of concerns
// Ready/quitting flags prevent rendering during state transitions
type Model struct {
	currentSection Section              // Holdings default optimizes for primary use case
	ready          bool                 // Terminal resize handling prevents corrupted layouts
	quitting       bool                 // Graceful shutdown preserves data integrity
	themeManager   *themes.ThemeManager // Live theme switching for extended trading sessions
	config         *config.Config       // Persistent preferences across market sessions
	width          int                  // Terminal width for responsive layout
	height         int                  // Terminal height for responsive layout
	showHelp       bool                 // Help overlay state
	selectedItem   int                  // Currently selected item within sections
}

// NewModelWithConfig prioritizes user workflow preferences over defaults
// Configuration cascade prevents frustrating resets during market hours
func NewModelWithConfig(cfg *config.Config) Model {
	// User-specified default section optimizes for individual trading patterns
	var defaultSection Section
	switch cfg.GetDefaultSection() {
	case "overview":
		defaultSection = SectionOverview
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

	return Model{
		currentSection: defaultSection,
		ready:          false,
		quitting:       false,
		themeManager:   themeManager,
		config:         cfg,
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

	case tea.KeyMsg:
		// Vim-style navigation reduces hand movement during intensive analysis
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
			m.currentSection = SectionOverview
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
			// Immediate persistence prevents theme reset during market volatility
			if m.config != nil {
				m.config.SetTheme(m.themeManager.GetCurrentThemeType())
				// PERF: Async save maintains UI responsiveness during rapid theme changes
				_ = config.Save(m.config)
			}

		// Vim-style navigation within sections
		case "h":
			// Move left - for future multi-column layouts
			// Currently no-op but reserved for holdings table navigation
		case "j":
			// Move down in current section
			m.selectedItem++
			// TODO: Add bounds checking when section content is implemented
		case "k":
			// Move up in current section
			if m.selectedItem > 0 {
				m.selectedItem--
			}
		case "l":
			// Move right - for future multi-column layouts
			// Currently no-op but reserved for holdings table navigation

		// Additional vim-style navigation
		case "g":
			// Go to top of current section
			m.selectedItem = 0
		case "G":
			// Go to bottom of current section
			// TODO: Set to max items when section content is implemented
			m.selectedItem = 10 // Placeholder
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

	// Structured layout ensures consistent information hierarchy for financial data
	return m.renderMainInterface()
}

// renderMainInterface implements responsive layout optimized for portfolio monitoring
// Layout adapts to terminal width for optimal information density
func (m Model) renderMainInterface() string {
	// Minimum size handling prevents corrupted layouts
	if m.width < 60 || m.height < 24 {
		return m.renderMinimumSizeWarning()
	}

	var content string

	// Header provides constant context awareness during section navigation
	content += m.renderSectionHeader()
	content += "\n\n"

	// Responsive layout based on terminal width
	if m.width >= 120 {
		// Wide layout (3-pane): Main content + sidebar + analytics
		content += m.renderThreePaneLayout()
	} else if m.width >= 80 {
		// Medium layout (2-pane): Main content + condensed sidebar
		content += m.renderTwoPaneLayout()
	} else {
		// Narrow layout (1-pane): Single pane with tab-style navigation
		content += m.renderSinglePaneLayout()
	}

	content += "\n\n"

	// Status bar enables rapid navigation without memorizing shortcuts
	content += m.renderStatusBar()

	return content
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

// renderSectionContent provides contextual placeholders during development phases
// Consistent styling prepares layout for real financial data integration
func (m Model) renderSectionContent() string {
	theme := m.themeManager.GetCurrentTheme()

	var content string
	var sectionIcon string
	var title string
	var description string

	switch m.currentSection {
	case SectionOverview:
		sectionIcon = "📊"
		title = "Overview Section"
		// NOTE: Portfolio aggregations will exclude T+3 pending NEPSE transactions
		description = "Portfolio summary and key statistics will be displayed here.\n" +
			"This section will show total portfolio value, daily changes, and performance metrics."

	case SectionHoldings:
		sectionIcon = "💼"
		title = "Holdings Section"
		// NOTE: P/L calculations will handle bonus shares per SEBON regulations
		description = "Current positions and holdings table will be displayed here.\n" +
			"This section will show individual stocks, quantities, current values, and P/L."

	case SectionAnalysis:
		sectionIcon = "📈"
		title = "Analysis Section"
		// TODO: RSI/MACD calculations adapted for NEPSE market characteristics
		description = "Portfolio analysis and metrics will be displayed here.\n" +
			"This section will show technical indicators, risk metrics, and performance analysis."

	case SectionHistory:
		sectionIcon = "📋"
		title = "History Section"
		// NOTE: Transaction timestamps will account for NEPSE trading hours (11:00-15:00 NPT)
		description = "Transaction history will be displayed here.\n" +
			"This section will show buy/sell transactions, dates, and historical performance."

	case SectionMarket:
		sectionIcon = "🌐"
		title = "Market Section"
		// TODO: Sharesansar integration with respectful rate limiting
		description = "Market data and information will be displayed here.\n" +
			"This section will show market indices, sector performance, and market news."

	default:
		sectionIcon = "❓"
		title = "Unknown Section"
		description = "This section is not recognized."
	}

	// Primary color styling establishes visual hierarchy for section identification
	styledTitle := theme.HighlightStyle().Render(sectionIcon + " " + title)

	// Content styling ensures readability across different theme palettes
	styledDescription := theme.ContentStyle().Render(description)

	// Consistent spacing prevents visual clutter during information scanning
	content = styledTitle + "\n\n" + styledDescription

	return content
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
		statusContent = "[1]Overview [2]Holdings [3]Analysis [4]History [5]Market | " +
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
  1-5           Switch to section directly (Overview, Holdings, Analysis, History, Market)
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

📍 RESPONSIVE LAYOUT:
  Wide (>120 cols)    3-pane layout with sidebar
  Medium (80-120)     2-pane layout
  Narrow (<80)        Single pane with tabs
  Minimum (60x24)     Required for full functionality

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

// renderThreePaneLayout creates wide-screen layout with main content, sidebar, and analytics
// Optimized for large terminals (≥120 columns) with maximum information density
func (m Model) renderThreePaneLayout() string {
	// Main content takes 60% width, sidebar 25%, analytics 15%
	mainWidth := (m.width * 60) / 100
	sidebarWidth := (m.width * 25) / 100
	analyticsWidth := m.width - mainWidth - sidebarWidth - 4 // Account for borders

	mainContent := m.renderSectionContentResized(mainWidth)
	sidebarContent := m.renderSidebar(sidebarWidth)
	analyticsContent := m.renderAnalyticsPanel(analyticsWidth)

	// Use lipgloss JoinHorizontal for side-by-side layout
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainContent,
		" ", // Separator
		sidebarContent,
		" ", // Separator
		analyticsContent,
	)
}

// renderTwoPaneLayout creates medium-screen layout with main content and condensed info
// Optimized for standard terminals (80-119 columns) balancing content and navigation
func (m Model) renderTwoPaneLayout() string {
	// Main content takes 70% width, sidebar 30%
	mainWidth := (m.width * 70) / 100
	sidebarWidth := m.width - mainWidth - 2 // Account for separator

	mainContent := m.renderSectionContentResized(mainWidth)
	sidebarContent := m.renderCondensedSidebar(sidebarWidth)

	// Use lipgloss JoinHorizontal for side-by-side layout
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainContent,
		" ", // Separator
		sidebarContent,
	)
}

// renderSinglePaneLayout creates narrow-screen layout with full-width content
// Optimized for small terminals (<80 columns) with clear single-focus display
func (m Model) renderSinglePaneLayout() string {
	// Full width for main content, navigation via sections only
	return m.renderSectionContentResized(m.width - 2)
}

// renderSectionContentResized provides section content adapted to specific width
// Ensures consistent styling across different layout modes
func (m Model) renderSectionContentResized(width int) string {
	theme := m.themeManager.GetCurrentTheme()

	var content string
	var sectionIcon string
	var title string
	var description string

	switch m.currentSection {
	case SectionOverview:
		sectionIcon = "📊"
		title = "Overview Section"
		// NOTE: Portfolio aggregations will exclude T+3 pending NEPSE transactions
		description = "Portfolio summary and key statistics will be displayed here.\n" +
			"This section will show total portfolio value, daily changes, and performance metrics."

	case SectionHoldings:
		sectionIcon = "💼"
		title = "Holdings Section"
		// NOTE: P/L calculations will handle bonus shares per SEBON regulations
		description = "Current positions and holdings table will be displayed here.\n" +
			"This section will show individual stocks, quantities, current values, and P/L."

	case SectionAnalysis:
		sectionIcon = "📈"
		title = "Analysis Section"
		// TODO: RSI/MACD calculations adapted for NEPSE market characteristics
		description = "Portfolio analysis and metrics will be displayed here.\n" +
			"This section will show technical indicators, risk metrics, and performance analysis."

	case SectionHistory:
		sectionIcon = "📋"
		title = "History Section"
		// NOTE: Transaction timestamps will account for NEPSE trading hours (11:00-15:00 NPT)
		description = "Transaction history will be displayed here.\n" +
			"This section will show buy/sell transactions, dates, and historical performance."

	case SectionMarket:
		sectionIcon = "🌐"
		title = "Market Section"
		// TODO: Sharesansar integration with respectful rate limiting
		description = "Market data and information will be displayed here.\n" +
			"This section will show market indices, sector performance, and market news."

	default:
		sectionIcon = "❓"
		title = "Unknown Section"
		description = "This section is not recognized."
	}

	// Primary color styling establishes visual hierarchy for section identification
	styledTitle := theme.HighlightStyle().Render(sectionIcon + " " + title)

	// Content styling with width constraints for responsive layouts
	styledDescription := theme.ContentStyle().
		Width(width - 4). // Account for padding
		Render(description)

	// Navigation hint for selected items (future use)
	navHint := ""
	if m.selectedItem > 0 {
		navHint = fmt.Sprintf("\n\n[Item %d selected - hjkl to navigate]", m.selectedItem)
		navHint = string(theme.Muted())
	}

	// Consistent spacing prevents visual clutter during information scanning
	content = styledTitle + "\n\n" + styledDescription + navHint

	return content
}

// renderSidebar provides supplementary information for wide layouts
// Contains quick stats, recent activity, and navigation context
func (m Model) renderSidebar(width int) string {
	theme := m.themeManager.GetCurrentTheme()

	sidebarTitle := theme.HighlightStyle().Render("📋 Quick Info")

	sidebarContent := `Portfolio Status:
• Total Value: ₹2,45,670 (+1.8%)
• Today's Change: +₹5,620
• Holdings: 5 stocks

Recent Activity:
• NABIL +10 @ ₹1,250
• EBL -20 @ ₹700
• HIDCL +50 @ ₹445

Market Status:
• NEPSE Index: 2,089.5
• Trading: Closed
• Next Session: Tomorrow 11:00`

	styledContent := theme.ContentStyle().
		Width(width - 2).
		Render(sidebarContent)

	return sidebarTitle + "\n\n" + styledContent
}

// renderCondensedSidebar provides essential information for medium layouts
// Condensed version focusing on most critical portfolio data
func (m Model) renderCondensedSidebar(width int) string {
	theme := m.themeManager.GetCurrentTheme()

	sidebarTitle := theme.HighlightStyle().Render("📊 Stats")

	sidebarContent := `Total: ₹2,45,670 (+1.8%)
Today: +₹5,620
Holdings: 5 stocks

NEPSE: 2,089.5
Status: Closed`

	styledContent := theme.ContentStyle().
		Width(width - 2).
		Render(sidebarContent)

	return sidebarTitle + "\n\n" + styledContent
}

// renderAnalyticsPanel provides technical analysis for wide layouts
// Advanced metrics and indicators for professional users
func (m Model) renderAnalyticsPanel(width int) string {
	theme := m.themeManager.GetCurrentTheme()

	analyticsTitle := theme.HighlightStyle().Render("📈 Analytics")

	analyticsContent := `Technical:
• RSI: 45.2
• MACD: Bullish
• MA(20): ₹1,245
• Support: ₹2,050

Risk:
• Beta: 1.2
• VaR: 2.1%
• Sharpe: 1.8

Sectors:
• Banking: 65%
• Hydro: 25%
• Others: 10%`

	styledContent := theme.ContentStyle().
		Width(width - 2).
		Render(analyticsContent)

	return analyticsTitle + "\n\n" + styledContent
}
