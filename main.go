package main

import (
	"context"
	"fmt"
	"log/slog"
	"ntx/internal/csv"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("NTX - Nepal Tax Portfolio Tracker")
		fmt.Println("Usage:")
		fmt.Println("  ntx import <csv-file>    Import transactions from CSV")
		fmt.Println("  ntx prices               Enter missing prices")
		fmt.Println("  ntx portfolio            Show current portfolio")
		fmt.Println("  ntx help                 Show this help message")
		return
	}

	command := os.Args[1]
	switch command {
	case "import":
		handleImport()
	case "prices":
		handlePrices()
	case "portfolio":
		handlePortfolio()
	case "help":
		handleHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'ntx help' for available commands")
	}
}

func handleImport() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ntx import <csv-file>")
		return
	}

	csvFile := os.Args[2]
	fmt.Printf("Importing transactions from %s...\n", csvFile)

	ctx := context.Background()
	transactions, err := csv.ParseCSV(ctx, csvFile)
	if err != nil {
		slog.Error("Failed to parse CSV file",
			"file", csvFile,
			"error", err)
		return
	}

	fmt.Printf("Successfully parsed %d transactions\n", len(transactions))

	// Count transaction types
	typeCounts := make(map[string]int)
	needsPriceCount := 0

	for _, tx := range transactions {
		typeCounts[tx.TransactionType.String()]++
		if tx.NeedsPrice() {
			needsPriceCount++
		}
	}

	fmt.Println("\nTransaction summary:")
	for txType, count := range typeCounts {
		fmt.Printf("  %s: %d transactions\n", txType, count)
	}

	fmt.Printf("\nTransactions needing price input: %d\n", needsPriceCount)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'ntx prices' to enter missing prices")
	fmt.Println("2. Run 'ntx portfolio' to view your holdings")
}

func handlePrices() {
	fmt.Println("Price input functionality - coming soon")
}

func handlePortfolio() {
	fmt.Println("Portfolio display functionality - coming soon")
}

func handleHelp() {
	fmt.Println("NTX - Nepal Tax Portfolio Tracker")
	fmt.Println("")
	fmt.Println("A terminal-based portfolio tracker for Nepali stock market")
	fmt.Println("that calculates holdings using FIFO method for tax purposes.")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  import <csv-file>    Import transactions from Meroshare CSV")
	fmt.Println("  prices               Enter missing transaction prices")
	fmt.Println("  portfolio            Display current portfolio with WAC")
	fmt.Println("  help                 Show this help message")
	fmt.Println("")
	fmt.Println("Phase 1: CSV Import & Basic Portfolio Display")
}

