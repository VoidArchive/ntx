package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("NTX - Nepal Tax Portfolio Tracker")
		fmt.Println("Usage:")
		fmt.Println("  ntx import <csv-file>    Import transactions from CSV")
		fmt.Println("  ntx prices               Enter missing prices")
		fmt.Println("  ntx portfolio            Show current portfolio")
		fmt.Println("  ntx test-db              Test database integration")
		fmt.Println("  ntx test-wac             Test WAC calculator")
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
	case "test-db":
		handleTestDB()
	case "test-wac":
		handleTestWAC()
	case "help":
		handleHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'ntx help' for available commands")
	}
}