package tui

import (
	"fmt"
	"strings"
)

func (m Model) viewSummary() string {
	if m.summary == nil {
		return dimStyle.Render("Loading summary...")
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Portfolio Summary"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", 35)))
	b.WriteString("\n\n")

	totalInv := float64(m.summary.TotalInvestment.Paisa) / 100
	currVal := float64(m.summary.CurrentValue.Paisa) / 100
	pnl := float64(m.summary.TotalUnrealizedPnl.Paisa) / 100
	pnlPct := m.summary.TotalUnrealizedPnlPercent

	b.WriteString(fmt.Sprintf("%-20s Rs. %12.2f\n", "Total Investment", totalInv))

	if currVal > 0 {
		b.WriteString(fmt.Sprintf("%-20s Rs. %12.2f\n", "Current Value", currVal))

		pnlStr := fmt.Sprintf("Rs. %12.2f (%+.2f%%)", pnl, pnlPct)
		if pnl >= 0 {
			pnlStr = profitStyle.Render(pnlStr)
		} else {
			pnlStr = lossStyle.Render(pnlStr)
		}
		b.WriteString(fmt.Sprintf("%-20s %s\n", "Unrealized P&L", pnlStr))
	} else {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Press 's' to sync prices and calculate P&L"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%-20s %d\n", "Holdings", m.summary.HoldingsCount))

	return b.String()
}
