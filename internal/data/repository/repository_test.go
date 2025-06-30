package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestRepository(t *testing.T) *Services {
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

	return NewServices(db)
}

func TestPortfolioRepository(t *testing.T) {
	services := setupTestRepository(t)
	defer services.Close()

	ctx := context.Background()
	repo := services.Repository

	// Test Create
	portfolio, err := repo.Portfolio.Create(ctx, CreatePortfolioRequest{
		Name:        "Test Portfolio",
		Description: stringPtr("Test Description"),
		Currency:    "NPR",
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	if portfolio.Name != "Test Portfolio" {
		t.Errorf("expected portfolio name 'Test Portfolio', got %s", portfolio.Name)
	}

	// Test GetByID
	retrieved, err := repo.Portfolio.GetByID(ctx, portfolio.ID)
	if err != nil {
		t.Fatalf("failed to get portfolio: %v", err)
	}

	if retrieved.ID != portfolio.ID {
		t.Errorf("expected portfolio ID %d, got %d", portfolio.ID, retrieved.ID)
	}

	// Test List
	portfolios, err := repo.Portfolio.List(ctx)
	if err != nil {
		t.Fatalf("failed to list portfolios: %v", err)
	}

	if len(portfolios) != 1 {
		t.Errorf("expected 1 portfolio, got %d", len(portfolios))
	}

	// Test Update
	updated, err := repo.Portfolio.Update(ctx, UpdatePortfolioRequest{
		ID:          portfolio.ID,
		Name:        "Updated Portfolio",
		Description: stringPtr("Updated Description"),
	})
	if err != nil {
		t.Fatalf("failed to update portfolio: %v", err)
	}

	if updated.Name != "Updated Portfolio" {
		t.Errorf("expected updated name 'Updated Portfolio', got %s", updated.Name)
	}
}

func TestHoldingRepository(t *testing.T) {
	services := setupTestRepository(t)
	defer services.Close()

	ctx := context.Background()
	repo := services.Repository

	// Create portfolio first
	portfolio, err := repo.Portfolio.Create(ctx, CreatePortfolioRequest{
		Name:     "Test Portfolio",
		Currency: "NPR",
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// Test Create Holding
	holding, err := repo.Holding.Create(ctx, CreateHoldingRequest{
		PortfolioID:       portfolio.ID,
		Symbol:           "NABIL",
		Quantity:         100,
		AverageCostPaisa: 125000,
		LastPricePaisa:   int64Ptr(130000),
	})
	if err != nil {
		t.Fatalf("failed to create holding: %v", err)
	}

	if holding.Symbol != "NABIL" {
		t.Errorf("expected symbol 'NABIL', got %s", holding.Symbol)
	}

	// Test GetBySymbol
	retrieved, err := repo.Holding.GetBySymbol(ctx, portfolio.ID, "NABIL")
	if err != nil {
		t.Fatalf("failed to get holding by symbol: %v", err)
	}

	if retrieved.ID != holding.ID {
		t.Errorf("expected holding ID %d, got %d", holding.ID, retrieved.ID)
	}

	// Test GetValue
	value, err := repo.Holding.GetValue(ctx, holding.ID)
	if err != nil {
		t.Fatalf("failed to get holding value: %v", err)
	}

	expectedCost := int64(100 * 125000) // 12,500,000 paisa
	if value.TotalCostPaisa != expectedCost {
		t.Errorf("expected total cost %d, got %d", expectedCost, value.TotalCostPaisa)
	}

	expectedValue := int64(100 * 130000) // 13,000,000 paisa
	if value.TotalValuePaisa != expectedValue {
		t.Errorf("expected total value %d, got %d", expectedValue, value.TotalValuePaisa)
	}

	expectedPnl := int64(100 * (130000 - 125000)) // 500,000 paisa profit
	if value.UnrealizedPnlPaisa != expectedPnl {
		t.Errorf("expected unrealized P/L %d, got %d", expectedPnl, value.UnrealizedPnlPaisa)
	}
}

func TestTransactionWithMultipleOperations(t *testing.T) {
	services := setupTestRepository(t)
	defer services.Close()

	ctx := context.Background()

	// Test transaction that creates portfolio and holding together
	err := services.Transactor.WithTx(ctx, func(ctx context.Context, repo *Repository) error {
		// Create portfolio
		portfolio, err := repo.Portfolio.Create(ctx, CreatePortfolioRequest{
			Name:     "Transaction Test Portfolio",
			Currency: "NPR",
		})
		if err != nil {
			return err
		}

		// Create transaction
		transactionDate := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
		_, err = repo.Transaction.Create(ctx, CreateTransactionRequest{
			PortfolioID:      portfolio.ID,
			Symbol:          "NABIL",
			TransactionType: "buy",
			Quantity:        100,
			PricePaisa:      125000,
			CommissionPaisa: 2500,
			TaxPaisa:        150,
			TransactionDate: transactionDate,
			Notes:           stringPtr("Initial purchase"),
		})
		if err != nil {
			return err
		}

		// Create holding
		_, err = repo.Holding.Create(ctx, CreateHoldingRequest{
			PortfolioID:       portfolio.ID,
			Symbol:           "NABIL",
			Quantity:         100,
			AverageCostPaisa: 125000,
		})
		return err
	})

	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	// Verify data was committed
	portfolios, err := services.Repository.Portfolio.List(ctx)
	if err != nil {
		t.Fatalf("failed to list portfolios: %v", err)
	}

	if len(portfolios) != 1 {
		t.Errorf("expected 1 portfolio, got %d", len(portfolios))
	}

	transactions, err := services.Repository.Transaction.ListByPortfolio(ctx, portfolios[0].ID)
	if err != nil {
		t.Fatalf("failed to list transactions: %v", err)
	}

	if len(transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactions))
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}