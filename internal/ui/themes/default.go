/**
 * NTX Portfolio Management TUI - Default Theme
 *
 * This file implements a sophisticated default theme with a custom color palette
 * that provides excellent readability and professional appearance. The theme
 * features a dark background with carefully chosen accent colors that work
 * well across different terminal environments.
 *
 * This theme serves as a high-quality alternative to themed options while
 * maintaining broad compatibility and elegant aesthetics.
 */

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// DefaultTheme implements the Theme interface with a custom sophisticated palette
// Provides a clean, professional aesthetic with carefully chosen colors
type DefaultTheme struct{}

// NewDefaultTheme creates a new default theme instance
// Returns a theme configured with a sophisticated custom palette for professional appearance
func NewDefaultTheme() Theme {
	return &DefaultTheme{}
}

// Name returns the human-readable name of the default theme
func (t *DefaultTheme) Name() string {
	return "Default"
}

// Type returns the theme type identifier for the default theme
func (t *DefaultTheme) Type() ThemeType {
	return ThemeDefault
}

// Color palette implementation using custom sophisticated colors

// Background returns the main background color (#141415 - very dark gray)
func (t *DefaultTheme) Background() lipgloss.Color {
	return lipgloss.Color("#141415")
}

// Foreground returns the primary text color (#cdcdcd - light gray)
func (t *DefaultTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#cdcdcd")
}

// Primary returns the accent/highlight color (#6e94b2 - soft blue)
func (t *DefaultTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#6e94b2")
}

// Success returns the success/positive indicator color (#7fa563 - green)
func (t *DefaultTheme) Success() lipgloss.Color {
	return lipgloss.Color("#7fa563")
}

// Warning returns the warning indicator color (#f3be7c - warm gold)
func (t *DefaultTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#f3be7c")
}

// Error returns the error/negative indicator color (#d8647e - soft red)
func (t *DefaultTheme) Error() lipgloss.Color {
	return lipgloss.Color("#d8647e")
}

// Muted returns the secondary/muted text color (#606079 - muted gray)
func (t *DefaultTheme) Muted() lipgloss.Color {
	return lipgloss.Color("#606079")
}

// Pre-configured styles for common UI elements

// HeaderStyle provides styling for section headers and titles
func (t *DefaultTheme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true).
		Padding(0, 1)
}

// ContentStyle provides styling for main content areas
func (t *DefaultTheme) ContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground()).
		Padding(1, 2)
}

// StatusBarStyle provides styling for the bottom status bar
func (t *DefaultTheme) StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground()).
		Background(t.Muted()).
		Padding(0, 1).
		Bold(false)
}

// BorderStyle provides styling for borders and separators
func (t *DefaultTheme) BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Primary()).
		Padding(1, 2)
}

// HighlightStyle provides styling for selected/focused items
func (t *DefaultTheme) HighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Background()).
		Background(t.Primary()).
		Bold(true).
		Padding(0, 1)
}

// ErrorStyle provides styling for error messages
func (t *DefaultTheme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Error()).
		Bold(true)
}

// SuccessStyle provides styling for success messages
func (t *DefaultTheme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Success()).
		Bold(true)
}
