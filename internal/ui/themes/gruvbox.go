/**
 * NTX Portfolio Management TUI - Gruvbox Theme
 *
 * This file implements the Gruvbox color scheme, a retro groove color scheme
 * that provides warm, earthy tones. Gruvbox is beloved by developers for its
 * excellent contrast and comfortable readability, featuring warm browns,
 * oranges, and greens that create a cozy coding environment.
 *
 * Color palette follows the classic Gruvbox dark specification for consistency
 * with the widely-used Gruvbox theme across development tools.
 */

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// GruvboxTheme implements the Theme interface with Gruvbox colors
// Provides warm, retro aesthetic with earthy tones
type GruvboxTheme struct{}

// NewGruvboxTheme creates a new Gruvbox theme instance
// Returns a theme configured with the classic Gruvbox dark color palette
func NewGruvboxTheme() Theme {
	return &GruvboxTheme{}
}

// Name returns the human-readable name of the Gruvbox theme
func (t *GruvboxTheme) Name() string {
	return "Gruvbox"
}

// Type returns the theme type identifier for Gruvbox
func (t *GruvboxTheme) Type() ThemeType {
	return ThemeGruvbox
}

// Color palette implementation (Classic Gruvbox Dark specification)

// Background returns the main background color (#282828 - warm dark brown)
func (t *GruvboxTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282828")
}

// Foreground returns the primary text color (#ebdbb2 - warm cream)
func (t *GruvboxTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#ebdbb2")
}

// Primary returns the accent/highlight color (#83a598 - soft blue)
func (t *GruvboxTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#83a598")
}

// Success returns the success/positive indicator color (#b8bb26 - bright green)
func (t *GruvboxTheme) Success() lipgloss.Color {
	return lipgloss.Color("#b8bb26")
}

// Warning returns the warning indicator color (#fabd2f - golden yellow)
func (t *GruvboxTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#fabd2f")
}

// Error returns the error/negative indicator color (#fb4934 - warm red)
func (t *GruvboxTheme) Error() lipgloss.Color {
	return lipgloss.Color("#fb4934")
}

// Muted returns the secondary/muted text color (#928374 - warm gray)
func (t *GruvboxTheme) Muted() lipgloss.Color {
	return lipgloss.Color("#928374")
}

// Pre-configured styles for common UI elements

// HeaderStyle provides styling for section headers and titles
func (t *GruvboxTheme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true).
		Padding(0, 1)
}

// ContentStyle provides styling for main content areas
func (t *GruvboxTheme) ContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground()).
		Padding(1, 2)
}

// StatusBarStyle provides styling for the bottom status bar
func (t *GruvboxTheme) StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground()).
		Background(t.Muted()).
		Padding(0, 1).
		Bold(false)
}

// BorderStyle provides styling for borders and separators
func (t *GruvboxTheme) BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Primary()).
		Padding(1, 2)
}

// HighlightStyle provides styling for selected/focused items
func (t *GruvboxTheme) HighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Background()).
		Background(t.Primary()).
		Bold(true).
		Padding(0, 1)
}

// ErrorStyle provides styling for error messages
func (t *GruvboxTheme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Error()).
		Bold(true)
}

// SuccessStyle provides styling for success messages
func (t *GruvboxTheme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Success()).
		Bold(true)
}
