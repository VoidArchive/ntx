package forms

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ntx/internal/ui/themes"
)

// FieldType defines the type of input field
type FieldType int

const (
	FieldTypeText FieldType = iota
	FieldTypeNumber
	FieldTypeDate
	FieldTypeSelect
)

// FormField represents a single form input field
type FormField struct {
	Label       string
	Value       string
	Placeholder string
	Required    bool
	Valid       bool
	ErrorMsg    string
	FieldType   FieldType
	Options     []string // For select fields
	Focused     bool
	Theme       themes.Theme
}

// NewFormField creates a new form field
func NewFormField(label, placeholder string, fieldType FieldType, theme themes.Theme) *FormField {
	return &FormField{
		Label:       label,
		Placeholder: placeholder,
		FieldType:   fieldType,
		Valid:       true,
		Theme:       theme,
	}
}

// SetRequired marks the field as required
func (f *FormField) SetRequired() *FormField {
	f.Required = true
	return f
}

// SetOptions sets options for select fields
func (f *FormField) SetOptions(options []string) *FormField {
	f.Options = options
	return f
}

// SetValue sets the field value
func (f *FormField) SetValue(value string) {
	f.Value = value
	f.Validate()
}

// Focus sets focus to this field
func (f *FormField) Focus() {
	f.Focused = true
}

// Blur removes focus from this field
func (f *FormField) Blur() {
	f.Focused = false
}

// Validate validates the field value with optional custom validator
func (f *FormField) Validate() {
	f.Valid = true
	f.ErrorMsg = ""

	// Check required
	if f.Required && strings.TrimSpace(f.Value) == "" {
		f.Valid = false
		f.ErrorMsg = "This field is required"
		return
	}

	// Type-specific validation
	switch f.FieldType {
	case FieldTypeNumber:
		if f.Value != "" {
			if !isValidNumber(f.Value) {
				f.Valid = false
				f.ErrorMsg = "Please enter a valid number"
			}
		}
	case FieldTypeDate:
		if f.Value != "" {
			if !isValidDate(f.Value) {
				f.Valid = false
				f.ErrorMsg = "Please enter a valid date (YYYY-MM-DD)"
			}
		}
	case FieldTypeSelect:
		if f.Value != "" && len(f.Options) > 0 {
			valid := false
			for _, option := range f.Options {
				if f.Value == option {
					valid = true
					break
				}
			}
			if !valid {
				f.Valid = false
				f.ErrorMsg = fmt.Sprintf("Must be one of: %s", strings.Join(f.Options, ", "))
			}
		}
	}
}

// Update handles keyboard input
func (f *FormField) Update(msg tea.Msg) {
	if !f.Focused {
		return
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if len(f.Value) > 0 {
				f.Value = f.Value[:len(f.Value)-1]
				f.Validate()
			}
		case "space":
			f.Value += " "
			f.Validate()
		default:
			// Add printable characters
			if len(msg.String()) == 1 {
				f.Value += msg.String()
				f.Validate()
			}
		}
	}
}

// View renders the form field
func (f *FormField) View() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(f.Theme.Foreground()).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(f.Theme.Muted()).
		Width(30).
		Padding(0, 1)

	if f.Focused {
		inputStyle = inputStyle.BorderForeground(f.Theme.Primary())
	}

	if !f.Valid {
		inputStyle = inputStyle.BorderForeground(f.Theme.Error())
	}

	errorStyle := lipgloss.NewStyle().
		Foreground(f.Theme.Error()).
		Italic(true)

	// Display value or placeholder
	displayValue := f.Value
	if displayValue == "" && f.Placeholder != "" {
		displayValue = f.Placeholder
		inputStyle = inputStyle.Foreground(f.Theme.Muted())
	}

	// Add cursor if focused
	if f.Focused {
		displayValue += "│"
	}

	label := labelStyle.Render(f.Label)
	if f.Required {
		label += " *"
	}

	input := inputStyle.Render(displayValue)
	
	parts := []string{label, input}
	
	// Add error message if invalid
	if !f.Valid && f.ErrorMsg != "" {
		parts = append(parts, errorStyle.Render(f.ErrorMsg))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// Helper functions for validation
func isValidNumber(s string) bool {
	if s == "" {
		return true
	}
	// Allow digits, decimal point, and negative sign
	for _, r := range s {
		if r < '0' || r > '9' {
			if r != '.' && r != '-' {
				return false
			}
		}
	}
	return true
}

func isValidDate(s string) bool {
	if s == "" {
		return true
	}
	// Basic date format validation (YYYY-MM-DD)
	if len(s) != 10 {
		return false
	}
	if s[4] != '-' || s[7] != '-' {
		return false
	}
	// Check if year, month, day are numbers
	year := s[0:4]
	month := s[5:7]
	day := s[8:10]
	
	return isValidNumber(year) && isValidNumber(month) && isValidNumber(day)
}