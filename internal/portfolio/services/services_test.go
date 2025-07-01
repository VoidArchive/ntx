package services

import (
	"context"
	"database/sql"
	"ntx/internal/data/repository"
	"ntx/internal/portfolio/models"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestServices(t *testing.T) (*PortfolioService, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create tables for testing
	createTables := `
	CREATE TABLE portfolios (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		currency TEXT DEFAULT 'NPR',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE holdings (
		id INTEGER PRIMARY KEY,
		portfolio_id INTEGER NOT NULL,
		symbol TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		average_cost_paisa INTEGER NOT NULL,
		last_price_paisa INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (portfolio_id) REFERENCES portfolios(id),
		UNIQUE(portfolio_id, symbol)
	);

	CREATE TABLE transactions (
		id INTEGER PRIMARY KEY,
		portfolio_id INTEGER NOT NULL,
		symbol TEXT NOT NULL,
		transaction_type TEXT NOT NULL CHECK (transaction_type IN ('buy', 'sell')),
		quantity INTEGER NOT NULL,
		price_paisa INTEGER NOT NULL,
		commission_paisa INTEGER DEFAULT 0,
		tax_paisa INTEGER DEFAULT 0,
		transaction_date DATE NOT NULL,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
	);
	`

	if _, err := db.Exec(createTables); err != nil {
		t.Fatalf("failed to create test tables: %v", err)
	}

	services := repository.NewServices(db)
	portfolioService := NewPortfolioService(services.Repository, services.Transactor)

	cleanup := func() {
		services.Close()
	}

	return portfolioService, cleanup
}

func TestMoneyOperations(t *testing.T) {
	// Test basic money operations
	m1 := models.NewMoney(1250.50)
	m2 := models.NewMoney(500.25)

	// Test addition
	sum := m1.Add(m2)
	expected := models.NewMoney(1750.75)
	if !sum.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.FormatSimple(), sum.FormatSimple())
	}

	// Test multiplication
	doubled := m1.MultiplyInt(2)
	expectedDoubled := models.NewMoney(2501.00)
	if !doubled.Equal(expectedDoubled) {
		t.Errorf("expected %s, got %s", expectedDoubled.FormatSimple(), doubled.FormatSimple())
	}

	// Test percentage calculation
	pctChange := m1.PercentageChange(models.NewMoney(1000.00))
	expectedPct := 25.05 // (1250.50 - 1000) / 1000 * 100
	if pctChange < expectedPct-0.01 || pctChange > expectedPct+0.01 {
		t.Errorf("expected percentage change ~%.2f%%, got %.2f%%", expectedPct, pctChange)
	}

	// Test formatting
	formatted := m1.FormatNPR()
	if formatted != "Rs.1,250.5" {
		t.Errorf("expected 'Rs.1,250.5', got '%s'", formatted)
	}
}

func TestPortfolioCreation(t *testing.T) {
	service, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()

	// Test portfolio creation
	portfolio, err := service.CreatePortfolio(ctx, CreatePortfolioRequest{
		Name:        "Test Portfolio",
		Description: stringPtr("A test portfolio"),
		Currency:    "NPR",
	})

	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	if portfolio.Name != "Test Portfolio" {
		t.Errorf("expected portfolio name 'Test Portfolio', got %s", portfolio.Name)
	}

	if portfolio.Currency != "NPR" {
		t.Errorf("expected currency 'NPR', got %s", portfolio.Currency)
	}
}

func TestTransactionExecution(t *testing.T) {
	service, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()

	// Create portfolio
	portfolio, err := service.CreatePortfolio(ctx, CreatePortfolioRequest{
		Name:     "Test Portfolio",
		Currency: "NPR",
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// Execute buy transaction
	transactionDate := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	result, err := service.ExecuteTransaction(ctx, ExecuteTransactionRequest{
		PortfolioID:     portfolio.ID,
		Symbol:          "NABIL",
		TransactionType: "buy",
		Quantity:        100,
		Price:           models.NewMoney(1250.00),
		Commission:      models.NewMoney(25.00),
		Tax:             models.NewMoney(1.50),
		TransactionDate: transactionDate,
		Notes:           stringPtr("Initial purchase"),
	})

	if err != nil {
		t.Fatalf("failed to execute transaction: %v", err)
	}

	// Verify impact
	if result.Impact.NewHolding == nil {
		t.Fatal("expected new holding to be created")
	}

	holding := result.Impact.NewHolding
	if holding.Symbol != "NABIL" {
		t.Errorf("expected symbol 'NABIL', got %s", holding.Symbol)
	}

	if holding.Quantity != 100 {
		t.Errorf("expected quantity 100, got %d", holding.Quantity)
	}

	expectedPrice := models.NewMoney(1250.00)
	if !holding.AverageCost.Equal(expectedPrice) {
		t.Errorf("expected average cost %s, got %s", expectedPrice.FormatSimple(), holding.AverageCost.FormatSimple())
	}

	// Test portfolio stats
	stats, err := service.GetPortfolioWithStats(ctx, portfolio.ID)
	if err != nil {
		t.Fatalf("failed to get portfolio stats: %v", err)
	}

	if stats.HoldingCount != 1 {
		t.Errorf("expected 1 holding, got %d", stats.HoldingCount)
	}

	expectedTotalCost := models.NewMoney(125000.00) // 100 * 1250
	if !stats.TotalCost.Equal(expectedTotalCost) {
		t.Errorf("expected total cost %s, got %s", expectedTotalCost.FormatSimple(), stats.TotalCost.FormatSimple())
	}
}

func TestAverageCostCalculation(t *testing.T) {
	service, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()

	// Create portfolio
	portfolio, err := service.CreatePortfolio(ctx, CreatePortfolioRequest{
		Name:     "Test Portfolio",
		Currency: "NPR",
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// First buy: 100 shares at Rs.1250
	transactionDate1 := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	_, err = service.ExecuteTransaction(ctx, ExecuteTransactionRequest{
		PortfolioID:     portfolio.ID,
		Symbol:          "NABIL",
		TransactionType: "buy",
		Quantity:        100,
		Price:           models.NewMoney(1250.00),
		Commission:      models.NewMoney(25.00),
		Tax:             models.NewMoney(1.50),
		TransactionDate: transactionDate1,
	})
	if err != nil {
		t.Fatalf("failed to execute first transaction: %v", err)
	}

	// Second buy: 50 shares at Rs.1300
	transactionDate2 := time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC)
	result, err := service.ExecuteTransaction(ctx, ExecuteTransactionRequest{
		PortfolioID:     portfolio.ID,
		Symbol:          "NABIL",
		TransactionType: "buy",
		Quantity:        50,
		Price:           models.NewMoney(1300.00),
		Commission:      models.NewMoney(25.00),
		Tax:             models.NewMoney(1.50),
		TransactionDate: transactionDate2,
	})
	if err != nil {
		t.Fatalf("failed to execute second transaction: %v", err)
	}

	// Check average cost calculation
	// (100 * 1250 + 50 * 1300) / 150 = (125000 + 65000) / 150 = 190000 / 150 = 1266.67
	expectedAvgCost := models.NewMoneyFromPaisa(126667) // Rs.1266.67 in paisa
	actualAvgCost := result.Impact.NewHolding.AverageCost

	// Allow small rounding differences
	diff := actualAvgCost.Sub(expectedAvgCost).Abs()
	if diff.Paisa > 1 { // Allow 1 paisa difference for rounding
		t.Errorf("expected average cost ~%s, got %s", expectedAvgCost.FormatSimple(), actualAvgCost.FormatSimple())
	}

	if result.Impact.NewHolding.Quantity != 150 {
		t.Errorf("expected total quantity 150, got %d", result.Impact.NewHolding.Quantity)
	}
}

func TestSellTransaction(t *testing.T) {
	service, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()

	// Create portfolio and initial buy
	portfolio, err := service.CreatePortfolio(ctx, CreatePortfolioRequest{
		Name:     "Test Portfolio",
		Currency: "NPR",
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// Buy 100 shares at Rs.1250
	buyDate := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	_, err = service.ExecuteTransaction(ctx, ExecuteTransactionRequest{
		PortfolioID:     portfolio.ID,
		Symbol:          "NABIL",
		TransactionType: "buy",
		Quantity:        100,
		Price:           models.NewMoney(1250.00),
		Commission:      models.NewMoney(25.00),
		Tax:             models.NewMoney(1.50),
		TransactionDate: buyDate,
	})
	if err != nil {
		t.Fatalf("failed to execute buy transaction: %v", err)
	}

	// Sell 30 shares at Rs.1350
	sellDate := time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC)
	result, err := service.ExecuteTransaction(ctx, ExecuteTransactionRequest{
		PortfolioID:     portfolio.ID,
		Symbol:          "NABIL",
		TransactionType: "sell",
		Quantity:        30,
		Price:           models.NewMoney(1350.00),
		Commission:      models.NewMoney(20.00),
		Tax:             models.NewMoney(2.50),
		TransactionDate: sellDate,
	})
	if err != nil {
		t.Fatalf("failed to execute sell transaction: %v", err)
	}

	// Check realized P/L
	// Cost basis: 30 * 1250 = 37,500
	// Sale proceeds: 30 * 1350 = 40,500 (before fees)
	// Realized P/L: 40,500 - 37,500 = 3,000
	expectedRealizedPnL := models.NewMoney(3000.00)
	if !result.Impact.RealizedPnL.Equal(expectedRealizedPnL) {
		t.Errorf("expected realized P/L %s, got %s",
			expectedRealizedPnL.FormatSimple(), result.Impact.RealizedPnL.FormatSimple())
	}

	// Check remaining holding
	if result.Impact.NewHolding.Quantity != 70 {
		t.Errorf("expected remaining quantity 70, got %d", result.Impact.NewHolding.Quantity)
	}

	// Average cost should remain the same
	expectedAvgCost := models.NewMoney(1250.00)
	if !result.Impact.NewHolding.AverageCost.Equal(expectedAvgCost) {
		t.Errorf("expected average cost %s, got %s",
			expectedAvgCost.FormatSimple(), result.Impact.NewHolding.AverageCost.FormatSimple())
	}
}

func TestInvalidTransactions(t *testing.T) {
	service, cleanup := setupTestServices(t)
	defer cleanup()

	ctx := context.Background()

	// Create portfolio
	portfolio, err := service.CreatePortfolio(ctx, CreatePortfolioRequest{
		Name:     "Test Portfolio",
		Currency: "NPR",
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// Try to sell without holding
	sellDate := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	_, err = service.ExecuteTransaction(ctx, ExecuteTransactionRequest{
		PortfolioID:     portfolio.ID,
		Symbol:          "NABIL",
		TransactionType: "sell",
		Quantity:        100,
		Price:           models.NewMoney(1250.00),
		Commission:      models.NewMoney(25.00),
		Tax:             models.NewMoney(1.50),
		TransactionDate: sellDate,
	})

	if err == nil {
		t.Error("expected error when selling without holding")
	}

	// Test invalid validation
	_, err = service.ExecuteTransaction(ctx, ExecuteTransactionRequest{
		PortfolioID:     portfolio.ID,
		Symbol:          "",
		TransactionType: "buy",
		Quantity:        100,
		Price:           models.NewMoney(1250.00),
		Commission:      models.NewMoney(25.00),
		Tax:             models.NewMoney(1.50),
		TransactionDate: sellDate,
	})

	if err == nil {
		t.Error("expected validation error for empty symbol")
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
