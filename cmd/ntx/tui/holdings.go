package tui

import (
	"fmt"
	"strings"
)

func (m Model) viewHoldings() string {
	if len(m.holdings) == 0 {
		return dimStyle.Render("No holdings found. Import transactions with: ntx import <file>")
	}

	// Check if any holding has price data
	hasPrices := false
	for _, h := range m.holdings {
		if h.CurrentPrice != nil {
			hasPrices = true
			break
		}
	}

	var b strings.Builder

	if hasPrices {
		// Header with prices
		header := fmt.Sprintf("%-8s %8s %12s %12s %14s %12s %8s",
			"SYMBOL", "QTY", "AVG COST", "LTP", "VALUE", "P&L", "P&L %")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(strings.Repeat("─", 78)))
		b.WriteString("\n")

		for _, h := range m.holdings {
			avgCost := float64(h.AverageCost.Paisa) / 100

			if h.CurrentPrice == nil {
				b.WriteString(fmt.Sprintf("%-8s %8.0f %12.2f %s\n",
					h.Stock.Symbol,
					h.Quantity,
					avgCost,
					dimStyle.Render("      -            -            -       -")))
				continue
			}

			curPrice := float64(h.CurrentPrice.Paisa) / 100
			curValue := float64(h.CurrentValue.Paisa) / 100
			pnl := float64(h.UnrealizedPnl.Paisa) / 100
			pnlPct := h.UnrealizedPnlPercent

			pnlStr := fmt.Sprintf("%12.2f", pnl)
			pnlPctStr := fmt.Sprintf("%7.2f%%", pnlPct)
			if pnl >= 0 {
				pnlStr = profitStyle.Render(pnlStr)
				pnlPctStr = profitStyle.Render(pnlPctStr)
			} else {
				pnlStr = lossStyle.Render(pnlStr)
				pnlPctStr = lossStyle.Render(pnlPctStr)
			}

			b.WriteString(fmt.Sprintf("%-8s %8.0f %12.2f %12.2f %14.2f %s %s\n",
				h.Stock.Symbol,
				h.Quantity,
				avgCost,
				curPrice,
				curValue,
				pnlStr,
				pnlPctStr))
		}
	} else {
		// Header without prices
		header := fmt.Sprintf("%-8s %10s %15s %15s",
			"SYMBOL", "QUANTITY", "AVG COST", "TOTAL COST")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(strings.Repeat("─", 52)))
		b.WriteString("\n")

		for _, h := range m.holdings {
			avgCost := float64(h.AverageCost.Paisa) / 100
			totalCost := float64(h.TotalCost.Paisa) / 100
			b.WriteString(fmt.Sprintf("%-8s %10.0f %15.2f %15.2f\n",
				h.Stock.Symbol,
				h.Quantity,
				avgCost,
				totalCost))
		}

		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Press 's' to sync prices from NEPSE"))
	}

	return b.String()
}
