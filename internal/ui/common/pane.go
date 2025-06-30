package common

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PaneType represents different types of dashboard panes
type PaneType string

const (
	PaneTypeDashboard    PaneType = "dashboard"
	PaneTypePortfolio    PaneType = "portfolio"
	PaneTypeWatchlist    PaneType = "watchlist"
	PaneTypeMarketStatus PaneType = "market_status"
)

// Pane represents a single pane in the multi-pane dashboard
type Pane interface {
	// Bubbletea model interface
	Init() tea.Cmd
	Update(msg tea.Msg) (Pane, tea.Cmd)
	View() string

	// Pane-specific interface
	GetType() PaneType
	GetTitle() string
	IsActive() bool
	SetActive(active bool)
	SetSize(width, height int)
	HandleKeypress(key string) tea.Cmd
	Refresh() tea.Cmd
}

// BasePane provides common functionality for all panes
type BasePane struct {
	Type       PaneType
	Title      string
	Width      int
	Height     int
	Active     bool
	LastUpdate time.Time
}

// GetType returns the pane type
func (bp *BasePane) GetType() PaneType {
	return bp.Type
}

// GetTitle returns the pane title
func (bp *BasePane) GetTitle() string {
	return bp.Title
}

// IsActive returns whether the pane is currently active/focused
func (bp *BasePane) IsActive() bool {
	return bp.Active
}

// SetActive sets the pane's active state
func (bp *BasePane) SetActive(active bool) {
	bp.Active = active
}

// SetSize sets the pane dimensions
func (bp *BasePane) SetSize(width, height int) {
	bp.Width = width
	bp.Height = height
}

// GetStyle returns the appropriate style for the pane based on active state
func (bp *BasePane) GetStyle() lipgloss.Style {
	if bp.Active {
		return ActivePanelStyle.Width(bp.Width).Height(bp.Height)
	}
	return PanelStyle.Width(bp.Width).Height(bp.Height)
}

// RenderTitle renders the pane title with status indicators
func (bp *BasePane) RenderTitle() string {
	title := TitleStyle.Render(bp.Title)

	// Add active indicator
	if bp.Active {
		indicator := lipgloss.NewStyle().
			Foreground(Primary).
			Render(" ◆")
		title = lipgloss.JoinHorizontal(lipgloss.Center, title, indicator)
	}

	// Add last update indicator if recent
	if !bp.LastUpdate.IsZero() {
		age := time.Since(bp.LastUpdate)
		if age < 30*time.Second {
			freshness := DataFreshnessIndicator(age)
			title = lipgloss.JoinHorizontal(lipgloss.Center, title, " ", freshness)
		}
	}

	return title
}

// MarkUpdated marks the pane as recently updated
func (bp *BasePane) MarkUpdated() {
	bp.LastUpdate = time.Now()
}

// PaneLayout defines the layout configuration for multi-pane dashboard
type PaneLayout string

const (
	LayoutQuadrant PaneLayout = "quadrant" // 2x2 grid
	LayoutSidebar  PaneLayout = "sidebar"  // Main + sidebar
	LayoutColumn   PaneLayout = "column"   // Vertical stack
	LayoutRow      PaneLayout = "row"      // Horizontal stack
)

// LayoutConfig defines how panes are arranged
type LayoutConfig struct {
	Type      PaneLayout
	PaneOrder []PaneType
	MainPane  PaneType // For sidebar layout
}

// DefaultLayouts provides predefined layout configurations
var DefaultLayouts = map[string]*LayoutConfig{
	"quadrant": {
		Type: LayoutQuadrant,
		PaneOrder: []PaneType{
			PaneTypeDashboard, PaneTypePortfolio,
			PaneTypeWatchlist, PaneTypeMarketStatus,
		},
	},
	"sidebar": {
		Type:      LayoutSidebar,
		MainPane:  PaneTypeDashboard,
		PaneOrder: []PaneType{PaneTypeDashboard, PaneTypePortfolio, PaneTypeWatchlist},
	},
	"column": {
		Type: LayoutColumn,
		PaneOrder: []PaneType{
			PaneTypeMarketStatus,
			PaneTypePortfolio,
			PaneTypeWatchlist,
		},
	},
}

// LayoutManager handles pane positioning and sizing
type LayoutManager struct {
	Config       *LayoutConfig
	TermWidth    int
	TermHeight   int
	HeaderHeight int
	FooterHeight int
}

// NewLayoutManager creates a new layout manager
func NewLayoutManager(config *LayoutConfig) *LayoutManager {
	return &LayoutManager{
		Config:       config,
		HeaderHeight: 3, // Header + spacing
		FooterHeight: 2, // Help text + spacing
	}
}

// SetTerminalSize updates the terminal dimensions
func (lm *LayoutManager) SetTerminalSize(width, height int) {
	lm.TermWidth = width
	lm.TermHeight = height
}

// CalculatePaneDimensions returns width and height for each pane based on layout
func (lm *LayoutManager) CalculatePaneDimensions() map[PaneType]struct{ Width, Height int } {
	contentHeight := lm.TermHeight - lm.HeaderHeight - lm.FooterHeight
	dimensions := make(map[PaneType]struct{ Width, Height int })

	switch lm.Config.Type {
	case LayoutQuadrant:
		// 2x2 grid layout
		paneWidth := (lm.TermWidth - 3) / 2   // Account for spacing
		paneHeight := (contentHeight - 1) / 2 // Account for spacing

		for _, paneType := range lm.Config.PaneOrder {
			dimensions[paneType] = struct{ Width, Height int }{paneWidth, paneHeight}
		}

	case LayoutSidebar:
		// Main pane + sidebar
		mainWidth := (lm.TermWidth * 2) / 3
		sidebarWidth := lm.TermWidth - mainWidth - 1

		for _, paneType := range lm.Config.PaneOrder {
			if paneType == lm.Config.MainPane {
				dimensions[paneType] = struct{ Width, Height int }{mainWidth, contentHeight}
			} else {
				sidebarHeight := (contentHeight - 1) / (len(lm.Config.PaneOrder) - 1)
				dimensions[paneType] = struct{ Width, Height int }{sidebarWidth, sidebarHeight}
			}
		}

	case LayoutColumn:
		// Vertical stack
		paneWidth := lm.TermWidth
		paneHeight := (contentHeight - len(lm.Config.PaneOrder) + 1) / len(lm.Config.PaneOrder)

		for _, paneType := range lm.Config.PaneOrder {
			dimensions[paneType] = struct{ Width, Height int }{paneWidth, paneHeight}
		}

	case LayoutRow:
		// Horizontal stack
		paneWidth := (lm.TermWidth - len(lm.Config.PaneOrder) + 1) / len(lm.Config.PaneOrder)
		paneHeight := contentHeight

		for _, paneType := range lm.Config.PaneOrder {
			dimensions[paneType] = struct{ Width, Height int }{paneWidth, paneHeight}
		}
	}

	return dimensions
}

// RenderPanes arranges panes according to the layout configuration
func (lm *LayoutManager) RenderPanes(panes map[PaneType]Pane) string {
	var rows []string

	switch lm.Config.Type {
	case LayoutQuadrant:
		// Render as 2x2 grid
		if len(lm.Config.PaneOrder) >= 4 {
			topRow := lipgloss.JoinHorizontal(lipgloss.Top,
				panes[lm.Config.PaneOrder[0]].View(),
				" ",
				panes[lm.Config.PaneOrder[1]].View(),
			)
			bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
				panes[lm.Config.PaneOrder[2]].View(),
				" ",
				panes[lm.Config.PaneOrder[3]].View(),
			)
			rows = []string{topRow, bottomRow}
		}

	case LayoutSidebar:
		// Render main pane + sidebar
		var sidebarPanes []string
		var mainPaneView string

		for _, paneType := range lm.Config.PaneOrder {
			if paneType == lm.Config.MainPane {
				mainPaneView = panes[paneType].View()
			} else {
				sidebarPanes = append(sidebarPanes, panes[paneType].View())
			}
		}

		sidebar := lipgloss.JoinVertical(lipgloss.Left, sidebarPanes...)
		row := lipgloss.JoinHorizontal(lipgloss.Top, mainPaneView, " ", sidebar)
		rows = []string{row}

	case LayoutColumn:
		// Render as vertical stack
		for _, paneType := range lm.Config.PaneOrder {
			rows = append(rows, panes[paneType].View())
		}

	case LayoutRow:
		// Render as horizontal stack
		var paneViews []string
		for _, paneType := range lm.Config.PaneOrder {
			paneViews = append(paneViews, panes[paneType].View())
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, paneViews...)
		rows = []string{row}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// GetNavigationHelp returns help text for pane navigation
func (lm *LayoutManager) GetNavigationHelp() string {
	help := []string{
		KeyStyle.Render("Tab") + "/Shift+Tab: Switch panes",
		KeyStyle.Render("R") + ": Refresh",
		KeyStyle.Render("L") + ": Change layout",
		KeyStyle.Render("Q") + ": Quit",
	}

	return HelpStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, help...))
}
