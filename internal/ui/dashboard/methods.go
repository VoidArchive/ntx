package dashboard

import (
	"context"
	"fmt"
	"ntx/internal/ui/common"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// renderLoading renders the loading screen
func (m Model) renderLoading() string {
	loading := common.InfoStyle.Render("Initializing NTX Dashboard...")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, loading)
}

// renderHeader renders the dashboard header
func (m Model) renderHeader() string {
	title := common.HeaderStyle.Render("🚀 NTX - NEPSE Power Terminal")

	// Add status indicators
	status := m.renderStatusIndicators()

	// Combine title and status
	headerContent := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", status)

	return headerContent
}

// renderStatusIndicators renders various status indicators in the header
func (m Model) renderStatusIndicators() string {
	var indicators []string

	// Current layout indicator
	layoutIndicator := common.MutedStyle.Render(fmt.Sprintf("Layout: %s", m.currentLayout))
	indicators = append(indicators, layoutIndicator)

	// Active pane indicator
	paneIndicator := common.InfoStyle.Render(fmt.Sprintf("Pane: %s", string(m.activePaneType)))
	indicators = append(indicators, paneIndicator)

	// Auto-update indicator
	if m.autoUpdate {
		updateIndicator := common.SuccessStyle.Render("● LIVE")
		indicators = append(indicators, updateIndicator)
	} else {
		updateIndicator := common.MutedStyle.Render("● MANUAL")
		indicators = append(indicators, updateIndicator)
	}

	// Last update time
	if !m.lastUpdate.IsZero() {
		age := time.Since(m.lastUpdate)
		timeIndicator := common.DataFreshnessIndicator(age)
		indicators = append(indicators, timeIndicator)
	}

	return strings.Join(indicators, " │ ")
}

// renderFooter renders the help and status footer
func (m Model) renderFooter() string {
	var sections []string

	// Navigation help
	navHelp := m.layoutManager.GetNavigationHelp()
	sections = append(sections, navHelp)

	// Error message if any
	if m.errorMessage != "" {
		errorMsg := common.ErrorStyle.Render("Error: " + m.errorMessage)
		sections = append(sections, errorMsg)
	}

	// Loading indicator
	if m.isLoading {
		loadingMsg := common.InfoStyle.Render("Loading...")
		sections = append(sections, loadingMsg)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// tickCmd creates a tick command for auto-updates
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return common.TickMsg(t)
	})
}

// refreshAllData triggers a refresh of all market data
func (m Model) refreshAllData() tea.Cmd {
	return func() tea.Msg {
		if m.marketService == nil {
			return common.NewErrorMsg("dashboard", fmt.Errorf("market service not available"))
		}

		// This would trigger the market service to refresh
		// For now, we'll simulate with a refresh message
		return common.RefreshMsg{Force: true}
	}
}

// fetchMarketData fetches market data from the service asynchronously (non-blocking)
func (m Model) fetchMarketData() tea.Cmd {
	return func() tea.Msg {
		if m.marketService == nil {
			return nil
		}

		// Use a short timeout to prevent UI blocking and launch async
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			defer func() {
				// Always recover from panics
				if r := recover(); r != nil {
					m.logger.Error("Market data fetch panic recovered", "error", r)
				}
			}()

			// Get market status asynchronously
			status, err := m.marketService.GetMarketStatus(ctx)
			if err != nil {
				m.logger.Warn("Failed to fetch market status", "error", err)
				return
			}

			m.logger.Debug("Market status updated",
				"is_open", status.IsOpen,
				"active_symbols", status.ActiveSymbols)
		}()

		// Return immediately to keep UI responsive
		return nil
	}
}

// switchToNextPane switches to the next pane in order
func (m Model) switchToNextPane() tea.Cmd {
	paneOrder := m.layoutManager.Config.PaneOrder
	currentIndex := -1

	// Find current pane index
	for i, paneType := range paneOrder {
		if paneType == m.activePaneType {
			currentIndex = i
			break
		}
	}

	// Move to next pane
	nextIndex := (currentIndex + 1) % len(paneOrder)
	nextPaneType := paneOrder[nextIndex]

	return m.switchToPane(nextPaneType)
}

// switchToPrevPane switches to the previous pane in order
func (m Model) switchToPrevPane() tea.Cmd {
	paneOrder := m.layoutManager.Config.PaneOrder
	currentIndex := -1

	// Find current pane index
	for i, paneType := range paneOrder {
		if paneType == m.activePaneType {
			currentIndex = i
			break
		}
	}

	// Move to previous pane
	prevIndex := currentIndex - 1
	if prevIndex < 0 {
		prevIndex = len(paneOrder) - 1
	}
	prevPaneType := paneOrder[prevIndex]

	return m.switchToPane(prevPaneType)
}

// switchToPane switches focus to a specific pane
func (m Model) switchToPane(paneType common.PaneType) tea.Cmd {
	// Deactivate current pane
	if currentPane, exists := m.panes[m.activePaneType]; exists {
		currentPane.SetActive(false)
	}

	// Activate new pane
	if newPane, exists := m.panes[paneType]; exists {
		newPane.SetActive(true)
		// Return a command to update the active pane type
		return func() tea.Msg {
			return common.PaneFocusMsg{PaneType: paneType}
		}
	}

	return nil
}

// switchLayout cycles through available layouts
func (m Model) switchLayout() tea.Cmd {
	layouts := []string{"quadrant", "sidebar", "column"}
	currentIndex := 0

	// Find current layout index
	for i, layout := range layouts {
		if layout == m.currentLayout {
			currentIndex = i
			break
		}
	}

	// Move to next layout
	nextIndex := (currentIndex + 1) % len(layouts)
	nextLayout := layouts[nextIndex]

	return common.NewLayoutChangeMsg(nextLayout)
}

// changeLayout changes the dashboard layout
func (m *Model) changeLayout(layoutName string) {
	if config, exists := common.DefaultLayouts[layoutName]; exists {
		m.layoutManager.Config = config
		m.currentLayout = layoutName

		// Recalculate pane dimensions
		dimensions := m.layoutManager.CalculatePaneDimensions()
		for paneType, pane := range m.panes {
			if dim, exists := dimensions[paneType]; exists {
				pane.SetSize(dim.Width, dim.Height)
			}
		}

		m.logger.Info("Layout changed", "layout", layoutName)
	}
}

// cleanup performs cleanup when shutting down
func (m *Model) cleanup() {
	if m.updateTicker != nil {
		m.updateTicker.Stop()
	}

	m.logger.Info("Dashboard shutting down")
}

// GetActivePaneType returns the currently active pane type
func (m Model) GetActivePaneType() common.PaneType {
	return m.activePaneType
}

// GetCurrentLayout returns the current layout name
func (m Model) GetCurrentLayout() string {
	return m.currentLayout
}

// SetAutoUpdate enables or disables auto-updates
func (m *Model) SetAutoUpdate(enabled bool) {
	m.autoUpdate = enabled

	if enabled && m.updateTicker == nil {
		m.updateTicker = time.NewTicker(5 * time.Second)
	} else if !enabled && m.updateTicker != nil {
		m.updateTicker.Stop()
		m.updateTicker = nil
	}
}

// IsReady returns whether the dashboard is fully initialized
func (m Model) IsReady() bool {
	return m.ready
}

// GetPaneCount returns the number of panes
func (m Model) GetPaneCount() int {
	return len(m.panes)
}

// HasError returns whether there's an active error
func (m Model) HasError() bool {
	return m.errorMessage != ""
}

// ClearError clears the current error message
func (m *Model) ClearError() {
	m.errorMessage = ""
}

// GetLastUpdateAge returns how long ago the last update occurred
func (m Model) GetLastUpdateAge() time.Duration {
	if m.lastUpdate.IsZero() {
		return 0
	}
	return time.Since(m.lastUpdate)
}

// GetDebugInfo returns debug information about the dashboard state
func (m Model) GetDebugInfo() map[string]any {
	info := map[string]any{
		"ready":          m.ready,
		"width":          m.width,
		"height":         m.height,
		"active_pane":    string(m.activePaneType),
		"current_layout": m.currentLayout,
		"auto_update":    m.autoUpdate,
		"pane_count":     len(m.panes),
		"has_error":      m.HasError(),
		"last_update":    m.lastUpdate,
		"is_loading":     m.isLoading,
	}

	return info
}
