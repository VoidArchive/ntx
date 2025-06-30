/**
 * NTX Portfolio Management TUI - Tokyo Night Theme
 *
 * This file implements the Tokyo Night color scheme as specified in FR3.2.
 * Tokyo Night is a beautiful dark theme with purple/blue aesthetics that
 * provides excellent readability and modern visual appeal.
 *
 * Color palette matches the official Tokyo Night specification for consistency
 * with popular development tools and editor themes.
 */

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// TokyoNightTheme implements the Theme interface with Tokyo Night colors
// Provides the dark purple/blue aesthetic specified in the requirements
type TokyoNightTheme struct{}

// NewTokyoNightTheme creates a new Tokyo Night theme instance
// Returns a theme configured with the official Tokyo Night color palette
func NewTokyoNightTheme() Theme {
	return &TokyoNightTheme{}
}

// Name returns the human-readable name of the Tokyo Night theme
func (t *TokyoNightTheme) Name() string {
	return "Tokyo Night"
}

// Type returns the theme type identifier for Tokyo Night
func (t *TokyoNightTheme) Type() ThemeType {
	return ThemeTokyoNight
}

// Color palette implementation (FR3.2)
// These colors are from the official Tokyo Night specification

// Background returns the main background color (#1a1b26)
func (t *TokyoNightTheme) Background() lipgloss.Color {
	return lipgloss.Color("#1a1b26")
}

// Foreground returns the primary text color (#c0caf5)
func (t *TokyoNightTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#c0caf5")
}

// Primary returns the accent/highlight color (#7aa2f7 - bright blue)
func (t *TokyoNightTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#7aa2f7")
}

// Success returns the success/positive indicator color (#9ece6a - green)
func (t *TokyoNightTheme) Success() lipgloss.Color {
	return lipgloss.Color("#9ece6a")
}

// Warning returns the warning indicator color (#e0af68 - yellow)
func (t *TokyoNightTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#e0af68")
}

// Error returns the error/negative indicator color (#f7768e - red)
func (t *TokyoNightTheme) Error() lipgloss.Color {
	return lipgloss.Color("#f7768e")
}

// Muted returns the secondary/muted text color (#565f89 - dark blue-gray)
func (t *TokyoNightTheme) Muted() lipgloss.Color {
	return lipgloss.Color("#565f89")
}

// Pre-configured styles for common UI elements (FR3.3)

// HeaderStyle provides styling for section headers and titles
func (t *TokyoNightTheme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true).
		Padding(0, 1)
}

// ContentStyle provides styling for main content areas
func (t *TokyoNightTheme) ContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground()).
		Padding(1, 2)
}

// StatusBarStyle provides styling for the bottom status bar
func (t *TokyoNightTheme) StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground()).
		Background(t.Muted()).
		Padding(0, 1).
		Bold(false)
}

// BorderStyle provides styling for borders and separators
func (t *TokyoNightTheme) BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Primary()).
		Padding(1, 2)
}

// HighlightStyle provides styling for selected/focused items
func (t *TokyoNightTheme) HighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Background()).
		Background(t.Primary()).
		Bold(true).
		Padding(0, 1)
}

// ErrorStyle provides styling for error messages
func (t *TokyoNightTheme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Error()).
		Bold(true)
}

// SuccessStyle provides styling for success messages
func (t *TokyoNightTheme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Success()).
		Bold(true)
}
