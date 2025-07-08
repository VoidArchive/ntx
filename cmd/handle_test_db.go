package main

import (
	"fmt"
	"ntx/internal/csv"
	"ntx/internal/db"
	"ntx/internal/money"
	"time"
)

func handleTestDB() {
	fmt.Println("Testing database integration (Steps 1-3)...")
	
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
	fmt.Println("✅ Database connection and migrations OK")
	
	// Create test transaction
	testTransaction := csv.Transaction{
		Scrip:           "API",
		Date:            time.Now().Truncate(24 * time.Hour),
		Quantity:        100,
		Price:           money.NewMoney(295.50),
		TransactionType: csv.TransactionTypeRegular,
		Description:     "Test transaction for integration",
	}
	
	// Test insertion
	fmt.Printf("Inserting test transaction: %s %d shares at %s\n", 
		testTransaction.Scrip, testTransaction.Quantity, testTransaction.Price)
	
	if err := database.InsertTransaction(testTransaction); err != nil {
		fmt.Printf("❌ Failed to insert transaction: %v\n", err)
		return
	}
	fmt.Println("✅ Transaction insertion OK")
	
	// Test retrieval
	allTransactions, err := database.GetAllTransactions()
	if err != nil {
		fmt.Printf("❌ Failed to retrieve transactions: %v\n", err)
		return
	}
	fmt.Printf("✅ Retrieved %d transactions from database\n", len(allTransactions))
	
	// Verify data integrity
	found := false
	for _, tx := range allTransactions {
		if tx.Scrip == testTransaction.Scrip && 
		   tx.Quantity == testTransaction.Quantity &&
		   tx.Price.Equal(testTransaction.Price) {
			found = true
			fmt.Printf("✅ Data integrity verified: %s %d shares at %s\n", 
				tx.Scrip, tx.Quantity, tx.Price)
			break
		}
	}
	
	if !found {
		fmt.Println("❌ Data integrity check failed - inserted data not found correctly")
		return
	}
	
	// Test scrip-specific retrieval
	apiTransactions, err := database.GetTransactionsByScrip("API")
	if err != nil {
		fmt.Printf("❌ Failed to retrieve API transactions: %v\n", err)
		return
	}
	fmt.Printf("✅ Retrieved %d API transactions\n", len(apiTransactions))
	
	fmt.Println("\n🎉 All database integration tests passed!")
	fmt.Println("Steps 1-3 integration verified:")
	fmt.Println("  ✓ CSV Transaction creation (Step 1-2)")
	fmt.Println("  ✓ Database storage (Step 3)")
	fmt.Println("  ✓ Data retrieval (Step 3)")
	fmt.Println("  ✓ Type conversion and integrity")
}