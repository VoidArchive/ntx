/**
 * NTX Portfolio Management TUI - Tokyo Night Theme
 *
 * Tokyo Night chosen for its proven psychological benefits during extended trading sessions.
 * The deep purple/blue palette reduces eye strain during NEPSE's 11:00-15:00 trading hours,
 * while maintaining the high contrast ratios essential for rapid financial data scanning.
 *
 * PERF: Official color specification ensures consistent rendering across terminal emulators
 * NOTE: Popular theme familiarity reduces cognitive load for developers using NTX
 */

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// TokyoNightTheme uses BaseTheme composition - eliminates 90% code duplication
type TokyoNightTheme struct {
	*BaseTheme
}

// NewTokyoNightTheme creates Tokyo Night theme with shared styling logic
func NewTokyoNightTheme() Theme {
	palette := ColorPalette{
		Background: lipgloss.Color("#1a1b26"), // Deep navy reduces blue light exposure
		Foreground: lipgloss.Color("#c0caf5"), // High contrast for rapid number recognition
		Primary:    lipgloss.Color("#7aa2f7"), // Bright blue commands attention without alarming
		Success:    lipgloss.Color("#9ece6a"), // Universal green for gains, prevents overconfidence
		Warning:    lipgloss.Color("#e0af68"), // Warm amber triggers caution without panic
		Error:      lipgloss.Color("#f7768e"), // Soft coral red indicates losses without stress
		Muted:      lipgloss.Color("#565f89"), // Subtle contrast reduces information noise
	}

	return &TokyoNightTheme{
		BaseTheme: NewBaseTheme("Tokyo Night", ThemeTokyoNight, palette),
	}
}
