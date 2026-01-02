package tui

import (
	"fmt"
	"strings"

	v1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
)

func (m Model) viewTransactions() string {
	if len(m.transactions) == 0 {
		return dimStyle.Render("No transactions found. Import with: ntx import <file>")
	}

	var b strings.Builder

	header := fmt.Sprintf("%-10s %-8s %-12s %10s %12s",
		"DATE", "SYMBOL", "TYPE", "QTY", "TOTAL")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", 56)))
	b.WriteString("\n")

	for _, tx := range m.transactions {
		total := float64(tx.Total.Paisa) / 100
		date := tx.Date.AsTime().Format("2006-01-02")
		txTypeDisplay := formatTxType(tx.Type)

		b.WriteString(fmt.Sprintf("%-10s %-8s %-12s %10.0f %12.2f\n",
			date,
			tx.Symbol,
			txTypeDisplay,
			tx.Quantity,
			total))
	}

	return b.String()
}

func formatTxType(t v1.TransactionType) string {
	switch t {
	case v1.TransactionType_TRANSACTION_TYPE_BUY:
		return profitStyle.Render("BUY")
	case v1.TransactionType_TRANSACTION_TYPE_SELL:
		return lossStyle.Render("SELL")
	case v1.TransactionType_TRANSACTION_TYPE_IPO:
		return "IPO"
	case v1.TransactionType_TRANSACTION_TYPE_BONUS:
		return "BONUS"
	case v1.TransactionType_TRANSACTION_TYPE_RIGHTS:
		return "RIGHTS"
	case v1.TransactionType_TRANSACTION_TYPE_MERGER_IN:
		return "MERGER_IN"
	case v1.TransactionType_TRANSACTION_TYPE_MERGER_OUT:
		return "MERGER_OUT"
	case v1.TransactionType_TRANSACTION_TYPE_DEMAT:
		return dimStyle.Render("DEMAT")
	case v1.TransactionType_TRANSACTION_TYPE_REARRANGEMENT:
		return "REARRANGE"
	default:
		return "UNKNOWN"
	}
}
