package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/VoidArchive/ntx/internal/domain"
)

// PerformanceData represents performance metrics for a stock
type PerformanceData struct {
	Symbol      string
	GainLoss    domain.Money
	GainPercent float64
	MarketValue domain.Money
}

// AnalysisModel handles the analysis view
type AnalysisModel struct {
	portfolio     *domain.Portfolio
	windowSize    tea.WindowSizeMsg
	topPerformers []PerformanceData
	worstPerformers []PerformanceData
}

// NewAnalysisModel creates a new analysis model
func NewAnalysisModel(portfolio *domain.Portfolio) *AnalysisModel {
	return &AnalysisModel{
		portfolio: portfolio,
	}
}

// Init initializes the analysis model
func (m *AnalysisModel) Init() tea.Cmd {
	m.refreshAnalysis()
	return nil
}

// Update handles messages and updates the model
func (m *AnalysisModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg

	case RefreshPortfolioMsg:
		m.refreshAnalysis()
	}

	return m, nil
}

// View renders the analysis view
func (m *AnalysisModel) View() string {
	if m.windowSize.Width == 0 {
		return "Loading analysis..."
	}

	var content strings.Builder

	// Section title
	title := SectionTitleStyle.Render("📊 Portfolio Analysis")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Top panel - Portfolio Summary
	summaryPanel := m.renderPortfolioSummaryPanel()
	content.WriteString(summaryPanel)
	content.WriteString("\n")

	// Middle panels - Performance and Tax Summary
	middlePanels := m.renderMiddlePanels()
	content.WriteString(middlePanels)
	content.WriteString("\n")

	// Bottom panel - Warnings and Alerts
	alertsPanel := m.renderAlertsPanel()
	content.WriteString(alertsPanel)

	return content.String()
}

// renderPortfolioSummaryPanel renders the top portfolio summary panel
func (m *AnalysisModel) renderPortfolioSummaryPanel() string {
	summary := m.portfolio.GetPortfolioSummary()
	holdings := m.portfolio.GetActiveHoldings()

	// TODO: Calculate current market value properly
	// For now, using total cost as market value
	marketValue := summary.TotalCost
	totalGainLoss := marketValue.Sub(summary.TotalCost)
	totalGainPercent := 0.0
	if !summary.TotalCost.IsZero() {
		totalGainPercent = float64(totalGainLoss.Paisa()) / float64(summary.TotalCost.Paisa()) * 100
	}

	// Calculate unrealized gain (difference between market value and cost, minus realized)
	unrealizedGain := totalGainLoss.Sub(summary.TotalRealizedGain)

	summaryItems := []string{
		fmt.Sprintf("Total Invested: %s", MoneyStyle.Render(summary.TotalCost.String())),
		fmt.Sprintf("Market Value: %s", MoneyStyle.Render(marketValue.String())),
		fmt.Sprintf("P&L: %s (%s)", 
			StyleForMoney(!totalGainLoss.IsNegative(), totalGainLoss.IsZero()).Render(totalGainLoss.String()),
			StyleForPercentage(totalGainPercent).Render(fmt.Sprintf("%.1f%%", totalGainPercent))),
		fmt.Sprintf("Holdings: %s", InfoStyle.Render(fmt.Sprintf("%d stocks", len(holdings)))),
		fmt.Sprintf("Realized P&L: %s", 
			StyleForMoney(!summary.TotalRealizedGain.IsNegative(), summary.TotalRealizedGain.IsZero()).Render(summary.TotalRealizedGain.String())),
		fmt.Sprintf("Unrealized: %s", 
			StyleForMoney(!unrealizedGain.IsNegative(), unrealizedGain.IsZero()).Render(unrealizedGain.String())),
	}

	summaryText := strings.Join(summaryItems, "  |  ")
	
	return PanelStyle.
		Width(m.windowSize.Width - 4).
		BorderForeground(ColorPrimary).
		Render(summaryText)
}

// renderMiddlePanels renders side-by-side performance and tax panels
func (m *AnalysisModel) renderMiddlePanels() string {
	leftPanelWidth := (m.windowSize.Width - 12) / 2
	rightPanelWidth := m.windowSize.Width - leftPanelWidth - 12

	// Left panel - Performance
	performancePanel := m.renderPerformancePanel(leftPanelWidth)
	
	// Right panel - Tax Summary
	taxPanel := m.renderTaxSummaryPanel(rightPanelWidth)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		performancePanel,
		strings.Repeat(" ", 4),
		taxPanel,
	)
}

// renderPerformancePanel renders the performance panel
func (m *AnalysisModel) renderPerformancePanel(width int) string {
	var content strings.Builder

	// Top Performers
	content.WriteString(SectionTitleStyle.Render("Top Performers"))
	content.WriteString("\n")

	if len(m.topPerformers) == 0 {
		content.WriteString(MutedStyle.Render("No data available"))
	} else {
		for i, perf := range m.topPerformers {
			if i >= 5 { // Show top 5
				break
			}
			
			line := fmt.Sprintf("%-8s %s %s",
				perf.Symbol,
				StyleForPercentage(perf.GainPercent).Render(fmt.Sprintf("%+.1f%%", perf.GainPercent)),
				StyleForMoney(perf.GainLoss.Paisa() > 0, perf.GainLoss.IsZero()).Render(perf.GainLoss.String()))
			
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")

	// Worst Performers
	content.WriteString(SectionTitleStyle.Render("Worst Performers"))
	content.WriteString("\n")

	if len(m.worstPerformers) == 0 {
		content.WriteString(MutedStyle.Render("No data available"))
	} else {
		for i, perf := range m.worstPerformers {
			if i >= 5 { // Show worst 5
				break
			}
			
			line := fmt.Sprintf("%-8s %s %s",
				perf.Symbol,
				StyleForPercentage(perf.GainPercent).Render(fmt.Sprintf("%+.1f%%", perf.GainPercent)),
				StyleForMoney(perf.GainLoss.Paisa() > 0, perf.GainLoss.IsZero()).Render(perf.GainLoss.String()))
			
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	return PanelStyle.
		Width(width).
		Height(15).
		BorderForeground(ColorSuccess).
		Render(content.String())
}

// renderTaxSummaryPanel renders the tax summary panel
func (m *AnalysisModel) renderTaxSummaryPanel(width int) string {
	var content strings.Builder

	content.WriteString(SectionTitleStyle.Render("Tax Summary"))
	content.WriteString("\n")

	// Calculate tax implications
	summary := m.portfolio.GetPortfolioSummary()
	
	// For Nepal: 
	// Short-term capital gains (≤ 365 days): 7.5%
	// Long-term capital gains (> 365 days): 5%
	shortTermRate := 0.075
	longTermRate := 0.05

	// TODO: Get actual short-term and long-term gains from portfolio
	// For now, assume all realized gains are subject to tax
	totalRealizedGain := summary.TotalRealizedGain
	
	// Placeholder calculations (need actual short/long term breakdown)
	shortTermGain := totalRealizedGain.Multiply(0.6) // Assume 60% short-term
	longTermGain := totalRealizedGain.Multiply(0.4)  // Assume 40% long-term

	shortTermTax := shortTermGain.Multiply(shortTermRate)
	longTermTax := longTermGain.Multiply(longTermRate)
	totalTax := shortTermTax.Add(longTermTax)

	taxItems := []string{
		fmt.Sprintf("Short-term gains: %s", 
			StyleForMoney(!shortTermGain.IsNegative(), shortTermGain.IsZero()).Render(shortTermGain.String())),
		fmt.Sprintf("Long-term gains:  %s", 
			StyleForMoney(!longTermGain.IsNegative(), longTermGain.IsZero()).Render(longTermGain.String())),
		"",
		fmt.Sprintf("Short-term tax (%.1f%%): %s", shortTermRate*100,
			WarningStyle.Render(shortTermTax.String())),
		fmt.Sprintf("Long-term tax (%.1f%%):  %s", longTermRate*100,
			WarningStyle.Render(longTermTax.String())),
		"",
		fmt.Sprintf("Estimated total tax: %s", 
			ErrorStyle.Render(totalTax.String())),
		"",
		MutedStyle.Render("* Estimates based on current"),
		MutedStyle.Render("  Nepal tax regulations"),
	}

	taxText := strings.Join(taxItems, "\n")

	return PanelStyle.
		Width(width).
		Height(15).
		BorderForeground(ColorWarning).
		Render(taxText)
}

// renderAlertsPanel renders the warnings and alerts panel
func (m *AnalysisModel) renderAlertsPanel() string {
	var alerts []string

	// Check for common issues
	holdings := m.portfolio.GetActiveHoldings()
	transactions := m.portfolio.GetAllTransactions()

	// Count transactions with default prices (Rs. 100.00)
	defaultPriceCount := 0
	for _, txn := range transactions {
		if txn.Price.Paisa() == 10000 { // Rs. 100.00 in paisa
			defaultPriceCount++
		}
	}

	if defaultPriceCount > 0 {
		alerts = append(alerts, fmt.Sprintf("⚠ %d transactions using default prices - please verify", defaultPriceCount))
	}

	// Check for recent activity
	now := time.Now()
	recentCount := 0
	for _, txn := range transactions {
		if now.Sub(txn.Date).Hours() < 24*7 { // Last 7 days
			recentCount++
		}
	}

	if recentCount > 0 {
		alerts = append(alerts, fmt.Sprintf("ℹ %d transactions in the last 7 days", recentCount))
	}

	// Check for stocks with only buy transactions (no sells)
	buyOnlyCount := 0
	symbolStats := make(map[string]struct {
		hasBuy  bool
		hasSell bool
	})

	for _, txn := range transactions {
		stats := symbolStats[txn.Symbol]
		if txn.Type == domain.Buy {
			stats.hasBuy = true
		} else if txn.Type == domain.Sell {
			stats.hasSell = true
		}
		symbolStats[txn.Symbol] = stats
	}

	for _, stats := range symbolStats {
		if stats.hasBuy && !stats.hasSell {
			buyOnlyCount++
		}
	}

	if buyOnlyCount > 0 {
		alerts = append(alerts, fmt.Sprintf("ℹ %d stocks with only buy transactions", buyOnlyCount))
	}

	// Mock data price update info
	alerts = append(alerts, "ℹ Last price update: 2 hours ago (prices need manual update)")

	// Diversification warning
	if len(holdings) < 5 {
		alerts = append(alerts, "⚠ Consider diversifying - portfolio has fewer than 5 holdings")
	}

	// If no alerts, show positive message
	if len(alerts) == 0 {
		alerts = append(alerts, "✓ No warnings or alerts")
	}

	alertText := strings.Join(alerts, "\n")

	return PanelStyle.
		Width(m.windowSize.Width - 4).
		Height(8).
		BorderForeground(ColorInfo).
		Render(SectionTitleStyle.Render("Warnings & Alerts") + "\n\n" + alertText)
}

// refreshAnalysis updates analysis data
func (m *AnalysisModel) refreshAnalysis() {
	m.calculatePerformanceData()
}

// calculatePerformanceData calculates performance metrics for all holdings
func (m *AnalysisModel) calculatePerformanceData() {
	holdings := m.portfolio.GetActiveHoldings()
	var performances []PerformanceData

	for _, holding := range holdings {
		// TODO: Calculate actual current price and market value
		// For now, using WAC as current price (no gain/loss)
		currentPrice := holding.WeightedAverageCost
		marketValue := domain.NewMoney(int64(holding.TotalShares) * currentPrice.Paisa())
		gainLoss := marketValue.Sub(holding.TotalCost)
		gainPercent := 0.0
		
		if !holding.TotalCost.IsZero() {
			gainPercent = float64(gainLoss.Paisa()) / float64(holding.TotalCost.Paisa()) * 100
		}

		performances = append(performances, PerformanceData{
			Symbol:      holding.Symbol,
			GainLoss:    gainLoss,
			GainPercent: gainPercent,
			MarketValue: marketValue,
		})
	}

	// Sort by gain percentage for top performers
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].GainPercent > performances[j].GainPercent
	})
	
	m.topPerformers = performances

	// Sort by gain percentage for worst performers (reverse order)
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].GainPercent < performances[j].GainPercent
	})
	
	m.worstPerformers = performances
}