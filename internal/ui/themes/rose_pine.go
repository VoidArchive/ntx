/**
 * NTX Portfolio Management TUI - Rose Pine Theme
 *
 * This file implements the Rose Pine color scheme, a warm and elegant theme
 * with rose and pine color tones. Rose Pine is known for its soothing,
 * nature-inspired palette that provides excellent readability while
 * maintaining a cozy, comfortable aesthetic.
 *
 * Color palette follows the official Rose Pine specification for consistency
 * with popular development environments and modern aesthetic preferences.
 */

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// RosePineTheme implements the Theme interface with Rose Pine colors
// Provides warm rose/pine aesthetic with nature-inspired tones
type RosePineTheme struct{}

// NewRosePineTheme creates a new Rose Pine theme instance
// Returns a theme configured with the official Rose Pine color palette
func NewRosePineTheme() Theme {
	return &RosePineTheme{}
}

// Name returns the human-readable name of the Rose Pine theme
func (t *RosePineTheme) Name() string {
	return "Rose Pine"
}

// Type returns the theme type identifier for Rose Pine
func (t *RosePineTheme) Type() ThemeType {
	return ThemeRosePine
}

// Color palette implementation (Official Rose Pine specification)

// Background returns the main background color (#191724 - deep pine)
func (t *RosePineTheme) Background() lipgloss.Color {
	return lipgloss.Color("#191724")
}

// Foreground returns the primary text color (#e0def4 - soft white)
func (t *RosePineTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#e0def4")
}

// Primary returns the accent/highlight color (#c4a7e7 - soft purple)
func (t *RosePineTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#c4a7e7")
}

// Success returns the success/positive indicator color (#9ccfd8 - soft teal)
func (t *RosePineTheme) Success() lipgloss.Color {
	return lipgloss.Color("#9ccfd8")
}

// Warning returns the warning indicator color (#f6c177 - warm gold)
func (t *RosePineTheme) Warning() lipgloss.Color {
	return lipgloss.Color("#f6c177")
}

// Error returns the error/negative indicator color (#eb6f92 - soft rose)
func (t *RosePineTheme) Error() lipgloss.Color {
	return lipgloss.Color("#eb6f92")
}

// Muted returns the secondary/muted text color (#6e6a86 - muted purple)
func (t *RosePineTheme) Muted() lipgloss.Color {
	return lipgloss.Color("#6e6a86")
}

// Pre-configured styles for common UI elements

// HeaderStyle provides styling for section headers and titles
func (t *RosePineTheme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true).
		Padding(0, 1)
}

// ContentStyle provides styling for main content areas
func (t *RosePineTheme) ContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground()).
		Padding(1, 2)
}

// StatusBarStyle provides styling for the bottom status bar
func (t *RosePineTheme) StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Foreground()).
		Background(t.Muted()).
		Padding(0, 1).
		Bold(false)
}

// BorderStyle provides styling for borders and separators
func (t *RosePineTheme) BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Primary()).
		Padding(1, 2)
}

// HighlightStyle provides styling for selected/focused items
func (t *RosePineTheme) HighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Background()).
		Background(t.Primary()).
		Bold(true).
		Padding(0, 1)
}

// ErrorStyle provides styling for error messages
func (t *RosePineTheme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Error()).
		Bold(true)
}

// SuccessStyle provides styling for success messages
func (t *RosePineTheme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Success()).
		Bold(true)
}
