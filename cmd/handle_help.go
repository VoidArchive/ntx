package main

import "fmt"

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
	fmt.Println("  test-db              Test database integration (Steps 1-3)")
	fmt.Println("  test-wac             Test WAC calculator (Step 4)")
	fmt.Println("  help                 Show this help message")
	fmt.Println("")
	fmt.Println("Phase 1: CSV Import & Basic Portfolio Display")
}