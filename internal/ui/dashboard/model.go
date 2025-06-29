package dashboard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the main dashboard model
type Model struct {
	width  int
	height int
	ready  bool
}

// NewModel creates a new dashboard model
func NewModel() Model {
	return Model{
		ready: false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the model
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Define styles
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Bold(true)

	contentStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Margin(1, 0)

	// Create the main content
	header := headerStyle.Render("NTX - NEPSE Power Terminal")
	
	welcome := `Welcome to NTX!

This is the foundation of your NEPSE Power Terminal.

Current Status:
✓ Go module initialized
✓ Project structure created  
✓ Basic application lifecycle implemented
✓ Configuration management setup
✓ Structured logging with slog
✓ Bubbletea TUI foundation ready

Next Steps:
• Implement configuration loading (TOML + encrypted credentials)
• Add portfolio management features
• Implement NEPSE data scraping
• Build multi-pane dashboard layout

The foundation is solid and ready for Phase 2 development!`

	content := contentStyle.Width(m.width - 4).Render(welcome)
	help := helpStyle.Render("Press 'q' or Ctrl+C to quit")

	// Combine all parts
	return fmt.Sprintf("%s\n\n%s\n\n%s", header, content, help)
}