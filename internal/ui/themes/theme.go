/*
NTX Portfolio Management TUI - Theme System Interface

Theme interface ensures visual consistency across components while enabling
personalization for different lighting conditions during trading sessions.

Extensible architecture supports diverse user preferences without code duplication
or breaking changes to existing components.
*/

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// ThemeType enum prevents invalid theme references and enables type-safe theme switching
type ThemeType string

const (
	ThemeTokyoNight ThemeType = "tokyo_night"
	ThemeRosePine   ThemeType = "rose_pine"
	ThemeGruvbox    ThemeType = "gruvbox"
	ThemeDefault    ThemeType = "default"
)

// ColorPalette defines the color scheme for a theme
// Separates color definitions from styling logic to eliminate duplication
type ColorPalette struct {
	Background lipgloss.Color
	Foreground lipgloss.Color
	Primary    lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Muted      lipgloss.Color
}

// BaseTheme provides shared styling methods using composition
// Eliminates ~90% code duplication across theme implementations
type BaseTheme struct {
	palette ColorPalette
	name    string
	type_   ThemeType
}

// NewBaseTheme creates base theme with shared styling logic
func NewBaseTheme(name string, type_ ThemeType, palette ColorPalette) *BaseTheme {
	return &BaseTheme{
		palette: palette,
		name:    name,
		type_:   type_,
	}
}

// Core theme interface methods
func (bt *BaseTheme) Name() string               { return bt.name }
func (bt *BaseTheme) Type() ThemeType            { return bt.type_ }
func (bt *BaseTheme) Background() lipgloss.Color { return bt.palette.Background }
func (bt *BaseTheme) Foreground() lipgloss.Color { return bt.palette.Foreground }
func (bt *BaseTheme) Primary() lipgloss.Color    { return bt.palette.Primary }
func (bt *BaseTheme) Success() lipgloss.Color    { return bt.palette.Success }
func (bt *BaseTheme) Warning() lipgloss.Color    { return bt.palette.Warning }
func (bt *BaseTheme) Error() lipgloss.Color      { return bt.palette.Error }
func (bt *BaseTheme) Muted() lipgloss.Color      { return bt.palette.Muted }

// Shared styling methods - previously duplicated across all themes
func (bt *BaseTheme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(bt.Primary()).
		Bold(true).
		Padding(0, 1)
}

func (bt *BaseTheme) ContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(bt.Foreground()).
		Padding(1, 2)
}

func (bt *BaseTheme) StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(bt.Foreground()).
		Background(bt.Muted()).
		Padding(0, 1).
		Bold(false)
}

func (bt *BaseTheme) BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(bt.Primary()).
		Padding(1, 2)
}

func (bt *BaseTheme) HighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(bt.Background()).
		Background(bt.Primary()).
		Bold(true).
		Padding(0, 1)
}

func (bt *BaseTheme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(bt.Error()).
		Bold(true)
}

func (bt *BaseTheme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(bt.Success()).
		Bold(true)
}

func (bt *BaseTheme) MutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(bt.Muted())
}

// Theme interface now focuses on color palette definition
// Styling logic centralized in BaseTheme eliminates duplication
type Theme interface {
	Name() string
	Type() ThemeType
	Background() lipgloss.Color
	Foreground() lipgloss.Color
	Primary() lipgloss.Color
	Success() lipgloss.Color
	Warning() lipgloss.Color
	Error() lipgloss.Color
	Muted() lipgloss.Color
	HeaderStyle() lipgloss.Style
	ContentStyle() lipgloss.Style
	StatusBarStyle() lipgloss.Style
	BorderStyle() lipgloss.Style
	HighlightStyle() lipgloss.Style
	ErrorStyle() lipgloss.Style
	SuccessStyle() lipgloss.Style
	MutedStyle() lipgloss.Style
}

// ThemeManager provides centralized theme state preventing inconsistent styling
// Registry pattern enables theme extensibility without core system changes
type ThemeManager struct {
	themes      map[ThemeType]Theme // Theme registry enables O(1) lookup and type safety
	currentType ThemeType           // Current theme state synchronized across all components
}

// NewThemeManager bootstraps theme system with curated professional themes
// Tokyo Night default provides excellent contrast for extended trading sessions
func NewThemeManager() *ThemeManager {
	manager := &ThemeManager{
		themes:      make(map[ThemeType]Theme),
		currentType: ThemeTokyoNight, // Tokyo Night optimizes for extended terminal usage
	}

	// Theme registration enables modular addition of color schemes
	manager.RegisterTheme(NewTokyoNightTheme())
	manager.RegisterTheme(NewRosePineTheme())
	manager.RegisterTheme(NewGruvboxTheme())
	manager.RegisterTheme(NewDefaultTheme())

	return manager
}

// RegisterTheme enables theme extensibility without modifying core architecture
// Plugin-style registration supports community themes and customization
func (tm *ThemeManager) RegisterTheme(theme Theme) {
	tm.themes[theme.Type()] = theme
}

// GetCurrentTheme provides single source of truth for component styling
// Centralized theme access ensures consistent application appearance
func (tm *ThemeManager) GetCurrentTheme() Theme {
	if theme, exists := tm.themes[tm.currentType]; exists {
		return theme
	}
	// Fallback prevents application crash from corrupted theme configuration
	return tm.themes[ThemeDefault]
}

// SwitchTheme enables live theme cycling for lighting condition adaptation
// Immediate switching improves usability during long trading sessions
func (tm *ThemeManager) SwitchTheme() ThemeType {
	// Fixed theme order provides predictable cycling behavior
	themeTypes := []ThemeType{ThemeTokyoNight, ThemeRosePine, ThemeGruvbox, ThemeDefault}

	// Index lookup enables circular theme navigation
	currentIndex := 0
	for i, themeType := range themeTypes {
		if themeType == tm.currentType {
			currentIndex = i
			break
		}
	}

	// Circular navigation prevents dead-end theme states
	nextIndex := (currentIndex + 1) % len(themeTypes)
	tm.currentType = themeTypes[nextIndex]

	return tm.currentType
}

// SetTheme enables programmatic theme control for configuration loading
// Validation prevents invalid theme states that could crash rendering
func (tm *ThemeManager) SetTheme(themeType ThemeType) bool {
	if _, exists := tm.themes[themeType]; exists {
		tm.currentType = themeType
		return true
	}
	return false
}

// SetThemeByString handles configuration file theme loading
// String conversion enables human-readable theme persistence
func (tm *ThemeManager) SetThemeByString(themeName string) bool {
	themeType := ThemeType(themeName)
	return tm.SetTheme(themeType)
}

// GetCurrentThemeName provides user-visible theme identification
// Status bar integration gives immediate theme feedback
func (tm *ThemeManager) GetCurrentThemeName() string {
	return tm.GetCurrentTheme().Name()
}

// GetCurrentThemeType returns the theme type identifier for configuration persistence
// Used for saving/loading theme preferences with proper theme type identifiers
func (tm *ThemeManager) GetCurrentThemeType() string {
	return string(tm.currentType)
}

// GetAvailableThemes enables theme discovery for configuration tools
// Dynamic theme listing supports extensible theme architecture
func (tm *ThemeManager) GetAvailableThemes() []ThemeType {
	themes := make([]ThemeType, 0, len(tm.themes))
	for themeType := range tm.themes {
		themes = append(themes, themeType)
	}
	return themes
}
