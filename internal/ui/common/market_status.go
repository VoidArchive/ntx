package common

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MarketStatusPane displays real-time market status and system health
type MarketStatusPane struct {
	BasePane

	// Market data
	IsMarketOpen  bool
	NextOpen      time.Time
	NextClose     time.Time
	ActiveSymbols int
	DataAge       time.Duration

	// System health
	SystemStatus   string
	ComponentsUp   int
	ComponentsDown int
	CacheHitRate   float64

	// Display state
	showHealth bool
	lastError  error
}

// NewMarketStatusPane creates a new market status pane
func NewMarketStatusPane() *MarketStatusPane {
	return &MarketStatusPane{
		BasePane: BasePane{
			Type:  PaneTypeMarketStatus,
			Title: "Market Status",
		},
		showHealth: true,
	}
}

// Init initializes the market status pane
func (msp *MarketStatusPane) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the pane
func (msp *MarketStatusPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case MarketStatusMsg:
		msp.IsMarketOpen = msg.IsOpen
		msp.NextOpen = msg.NextOpen
		msp.NextClose = msg.NextClose
		msp.ActiveSymbols = msg.ActiveSymbols
		msp.DataAge = msg.DataAge
		msp.MarkUpdated()

	case SystemHealthMsg:
		msp.SystemStatus = msg.Status
		msp.ComponentsUp = msg.ComponentsUp
		msp.ComponentsDown = msg.ComponentsDown
		msp.CacheHitRate = msg.CacheHitRate
		msp.MarkUpdated()

	case ErrorMsg:
		msp.lastError = msg.Error

	case tea.KeyMsg:
		if msp.Active {
			return msp, msp.HandleKeypress(msg.String())
		}
	}

	return msp, nil
}

// View renders the market status pane
func (msp *MarketStatusPane) View() string {
	style := msp.GetStyle()
	content := msp.renderContent()

	return style.Render(content)
}

// HandleKeypress handles pane-specific key presses
func (msp *MarketStatusPane) HandleKeypress(key string) tea.Cmd {
	switch key {
	case "h":
		msp.showHealth = !msp.showHealth
		return nil
	}
	return nil
}

// Refresh triggers a refresh of market status data
func (msp *MarketStatusPane) Refresh() tea.Cmd {
	return NewRefreshMsg(false)
}

// renderContent renders the pane content
func (msp *MarketStatusPane) renderContent() string {
	var sections []string

	// Title
	sections = append(sections, msp.RenderTitle())

	// Market status section
	sections = append(sections, msp.renderMarketStatus())

	// System health section (if enabled)
	if msp.showHealth {
		sections = append(sections, msp.renderSystemHealth())
	}

	// Error section (if any)
	if msp.lastError != nil {
		sections = append(sections, msp.renderError())
	}

	// Help section
	sections = append(sections, msp.renderHelp())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderMarketStatus renders the current market status
func (msp *MarketStatusPane) renderMarketStatus() string {
	var lines []string

	// Market open/closed status
	statusLine := "Status: " + MarketStatusIndicator(msp.IsMarketOpen)
	lines = append(lines, statusLine)

	// Next market event
	now := time.Now()
	if msp.IsMarketOpen {
		if !msp.NextClose.IsZero() {
			remaining := msp.NextClose.Sub(now)
			nextEvent := fmt.Sprintf("Closes in: %s", formatDuration(remaining))
			lines = append(lines, InfoStyle.Render(nextEvent))
		}
	} else {
		if !msp.NextOpen.IsZero() {
			until := msp.NextOpen.Sub(now)
			if until > 0 {
				nextEvent := fmt.Sprintf("Opens in: %s", formatDuration(until))
				lines = append(lines, InfoStyle.Render(nextEvent))
			} else {
				lines = append(lines, WarningStyle.Render("Market should be open"))
			}
		}
	}

	// Data freshness
	if msp.DataAge > 0 {
		freshness := fmt.Sprintf("Data age: %s", DataFreshnessIndicator(msp.DataAge))
		lines = append(lines, freshness)
	}

	// Active symbols count
	if msp.ActiveSymbols > 0 {
		symbolCount := fmt.Sprintf("Active symbols: %s",
			InfoStyle.Render(fmt.Sprintf("%d", msp.ActiveSymbols)))
		lines = append(lines, symbolCount)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderSystemHealth renders system health metrics
func (msp *MarketStatusPane) renderSystemHealth() string {
	var lines []string

	// Section header
	healthHeader := SubHeaderStyle.Render(" System Health ")
	lines = append(lines, healthHeader)

	// Overall status
	var statusStyle lipgloss.Style
	switch strings.ToLower(msp.SystemStatus) {
	case "healthy":
		statusStyle = SuccessStyle
	case "degraded":
		statusStyle = WarningStyle
	case "unhealthy":
		statusStyle = ErrorStyle
	default:
		statusStyle = MutedStyle
	}

	statusLine := fmt.Sprintf("Status: %s", statusStyle.Render(msp.SystemStatus))
	lines = append(lines, statusLine)

	// Component status
	if msp.ComponentsUp > 0 || msp.ComponentsDown > 0 {
		total := msp.ComponentsUp + msp.ComponentsDown
		upPercent := float64(msp.ComponentsUp) / float64(total) * 100

		componentLine := fmt.Sprintf("Components: %s up, %s down (%.1f%%)",
			SuccessStyle.Render(fmt.Sprintf("%d", msp.ComponentsUp)),
			ErrorStyle.Render(fmt.Sprintf("%d", msp.ComponentsDown)),
			upPercent)
		lines = append(lines, componentLine)
	}

	// Cache hit rate
	if msp.CacheHitRate > 0 {
		var cacheStyle lipgloss.Style
		if msp.CacheHitRate >= 90 {
			cacheStyle = SuccessStyle
		} else if msp.CacheHitRate >= 70 {
			cacheStyle = WarningStyle
		} else {
			cacheStyle = ErrorStyle
		}

		cacheLine := fmt.Sprintf("Cache hit rate: %s",
			cacheStyle.Render(fmt.Sprintf("%.1f%%", msp.CacheHitRate)))
		lines = append(lines, cacheLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderError renders the last error if any
func (msp *MarketStatusPane) renderError() string {
	if msp.lastError == nil {
		return ""
	}

	errorHeader := ErrorStyle.Render(" Last Error ")
	errorText := MutedStyle.Render(msp.lastError.Error())

	return lipgloss.JoinVertical(lipgloss.Left, errorHeader, errorText)
}

// renderHelp renders pane-specific help
func (msp *MarketStatusPane) renderHelp() string {
	if !msp.Active {
		return ""
	}

	help := []string{
		KeyStyle.Render("h") + ": Toggle health view",
		KeyStyle.Render("r") + ": Refresh status",
	}

	return HelpStyle.Render(strings.Join(help, " • "))
}

// Helper functions

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "overdue"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}
