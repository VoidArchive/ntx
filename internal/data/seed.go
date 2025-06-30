package data

import (
	"context"
	"fmt"
	"ntx/internal/data/repository"
	"time"
)

// SeedData populates the database with sample data for testing and development
func SeedData(db *Database) error {
	services := repository.NewServices(db.DB)
	// NOTE: Don't close services here as it would close the passed-in database connection

	ctx := context.Background()

	// Create sample portfolios
	portfolio1, err := services.Repository.Portfolio.Create(ctx, repository.CreatePortfolioRequest{
		Name:        "NEPSE Growth Portfolio",
		Description: stringPtr("High-growth banking and hydropower stocks"),
		Currency:    "NPR",
	})
	if err != nil {
		return fmt.Errorf("failed to create first portfolio: %w", err)
	}

	portfolio2, err := services.Repository.Portfolio.Create(ctx, repository.CreatePortfolioRequest{
		Name:        "Conservative Holdings",
		Description: stringPtr("Stable dividend-paying stocks"),
		Currency:    "NPR",
	})
	if err != nil {
		return fmt.Errorf("failed to create second portfolio: %w", err)
	}

	// Sample transaction dates
	nov15 := time.Date(2024, 11, 15, 0, 0, 0, 0, time.UTC)
	nov20 := time.Date(2024, 11, 20, 0, 0, 0, 0, time.UTC)
	nov22 := time.Date(2024, 11, 22, 0, 0, 0, 0, time.UTC)
	nov25 := time.Date(2024, 11, 25, 0, 0, 0, 0, time.UTC)
	nov28 := time.Date(2024, 11, 28, 0, 0, 0, 0, time.UTC)
	dec1 := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	dec5 := time.Date(2024, 12, 5, 0, 0, 0, 0, time.UTC)

	// Portfolio 1 transactions and holdings
	transactions1 := []struct {
		symbol     string
		txType     string
		quantity   int64
		price      int64
		commission int64
		tax        int64
		date       time.Time
		notes      string
	}{
		{"NABIL", "buy", 50, 125000, 1250, 75, nov15, "Initial NABIL purchase"},
		{"EBL", "buy", 30, 68000, 680, 40, nov20, "EBL investment"},
		{"HIDCL", "buy", 100, 42000, 1000, 60, nov25, "Hydropower diversification"},
		{"NABIL", "buy", 25, 130000, 650, 39, dec1, "NABIL additional purchase"},
		{"KTM", "buy", 40, 89000, 890, 53, dec5, "KTM Holdings entry"},
	}

	for _, tx := range transactions1 {
		_, err := services.Repository.Transaction.Create(ctx, repository.CreateTransactionRequest{
			PortfolioID:      portfolio1.ID,
			Symbol:          tx.symbol,
			TransactionType: tx.txType,
			Quantity:        tx.quantity,
			PricePaisa:      tx.price,
			CommissionPaisa: tx.commission,
			TaxPaisa:        tx.tax,
			TransactionDate: tx.date,
			Notes:           stringPtr(tx.notes),
		})
		if err != nil {
			return fmt.Errorf("failed to create transaction for %s: %w", tx.symbol, err)
		}
	}

	// Portfolio 1 holdings (aggregated positions)
	holdings1 := []struct {
		symbol           string
		quantity         int64
		avgCostPaisa     int64
		lastPricePaisa   int64
	}{
		{"NABIL", 75, 126667, 132000},  // 75 shares avg ₹1,266.67, LTP ₹1,320
		{"EBL", 30, 68000, 71000},      // 30 shares at ₹680, LTP ₹710
		{"HIDCL", 100, 42000, 44500},   // 100 shares at ₹420, LTP ₹445
		{"KTM", 40, 89000, 92000},      // 40 shares at ₹890, LTP ₹920
	}

	for _, h := range holdings1 {
		_, err := services.Repository.Holding.Create(ctx, repository.CreateHoldingRequest{
			PortfolioID:       portfolio1.ID,
			Symbol:           h.symbol,
			Quantity:         h.quantity,
			AverageCostPaisa: h.avgCostPaisa,
			LastPricePaisa:   int64Ptr(h.lastPricePaisa),
		})
		if err != nil {
			return fmt.Errorf("failed to create holding for %s: %w", h.symbol, err)
		}
	}

	// Portfolio 2 transactions and holdings
	transactions2 := []struct {
		symbol     string
		txType     string
		quantity   int64
		price      int64
		commission int64
		tax        int64
		date       time.Time
		notes      string
	}{
		{"ADBL", "buy", 60, 55000, 825, 49, nov15, "ADBL conservative entry"},
		{"PRVU", "buy", 25, 78000, 487, 29, nov22, "PRVU stable banking"},
		{"SBI", "buy", 80, 41000, 820, 49, nov28, "SBI diversification"},
	}

	for _, tx := range transactions2 {
		_, err := services.Repository.Transaction.Create(ctx, repository.CreateTransactionRequest{
			PortfolioID:      portfolio2.ID,
			Symbol:          tx.symbol,
			TransactionType: tx.txType,
			Quantity:        tx.quantity,
			PricePaisa:      tx.price,
			CommissionPaisa: tx.commission,
			TaxPaisa:        tx.tax,
			TransactionDate: tx.date,
			Notes:           stringPtr(tx.notes),
		})
		if err != nil {
			return fmt.Errorf("failed to create transaction for %s: %w", tx.symbol, err)
		}
	}

	// Portfolio 2 holdings
	holdings2 := []struct {
		symbol           string
		quantity         int64
		avgCostPaisa     int64
		lastPricePaisa   int64
	}{
		{"ADBL", 60, 55000, 56500},     // 60 shares at ₹550, LTP ₹565
		{"PRVU", 25, 78000, 79500},     // 25 shares at ₹780, LTP ₹795
		{"SBI", 80, 41000, 42200},      // 80 shares at ₹410, LTP ₹422
	}

	for _, h := range holdings2 {
		_, err := services.Repository.Holding.Create(ctx, repository.CreateHoldingRequest{
			PortfolioID:       portfolio2.ID,
			Symbol:           h.symbol,
			Quantity:         h.quantity,
			AverageCostPaisa: h.avgCostPaisa,
			LastPricePaisa:   int64Ptr(h.lastPricePaisa),
		})
		if err != nil {
			return fmt.Errorf("failed to create holding for %s: %w", h.symbol, err)
		}
	}

	// Sample corporate actions
	corporateActions := []struct {
		symbol           string
		actionType       string
		announcementDate time.Time
		recordDate       time.Time
		executionDate    *time.Time
		ratioFrom        *int64
		ratioTo          *int64
		amountPaisa      *int64
		notes            string
	}{
		{
			symbol:           "NABIL",
			actionType:       "dividend",
			announcementDate: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			recordDate:       time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC),
			executionDate:    timePtr(time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)),
			amountPaisa:      int64Ptr(2000), // ₹20 per share dividend
			notes:            "Annual cash dividend payment",
		},
		{
			symbol:           "HIDCL",
			actionType:       "bonus",
			announcementDate: time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC),
			recordDate:       time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			executionDate:    timePtr(time.Date(2024, 10, 20, 0, 0, 0, 0, time.UTC)),
			ratioFrom:        int64Ptr(1),
			ratioTo:          int64Ptr(10), // 1:10 bonus shares
			notes:            "1 bonus share for every 10 held",
		},
		{
			symbol:           "EBL",
			actionType:       "rights",
			announcementDate: time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC),
			recordDate:       time.Date(2024, 11, 15, 0, 0, 0, 0, time.UTC),
			ratioFrom:        int64Ptr(1),
			ratioTo:          int64Ptr(5), // 1:5 rights issue
			amountPaisa:      int64Ptr(50000), // ₹500 per share rights price
			notes:            "Rights issue at ₹500 per share",
		},
	}

	for _, ca := range corporateActions {
		_, err := services.Repository.CorporateAction.Create(ctx, repository.CreateCorporateActionRequest{
			Symbol:           ca.symbol,
			ActionType:       ca.actionType,
			AnnouncementDate: ca.announcementDate,
			RecordDate:       ca.recordDate,
			ExecutionDate:    ca.executionDate,
			RatioFrom:        ca.ratioFrom,
			RatioTo:          ca.ratioTo,
			AmountPaisa:      ca.amountPaisa,
			Notes:            stringPtr(ca.notes),
		})
		if err != nil {
			return fmt.Errorf("failed to create corporate action for %s: %w", ca.symbol, err)
		}
	}

	fmt.Printf("Successfully seeded database with:\n")
	fmt.Printf("  - 2 portfolios\n")
	fmt.Printf("  - 8 transactions\n")
	fmt.Printf("  - 7 holdings\n")
	fmt.Printf("  - 3 corporate actions\n")
	fmt.Printf("\nPortfolios:\n")
	fmt.Printf("  1. %s (ID: %d)\n", portfolio1.Name, portfolio1.ID)
	fmt.Printf("  2. %s (ID: %d)\n", portfolio2.Name, portfolio2.ID)

	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}