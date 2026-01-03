package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
)

// BackfillCmd fetches historical data for all symbols.
type BackfillCmd struct {
	Prices    bool `help:"Backfill price history"                 default:"true"`
	Reports   bool `help:"Backfill financial reports"             default:"true"`
	Dividends bool `help:"Backfill dividend history"              default:"true"`
	Profiles  bool `help:"Backfill company profiles/descriptions" default:"true"`
}

func (c *BackfillCmd) Run() error {
	ctx := context.Background()

	dbPath := database.DefaultServerPath()

	db, err := database.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := database.AutoMigrate(db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	queries := sqlc.New(db)

	client, err := nepse.NewClient()
	if err != nil {
		return fmt.Errorf("create nepse client: %w", err)
	}
	defer func() { _ = client.Close() }()

	// Fetch companies (for sync - has sector, market cap, etc.)
	fmt.Println("Fetching company list...")
	companies, err := client.Companies(ctx)
	if err != nil {
		return fmt.Errorf("fetch companies: %w", err)
	}
	fmt.Printf("Found %d companies\n", len(companies))

	// Fetch securities (actively tradable - use for price/reports/dividends)
	fmt.Println("Fetching securities list...")
	securities, err := client.Securities(ctx)
	if err != nil {
		return fmt.Errorf("fetch securities: %w", err)
	}
	fmt.Printf("Found %d tradable securities\n\n", len(securities))

	// Always sync companies first (required for foreign keys)
	syncCompanies(ctx, queries, companies)

	if c.Prices {
		backfillPrices(ctx, client, queries, securities)
	}

	if c.Reports {
		backfillReports(ctx, client, queries, securities)
	}

	if c.Dividends {
		backfillDividends(ctx, client, queries, securities)
	}

	if c.Profiles {
		backfillProfiles(ctx, client, queries, securities)
	}

	fmt.Println("\nBackfill complete!")
	return nil
}

// Sector name mapping from NEPSE API to database integers.
var sectorMap = map[string]int64{
	"Commercial Banks":             1,
	"Development Banks":            2,
	"Finance":                      3,
	"Microfinance":                 4,
	"Life Insurance":               5,
	"Non Life Insurance":           6,
	"Hydro Power":                  7,
	"Manufacturing And Processing": 8,
	"Hotels And Tourism":           9,
	"Trading":                      10,
	"Investment":                   11,
	"Mutual Fund":                  12,
	"Others":                       13,
}

func sectorToInt(name string) int64 {
	if id, ok := sectorMap[name]; ok {
		return id
	}
	return 0
}

func syncCompanies(ctx context.Context, queries *sqlc.Queries, companies []nepse.Company) {
	fmt.Println("=== Syncing Companies ===")

	for _, c := range companies {
		err := queries.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
			Symbol:      c.Symbol,
			Name:        c.Name,
			Sector:      sectorToInt(c.Sector),
			Description: "",
			LogoUrl:     "",
		})
		if err != nil {
			slog.Error("failed to upsert company", "symbol", c.Symbol, "error", err)
		}
	}

	fmt.Printf("Synced %d companies\n\n", len(companies))
}

func backfillPrices(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, securities []nepse.Security) {
	now := time.Now()
	defaultFrom := now.AddDate(-1, 0, 0).Format("2006-01-02")
	to := now.Format("2006-01-02")

	fmt.Println("=== Backfilling Prices ===")

	// Get existing latest dates for all symbols
	latestDates, err := queries.GetLatestPriceDates(ctx)
	if err != nil {
		slog.Error("failed to get latest dates", "error", err)
	}

	latestMap := make(map[string]string)
	for _, ld := range latestDates {
		if ld.LatestDate != nil {
			if s, ok := ld.LatestDate.(string); ok {
				latestMap[ld.Symbol] = s
			}
		}
	}

	total := len(securities)
	priceCount := 0
	errorCount := 0
	skipped := 0

	for i, sec := range securities {
		symbol := sec.Symbol
		from := defaultFrom

		// If we have existing data, start from the day after
		if latest, ok := latestMap[symbol]; ok {
			if latest >= to {
				fmt.Printf("[%d/%d] %s skipped (up to date)\n", i+1, total, symbol)
				skipped++
				continue
			}
			// Start from the day after the latest
			t, err := time.Parse("2006-01-02", latest)
			if err == nil {
				from = t.AddDate(0, 0, 1).Format("2006-01-02")
			}
		}

		fmt.Printf("[%d/%d] %s (%s to %s)...", i+1, total, symbol, from, to)

		history, err := client.PriceHistory(ctx, symbol, from, to)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				fmt.Printf(" not found\n")
			} else {
				fmt.Printf(" error: %v\n", err)
			}
			errorCount++
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for _, candle := range history {
			err := queries.UpsertPrice(ctx, sqlc.UpsertPriceParams{
				Symbol:        symbol,
				Date:          candle.Date,
				Open:          candle.Open,
				High:          candle.High,
				Low:           candle.Low,
				Close:         candle.Close,
				PreviousClose: sql.NullFloat64{Valid: false},
				Volume:        candle.Volume,
				Turnover:      sql.NullInt64{Int64: int64(candle.Turnover), Valid: true},
				IsComplete:    1,
			})
			if err != nil {
				slog.Error("failed to upsert price", "symbol", symbol, "date", candle.Date, "error", err)
				continue
			}
			priceCount++
		}

		fmt.Printf(" %d\n", len(history))
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Prices: %d new records, %d skipped, %d errors\n\n", priceCount, skipped, errorCount)
}

func backfillReports(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, securities []nepse.Security) {
	fmt.Println("=== Backfilling Reports ===")

	total := len(securities)
	reportCount := 0
	errorCount := 0

	for i, sec := range securities {
		symbol := sec.Symbol
		fmt.Printf("[%d/%d] %s reports...", i+1, total, symbol)

		reports, err := client.Reports(ctx, symbol)
		if err != nil {
			fmt.Printf(" error: %v\n", err)
			errorCount++
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for _, r := range reports {
			reportType := int64(1) // quarterly
			if r.ReportType == "annual" {
				reportType = 2
			}

			fiscalYear := parseFiscalYear(r.FiscalYear)

			err := queries.InsertReport(ctx, sqlc.InsertReportParams{
				Symbol:      symbol,
				Type:        reportType,
				FiscalYear:  fiscalYear,
				Quarter:     int64(r.Quarter),
				Eps:         sql.NullFloat64{Float64: r.EPS, Valid: r.EPS != 0},
				BookValue:   sql.NullFloat64{Float64: r.BookValue, Valid: r.BookValue != 0},
				NetIncome:   sql.NullFloat64{Float64: r.Profit, Valid: r.Profit != 0},
				PublishedAt: sql.NullString{String: r.PublishedAt, Valid: r.PublishedAt != ""},
			})
			if err != nil {
				slog.Error("failed to insert report", "symbol", symbol, "fy", r.FiscalYear, "error", err)
				continue
			}
			reportCount++
		}

		fmt.Printf(" %d\n", len(reports))
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Reports: %d symbols, %d records, %d errors\n\n", total-errorCount, reportCount, errorCount)
}

func backfillDividends(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, securities []nepse.Security) {
	fmt.Println("=== Backfilling Dividends ===")

	total := len(securities)
	divCount := 0
	errorCount := 0

	for i, company := range securities {
		symbol := company.Symbol
		fmt.Printf("[%d/%d] %s dividends...", i+1, total, symbol)

		dividends, err := client.Dividends(ctx, symbol)
		if err != nil {
			fmt.Printf(" error: %v\n", err)
			errorCount++
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for _, d := range dividends {
			err := queries.UpsertDividend(ctx, sqlc.UpsertDividendParams{
				Symbol:       symbol,
				FiscalYear:   d.FiscalYear,
				CashPercent:  d.CashPercent,
				BonusPercent: d.BonusPercent,
				Headline:     sql.NullString{String: d.Headline, Valid: d.Headline != ""},
				PublishedAt:  sql.NullString{String: d.PublishedAt, Valid: d.PublishedAt != ""},
			})
			if err != nil {
				slog.Error("failed to upsert dividend", "symbol", symbol, "fy", d.FiscalYear, "error", err)
				continue
			}
			divCount++
		}

		fmt.Printf(" %d\n", len(dividends))
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Dividends: %d symbols, %d records, %d errors\n\n", total-errorCount, divCount, errorCount)
}

func backfillProfiles(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, securities []nepse.Security) {
	fmt.Println("=== Backfilling Company Profiles ===")

	total := len(securities)
	profileCount := 0
	errorCount := 0

	for i, company := range securities {
		symbol := company.Symbol
		fmt.Printf("[%d/%d] %s profile...", i+1, total, symbol)

		profile, err := client.CompanyProfile(ctx, symbol)
		if err != nil {
			fmt.Printf(" error: %v\n", err)
			errorCount++
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if profile.Profile == "" {
			fmt.Printf(" empty\n")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		err = queries.UpdateCompanyDescription(ctx, sqlc.UpdateCompanyDescriptionParams{
			Description: profile.Profile,
			Symbol:      symbol,
		})
		if err != nil {
			slog.Error("failed to update description", "symbol", symbol, "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		profileCount++
		fmt.Printf(" ok\n")
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Profiles: %d symbols, %d updated, %d errors\n\n", total, profileCount, errorCount)
}

// parseFiscalYear extracts numeric year from Nepali fiscal year format (e.g., "2080/81" -> 2080).
func parseFiscalYear(fy string) int64 {
	fy = strings.TrimSpace(fy)
	if fy == "" {
		return 0
	}

	parts := strings.Split(fy, "/")
	if len(parts) > 0 {
		year, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			return year
		}
	}

	year, _ := strconv.ParseInt(fy, 10, 64)
	return year
}
