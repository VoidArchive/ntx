package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/VoidArchive/ntx/internal/domain"
)

// SortMode defines how the portfolio table is sorted
type SortMode int

const (
	SortBySymbol SortMode = iota
	SortByGainPercent
	SortByMarketValue
	SortByGainLoss
)

func (s SortMode) String() string {
	switch s {
	case SortBySymbol:
		return "Symbol"
	case SortByGainPercent:
		return "Gain%"
	case SortByMarketValue:
		return "Market Value"
	case SortByGainLoss:
		return "Gain/Loss"
	default:
		return "Unknown"
	}
}

// PortfolioModel handles the portfolio view
type PortfolioModel struct {
	portfolio      *domain.Portfolio
	holdings       []domain.Holding
	selectedIndex  int
	sortMode       SortMode
	windowSize     tea.WindowSizeMsg
	showDetail     bool
	detailSymbol   string
}

// NewPortfolioModel creates a new portfolio model
func NewPortfolioModel(portfolio *domain.Portfolio) *PortfolioModel {
	return &PortfolioModel{
		portfolio:     portfolio,
		holdings:      []domain.Holding{},
		selectedIndex: 0,
		sortMode:      SortBySymbol,
	}
}

// Init initializes the portfolio model
func (m *PortfolioModel) Init() tea.Cmd {
	m.refreshHoldings()
	return nil
}

// Update handles messages and updates the model
func (m *PortfolioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showDetail {
			return m.handleDetailKeys(msg)
		}
		return m.handleTableKeys(msg)

	case tea.WindowSizeMsg:
		m.windowSize = msg

	case RefreshPortfolioMsg:
		m.refreshHoldings()
	}

	return m, nil
}

// handleTableKeys handles key presses when showing the table view
func (m *PortfolioModel) handleTableKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < len(m.holdings)-1 {
			m.selectedIndex++
		}
	case "enter":
		if len(m.holdings) > 0 && m.selectedIndex < len(m.holdings) {
			m.showDetail = true
			m.detailSymbol = m.holdings[m.selectedIndex].StockSymbol
		}
	case "s":
		// Cycle through sort modes
		m.sortMode = SortMode((int(m.sortMode) + 1) % 4)
		m.refreshHoldings()
	case "i":
		// TODO: Trigger CSV import (this should be handled by parent)
		return m, func() tea.Msg {
			return SuccessMsg{Message: "Import functionality will be available in Transactions section"}
		}
	case "e":
		// TODO: Edit selected stock price
		if len(m.holdings) > 0 && m.selectedIndex < len(m.holdings) {
			symbol := m.holdings[m.selectedIndex].StockSymbol
			return m, func() tea.Msg {
				return SuccessMsg{Message: fmt.Sprintf("Edit price for %s (feature coming soon)", symbol)}
			}
		}
	}
	return m, nil
}

// handleDetailKeys handles key presses when showing detail view
func (m *PortfolioModel) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.showDetail = false
		m.detailSymbol = ""
	}
	return m, nil
}

// View renders the portfolio view
func (m *PortfolioModel) View() string {
	if m.windowSize.Width == 0 {
		return "Loading portfolio..."
	}

	if m.showDetail {
		return m.renderDetailView()
	}

	return m.renderTableView()
}

// renderTableView renders the main portfolio table
func (m *PortfolioModel) renderTableView() string {
	var content strings.Builder

	// Section title
	title := SectionTitleStyle.Render("📊 Portfolio Holdings")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Holdings table
	if len(m.holdings) == 0 {
		noData := PanelStyle.
			Width(m.windowSize.Width - 8).
			Align(lipgloss.Center).
			Render("No holdings found. Import transactions to get started!")
		content.WriteString(noData)
	} else {
		table := m.renderHoldingsTable()
		content.WriteString(table)
	}

	content.WriteString("\n")

	// Portfolio summary
	summary := m.renderPortfolioSummary()
	content.WriteString(summary)

	content.WriteString("\n")

	// Help text
	helpText := m.renderHelpText()
	content.WriteString(helpText)

	return content.String()
}

// renderHoldingsTable renders the holdings table
func (m *PortfolioModel) renderHoldingsTable() string {
	headers := []string{"Symbol", "Shares", "WAC", "Current", "Market Value", "Gain/Loss", "Gain%", "Status"}
	colWidths := []int{8, 8, 12, 12, 15, 15, 8, 8}

	// Render headers
	var headerCells []string
	for i, header := range headers {
		if int(m.sortMode) == i || (m.sortMode == SortByGainLoss && i == 5) {
			// Highlight current sort column
			cell := TableHeaderStyle.
				Foreground(ColorAccent).
				Width(colWidths[i]).
				Render(header + " ↓")
			headerCells = append(headerCells, cell)
		} else {
			cell := TableHeaderStyle.Width(colWidths[i]).Render(header)
			headerCells = append(headerCells, cell)
		}
	}
	headerRow := lipgloss.JoinHorizontal(lipgloss.Left, headerCells...)

	// Render data rows
	var rows []string
	rows = append(rows, headerRow)

	for i, holding := range m.holdings {
		// Use values directly from the holding struct
		gainLoss := holding.UnrealizedGainLoss
		gainPercent := holding.UnrealizedGainPct

		// Style based on selection
		rowStyle := StyleForTableRow(i, i == m.selectedIndex)

		cells := []string{
			rowStyle.Width(colWidths[0]).Render(holding.StockSymbol),
			rowStyle.Width(colWidths[1]).Align(lipgloss.Right).Render(strconv.Itoa(holding.TotalShares)),
			rowStyle.Width(colWidths[2]).Align(lipgloss.Right).Render(holding.WeightedAvgCost.String()),
			rowStyle.Width(colWidths[3]).Align(lipgloss.Right).Render(holding.CurrentPrice.String()),
			rowStyle.Width(colWidths[4]).Align(lipgloss.Right).Render(holding.MarketValue.String()),
		}

		// Style gain/loss based on value
		gainLossStyle := StyleForMoney(!gainLoss.IsNegative(), gainLoss.IsZero())
		if i == m.selectedIndex {
			gainLossStyle = TableSelectedRowStyle
		}
		cells = append(cells, gainLossStyle.Width(colWidths[5]).Align(lipgloss.Right).Render(gainLoss.String()))

		// Style percentage
		percentStyle := StyleForPercentage(gainPercent)
		if i == m.selectedIndex {
			percentStyle = TableSelectedRowStyle
		}
		percentText := fmt.Sprintf("%.1f%%", gainPercent)
		cells = append(cells, percentStyle.Width(colWidths[6]).Align(lipgloss.Right).Render(percentText))

		// Status (show warning for default prices)
		status := "✓"
		statusStyle := rowStyle.Foreground(ColorSuccess)
		// TODO: Check if using default price from CSV import
		// if holding.HasDefaultPrice {
		//     status = "⚠"
		//     statusStyle = rowStyle.Foreground(ColorWarning)
		// }
		cells = append(cells, statusStyle.Width(colWidths[7]).Align(lipgloss.Center).Render(status))

		row := lipgloss.JoinHorizontal(lipgloss.Left, cells...)
		rows = append(rows, row)
	}

	table := strings.Join(rows, "\n")
	return PanelStyle.Width(m.windowSize.Width - 4).Render(table)
}

// renderPortfolioSummary renders the portfolio summary at the bottom
func (m *PortfolioModel) renderPortfolioSummary() string {
	currentPrices := make(map[string]domain.Money)
	summary := m.portfolio.GetPortfolioSummary(currentPrices)

	// Calculate total gain/loss
	totalGainLoss := summary.TotalMarketValue.Subtract(summary.TotalInvested)
	totalGainPercent := 0.0
	if !summary.TotalInvested.IsZero() {
		totalGainPercent = float64(totalGainLoss.Paisa()) / float64(summary.TotalInvested.Paisa()) * 100
	}

	summaryItems := []string{
		fmt.Sprintf("Holdings: %s", InfoStyle.Render(fmt.Sprintf("%d stocks", len(m.holdings)))),
		fmt.Sprintf("Total Invested: %s", MoneyStyle.Render(summary.TotalInvested.String())),
		fmt.Sprintf("Market Value: %s", MoneyStyle.Render(summary.TotalMarketValue.String())),
		fmt.Sprintf("Total P&L: %s", StyleForMoney(!totalGainLoss.IsNegative(), totalGainLoss.IsZero()).Render(totalGainLoss.String())),
		fmt.Sprintf("Total Gain: %s", StyleForPercentage(totalGainPercent).Render(fmt.Sprintf("%.1f%%", totalGainPercent))),
		fmt.Sprintf("Realized P&L: %s", StyleForMoney(!summary.TotalRealizedPL.IsNegative(), summary.TotalRealizedPL.IsZero()).Render(summary.TotalRealizedPL.String())),
	}

	summaryText := strings.Join(summaryItems, "  |  ")
	return PanelStyle.
		Width(m.windowSize.Width - 4).
		BorderForeground(ColorSuccess).
		Render(summaryText)
}

// renderDetailView renders the detailed view for a selected stock
func (m *PortfolioModel) renderDetailView() string {
	// Find the holding for the selected symbol
	var selectedHolding *domain.Holding
	for _, holding := range m.holdings {
		if holding.StockSymbol == m.detailSymbol {
			selectedHolding = &holding
			break
		}
	}

	if selectedHolding == nil {
		return PanelStyle.Render("Stock not found")
	}

	var content strings.Builder

	// Title
	title := SectionTitleStyle.Render(fmt.Sprintf("📈 %s - Stock Detail", selectedHolding.StockSymbol))
	content.WriteString(title)
	content.WriteString("\n\n")

	// Stock details
	details := []string{
		fmt.Sprintf("Total Shares: %s", MoneyStyle.Render(strconv.Itoa(selectedHolding.TotalShares))),
		fmt.Sprintf("Weighted Average Cost: %s", MoneyStyle.Render(selectedHolding.WeightedAvgCost.String())),
		fmt.Sprintf("Total Investment: %s", MoneyStyle.Render(selectedHolding.TotalCost.String())),
		"",
		"Recent Transactions:",
		// TODO: Show recent transactions for this symbol
		MutedStyle.Render("(Transaction history will be shown here)"),
		"",
		"Lot Information:",
		// TODO: Show lot details from FIFO queue
		MutedStyle.Render("(Lot breakdown will be shown here)"),
	}

	detailText := strings.Join(details, "\n")
	detailPanel := PanelStyle.
		Width(m.windowSize.Width - 4).
		Height(m.windowSize.Height - 10).
		Render(detailText)

	content.WriteString(detailPanel)
	content.WriteString("\n")

	// Help text
	helpText := HelpStyle.Render("Press ESC or q to return to portfolio view")
	content.WriteString(helpText)

	return content.String()
}

// renderHelpText renders help text for the portfolio view
func (m *PortfolioModel) renderHelpText() string {
	helpItems := []string{
		KeybindStyle.Render("↑/↓") + " navigate",
		KeybindStyle.Render("Enter") + " details",
		KeybindStyle.Render("s") + " sort (" + m.sortMode.String() + ")",
		KeybindStyle.Render("i") + " import",
		KeybindStyle.Render("e") + " edit price",
	}

	return HelpStyle.Render(strings.Join(helpItems, " | "))
}

// refreshHoldings updates the holdings list from the portfolio
func (m *PortfolioModel) refreshHoldings() {
	currentPrices := make(map[string]domain.Money)
	m.holdings = m.portfolio.GetActiveHoldings(currentPrices)
	m.sortHoldings()

	// Adjust selected index if needed
	if m.selectedIndex >= len(m.holdings) {
		m.selectedIndex = len(m.holdings) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

// sortHoldings sorts the holdings based on the current sort mode
func (m *PortfolioModel) sortHoldings() {
	switch m.sortMode {
	case SortBySymbol:
		sort.Slice(m.holdings, func(i, j int) bool {
			return m.holdings[i].StockSymbol < m.holdings[j].StockSymbol
		})
	case SortByGainPercent:
		sort.Slice(m.holdings, func(i, j int) bool {
			return m.holdings[i].UnrealizedGainPct > m.holdings[j].UnrealizedGainPct
		})
	case SortByMarketValue:
		sort.Slice(m.holdings, func(i, j int) bool {
			return m.holdings[i].MarketValue.Paisa() > m.holdings[j].MarketValue.Paisa()
		})
	case SortByGainLoss:
		sort.Slice(m.holdings, func(i, j int) bool {
			return m.holdings[i].UnrealizedGainLoss.Paisa() > m.holdings[j].UnrealizedGainLoss.Paisa()
		})
	}
}