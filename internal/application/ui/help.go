package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpSection represents different help sections
type HelpSection int

const (
	GlobalKeybindsSection HelpSection = iota
	PortfolioHelpSection
	TransactionsHelpSection
	AnalysisHelpSection
	SettingsHelpSection
	AboutSection
)

func (h HelpSection) String() string {
	switch h {
	case GlobalKeybindsSection:
		return "Global Keybinds"
	case PortfolioHelpSection:
		return "Portfolio View"
	case TransactionsHelpSection:
		return "Transactions View"
	case AnalysisHelpSection:
		return "Analysis View"
	case SettingsHelpSection:
		return "Settings View"
	case AboutSection:
		return "About NTX"
	default:
		return "Unknown"
	}
}

// HelpModel handles the help view
type HelpModel struct {
	selectedSection HelpSection
	windowSize      tea.WindowSizeMsg
}

// NewHelpModel creates a new help model
func NewHelpModel() *HelpModel {
	return &HelpModel{
		selectedSection: GlobalKeybindsSection,
	}
}

// Init initializes the help model
func (m *HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedSection > 0 {
				m.selectedSection--
			}
		case "down", "j":
			if m.selectedSection < AboutSection {
				m.selectedSection++
			}
		case "1":
			m.selectedSection = GlobalKeybindsSection
		case "2":
			m.selectedSection = PortfolioHelpSection
		case "3":
			m.selectedSection = TransactionsHelpSection
		case "4":
			m.selectedSection = AnalysisHelpSection
		case "5":
			m.selectedSection = SettingsHelpSection
		case "6":
			m.selectedSection = AboutSection
		}

	case tea.WindowSizeMsg:
		m.windowSize = msg
	}

	return m, nil
}

// View renders the help view
func (m *HelpModel) View() string {
	if m.windowSize.Width == 0 {
		return "Loading help..."
	}

	var content strings.Builder

	// Section title
	title := SectionTitleStyle.Render("❓ NTX Help & Reference")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Split layout: navigation on left, content on right
	leftPanelWidth := 25
	rightPanelWidth := m.windowSize.Width - leftPanelWidth - 8

	// Left panel - navigation
	leftPanel := m.renderNavigationPanel(leftPanelWidth)
	
	// Right panel - help content
	rightPanel := m.renderHelpContent(rightPanelWidth)

	// Join panels horizontally
	splitView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		strings.Repeat(" ", 4),
		rightPanel,
	)
	
	content.WriteString(splitView)

	return content.String()
}

// renderNavigationPanel renders the help section navigation
func (m *HelpModel) renderNavigationPanel(width int) string {
	var nav strings.Builder

	nav.WriteString(SectionTitleStyle.Render("Sections"))
	nav.WriteString("\n\n")

	sections := []HelpSection{
		GlobalKeybindsSection,
		PortfolioHelpSection,
		TransactionsHelpSection,
		AnalysisHelpSection,
		SettingsHelpSection,
		AboutSection,
	}

	for i, section := range sections {
		sectionText := section.String()
		
		if section == m.selectedSection {
			nav.WriteString(SelectedStyle.Render("► " + sectionText))
		} else {
			nav.WriteString("  " + sectionText)
		}
		nav.WriteString("\n")

		// Add shortcut hint
		if i < 5 {
			nav.WriteString(MutedStyle.Render(lipgloss.NewStyle().MarginLeft(4).Render("(" + string(rune('1'+i)) + ")")))
			nav.WriteString("\n")
		}
	}

	return PanelStyle.
		Width(width).
		Height(m.windowSize.Height - 8).
		Render(nav.String())
}

// renderHelpContent renders the help content for the selected section
func (m *HelpModel) renderHelpContent(width int) string {
	var content string

	switch m.selectedSection {
	case GlobalKeybindsSection:
		content = m.renderGlobalKeybinds()
	case PortfolioHelpSection:
		content = m.renderPortfolioHelp()
	case TransactionsHelpSection:
		content = m.renderTransactionsHelp()
	case AnalysisHelpSection:
		content = m.renderAnalysisHelp()
	case SettingsHelpSection:
		content = m.renderSettingsHelp()
	case AboutSection:
		content = m.renderAbout()
	}

	return PanelStyle.
		Width(width).
		Height(m.windowSize.Height - 8).
		Render(content)
}

// renderGlobalKeybinds renders global keybinds help
func (m *HelpModel) renderGlobalKeybinds() string {
	keybinds := [][]string{
		{"Navigation", ""},
		{"1-5", "Switch between sections (Portfolio, Transactions, Analysis, Settings, Help)"},
		{"?", "Toggle help overlay (available in any section)"},
		{"r", "Refresh portfolio data and recalculate metrics"},
		{"", ""},
		{"Application", ""},
		{"q", "Quit application"},
		{"Ctrl+C", "Force quit application"},
		{"", ""},
		{"Movement", ""},
		{"↑/k", "Move up in lists and menus"},
		{"↓/j", "Move down in lists and menus"},
		{"Enter", "Select item or confirm action"},
		{"ESC", "Cancel action or go back"},
		{"Space", "Toggle boolean options"},
		{"Tab", "Navigate between panels (where applicable)"},
	}

	return m.formatKeybindTable(keybinds)
}

// renderPortfolioHelp renders portfolio view help
func (m *HelpModel) renderPortfolioHelp() string {
	content := []string{
		SectionTitleStyle.Render("Portfolio View - Holdings Dashboard"),
		"",
		"The Portfolio view shows your current stock holdings with real-time",
		"profit/loss calculations and portfolio summary.",
		"",
		HelpKeyStyle.Render("Key Features:"),
		"• Holdings table with WAC, current price, and P&L",
		"• Color-coded gains (green) and losses (red)",
		"• Portfolio summary with total invested and market value",
		"• Warning indicators for stocks using default prices",
		"",
	}

	keybinds := [][]string{
		{"Navigation", ""},
		{"↑/↓", "Navigate through holdings list"},
		{"Enter", "View detailed information for selected stock"},
		{"ESC", "Return to main view from detail view"},
		{"", ""},
		{"Actions", ""},
		{"s", "Cycle sort modes (Symbol, Gain%, Market Value, Gain/Loss)"},
		{"i", "Quick access to import CSV (redirects to Transactions)"},
		{"e", "Edit current price for selected stock"},
		{"", ""},
		{"Display", ""},
		{"", "Green text indicates profits"},
		{"", "Red text indicates losses"},
		{"", "⚠ symbol indicates default price needs verification"},
		{"", "* after price indicates default Rs.100 value"},
	}

	content = append(content, m.formatKeybindTable(keybinds))

	return strings.Join(content, "\n")
}

// renderTransactionsHelp renders transactions view help
func (m *HelpModel) renderTransactionsHelp() string {
	content := []string{
		SectionTitleStyle.Render("Transactions View - Import & History"),
		"",
		"The Transactions view handles CSV import from MeroShare and displays",
		"your complete transaction history with filtering options.",
		"",
		HelpKeyStyle.Render("Key Features:"),
		"• Import MeroShare CSV files with progress tracking",
		"• View all transactions with type, date, and amount",
		"• Filter by symbol or transaction type",
		"• Edit transactions and correct default prices",
		"",
	}

	keybinds := [][]string{
		{"CSV Import", ""},
		{"i", "Open CSV import dialog"},
		{"", "Enter file path and press Enter to import"},
		{"", "ESC to cancel import"},
		{"", ""},
		{"Navigation", ""},
		{"↑/↓", "Navigate through transaction list"},
		{"Tab", "Switch between action panel and transaction list"},
		{"", ""},
		{"Actions", ""},
		{"a", "Add manual transaction"},
		{"e", "Edit selected transaction"},
		{"d", "Delete selected transaction"},
		{"f", "Toggle filter options"},
		{"", ""},
		{"Import Notes", ""},
		{"", "Supports MeroShare export format"},
		{"", "Default price Rs.100 applied when CSV lacks price"},
		{"", "Progress bar shows import status"},
		{"", "Warnings displayed for data that needs review"},
	}

	content = append(content, m.formatKeybindTable(keybinds))

	return strings.Join(content, "\n")
}

// renderAnalysisHelp renders analysis view help
func (m *HelpModel) renderAnalysisHelp() string {
	content := []string{
		SectionTitleStyle.Render("Analysis View - Performance Metrics"),
		"",
		"The Analysis view provides comprehensive portfolio metrics, tax",
		"calculations, and performance analysis for informed decision making.",
		"",
		HelpKeyStyle.Render("Key Features:"),
		"• Portfolio summary with total P&L and market value",
		"• Top and worst performing stocks",
		"• Nepal tax calculations (7.5% short-term, 5% long-term)",
		"• Warnings and alerts for portfolio issues",
		"",
		HelpKeyStyle.Render("Metrics Explained:"),
		"",
		WarningStyle.Render("Total Invested:") + " Sum of all purchase costs",
		WarningStyle.Render("Market Value:") + " Current value of all holdings",
		WarningStyle.Render("Realized P&L:") + " Profit/loss from completed sales",
		WarningStyle.Render("Unrealized P&L:") + " Paper gains/losses on current holdings",
		"",
		HelpKeyStyle.Render("Tax Information:"),
		"• Short-term: Holdings ≤ 365 days (7.5% tax)",
		"• Long-term: Holdings > 365 days (5% tax)",
		"• Tax calculated on realized gains only",
		"• Estimates based on current Nepal regulations",
		"",
		HelpKeyStyle.Render("Alerts Monitor:"),
		"• Default price warnings",
		"• Diversification recommendations",
		"• Recent trading activity",
		"• Stocks with only buy transactions",
	}

	return strings.Join(content, "\n")
}

// renderSettingsHelp renders settings view help
func (m *HelpModel) renderSettingsHelp() string {
	content := []string{
		SectionTitleStyle.Render("Settings View - Configuration"),
		"",
		"Configure NTX behavior, display preferences, and import settings",
		"to match your workflow and requirements.",
		"",
	}

	keybinds := [][]string{
		{"Navigation", ""},
		{"↑/↓", "Navigate through settings"},
		{"Enter/Space", "Edit selected setting"},
		{"", ""},
		{"Editing", ""},
		{"Type", "Enter new value for numeric fields"},
		{"Space/Enter", "Toggle boolean options"},
		{"Enter", "Save current edit"},
		{"ESC", "Cancel edit without saving"},
		{"", ""},
		{"Actions", ""},
		{"s", "Save all settings"},
		{"r", "Reset all settings to defaults"},
		{"", ""},
		{"Settings Categories", ""},
		{"Default Price", "Rs.100.00 applied to CSV imports without price"},
		{"Tax Rates", "Nepal capital gains tax rates"},
		{"Display", "Colors, formatting, and visual preferences"},
		{"Import", "CSV processing batch size and options"},
	}

	content = append(content, m.formatKeybindTable(keybinds))

	return strings.Join(content, "\n")
}

// renderAbout renders about information
func (m *HelpModel) renderAbout() string {
	return strings.Join([]string{
		SectionTitleStyle.Render("About NTX"),
		"",
		HelpKeyStyle.Render("NEPSE Portfolio Management TUI"),
		"",
		"NTX is a terminal-based portfolio management application specifically",
		"designed for Nepal Stock Exchange (NEPSE) investors. It provides",
		"comprehensive tools for tracking investments, calculating taxes,",
		"and analyzing portfolio performance.",
		"",
		HelpKeyStyle.Render("Key Features:"),
		"• MeroShare CSV import with progress tracking",
		"• FIFO-based cost basis calculations",
		"• Nepal-specific tax calculations (short/long term)",
		"• Real-time portfolio metrics and analysis",
		"• Professional terminal user interface",
		"",
		HelpKeyStyle.Render("Domain Expertise:"),
		"• Weighted Average Cost (WAC) calculations",
		"• Corporate actions (bonus, rights, splits, mergers)",
		"• Realized vs unrealized gains tracking",
		"• Tax-efficient portfolio analysis",
		"",
		HelpKeyStyle.Render("Technical Foundation:"),
		"• Built with Go and Bubble Tea TUI framework",
		"• Precise money calculations (no floating point errors)",
		"• Clean architecture with comprehensive testing",
		"• Memory-efficient CSV processing",
		"",
		HelpKeyStyle.Render("Version Information:"),
		"• Version: 1.0.0 (TUI Implementation)",
		"• Go Version: 1.24+",
		"• Platform: Cross-platform terminal application",
		"",
		MutedStyle.Render("Designed for NEPSE investors by developers who understand"),
		MutedStyle.Render("the unique requirements of Nepal's stock market."),
	}, "\n")
}

// formatKeybindTable formats keybind information as a table
func (m *HelpModel) formatKeybindTable(keybinds [][]string) string {
	var lines []string

	for _, row := range keybinds {
		if len(row) != 2 {
			continue
		}

		key := row[0]
		description := row[1]

		if key == "" && description == "" {
			lines = append(lines, "")
			continue
		}

		if description == "" {
			// This is a section header
			lines = append(lines, HelpKeyStyle.Render(key+":"))
			continue
		}

		// Format as keybind
		if key == "" {
			// Description only (note or continuation)
			lines = append(lines, "  " + MutedStyle.Render(description))
		} else {
			// Key + description
			keyPart := KeybindStyle.Render(key)
			padding := strings.Repeat(" ", 12-lipgloss.Width(key))
			line := "  " + keyPart + padding + description
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}