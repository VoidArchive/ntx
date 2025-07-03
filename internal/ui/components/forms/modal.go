package forms

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ntx/internal/ui/themes"
)

// Modal represents a modal dialog overlay
type Modal struct {
	Active      bool
	Content     tea.Model
	Title       string
	Width       int
	Height      int
	Theme       themes.Theme
	OnClose     func()
	OnSubmit    func() error
}

// NewModal creates a new modal with the given content
func NewModal(title string, content tea.Model, theme themes.Theme) *Modal {
	return &Modal{
		Active:  false,
		Content: content,
		Title:   title,
		Width:   60,
		Height:  20,
		Theme:   theme,
	}
}

// Show displays the modal
func (m *Modal) Show() {
	m.Active = true
}

// Hide dismisses the modal
func (m *Modal) Hide() {
	m.Active = false
	if m.OnClose != nil {
		m.OnClose()
	}
}

// Init implements tea.Model
func (m *Modal) Init() tea.Cmd {
	if m.Content != nil {
		return m.Content.Init()
	}
	return nil
}

// Update implements tea.Model
func (m *Modal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Hide()
			return m, nil
		case "enter":
			if m.OnSubmit != nil {
				if err := m.OnSubmit(); err == nil {
					m.Hide()
				}
			}
			return m, nil
		}
	}

	// Forward other messages to content
	if m.Content != nil {
		var cmd tea.Cmd
		m.Content, cmd = m.Content.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model
func (m *Modal) View() string {
	if !m.Active {
		return ""
	}

	// Create modal overlay
	content := ""
	if m.Content != nil {
		content = m.Content.View()
	}

	// Create bordered modal window
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.Theme.Primary()).
		Width(m.Width).
		Height(m.Height).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Foreground(m.Theme.Primary()).
		Bold(true)

	title := titleStyle.Render(m.Title)
	modalContent := modalStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		content,
	))

	// Center the modal on screen
	return lipgloss.Place(
		lipgloss.Width(modalContent), lipgloss.Height(modalContent),
		lipgloss.Center, lipgloss.Center,
		modalContent,
	)
}