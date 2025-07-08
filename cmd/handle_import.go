package main

import (
	"context"
	"fmt"
	"log/slog"
	"ntx/internal/csv"
	"ntx/internal/db"
	"os"
)

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

	// Initialize database
	database, err := db.NewDB("ntx.db")
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer database.Close()

	// Run migrations
	if err := database.RunMigrations("internal/db/migrations"); err != nil {
		fmt.Printf("❌ Failed to run migrations: %v\n", err)
		return
	}

	// Insert transactions into database
	fmt.Println("Saving transactions to database...")
	insertedCount := 0
	errorCount := 0

	for _, tx := range transactions {
		if err := database.InsertTransaction(tx); err != nil {
			fmt.Printf("⚠️  Error inserting transaction %s: %v\n", tx.Scrip, err)
			errorCount++
		} else {
			insertedCount++
		}
	}

	fmt.Printf("✅ Successfully saved %d transactions to database\n", insertedCount)
	if errorCount > 0 {
		fmt.Printf("⚠️  %d transactions had errors\n", errorCount)
	}

	// Count transaction types and price needs
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