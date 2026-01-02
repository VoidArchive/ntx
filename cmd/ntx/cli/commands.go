package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	v1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/portfolio"
)

var (
	// Colors
	green  = lipgloss.Color("2")
	red    = lipgloss.Color("1")
	dim    = lipgloss.Color("8")
	accent = lipgloss.Color("6")

	// Styles
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(accent)
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(dim)
	profitStyle = lipgloss.NewStyle().Foreground(green)
	lossStyle   = lipgloss.NewStyle().Foreground(red)
	dimStyle    = lipgloss.NewStyle().Foreground(dim)
)

func Import(service *portfolio.Service, filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportCSV(ctx, data)
	if err != nil {
		log.Fatalf("import failed: %v", err)
	}

	fmt.Printf("Imported %d transactions\n", result.Imported)
	if result.Skipped > 0 {
		fmt.Printf("Skipped %d transactions\n", result.Skipped)
	}
	if len(result.Errors) > 0 {
		fmt.Printf("%d errors:\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}
}

func ImportWacc(service *portfolio.Service, filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportWACC(ctx, data)
	if err != nil {
		log.Fatalf("import failed: %v", err)
	}

	if result.Updated > 0 {
		fmt.Println(profitStyle.Render(fmt.Sprintf("Updated %d holdings with cost data", result.Updated)))
	}
	if result.Skipped > 0 {
		fmt.Printf("Skipped %d entries (no matching holdings)\n", result.Skipped)
	}
}

func Holdings(service *portfolio.Service) {
	ctx := context.Background()

	holdings, err := service.ListHoldings(ctx)
	if err != nil {
		log.Fatalf("failed to list holdings: %v", err)
	}

	if len(holdings) == 0 {
		fmt.Println("No holdings found")
		return
	}

	// Check if any holding has price data
	hasPrices := false
	for _, h := range holdings {
		if h.CurrentPrice != nil {
			hasPrices = true
			break
		}
	}

	if hasPrices {
		printHoldingsWithPrices(holdings)
	} else {
		printHoldingsBasic(holdings)
		fmt.Println()
		fmt.Println(dimStyle.Render("Run 'ntx sync' to fetch current prices"))
	}
}

func printHoldingsWithPrices(holdings []*v1.Holding) {
	// Header
	header := fmt.Sprintf("%-8s %8s %12s %12s %14s %12s %8s",
		"SYMBOL", "QTY", "AVG COST", "LTP", "VALUE", "P&L", "P&L %")
	fmt.Println(headerStyle.Render(header))
	fmt.Println(dimStyle.Render(strings.Repeat("─", 78)))

	for _, h := range holdings {
		avgCost := float64(h.AverageCost.Paisa) / 100

		// Handle holdings without price data
		if h.CurrentPrice == nil {
			fmt.Printf("%-8s %8.0f %12.2f %s\n",
				h.Stock.Symbol,
				h.Quantity,
				avgCost,
				dimStyle.Render("      -            -            -       -"))
			continue
		}

		curPrice := float64(h.CurrentPrice.Paisa) / 100
		curValue := float64(h.CurrentValue.Paisa) / 100
		pnl := float64(h.UnrealizedPnl.Paisa) / 100
		pnlPct := h.UnrealizedPnlPercent

		// Format P&L with color
		pnlStr := fmt.Sprintf("%12.2f", pnl)
		pnlPctStr := fmt.Sprintf("%7.2f%%", pnlPct)
		if pnl >= 0 {
			pnlStr = profitStyle.Render(pnlStr)
			pnlPctStr = profitStyle.Render(pnlPctStr)
		} else {
			pnlStr = lossStyle.Render(pnlStr)
			pnlPctStr = lossStyle.Render(pnlPctStr)
		}

		fmt.Printf("%-8s %8.0f %12.2f %12.2f %14.2f %s %s\n",
			h.Stock.Symbol,
			h.Quantity,
			avgCost,
			curPrice,
			curValue,
			pnlStr,
			pnlPctStr)
	}
}

func printHoldingsBasic(holdings []*v1.Holding) {
	header := fmt.Sprintf("%-8s %10s %15s %15s",
		"SYMBOL", "QUANTITY", "AVG COST", "TOTAL COST")
	fmt.Println(headerStyle.Render(header))
	fmt.Println(dimStyle.Render(strings.Repeat("─", 52)))

	for _, h := range holdings {
		avgCost := float64(h.AverageCost.Paisa) / 100
		totalCost := float64(h.TotalCost.Paisa) / 100
		fmt.Printf("%-8s %10.0f %15.2f %15.2f\n",
			h.Stock.Symbol,
			h.Quantity,
			avgCost,
			totalCost)
	}
}

func Summary(service *portfolio.Service) {
	ctx := context.Background()

	summary, err := service.Summary(ctx)
	if err != nil {
		log.Fatalf("failed to get summary: %v", err)
	}

	totalInv := float64(summary.TotalInvestment.Paisa) / 100
	currVal := float64(summary.CurrentValue.Paisa) / 100
	pnl := float64(summary.TotalUnrealizedPnl.Paisa) / 100
	pnlPct := summary.TotalUnrealizedPnlPercent

	fmt.Println(titleStyle.Render("Portfolio Summary"))
	fmt.Println(dimStyle.Render(strings.Repeat("─", 30)))

	fmt.Printf("%-18s Rs.%12.2f\n", "Total Investment", totalInv)

	if currVal > 0 {
		fmt.Printf("%-18s Rs.%12.2f\n", "Current Value", currVal)

		pnlStr := fmt.Sprintf("Rs.%12.2f (%+.2f%%)", pnl, pnlPct)
		if pnl >= 0 {
			pnlStr = profitStyle.Render(pnlStr)
		} else {
			pnlStr = lossStyle.Render(pnlStr)
		}
		fmt.Printf("%-18s %s\n", "Unrealized P&L", pnlStr)
	}

	fmt.Printf("%-18s %d\n", "Holdings", summary.HoldingsCount)
}

func Sync(service *portfolio.Service) {
	ctx := context.Background()

	fmt.Println("Fetching prices from NEPSE...")

	result, err := service.SyncPrices(ctx)
	if err != nil {
		log.Fatalf("sync failed: %v", err)
	}

	if result.Updated > 0 {
		fmt.Println(profitStyle.Render(fmt.Sprintf("Updated %d holdings", result.Updated)))
	}
	if result.Failed > 0 {
		fmt.Println(lossStyle.Render(fmt.Sprintf("Failed %d holdings:", result.Failed)))
		for _, e := range result.Errors {
			fmt.Printf("  %s\n", dimStyle.Render(e))
		}
	}
}

func Transactions(service *portfolio.Service, symbol, txTypeStr string, limit, offset int) {
	ctx := context.Background()

	txType := v1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
	if txTypeStr != "" {
		switch strings.ToUpper(txTypeStr) {
		case "BUY":
			txType = v1.TransactionType_TRANSACTION_TYPE_BUY
		case "SELL":
			txType = v1.TransactionType_TRANSACTION_TYPE_SELL
		case "IPO":
			txType = v1.TransactionType_TRANSACTION_TYPE_IPO
		case "BONUS":
			txType = v1.TransactionType_TRANSACTION_TYPE_BONUS
		case "RIGHTS":
			txType = v1.TransactionType_TRANSACTION_TYPE_RIGHTS
		case "MERGER_IN":
			txType = v1.TransactionType_TRANSACTION_TYPE_MERGER_IN
		case "MERGER_OUT":
			txType = v1.TransactionType_TRANSACTION_TYPE_MERGER_OUT
		case "DEMAT":
			txType = v1.TransactionType_TRANSACTION_TYPE_DEMAT
		case "REARRANGEMENT":
			txType = v1.TransactionType_TRANSACTION_TYPE_REARRANGEMENT
		default:
			log.Fatalf("unknown transaction type: %s", txTypeStr)
		}
	}

	txs, _, err := service.ListTransactions(ctx, symbol, txType, int32(limit), int32(offset))
	if err != nil {
		log.Fatalf("failed to list transactions: %v", err)
	}

	if len(txs) == 0 {
		fmt.Println("No transactions found")
		return
	}

	header := fmt.Sprintf("%-10s %-8s %-12s %10s %12s %s",
		"DATE", "SYMBOL", "TYPE", "QTY", "TOTAL", "ID")
	fmt.Println(headerStyle.Render(header))
	fmt.Println(dimStyle.Render(strings.Repeat("─", 80)))

	for _, tx := range txs {
		total := float64(tx.Total.Paisa) / 100
		date := tx.Date.AsTime().Format("2006-01-02")
		txTypeDisplay := formatTxType(tx.Type)

		fmt.Printf("%-10s %-8s %-12s %10.0f %12.2f %s\n",
			date,
			tx.Symbol,
			txTypeDisplay,
			tx.Quantity,
			total,
			dimStyle.Render(tx.Id[:8]))
	}
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
