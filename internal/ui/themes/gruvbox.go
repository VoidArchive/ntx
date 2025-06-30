/**
 * NTX Portfolio Management TUI - Gruvbox Theme
 *
 * Gruvbox chosen for its proven effectiveness in reducing eye strain during
 * extended development sessions - equally valuable for marathon trading days.
 * The warm, earthy palette creates psychological comfort while maintaining
 * the high contrast ratios essential for rapid financial data processing.
 *
 * NOTE: Warm palette may increase comfort but can affect color temperature perception
 * PERF: Widely supported palette ensures consistent rendering across terminals
 */

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// GruvboxTheme uses BaseTheme composition - eliminates 90% code duplication
type GruvboxTheme struct {
	*BaseTheme
}

// NewGruvboxTheme creates Gruvbox theme with shared styling logic
func NewGruvboxTheme() Theme {
	palette := ColorPalette{
		Background: lipgloss.Color("#282828"), // Warm brown creates cozy environment
		Foreground: lipgloss.Color("#ebdbb2"), // Cream text ensures excellent readability
		Primary:    lipgloss.Color("#83a598"), // Muted blue provides focus without temperature conflict
		Success:    lipgloss.Color("#b8bb26"), // Vibrant green celebrates gains appropriately
		Warning:    lipgloss.Color("#fabd2f"), // Rich gold commands attention for risk management
		Error:      lipgloss.Color("#fb4934"), // Warm red indicates losses without harsh alarm
		Muted:      lipgloss.Color("#928374"), // Warm gray preserves theme coherence
	}

	return &GruvboxTheme{
		BaseTheme: NewBaseTheme("Gruvbox", ThemeGruvbox, palette),
	}
}
