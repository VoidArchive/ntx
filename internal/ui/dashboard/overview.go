package dashboard

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ntx/internal/ui/common"
)

// OverviewPane displays a comprehensive market overview and key metrics
type OverviewPane struct {
	common.BasePane

	// Market summary data
	MarketIndices map[string]MarketIndex
	TopGainers    []StockSummary
	TopLosers     []StockSummary
	MostActive    []StockSummary

	// Statistics
	TotalVolume    int64
	TotalTurnover  float64
	AdvancingCount int
	DecliningCount int
	UnchangedCount int

	// Display preferences
	showIndices   bool
	showTopMovers bool
	showStats     bool
	refreshRate   time.Duration
}

// MarketIndex represents a market index
type MarketIndex struct {
	Name          string
	Value         float64
	Change        float64
	ChangePercent float64
	LastUpdate    time.Time
}

// StockSummary represents a stock with key metrics
type StockSummary struct {
	Symbol        string
	LastPrice     float64
	Change        float64
	ChangePercent float64
	Volume        int64
	Turnover      float64
}

// NewOverviewPane creates a new overview pane
func NewOverviewPane() *OverviewPane {
	return &OverviewPane{
		BasePane: common.BasePane{
			Type:  common.PaneTypeDashboard,
			Title: "Market Overview",
		},
		MarketIndices: make(map[string]MarketIndex),
		showIndices:   true,
		showTopMovers: true,
		showStats:     true,
		refreshRate:   30 * time.Second,
	}
}

// Init initializes the overview pane
func (op *OverviewPane) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the pane
func (op *OverviewPane) Update(msg tea.Msg) (common.Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case common.MarketDataMsg:
		op.updateStockData(msg)
		op.MarkUpdated()

	case MarketOverviewMsg:
		op.updateOverviewData(msg)
		op.MarkUpdated()

	case tea.KeyMsg:
		if op.Active {
			return op, op.HandleKeypress(msg.String())
		}
	}

	return op, nil
}

// View renders the overview pane
func (op *OverviewPane) View() string {
	style := op.GetStyle()
	content := op.renderContent()

	return style.Render(content)
}

// HandleKeypress handles pane-specific key presses
func (op *OverviewPane) HandleKeypress(key string) tea.Cmd {
	switch key {
	case "i":
		op.showIndices = !op.showIndices
		return nil
	case "m":
		op.showTopMovers = !op.showTopMovers
		return nil
	case "s":
		op.showStats = !op.showStats
		return nil
	}
	return nil
}

// Refresh triggers a refresh of overview data
func (op *OverviewPane) Refresh() tea.Cmd {
	return common.NewRefreshMsg(false)
}

// renderContent renders the pane content
func (op *OverviewPane) renderContent() string {
	var sections []string

	// Title
	sections = append(sections, op.RenderTitle())

	// Market indices section
	if op.showIndices {
		sections = append(sections, op.renderMarketIndices())
	}

	// Market statistics
	if op.showStats {
		sections = append(sections, op.renderMarketStats())
	}

	// Top movers section
	if op.showTopMovers {
		sections = append(sections, op.renderTopMovers())
	}

	// Help section
	sections = append(sections, op.renderHelp())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderMarketIndices renders major market indices
func (op *OverviewPane) renderMarketIndices() string {
	if len(op.MarketIndices) == 0 {
		return common.MutedStyle.Render("No index data available")
	}

	var lines []string

	// Section header
	header := common.SubHeaderStyle.Render(" Market Indices ")
	lines = append(lines, header)

	// Render each index
	for name, index := range op.MarketIndices {
		indexLine := op.formatIndexLine(name, index)
		lines = append(lines, indexLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderMarketStats renders overall market statistics
func (op *OverviewPane) renderMarketStats() string {
	var lines []string

	// Section header
	header := common.SubHeaderStyle.Render(" Market Statistics ")
	lines = append(lines, header)

	// Total volume and turnover
	if op.TotalVolume > 0 {
		volumeLine := fmt.Sprintf("Volume: %s",
			common.InfoStyle.Render(formatLargeNumber(op.TotalVolume)))
		lines = append(lines, volumeLine)
	}

	if op.TotalTurnover > 0 {
		turnoverLine := fmt.Sprintf("Turnover: %s",
			common.InfoStyle.Render(common.FormatMoney(op.TotalTurnover, false)))
		lines = append(lines, turnoverLine)
	}

	// Advance/decline statistics
	total := op.AdvancingCount + op.DecliningCount + op.UnchangedCount
	if total > 0 {
		advanceLine := fmt.Sprintf("Advancing: %s",
			common.GainStyle.Render(fmt.Sprintf("%d", op.AdvancingCount)))
		decliningLine := fmt.Sprintf("Declining: %s",
			common.LossStyle.Render(fmt.Sprintf("%d", op.DecliningCount)))
		unchangedLine := fmt.Sprintf("Unchanged: %s",
			common.NeutralStyle.Render(fmt.Sprintf("%d", op.UnchangedCount)))

		adRatio := float64(op.AdvancingCount) / float64(op.DecliningCount)
		ratioStyle := common.NeutralStyle
		if adRatio > 1.5 {
			ratioStyle = common.GainStyle
		} else if adRatio < 0.67 {
			ratioStyle = common.LossStyle
		}

		ratioLine := fmt.Sprintf("A/D Ratio: %s",
			ratioStyle.Render(fmt.Sprintf("%.2f", adRatio)))

		lines = append(lines, advanceLine, decliningLine, unchangedLine, ratioLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderTopMovers renders top gainers, losers, and most active stocks
func (op *OverviewPane) renderTopMovers() string {
	var sections []string

	// Top gainers
	if len(op.TopGainers) > 0 {
		gainersHeader := common.GainStyle.Render(" Top Gainers ")
		sections = append(sections, gainersHeader)

		for i, stock := range op.TopGainers {
			if i >= 3 { // Show top 3
				break
			}
			stockLine := op.formatStockLine(stock)
			sections = append(sections, stockLine)
		}
	}

	// Top losers
	if len(op.TopLosers) > 0 {
		losersHeader := common.LossStyle.Render(" Top Losers ")
		sections = append(sections, losersHeader)

		for i, stock := range op.TopLosers {
			if i >= 3 { // Show top 3
				break
			}
			stockLine := op.formatStockLine(stock)
			sections = append(sections, stockLine)
		}
	}

	// Most active
	if len(op.MostActive) > 0 {
		activeHeader := common.InfoStyle.Render(" Most Active ")
		sections = append(sections, activeHeader)

		for i, stock := range op.MostActive {
			if i >= 3 { // Show top 3
				break
			}
			stockLine := op.formatStockLine(stock)
			sections = append(sections, stockLine)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHelp renders pane-specific help
func (op *OverviewPane) renderHelp() string {
	if !op.Active {
		return ""
	}

	help := []string{
		common.KeyStyle.Render("i") + ": Toggle indices",
		common.KeyStyle.Render("m") + ": Toggle movers",
		common.KeyStyle.Render("s") + ": Toggle stats",
	}

	return common.HelpStyle.Render(strings.Join(help, " • "))
}

// Helper methods

// formatIndexLine formats a market index display line
func (op *OverviewPane) formatIndexLine(name string, index MarketIndex) string {
	nameStyle := common.ContentStyle.Render(padRight(name, 12))
	valueStyle := common.InfoStyle.Render(fmt.Sprintf("%.2f", index.Value))
	changeStyle := common.FormatMoney(index.Change, true)
	percentStyle := common.FormatPercentage(index.ChangePercent)

	return fmt.Sprintf("%s %s %s %s", nameStyle, valueStyle, changeStyle, percentStyle)
}

// formatStockLine formats a stock summary display line
func (op *OverviewPane) formatStockLine(stock StockSummary) string {
	symbolStyle := common.ContentStyle.Render(padRight(stock.Symbol, 8))
	priceStyle := common.InfoStyle.Render(fmt.Sprintf("%.2f", stock.LastPrice))
	changeStyle := common.FormatMoney(stock.Change, true)
	percentStyle := common.FormatPercentage(stock.ChangePercent)

	return fmt.Sprintf("%s %s %s %s", symbolStyle, priceStyle, changeStyle, percentStyle)
}

// updateStockData updates individual stock data
func (op *OverviewPane) updateStockData(msg common.MarketDataMsg) {
	// Update top movers lists based on the new data
	// This would normally integrate with a more comprehensive data source

	stock := StockSummary{
		Symbol:        msg.Symbol,
		LastPrice:     msg.LastPrice,
		Change:        msg.Change,
		ChangePercent: (msg.Change / (msg.LastPrice - msg.Change)) * 100,
		Volume:        msg.Volume,
	}

	// Simple logic to update top movers (in practice, this would be more sophisticated)
	if stock.ChangePercent > 5 {
		op.addToTopGainers(stock)
	} else if stock.ChangePercent < -5 {
		op.addToTopLosers(stock)
	}

	if stock.Volume > 100000 {
		op.addToMostActive(stock)
	}
}

// updateOverviewData updates comprehensive overview data
func (op *OverviewPane) updateOverviewData(msg MarketOverviewMsg) {
	op.MarketIndices = msg.Indices
	op.TopGainers = msg.TopGainers
	op.TopLosers = msg.TopLosers
	op.MostActive = msg.MostActive
	op.TotalVolume = msg.TotalVolume
	op.TotalTurnover = msg.TotalTurnover
	op.AdvancingCount = msg.AdvancingCount
	op.DecliningCount = msg.DecliningCount
	op.UnchangedCount = msg.UnchangedCount
}

// addToTopGainers adds a stock to top gainers list
func (op *OverviewPane) addToTopGainers(stock StockSummary) {
	// Insert and maintain sorted order (top 5)
	inserted := false
	for i, existing := range op.TopGainers {
		if stock.ChangePercent > existing.ChangePercent {
			// Insert here
			op.TopGainers = append(op.TopGainers[:i], append([]StockSummary{stock}, op.TopGainers[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		op.TopGainers = append(op.TopGainers, stock)
	}

	// Keep only top 5
	if len(op.TopGainers) > 5 {
		op.TopGainers = op.TopGainers[:5]
	}
}

// addToTopLosers adds a stock to top losers list
func (op *OverviewPane) addToTopLosers(stock StockSummary) {
	// Insert and maintain sorted order (lowest first)
	inserted := false
	for i, existing := range op.TopLosers {
		if stock.ChangePercent < existing.ChangePercent {
			// Insert here
			op.TopLosers = append(op.TopLosers[:i], append([]StockSummary{stock}, op.TopLosers[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		op.TopLosers = append(op.TopLosers, stock)
	}

	// Keep only top 5
	if len(op.TopLosers) > 5 {
		op.TopLosers = op.TopLosers[:5]
	}
}

// addToMostActive adds a stock to most active list
func (op *OverviewPane) addToMostActive(stock StockSummary) {
	// Insert and maintain sorted order by volume
	inserted := false
	for i, existing := range op.MostActive {
		if stock.Volume > existing.Volume {
			// Insert here
			op.MostActive = append(op.MostActive[:i], append([]StockSummary{stock}, op.MostActive[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		op.MostActive = append(op.MostActive, stock)
	}

	// Keep only top 5
	if len(op.MostActive) > 5 {
		op.MostActive = op.MostActive[:5]
	}
}

// Utility functions

// padRight pads a string to the right with spaces
func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

// formatLargeNumber formats large numbers with K/M/B suffixes
func formatLargeNumber(num int64) string {
	if num >= 1000000000 {
		return fmt.Sprintf("%.2fB", float64(num)/1000000000)
	} else if num >= 1000000 {
		return fmt.Sprintf("%.2fM", float64(num)/1000000)
	} else if num >= 1000 {
		return fmt.Sprintf("%.2fK", float64(num)/1000)
	}
	return fmt.Sprintf("%d", num)
}

// MarketOverviewMsg carries comprehensive market overview data
type MarketOverviewMsg struct {
	Indices        map[string]MarketIndex
	TopGainers     []StockSummary
	TopLosers      []StockSummary
	MostActive     []StockSummary
	TotalVolume    int64
	TotalTurnover  float64
	AdvancingCount int
	DecliningCount int
	UnchangedCount int
	LastUpdate     time.Time
}
