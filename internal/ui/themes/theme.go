/**
 * NTX Portfolio Management TUI - Theme System Interface
 *
 * This file defines the theme interface and core theme system that enables
 * multiple color schemes for the TUI. The interface allows for easy theme
 * switching and consistent styling across all UI components.
 *
 * The theme system supports the Tokyo Night theme as the primary theme,
 * with extensibility for additional themes in future phases.
 */

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// ThemeType represents the available theme types for easy identification
type ThemeType string

const (
	ThemeTokyoNight ThemeType = "tokyo_night"
	ThemeRosePine   ThemeType = "rose_pine"
	ThemeGruvbox    ThemeType = "gruvbox"
	ThemeDefault    ThemeType = "default"
)

// Theme defines the interface that all themes must implement
// This provides a consistent color palette and styling approach across themes
type Theme interface {
	// Name returns the human-readable name of the theme
	Name() string

	// Type returns the theme type identifier
	Type() ThemeType

	// Color palette methods - core colors used throughout the interface
	Background() lipgloss.Color // Main background color
	Foreground() lipgloss.Color // Primary text color
	Primary() lipgloss.Color    // Accent/highlight color
	Success() lipgloss.Color    // Positive indicators (gains, success)
	Warning() lipgloss.Color    // Warning indicators
	Error() lipgloss.Color      // Error indicators (losses, errors)
	Muted() lipgloss.Color      // Secondary/muted text

	// Styling methods - pre-configured styles for common UI elements
	HeaderStyle() lipgloss.Style    // Section headers and titles
	ContentStyle() lipgloss.Style   // Main content area styling
	StatusBarStyle() lipgloss.Style // Bottom status bar styling
	BorderStyle() lipgloss.Style    // Border and separator styling
	HighlightStyle() lipgloss.Style // Selected/focused item styling
	ErrorStyle() lipgloss.Style     // Error message styling
	SuccessStyle() lipgloss.Style   // Success message styling
}

// ThemeManager manages the available themes and current theme selection
// Provides centralized theme switching and theme registry functionality
type ThemeManager struct {
	themes      map[ThemeType]Theme // Registry of available themes
	currentType ThemeType           // Currently active theme type
}

// NewThemeManager creates a new theme manager with default themes registered
// Returns a manager initialized with Tokyo Night as the default theme
func NewThemeManager() *ThemeManager {
	manager := &ThemeManager{
		themes:      make(map[ThemeType]Theme),
		currentType: ThemeTokyoNight, // Default to Tokyo Night as specified in FR3.2
	}

	// Register all available themes
	manager.RegisterTheme(NewTokyoNightTheme())
	manager.RegisterTheme(NewRosePineTheme())
	manager.RegisterTheme(NewGruvboxTheme())
	manager.RegisterTheme(NewDefaultTheme())

	return manager
}

// RegisterTheme adds a theme to the manager's registry
// Allows for easy addition of new themes in future phases
func (tm *ThemeManager) RegisterTheme(theme Theme) {
	tm.themes[theme.Type()] = theme
}

// GetCurrentTheme returns the currently active theme
// This is the primary method used by UI components to get styling
func (tm *ThemeManager) GetCurrentTheme() Theme {
	if theme, exists := tm.themes[tm.currentType]; exists {
		return theme
	}
	// Fallback to default theme if current theme is not found
	return tm.themes[ThemeDefault]
}

// SwitchTheme cycles to the next available theme
// Implements the theme switching functionality for the 't' key (FR3.4)
func (tm *ThemeManager) SwitchTheme() ThemeType {
	// Get all available theme types in order
	themeTypes := []ThemeType{ThemeTokyoNight, ThemeRosePine, ThemeGruvbox, ThemeDefault}

	// Find current theme index
	currentIndex := 0
	for i, themeType := range themeTypes {
		if themeType == tm.currentType {
			currentIndex = i
			break
		}
	}

	// Switch to next theme (cycle back to first if at end)
	nextIndex := (currentIndex + 1) % len(themeTypes)
	tm.currentType = themeTypes[nextIndex]

	return tm.currentType
}

// SetTheme sets the current theme to the specified type
// Used for configuration-based theme setting and explicit theme selection
func (tm *ThemeManager) SetTheme(themeType ThemeType) bool {
	if _, exists := tm.themes[themeType]; exists {
		tm.currentType = themeType
		return true
	}
	return false
}

// GetCurrentThemeName returns the name of the currently active theme
// Used for display in status bars and configuration interfaces
func (tm *ThemeManager) GetCurrentThemeName() string {
	return tm.GetCurrentTheme().Name()
}

// GetAvailableThemes returns a list of all available theme types
// Useful for configuration interfaces and theme selection menus
func (tm *ThemeManager) GetAvailableThemes() []ThemeType {
	themes := make([]ThemeType, 0, len(tm.themes))
	for themeType := range tm.themes {
		themes = append(themes, themeType)
	}
	return themes
}
