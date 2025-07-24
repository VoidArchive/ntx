package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SettingsField represents different settings fields
type SettingsField int

const (
	DefaultPriceField SettingsField = iota
	ShortTermTaxField
	LongTermTaxField
	ShowColorsField
	UseCommasField
	ShowPercentagesField
	BatchSizeField
	AutoSortField
	ShowWarningsField
)

// AppSettings represents application settings
type AppSettings struct {
	DefaultPrice     float64
	ShortTermTaxRate float64
	LongTermTaxRate  float64
	ShowColors       bool
	UseCommas        bool
	ShowPercentages  bool
	BatchSize        int
	AutoSort         bool
	ShowWarnings     bool
}

// DefaultSettings returns default application settings
func DefaultSettings() AppSettings {
	return AppSettings{
		DefaultPrice:     100.00,
		ShortTermTaxRate: 7.5,
		LongTermTaxRate:  5.0,
		ShowColors:       true,
		UseCommas:        true,
		ShowPercentages:  true,
		BatchSize:        1000,
		AutoSort:         true,
		ShowWarnings:     true,
	}
}

// SettingsModel handles the settings view
type SettingsModel struct {
	settings      AppSettings
	selectedField SettingsField
	editing       bool
	editValue     string
	windowSize    tea.WindowSizeMsg
	modified      bool
}

// NewSettingsModel creates a new settings model
func NewSettingsModel() *SettingsModel {
	return &SettingsModel{
		settings:      DefaultSettings(),
		selectedField: DefaultPriceField,
	}
}

// Init initializes the settings model
func (m *SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			return m.handleEditingKeys(msg)
		}
		return m.handleNavigationKeys(msg)

	case tea.WindowSizeMsg:
		m.windowSize = msg
	}

	return m, nil
}

// handleNavigationKeys handles navigation when not editing
func (m *SettingsModel) handleNavigationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedField > 0 {
			m.selectedField--
		}
	case "down", "j":
		if m.selectedField < ShowWarningsField {
			m.selectedField++
		}
	case "enter", " ":
		return m.startEditing()
	case "r":
		// Reset to defaults
		m.settings = DefaultSettings()
		m.modified = true
		return m, func() tea.Msg {
			return SuccessMsg{Message: "Settings reset to defaults"}
		}
	case "s":
		// Save settings
		if m.modified {
			// TODO: Implement settings persistence
			return m, func() tea.Msg {
				return SuccessMsg{Message: "Settings saved successfully"}
			}
		}
		return m, func() tea.Msg {
			return SuccessMsg{Message: "No changes to save"}
		}
	}
	return m, nil
}

// handleEditingKeys handles keys when editing a field
func (m *SettingsModel) handleEditingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.finishEditing()
	case "esc":
		m.editing = false
		m.editValue = ""
		return m, nil
	case "backspace":
		if len(m.editValue) > 0 {
			m.editValue = m.editValue[:len(m.editValue)-1]
		}
	default:
		// Add character for numeric/text input
		if len(msg.String()) == 1 {
			char := msg.String()
			// Validate input based on field type
			if m.isNumericField() {
				if (char >= "0" && char <= "9") || char == "." {
					m.editValue += char
				}
			} else {
				m.editValue += char
			}
		}
	}
	return m, nil
}

// startEditing begins editing a field
func (m *SettingsModel) startEditing() (tea.Model, tea.Cmd) {
	m.editing = true
	
	switch m.selectedField {
	case DefaultPriceField:
		m.editValue = fmt.Sprintf("%.2f", m.settings.DefaultPrice)
	case ShortTermTaxField:
		m.editValue = fmt.Sprintf("%.1f", m.settings.ShortTermTaxRate)
	case LongTermTaxField:
		m.editValue = fmt.Sprintf("%.1f", m.settings.LongTermTaxRate)
	case BatchSizeField:
		m.editValue = fmt.Sprintf("%d", m.settings.BatchSize)
	case ShowColorsField:
		m.settings.ShowColors = !m.settings.ShowColors
		m.modified = true
		m.editing = false
		return m, nil
	case UseCommasField:
		m.settings.UseCommas = !m.settings.UseCommas
		m.modified = true
		m.editing = false
		return m, nil
	case ShowPercentagesField:
		m.settings.ShowPercentages = !m.settings.ShowPercentages
		m.modified = true
		m.editing = false
		return m, nil
	case AutoSortField:
		m.settings.AutoSort = !m.settings.AutoSort
		m.modified = true
		m.editing = false
		return m, nil
	case ShowWarningsField:
		m.settings.ShowWarnings = !m.settings.ShowWarnings
		m.modified = true
		m.editing = false
		return m, nil
	}
	
	return m, nil
}

// finishEditing completes editing and saves the value
func (m *SettingsModel) finishEditing() (tea.Model, tea.Cmd) {
	m.editing = false
	
	switch m.selectedField {
	case DefaultPriceField:
		if val, err := strconv.ParseFloat(m.editValue, 64); err == nil && val > 0 {
			m.settings.DefaultPrice = val
			m.modified = true
		}
	case ShortTermTaxField:
		if val, err := strconv.ParseFloat(m.editValue, 64); err == nil && val >= 0 && val <= 100 {
			m.settings.ShortTermTaxRate = val
			m.modified = true
		}
	case LongTermTaxField:
		if val, err := strconv.ParseFloat(m.editValue, 64); err == nil && val >= 0 && val <= 100 {
			m.settings.LongTermTaxRate = val
			m.modified = true
		}
	case BatchSizeField:
		if val, err := strconv.Atoi(m.editValue); err == nil && val > 0 {
			m.settings.BatchSize = val
			m.modified = true
		}
	}
	
	m.editValue = ""
	return m, nil
}

// View renders the settings view
func (m *SettingsModel) View() string {
	if m.windowSize.Width == 0 {
		return "Loading settings..."
	}

	var content strings.Builder

	// Section title
	title := SectionTitleStyle.Render("⚙️ NTX Settings")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Settings form
	form := m.renderSettingsForm()
	content.WriteString(form)
	content.WriteString("\n")

	// Action buttons
	buttons := m.renderActionButtons()
	content.WriteString(buttons)
	content.WriteString("\n")

	// Help text
	helpText := m.renderHelpText()
	content.WriteString(helpText)

	return content.String()
}

// renderSettingsForm renders the settings form
func (m *SettingsModel) renderSettingsForm() string {
	var formItems []string

	// Default Stock Price
	formItems = append(formItems, m.renderField(
		DefaultPriceField,
		"Default Stock Price:",
		fmt.Sprintf("Rs. %.2f", m.settings.DefaultPrice),
		"(for CSV imports without price data)",
	))

	// Tax Rates section
	formItems = append(formItems, "")
	formItems = append(formItems, SectionTitleStyle.Render("Tax Rates:"))
	
	formItems = append(formItems, m.renderField(
		ShortTermTaxField,
		"  Short-term (≤1 year):",
		fmt.Sprintf("%.1f%%", m.settings.ShortTermTaxRate),
		"",
	))
	
	formItems = append(formItems, m.renderField(
		LongTermTaxField,
		"  Long-term (>1 year):",
		fmt.Sprintf("%.1f%%", m.settings.LongTermTaxRate),
		"",
	))

	// Display Preferences section
	formItems = append(formItems, "")
	formItems = append(formItems, SectionTitleStyle.Render("Display Preferences:"))
	
	formItems = append(formItems, m.renderBooleanField(
		ShowColorsField,
		"  Show colors for gains/losses",
		m.settings.ShowColors,
	))
	
	formItems = append(formItems, m.renderBooleanField(
		UseCommasField,
		"  Use comma separators in numbers",
		m.settings.UseCommas,
	))
	
	formItems = append(formItems, m.renderBooleanField(
		ShowPercentagesField,
		"  Show percentage changes",
		m.settings.ShowPercentages,
	))

	// Import Settings section
	formItems = append(formItems, "")
	formItems = append(formItems, SectionTitleStyle.Render("Import Settings:"))
	
	formItems = append(formItems, m.renderField(
		BatchSizeField,
		"  Batch size:",
		fmt.Sprintf("%d transactions", m.settings.BatchSize),
		"",
	))
	
	formItems = append(formItems, m.renderBooleanField(
		AutoSortField,
		"  Auto-sort by date",
		m.settings.AutoSort,
	))
	
	formItems = append(formItems, m.renderBooleanField(
		ShowWarningsField,
		"  Show import warnings",
		m.settings.ShowWarnings,
	))

	formText := strings.Join(formItems, "\n")
	
	return PanelStyle.
		Width(m.windowSize.Width - 4).
		Height(m.windowSize.Height - 12).
		Render(formText)
}

// renderField renders a single settings field
func (m *SettingsModel) renderField(field SettingsField, label, value, description string) string {
	isSelected := field == m.selectedField
	
	var line strings.Builder
	
	if isSelected {
		if m.editing {
			// Show editing state
			line.WriteString(FocusedStyle.Render(label))
			line.WriteString(" ")
			editingValue := m.editValue
			if editingValue == "" {
				editingValue = "..."
			}
			line.WriteString(FocusedInputStyle.Width(20).Render(editingValue))
		} else {
			// Show selected state
			line.WriteString(SelectedStyle.Render(label))
			line.WriteString(" ")
			line.WriteString(SelectedStyle.Render(fmt.Sprintf("[%s]", value)))
		}
	} else {
		// Show normal state
		line.WriteString(label)
		line.WriteString(" ")
		line.WriteString(InputStyle.Width(20).Render(value))
	}
	
	if description != "" {
		line.WriteString(" ")
		line.WriteString(MutedStyle.Render(description))
	}
	
	return line.String()
}

// renderBooleanField renders a boolean settings field
func (m *SettingsModel) renderBooleanField(field SettingsField, label string, value bool) string {
	isSelected := field == m.selectedField
	
	checkbox := "[ ]"
	if value {
		checkbox = "[x]"
	}
	
	var line strings.Builder
	
	if isSelected {
		line.WriteString(SelectedStyle.Render(checkbox))
		line.WriteString(" ")
		line.WriteString(SelectedStyle.Render(label))
	} else {
		line.WriteString(checkbox)
		line.WriteString(" ")
		line.WriteString(label)
	}
	
	return line.String()
}

// renderActionButtons renders save and reset buttons
func (m *SettingsModel) renderActionButtons() string {
	var buttons []string
	
	if m.modified {
		buttons = append(buttons, ActiveButtonStyle.Render("Save Settings (s)"))
	} else {
		buttons = append(buttons, InactiveButtonStyle.Render("Save Settings (s)"))
	}
	
	buttons = append(buttons, DangerButtonStyle.Render("Reset to Defaults (r)"))
	
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Left, buttons...)
	
	return PanelStyle.
		Width(m.windowSize.Width - 4).
		Align(lipgloss.Center).
		Render(buttonRow)
}

// renderHelpText renders help text for settings
func (m *SettingsModel) renderHelpText() string {
	if m.editing {
		return HelpStyle.Render("Enter to save, ESC to cancel")
	}
	
	helpItems := []string{
		KeybindStyle.Render("↑/↓") + " navigate",
		KeybindStyle.Render("Enter/Space") + " edit",
		KeybindStyle.Render("s") + " save",
		KeybindStyle.Render("r") + " reset",
	}
	
	return HelpStyle.Render(strings.Join(helpItems, " | "))
}

// isNumericField returns true if the current field accepts numeric input
func (m *SettingsModel) isNumericField() bool {
	return m.selectedField == DefaultPriceField ||
		m.selectedField == ShortTermTaxField ||
		m.selectedField == LongTermTaxField ||
		m.selectedField == BatchSizeField
}

// GetSettings returns the current settings (for external use)
func (m *SettingsModel) GetSettings() AppSettings {
	return m.settings
}