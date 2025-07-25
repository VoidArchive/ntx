package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/VoidArchive/ntx/internal/domain"
	"github.com/VoidArchive/ntx/internal/infrastructure/importer"
)

// Section represents different sections of the TUI
type Section int

const (
	PortfolioSection Section = iota
	TransactionsSection
	AnalysisSection
	SettingsSection
	HelpSectionMain
)

// String returns the string representation of a section
func (s Section) String() string {
	switch s {
	case PortfolioSection:
		return "Portfolio"
	case TransactionsSection:
		return "Transactions"
	case AnalysisSection:
		return "Analysis"
	case SettingsSection:
		return "Settings"
	case HelpSectionMain:
		return "Help"
	default:
		return "Unknown"
	}
}

// Global error message
type ErrorMsg struct {
	Error error
}

// Global success message
type SuccessMsg struct {
	Message string
}

// Message for portfolio data refresh
type RefreshPortfolioMsg struct{}

// RootModel is the main model that manages navigation between sections
type RootModel struct {
	currentSection Section
	sections       map[Section]tea.Model
	windowSize     tea.WindowSizeMsg
	portfolio      *domain.Portfolio
	importer       *importer.CSVImporter
	errorMsg       string
	successMsg     string
	showHelp       bool
}

// NewRootModel creates a new root model
func NewRootModel() *RootModel {
	portfolio := domain.NewPortfolio()
	// Create CSV importer with default price of Rs. 100.00
	defaultPrice := domain.NewMoney(10000) // Rs. 100.00 in paisa
	csvImporter := importer.NewCSVImporter(defaultPrice)

	m := &RootModel{
		currentSection: PortfolioSection,
		sections:       make(map[Section]tea.Model),
		portfolio:      portfolio,
		importer:       csvImporter,
	}

	// Initialize sections
	m.sections[PortfolioSection] = NewPortfolioModel(portfolio)
	m.sections[TransactionsSection] = NewTransactionsModel(portfolio, csvImporter)
	m.sections[AnalysisSection] = NewAnalysisModel(portfolio)
	m.sections[SettingsSection] = NewSettingsModel()
	m.sections[HelpSectionMain] = NewHelpModel()

	return m
}

// Init initializes the root model
func (m *RootModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	
	// Initialize all sections
	for _, section := range m.sections {
		if cmd := section.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	return tea.Batch(cmds...)
}

// Update handles messages and updates the model
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "1":
			m.currentSection = PortfolioSection
			return m, nil
		case "2":
			m.currentSection = TransactionsSection
			return m, nil
		case "3":
			m.currentSection = AnalysisSection
			return m, nil
		case "4":
			m.currentSection = SettingsSection
			return m, nil
		case "5":
			m.currentSection = HelpSectionMain
			return m, nil
		case "r":
			// Refresh portfolio data
			cmd = func() tea.Msg { return RefreshPortfolioMsg{} }
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.windowSize = msg
		// Forward window size to all sections
		for section, model := range m.sections {
			model, cmd = model.Update(msg)
			m.sections[section] = model
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case ErrorMsg:
		m.errorMsg = msg.Error.Error()
		m.successMsg = ""
		// Clear error after 5 seconds
		cmd = tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
		cmds = append(cmds, cmd)

	case SuccessMsg:
		m.successMsg = msg.Message
		m.errorMsg = ""
		// Clear success message after 3 seconds
		cmd = tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return clearSuccessMsg{}
		})
		cmds = append(cmds, cmd)

	case clearErrorMsg:
		m.errorMsg = ""

	case clearSuccessMsg:
		m.successMsg = ""

	case RefreshPortfolioMsg:
		// TODO: Implement portfolio refresh logic
		// This could involve recalculating holdings, updating prices, etc.
		m.successMsg = "Portfolio data refreshed"
		cmd = tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return clearSuccessMsg{}
		})
		cmds = append(cmds, cmd)
	}

	// Update current section
	if currentModel, exists := m.sections[m.currentSection]; exists {
		updatedModel, sectionCmd := currentModel.Update(msg)
		m.sections[m.currentSection] = updatedModel
		if sectionCmd != nil {
			cmds = append(cmds, sectionCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *RootModel) View() string {
	if m.windowSize.Width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Render navigation
	content.WriteString(m.renderNavigation())
	content.WriteString("\n")

	// Render status messages
	if statusMsg := m.renderStatusMessages(); statusMsg != "" {
		content.WriteString(statusMsg)
		content.WriteString("\n")
	}

	// Render current section
	if m.showHelp {
		content.WriteString(m.renderHelpOverlay())
	} else {
		content.WriteString(m.renderCurrentSection())
	}

	// Render footer
	content.WriteString("\n")
	content.WriteString(m.renderFooter())

	return content.String()
}

// renderNavigation renders the navigation bar
func (m *RootModel) renderNavigation() string {
	var tabs []string

	sections := []Section{
		PortfolioSection,
		TransactionsSection,
		AnalysisSection,
		SettingsSection,
		HelpSectionMain,
	}

	for i, section := range sections {
		tabText := fmt.Sprintf("%d. %s", i+1, section.String())
		
		if section == m.currentSection {
			tabs = append(tabs, ActiveTabStyle.Render(tabText))
		} else {
			tabs = append(tabs, InactiveTabStyle.Render(tabText))
		}
	}

	navigation := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
	return NavigationStyle.Width(m.windowSize.Width - 4).Render(navigation)
}

// renderStatusMessages renders error and success messages
func (m *RootModel) renderStatusMessages() string {
	if m.errorMsg != "" {
		errorBox := DialogStyle.
			BorderForeground(ColorDanger).
			Width(m.windowSize.Width - 8).
			Render(ErrorStyle.Render("Error: " + m.errorMsg))
		return errorBox
	}

	if m.successMsg != "" {
		successBox := DialogStyle.
			BorderForeground(ColorSuccess).
			Width(m.windowSize.Width - 8).
			Render(SuccessStyle.Render("Success: " + m.successMsg))
		return successBox
	}

	return ""
}

// renderCurrentSection renders the current section
func (m *RootModel) renderCurrentSection() string {
	if currentModel, exists := m.sections[m.currentSection]; exists {
		return currentModel.View()
	}
	return "Section not found"
}

// renderHelpOverlay renders the help overlay
func (m *RootModel) renderHelpOverlay() string {
	helpContent := []string{
		DialogTitleStyle.Render("NTX - NEPSE Portfolio Manager Help"),
		"",
		HelpKeyStyle.Render("Global Keybinds:"),
		"  " + KeybindStyle.Render("1-5") + "    " + HelpDescStyle.Render("Switch between sections"),
		"  " + KeybindStyle.Render("q") + "      " + HelpDescStyle.Render("Quit application"),
		"  " + KeybindStyle.Render("Ctrl+C") + " " + HelpDescStyle.Render("Force quit"),
		"  " + KeybindStyle.Render("?") + "      " + HelpDescStyle.Render("Toggle this help"),
		"  " + KeybindStyle.Render("r") + "      " + HelpDescStyle.Render("Refresh portfolio data"),
		"",
		HelpKeyStyle.Render("Sections:"),
		"  " + KeybindStyle.Render("1. Portfolio") + "    " + HelpDescStyle.Render("View current holdings and P&L"),
		"  " + KeybindStyle.Render("2. Transactions") + " " + HelpDescStyle.Render("Import CSV, view transaction history"),
		"  " + KeybindStyle.Render("3. Analysis") + "     " + HelpDescStyle.Render("Metrics, tax calculations, performance"),
		"  " + KeybindStyle.Render("4. Settings") + "     " + HelpDescStyle.Render("Configuration and preferences"),
		"  " + KeybindStyle.Render("5. Help") + "        " + HelpDescStyle.Render("Detailed keybind reference"),
		"",
		HelpDescStyle.Render("Press ? again to close this help"),
	}

	content := strings.Join(helpContent, "\n")
	
	width := m.windowSize.Width - 8
	height := m.windowSize.Height - 8
	
	return DialogStyle.
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// renderFooter renders the footer with current status
func (m *RootModel) renderFooter() string {
	leftInfo := fmt.Sprintf("Section: %s", m.currentSection.String())
	rightInfo := "Press ? for help | q to quit"
	
	footerWidth := m.windowSize.Width - 4
	spacing := footerWidth - lipgloss.Width(leftInfo) - lipgloss.Width(rightInfo)
	if spacing < 0 {
		spacing = 0
	}
	
	footer := lipgloss.JoinHorizontal(
		lipgloss.Left,
		MutedStyle.Render(leftInfo),
		strings.Repeat(" ", spacing),
		MutedStyle.Render(rightInfo),
	)
	
	return NavigationStyle.Width(m.windowSize.Width - 4).Render(footer)
}

// Helper message types for clearing status messages
type clearErrorMsg struct{}
type clearSuccessMsg struct{}