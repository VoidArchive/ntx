/**
 * NTX Portfolio Management TUI - Default Theme
 *
 * Default theme designed for maximum terminal compatibility and professional appearance
 * in corporate environments. Custom palette avoids brand associations while maintaining
 * the contrast ratios essential for financial data interpretation.
 *
 * WARN: Must maintain highest contrast ratios as other themes may not be available
 * TEST: Verify readability across terminals with varying gamma/contrast settings
 */

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// DefaultTheme uses BaseTheme composition - eliminates 90% code duplication
type DefaultTheme struct {
	*BaseTheme
}

// NewDefaultTheme creates Default theme with shared styling logic
func NewDefaultTheme() Theme {
	palette := ColorPalette{
		Background: lipgloss.Color("#141415"), // Near-black provides maximum contrast foundation
		Foreground: lipgloss.Color("#cdcdcd"), // High contrast light gray ensures number readability
		Primary:    lipgloss.Color("#6e94b2"), // Professional blue commands attention without distraction
		Success:    lipgloss.Color("#7fa563"), // Conventional green clearly indicates positive performance
		Warning:    lipgloss.Color("#f3be7c"), // Professional gold suggests caution appropriately
		Error:      lipgloss.Color("#d8647e"), // Muted red indicates losses without psychological devastation
		Muted:      lipgloss.Color("#606079"), // Neutral gray maintains information hierarchy
	}

	return &DefaultTheme{
		BaseTheme: NewBaseTheme("Default", ThemeDefault, palette),
	}
}
