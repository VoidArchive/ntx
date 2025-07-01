package database

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *Service {
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
	`

	if _, err := db.Exec(createTables); err != nil {
		t.Fatalf("failed to create test tables: %v", err)
	}

	return NewService(db)
}

func TestCreatePortfolio(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	portfolio, err := db.CreatePortfolio(ctx, CreatePortfolioParams{
		Name:        "Test Portfolio",
		Description: sql.NullString{String: "Test Description", Valid: true},
		Currency:    sql.NullString{String: "NPR", Valid: true},
	})

	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	if portfolio.Name != "Test Portfolio" {
		t.Errorf("expected portfolio name 'Test Portfolio', got %s", portfolio.Name)
	}

	if !portfolio.Description.Valid || portfolio.Description.String != "Test Description" {
		t.Errorf("expected description 'Test Description', got %v", portfolio.Description)
	}
}

func TestCreateHolding(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// First create a portfolio
	portfolio, err := db.CreatePortfolio(ctx, CreatePortfolioParams{
		Name:     "Test Portfolio",
		Currency: sql.NullString{String: "NPR", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// Create a holding
	holding, err := db.CreateHolding(ctx, CreateHoldingParams{
		PortfolioID:      portfolio.ID,
		Symbol:           "NABIL",
		Quantity:         100,
		AverageCostPaisa: 125000,                                    // Rs.1250.00
		LastPricePaisa:   sql.NullInt64{Int64: 130000, Valid: true}, // Rs.1300.00
	})

	if err != nil {
		t.Fatalf("failed to create holding: %v", err)
	}

	if holding.Symbol != "NABIL" {
		t.Errorf("expected symbol 'NABIL', got %s", holding.Symbol)
	}

	if holding.Quantity != 100 {
		t.Errorf("expected quantity 100, got %d", holding.Quantity)
	}

	if holding.AverageCostPaisa != 125000 {
		t.Errorf("expected average cost 125000 paisa, got %d", holding.AverageCostPaisa)
	}
}

func TestPortfolioStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create portfolio with holdings
	portfolio, err := db.CreatePortfolio(ctx, CreatePortfolioParams{
		Name:     "Test Portfolio",
		Currency: sql.NullString{String: "NPR", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create portfolio: %v", err)
	}

	// Create multiple holdings
	holdings := []CreateHoldingParams{
		{
			PortfolioID:      portfolio.ID,
			Symbol:           "NABIL",
			Quantity:         100,
			AverageCostPaisa: 125000,
			LastPricePaisa:   sql.NullInt64{Int64: 130000, Valid: true},
		},
		{
			PortfolioID:      portfolio.ID,
			Symbol:           "EBL",
			Quantity:         50,
			AverageCostPaisa: 68000,
			LastPricePaisa:   sql.NullInt64{Int64: 71000, Valid: true},
		},
	}

	for _, h := range holdings {
		_, err := db.CreateHolding(ctx, h)
		if err != nil {
			t.Fatalf("failed to create holding: %v", err)
		}
	}

	// Get portfolio stats
	stats, err := db.GetPortfolioStats(ctx, portfolio.ID)
	if err != nil {
		t.Fatalf("failed to get portfolio stats: %v", err)
	}

	expectedHoldings := int64(2)
	if stats.HoldingCount != expectedHoldings {
		t.Errorf("expected %d holdings, got %d", expectedHoldings, stats.HoldingCount)
	}

	expectedCost := int64(100*125000 + 50*68000) // 12,500,000 + 3,400,000 = 15,900,000 paisa
	if stats.TotalCostPaisa != expectedCost {
		t.Errorf("expected total cost %d paisa, got %d", expectedCost, stats.TotalCostPaisa)
	}

	expectedValue := int64(100*130000 + 50*71000) // 13,000,000 + 3,550,000 = 16,550,000 paisa
	if stats.TotalValuePaisa != expectedValue {
		t.Errorf("expected total value %d paisa, got %d", expectedValue, stats.TotalValuePaisa)
	}
}
