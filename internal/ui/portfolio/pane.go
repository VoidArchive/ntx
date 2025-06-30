package portfolio

import (
	"fmt"
	"ntx/internal/ui/common"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PortfolioPane displays portfolio holdings, performance, and P&L
type PortfolioPane struct {
	common.BasePane

	// Portfolio data
	TotalValue float64
	DayChange  float64
	TotalGain  float64
	Holdings   []common.PortfolioHolding

	// Display preferences
	sortBy        SortOption
	showDetails   bool
	selectedIndex int
	showSummary   bool

	// Performance tracking
	PreviousValue float64
	HighWaterMark float64
	MaxDrawdown   float64
}

// SortOption defines how to sort portfolio holdings
type SortOption string

const (
	SortBySymbol      SortOption = "symbol"
	SortByValue       SortOption = "value"
	SortByGain        SortOption = "gain"
	SortByGainPercent SortOption = "gain_percent"
	SortByQuantity    SortOption = "quantity"
)

// NewPortfolioPane creates a new portfolio pane
func NewPortfolioPane() *PortfolioPane {
	return &PortfolioPane{
		BasePane: common.BasePane{
			Type:  common.PaneTypePortfolio,
			Title: "Portfolio",
		},
		Holdings:      make([]common.PortfolioHolding, 0),
		sortBy:        SortByValue,
		showDetails:   true,
		showSummary:   true,
		selectedIndex: 0,
	}
}

// Init initializes the portfolio pane
func (pp *PortfolioPane) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the pane
func (pp *PortfolioPane) Update(msg tea.Msg) (common.Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case common.PortfolioUpdateMsg:
		pp.updatePortfolioData(msg)
		pp.MarkUpdated()

	case common.MarketDataMsg:
		pp.updateHoldingPrice(msg)
		pp.MarkUpdated()

	case tea.KeyMsg:
		if pp.Active {
			return pp, pp.HandleKeypress(msg.String())
		}
	}

	return pp, nil
}

// View renders the portfolio pane
func (pp *PortfolioPane) View() string {
	style := pp.GetStyle()
	content := pp.renderContent()

	return style.Render(content)
}

// HandleKeypress handles pane-specific key presses
func (pp *PortfolioPane) HandleKeypress(key string) tea.Cmd {
	switch key {
	case "j", "down":
		if pp.selectedIndex < len(pp.Holdings)-1 {
			pp.selectedIndex++
		}

	case "k", "up":
		if pp.selectedIndex > 0 {
			pp.selectedIndex--
		}

	case "d":
		pp.showDetails = !pp.showDetails

	case "s":
		pp.showSummary = !pp.showSummary

	case "1":
		pp.sortBy = SortBySymbol
		pp.sortHoldings()

	case "2":
		pp.sortBy = SortByValue
		pp.sortHoldings()

	case "3":
		pp.sortBy = SortByGain
		pp.sortHoldings()

	case "4":
		pp.sortBy = SortByGainPercent
		pp.sortHoldings()
	}

	return nil
}

// Refresh triggers a refresh of portfolio data
func (pp *PortfolioPane) Refresh() tea.Cmd {
	return common.NewRefreshMsg(false)
}

// renderContent renders the pane content
func (pp *PortfolioPane) renderContent() string {
	var sections []string

	// Title
	sections = append(sections, pp.RenderTitle())

	// Portfolio summary
	if pp.showSummary {
		sections = append(sections, pp.renderSummary())
	}

	// Holdings table
	sections = append(sections, pp.renderHoldings())

	// Selected holding details
	if pp.showDetails && len(pp.Holdings) > 0 && pp.selectedIndex < len(pp.Holdings) {
		sections = append(sections, pp.renderHoldingDetails())
	}

	// Help section
	sections = append(sections, pp.renderHelp())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderSummary renders portfolio summary statistics
func (pp *PortfolioPane) renderSummary() string {
	var lines []string

	// Section header
	header := common.SubHeaderStyle.Render(" Portfolio Summary ")
	lines = append(lines, header)

	// Total value
	totalValueLine := fmt.Sprintf("Total Value: %s",
		common.InfoStyle.Render(common.FormatMoney(pp.TotalValue, false)))
	lines = append(lines, totalValueLine)

	// Day's change
	dayChangeLine := fmt.Sprintf("Day's Change: %s",
		common.FormatMoney(pp.DayChange, true))
	lines = append(lines, dayChangeLine)

	// Total gain/loss
	totalGainLine := fmt.Sprintf("Total Gain: %s",
		common.FormatMoney(pp.TotalGain, true))
	lines = append(lines, totalGainLine)

	// Performance metrics
	if pp.TotalValue > 0 && pp.PreviousValue > 0 {
		dayPercent := (pp.DayChange / pp.PreviousValue) * 100
		dayPercentLine := fmt.Sprintf("Day's %%: %s",
			common.FormatPercentage(dayPercent))
		lines = append(lines, dayPercentLine)
	}

	// Holdings count
	holdingsCount := fmt.Sprintf("Holdings: %s",
		common.InfoStyle.Render(fmt.Sprintf("%d", len(pp.Holdings))))
	lines = append(lines, holdingsCount)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderHoldings renders the holdings table
func (pp *PortfolioPane) renderHoldings() string {
	if len(pp.Holdings) == 0 {
		return common.MutedStyle.Render("No holdings in portfolio")
	}

	var lines []string

	// Table header
	headerLine := pp.renderTableHeader()
	lines = append(lines, headerLine)

	// Holdings rows
	for i, holding := range pp.Holdings {
		isSelected := i == pp.selectedIndex && pp.Active
		holdingLine := pp.renderHoldingRow(holding, isSelected)
		lines = append(lines, holdingLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderTableHeader renders the holdings table header
func (pp *PortfolioPane) renderTableHeader() string {
	headers := []string{
		padCenter("Symbol", 8),
		padCenter("Qty", 8),
		padCenter("Avg Price", 10),
		padCenter("LTP", 8),
		padCenter("Value", 12),
		padCenter("P&L", 10),
		padCenter("P&L%", 8),
	}

	headerRow := strings.Join(headers, " ")
	return common.TableHeaderStyle.Render(headerRow)
}

// renderHoldingRow renders a single holding row
func (pp *PortfolioPane) renderHoldingRow(holding common.PortfolioHolding, isSelected bool) string {
	style := common.TableCellStyle
	if isSelected {
		style = common.TableSelectedStyle
	}

	cells := []string{
		padCenter(holding.Symbol, 8),
		padCenter(fmt.Sprintf("%d", holding.Quantity), 8),
		padCenter(fmt.Sprintf("%.2f", holding.AvgPrice), 10),
		padCenter(fmt.Sprintf("%.2f", holding.CurrentPrice), 8),
		padCenter(fmt.Sprintf("%.2f", holding.Value), 12),
		padCenter(fmt.Sprintf("%.2f", holding.Gain), 10),
		padCenter(fmt.Sprintf("%.2f%%", holding.GainPercent), 8),
	}

	row := strings.Join(cells, " ")
	return style.Render(row)
}

// renderHoldingDetails renders detailed information about the selected holding
func (pp *PortfolioPane) renderHoldingDetails() string {
	holding := pp.Holdings[pp.selectedIndex]

	var lines []string

	// Section header
	header := common.SubHeaderStyle.Render(fmt.Sprintf(" %s Details ", holding.Symbol))
	lines = append(lines, header)

	// Basic info
	qtyLine := fmt.Sprintf("Quantity: %s",
		common.InfoStyle.Render(fmt.Sprintf("%d shares", holding.Quantity)))
	lines = append(lines, qtyLine)

	avgPriceLine := fmt.Sprintf("Average Price: %s",
		common.InfoStyle.Render(fmt.Sprintf("%.2f", holding.AvgPrice)))
	lines = append(lines, avgPriceLine)

	currentPriceLine := fmt.Sprintf("Current Price: %s",
		common.InfoStyle.Render(fmt.Sprintf("%.2f", holding.CurrentPrice)))
	lines = append(lines, currentPriceLine)

	// Investment metrics
	invested := holding.AvgPrice * float64(holding.Quantity)
	investedLine := fmt.Sprintf("Invested: %s",
		common.InfoStyle.Render(fmt.Sprintf("%.2f", invested)))
	lines = append(lines, investedLine)

	currentValueLine := fmt.Sprintf("Current Value: %s",
		common.InfoStyle.Render(fmt.Sprintf("%.2f", holding.Value)))
	lines = append(lines, currentValueLine)

	// Gain/Loss
	gainLine := fmt.Sprintf("Unrealized P&L: %s (%s)",
		common.FormatMoney(holding.Gain, true),
		common.FormatPercentage(holding.GainPercent))
	lines = append(lines, gainLine)

	// Portfolio allocation
	if pp.TotalValue > 0 {
		allocation := (holding.Value / pp.TotalValue) * 100
		allocationLine := fmt.Sprintf("Portfolio Weight: %s",
			common.InfoStyle.Render(fmt.Sprintf("%.2f%%", allocation)))
		lines = append(lines, allocationLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderHelp renders pane-specific help
func (pp *PortfolioPane) renderHelp() string {
	if !pp.Active {
		return ""
	}

	help := []string{
		common.KeyStyle.Render("↑↓") + ": Select holding",
		common.KeyStyle.Render("d") + ": Toggle details",
		common.KeyStyle.Render("s") + ": Toggle summary",
		common.KeyStyle.Render("1-4") + ": Sort options",
	}

	return common.HelpStyle.Render(strings.Join(help, " • "))
}

// Helper methods

// updatePortfolioData updates portfolio data from message
func (pp *PortfolioPane) updatePortfolioData(msg common.PortfolioUpdateMsg) {
	pp.PreviousValue = pp.TotalValue
	pp.TotalValue = msg.TotalValue
	pp.DayChange = msg.DayChange
	pp.TotalGain = msg.TotalGain
	pp.Holdings = msg.Holdings

	// Update performance tracking
	if pp.TotalValue > pp.HighWaterMark {
		pp.HighWaterMark = pp.TotalValue
	}

	if pp.HighWaterMark > 0 {
		drawdown := (pp.HighWaterMark - pp.TotalValue) / pp.HighWaterMark * 100
		if drawdown > pp.MaxDrawdown {
			pp.MaxDrawdown = drawdown
		}
	}

	pp.sortHoldings()
}

// updateHoldingPrice updates a specific holding's current price
func (pp *PortfolioPane) updateHoldingPrice(msg common.MarketDataMsg) {
	for i := range pp.Holdings {
		if pp.Holdings[i].Symbol == msg.Symbol {
			pp.Holdings[i].CurrentPrice = msg.LastPrice
			pp.Holdings[i].Value = float64(pp.Holdings[i].Quantity) * msg.LastPrice
			pp.Holdings[i].Gain = pp.Holdings[i].Value - (pp.Holdings[i].AvgPrice * float64(pp.Holdings[i].Quantity))

			if pp.Holdings[i].AvgPrice > 0 && pp.Holdings[i].Quantity > 0 {
				costBasis := pp.Holdings[i].AvgPrice * float64(pp.Holdings[i].Quantity)
				if costBasis > 0 {
					pp.Holdings[i].GainPercent = (pp.Holdings[i].Gain / costBasis) * 100
				}
			}
			break
		}
	}

	// Recalculate total value
	pp.recalculateTotalValue()
}

// recalculateTotalValue recalculates the total portfolio value
func (pp *PortfolioPane) recalculateTotalValue() {
	total := 0.0
	for _, holding := range pp.Holdings {
		total += holding.Value
	}

	dayChange := total - pp.PreviousValue
	totalGain := 0.0
	for _, holding := range pp.Holdings {
		totalGain += holding.Gain
	}

	pp.TotalValue = total
	pp.DayChange = dayChange
	pp.TotalGain = totalGain
}

// sortHoldings sorts the holdings based on the selected sort option
func (pp *PortfolioPane) sortHoldings() {
	sort.SliceStable(pp.Holdings, func(i, j int) bool {
		switch pp.sortBy {
		case SortBySymbol:
			return pp.Holdings[i].Symbol < pp.Holdings[j].Symbol
		case SortByValue:
			return pp.Holdings[i].Value > pp.Holdings[j].Value
		case SortByGain:
			return pp.Holdings[i].Gain > pp.Holdings[j].Gain
		case SortByGainPercent:
			return pp.Holdings[i].GainPercent > pp.Holdings[j].GainPercent
		case SortByQuantity:
			return pp.Holdings[i].Quantity > pp.Holdings[j].Quantity
		default:
			return false
		}
	})
}

// Utility functions

// padCenter pads a string to center it within the given width
func padCenter(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}

	padding := width - len(s)
	leftPad := padding / 2
	rightPad := padding - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}
