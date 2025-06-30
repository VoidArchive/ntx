/**
 * NTX Portfolio Management TUI - Application Model
 *
 * This file defines the main application model that implements the Bubbletea
 * Model-View-Update pattern. It manages the application state, handles user
 * input, and coordinates between different UI sections.
 *
 * The model follows a section-based navigation approach inspired by btop,
 * with 5 main sections accessible via 1-5 keys and vim-style navigation.
 */

package app

import (
	"ntx/internal/ui/themes"

	tea "github.com/charmbracelet/bubbletea"
)

// Section represents the different main sections of the application
type Section int

const (
	SectionOverview Section = iota // [1] Portfolio summary and key statistics
	SectionHoldings                // [2] Current positions (default focus)
	SectionAnalysis                // [3] Placeholder for future metrics
	SectionHistory                 // [4] Placeholder for transaction history
	SectionMarket                  // [5] Placeholder for market data
)

// sectionNames provides human-readable names for each section
var sectionNames = map[Section]string{
	SectionOverview: "Overview",
	SectionHoldings: "Holdings",
	SectionAnalysis: "Analysis",
	SectionHistory:  "History",
	SectionMarket:   "Market",
}

// Model represents the main application state following Bubbletea's Model interface
// It tracks the current section, navigation state, theme management, and global application data
type Model struct {
	currentSection Section              // Currently active section
	ready          bool                 // Whether the application has finished initializing
	quitting       bool                 // Whether the application is in the process of quitting
	themeManager   *themes.ThemeManager // Theme management and switching functionality
}

// NewModel creates and initializes a new application model with default values
// Returns a Model configured with Holdings as the default section and Tokyo Night theme per requirements
func NewModel() Model {
	return Model{
		currentSection: SectionHoldings, // Default focus as specified in FR2.2
		ready:          false,
		quitting:       false,
		themeManager:   themes.NewThemeManager(), // Initialize with Tokyo Night theme (FR3.2)
	}
}

// Init satisfies the Bubbletea Model interface
// Returns any commands that should be run when the application starts
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model state accordingly
// This is the core of the Model-View-Update pattern, processing user input and state changes
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle terminal resize events gracefully without restart (FR6.3)
		m.ready = true

	case tea.KeyMsg:
		// Handle keyboard navigation (FR5: Keyboard Navigation)
		switch msg.String() {
		case "ctrl+c", "q":
			// Quit application cleanly (FR5.4)
			m.quitting = true
			return m, tea.Quit

		case "1":
			// Section switching: 1-5 keys for direct section access (FR5.1)
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
			// Tab for section cycling forward (FR5.3)
			m.currentSection = (m.currentSection + 1) % 5
		case "shift+tab":
			// Shift+Tab for section cycling backward (FR5.3)
			if m.currentSection == 0 {
				m.currentSection = 4
			} else {
				m.currentSection = m.currentSection - 1
			}

		case "t":
			// Toggle theme with 't' key (FR3.4)
			m.themeManager.SwitchTheme()
		}
	}

	return m, nil
}

// View renders the current application state as a string for display
// This implements the View part of the Model-View-Update pattern
func (m Model) View() string {
	if m.quitting {
		return "Thanks for using NTX Portfolio Management TUI!\n"
	}

	if !m.ready {
		return "Initializing NTX Portfolio Management TUI...\n"
	}

	// Build the main interface with current section content and status bar
	return m.renderMainInterface()
}

// renderMainInterface constructs the main TUI interface with section content and status bar
// This organizes the layout as specified in FR6.2: Header, Main content, Status bar
func (m Model) renderMainInterface() string {
	var content string

	// Add section header
	content += m.renderSectionHeader()
	content += "\n\n"

	// Add main section content (FR2.5: Display placeholder content)
	content += m.renderSectionContent()
	content += "\n\n"

	// Add status bar with navigation help (FR2.4)
	content += m.renderStatusBar()

	return content
}

// renderSectionHeader creates the styled header showing the current section name
// Uses theme styling for consistent appearance across different themes (FR3.3)
func (m Model) renderSectionHeader() string {
	theme := m.themeManager.GetCurrentTheme()
	sectionName := sectionNames[m.currentSection]

	// Create styled header with theme colors
	header := theme.HeaderStyle().Render("NTX Portfolio Management - " + sectionName + " Section")

	// Add decorative border using theme colors
	borderStyle := theme.BorderStyle().
		BorderTop(true).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		Width(78)

	return borderStyle.Render(header)
}

// renderSectionContent displays styled placeholder content for the current section
// Each section shows its name and purpose with theme-appropriate styling (FR2.2, FR3.3)
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
		description = "Portfolio summary and key statistics will be displayed here.\n" +
			"This section will show total portfolio value, daily changes, and performance metrics."

	case SectionHoldings:
		sectionIcon = "💼"
		title = "Holdings Section"
		description = "Current positions and holdings table will be displayed here.\n" +
			"This section will show individual stocks, quantities, current values, and P/L."

	case SectionAnalysis:
		sectionIcon = "📈"
		title = "Analysis Section"
		description = "Portfolio analysis and metrics will be displayed here.\n" +
			"This section will show technical indicators, risk metrics, and performance analysis."

	case SectionHistory:
		sectionIcon = "📋"
		title = "History Section"
		description = "Transaction history will be displayed here.\n" +
			"This section will show buy/sell transactions, dates, and historical performance."

	case SectionMarket:
		sectionIcon = "🌐"
		title = "Market Section"
		description = "Market data and information will be displayed here.\n" +
			"This section will show market indices, sector performance, and market news."

	default:
		sectionIcon = "❓"
		title = "Unknown Section"
		description = "This section is not recognized."
	}

	// Style the section title with primary color
	styledTitle := theme.HighlightStyle().Render(sectionIcon + " " + title)

	// Style the description with content styling
	styledDescription := theme.ContentStyle().Render(description)

	// Combine with proper spacing
	content = styledTitle + "\n\n" + styledDescription

	return content
}

// renderStatusBar creates the styled bottom status bar with navigation help and current section indicator
// Shows section navigation, current section, theme info, and key shortcuts with theme styling (FR2.4, FR3.3)
func (m Model) renderStatusBar() string {
	theme := m.themeManager.GetCurrentTheme()
	currentSectionName := sectionNames[m.currentSection]
	currentThemeName := m.themeManager.GetCurrentThemeName()

	// Build status bar content with navigation and current state
	statusContent := "[1]Overview [2]Holdings [3]Analysis [4]History [5]Market | " +
		"Current: " + currentSectionName + " | " +
		"Theme: " + currentThemeName + " | " +
		"Tab/Shift+Tab: Cycle | t: Theme | q: Quit"

	// Apply status bar styling with full width
	styledStatusBar := theme.StatusBarStyle().
		Width(78).
		Render(statusContent)

	// Add separator line using theme border color
	separator := theme.BorderStyle().
		BorderTop(true).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		Width(78).
		Render("")

	return separator + "\n" + styledStatusBar
}
