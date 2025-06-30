/**
 * NTX Portfolio Management TUI - Rose Pine Theme
 *
 * Rose Pine selected for users who prefer warm, nature-inspired aesthetics during trading.
 * The muted, organic palette reduces visual aggression that can compound stress during
 * market downturns, particularly important in NEPSE's emotionally volatile environment.
 *
 * NOTE: Warmer tones may increase comfort but require careful contrast management
 * WARN: Lower contrast than cooler themes - verify readability in bright terminals
 */

package themes

import (
	"github.com/charmbracelet/lipgloss"
)

// RosePineTheme uses BaseTheme composition - eliminates 90% code duplication
type RosePineTheme struct {
	*BaseTheme
}

// NewRosePineTheme creates Rose Pine theme with shared styling logic
func NewRosePineTheme() Theme {
	palette := ColorPalette{
		Background: lipgloss.Color("#191724"), // Deep pine creates cozy environment
		Foreground: lipgloss.Color("#e0def4"), // Soft white ensures readability without harsh glare
		Primary:    lipgloss.Color("#c4a7e7"), // Gentle purple creates focus without demanding attention
		Success:    lipgloss.Color("#9ccfd8"), // Calming teal celebrates gains without overconfidence
		Warning:    lipgloss.Color("#f6c177"), // Warm gold suggests caution with optimistic undertones
		Error:      lipgloss.Color("#eb6f92"), // Muted rose indicates losses without emotional devastation
		Muted:      lipgloss.Color("#6e6a86"), // Subtle purple-gray maintains theme coherence
	}

	return &RosePineTheme{
		BaseTheme: NewBaseTheme("Rose Pine", ThemeRosePine, palette),
	}
}
