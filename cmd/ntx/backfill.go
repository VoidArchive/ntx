package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
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

	fmt.Println("=== Backfilling Prices (parallel, 5 workers) ===")

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
	priceCount := int64(0)
	errorCount := int64(0)
	skipped := int64(0)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, 5)
	progressChan := make(chan string, total)

	// Progress printer goroutine
	go func() {
		for msg := range progressChan {
			fmt.Println(msg)
		}
		close(progressChan)
	}()

	for i, sec := range securities {
		wg.Add(1)
		go func(idx int, s nepse.Security) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			symbol := s.Symbol

			if err := ensureCompanyExists(ctx, queries, symbol, s.Name); err != nil {
				slog.Error("failed to ensure company exists", "symbol", symbol, "error", err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			from := defaultFrom

			if latest, ok := latestMap[symbol]; ok {
				if latest >= to {
					progressChan <- fmt.Sprintf("[%d/%d] %s skipped (up to date)", idx+1, total, symbol)
					mu.Lock()
					skipped++
					mu.Unlock()
					return
				}
				// Start from the day after the latest
				t, err := time.Parse("2006-01-02", latest)
				if err == nil {
					from = t.AddDate(0, 0, 1).Format("2006-01-02")
				}
			}

			progressChan <- fmt.Sprintf("[%d/%d] %s fetching (%s to %s)...", idx+1, total, symbol, from, to)

			history, err := client.PriceHistory(ctx, symbol, from, to)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					progressChan <- fmt.Sprintf("[%d/%d] %s not found", idx+1, total, symbol)
				} else {
					progressChan <- fmt.Sprintf("[%d/%d] %s error: %v", idx+1, total, symbol, err)
					mu.Lock()
					errorCount++
					mu.Unlock()
				}
				time.Sleep(200 * time.Millisecond)
				return
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

			progressChan <- fmt.Sprintf("[%d/%d] %s done (%d records)", idx+1, total, symbol, len(history))
			time.Sleep(200 * time.Millisecond)
		}(i, sec)
	}

	wg.Wait()
	close(progressChan)

	fmt.Printf("\nPrices: %d new records, %d skipped, %d errors\n\n", priceCount, skipped, errorCount)
}

func backfillReports(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, securities []nepse.Security) {
	fmt.Println("=== Backfilling Reports (parallel, 5 workers) ===")

	total := len(securities)
	reportCount := int64(0)
	errorCount := int64(0)
	skipped := int64(0)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, 5)
	progressChan := make(chan string, total)

	// Progress printer goroutine
	go func() {
		for msg := range progressChan {
			fmt.Println(msg)
		}
		close(progressChan)
	}()

	for i, sec := range securities {
		wg.Add(1)
		go func(idx int, s nepse.Security) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			symbol := s.Symbol

			if err := ensureCompanyExists(ctx, queries, symbol, s.Name); err != nil {
				slog.Error("failed to ensure company exists", "symbol", symbol, "error", err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			latest, err := queries.GetLatestReport(ctx, symbol)
			if err == nil {
				latestFY := latest.FiscalYear
				progressChan <- fmt.Sprintf("[%d/%d] %s skipped (has data through FY %d)", idx+1, total, symbol, latestFY)
				mu.Lock()
				skipped++
				mu.Unlock()
				return
			}

			progressChan <- fmt.Sprintf("[%d/%d] %s fetching reports...", idx+1, total, symbol)

			reports, err := client.Reports(ctx, symbol)
			if err != nil {
				progressChan <- fmt.Sprintf("[%d/%d] %s error: %v", idx+1, total, symbol, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				time.Sleep(200 * time.Millisecond)
				return
			}

			for _, r := range reports {
				reportType := int64(1)
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

			progressChan <- fmt.Sprintf("[%d/%d] %s done (%d reports)", idx+1, total, symbol, len(reports))
			time.Sleep(200 * time.Millisecond)
		}(i, sec)
	}

	wg.Wait()
	close(progressChan)

	fmt.Printf("\nReports: %d skipped, %d records, %d errors\n\n", skipped, reportCount, errorCount)
}

func backfillDividends(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, securities []nepse.Security) {
	fmt.Println("=== Backfilling Dividends (parallel, 5 workers) ===")

	total := len(securities)
	divCount := int64(0)
	errorCount := int64(0)
	skipped := int64(0)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, 5)
	progressChan := make(chan string, total)

	// Progress printer goroutine
	go func() {
		for msg := range progressChan {
			fmt.Println(msg)
		}
		close(progressChan)
	}()

	for i, sec := range securities {
		wg.Add(1)
		go func(idx int, s nepse.Security) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			symbol := s.Symbol

			if err := ensureCompanyExists(ctx, queries, symbol, s.Name); err != nil {
				slog.Error("failed to ensure company exists", "symbol", symbol, "error", err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			latest, err := queries.GetLatestDividend(ctx, symbol)
			if err == nil {
				latestFY := latest.FiscalYear
				progressChan <- fmt.Sprintf("[%d/%d] %s skipped (has data through FY %s)", idx+1, total, symbol, latestFY)
				mu.Lock()
				skipped++
				mu.Unlock()
				return
			}

			progressChan <- fmt.Sprintf("[%d/%d] %s fetching dividends...", idx+1, total, symbol)

			dividends, err := client.Dividends(ctx, symbol)
			if err != nil {
				progressChan <- fmt.Sprintf("[%d/%d] %s error: %v", idx+1, total, symbol, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				time.Sleep(200 * time.Millisecond)
				return
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

			progressChan <- fmt.Sprintf("[%d/%d] %s done (%d dividends)", idx+1, total, symbol, len(dividends))
			time.Sleep(200 * time.Millisecond)
		}(i, sec)
	}

	wg.Wait()
	close(progressChan)

	fmt.Printf("\nDividends: %d skipped, %d records, %d errors\n\n", skipped, divCount, errorCount)
}

func backfillProfiles(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, securities []nepse.Security) {
	fmt.Println("=== Backfilling Company Profiles (parallel, 5 workers) ===")

	total := len(securities)
	profileCount := int64(0)
	errorCount := int64(0)
	skipped := int64(0)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, 5)
	progressChan := make(chan string, total)

	// Progress printer goroutine
	go func() {
		for msg := range progressChan {
			fmt.Println(msg)
		}
		close(progressChan)
	}()

	for i, sec := range securities {
		wg.Add(1)
		go func(idx int, s nepse.Security) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			symbol := s.Symbol

			if err := ensureCompanyExists(ctx, queries, symbol, s.Name); err != nil {
				slog.Error("failed to ensure company exists", "symbol", symbol, "error", err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			company, err := queries.GetCompany(ctx, symbol)
			if err == nil && company.Description != "" {
				progressChan <- fmt.Sprintf("[%d/%d] %s skipped (has description)", idx+1, total, symbol)
				mu.Lock()
				skipped++
				mu.Unlock()
				return
			}

			progressChan <- fmt.Sprintf("[%d/%d] %s fetching profile...", idx+1, total, symbol)

			profile, err := client.CompanyProfile(ctx, symbol)
			if err != nil {
				progressChan <- fmt.Sprintf("[%d/%d] %s error: %v", idx+1, total, symbol, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				time.Sleep(200 * time.Millisecond)
				return
			}

			if profile.Profile == "" {
				progressChan <- fmt.Sprintf("[%d/%d] %s empty", idx+1, total, symbol)
				time.Sleep(200 * time.Millisecond)
				return
			}

			err = queries.UpdateCompanyDescription(ctx, sqlc.UpdateCompanyDescriptionParams{
				Description: profile.Profile,
				Symbol:      symbol,
			})
			if err != nil {
				slog.Error("failed to update description", "symbol", symbol, "error", err)
				time.Sleep(200 * time.Millisecond)
				return
			}

			progressChan <- fmt.Sprintf("[%d/%d] %s done", idx+1, total, symbol)
			mu.Lock()
			profileCount++
			mu.Unlock()
			time.Sleep(200 * time.Millisecond)
		}(i, sec)
	}

	wg.Wait()
	close(progressChan)

	fmt.Printf("\nProfiles: %d skipped, %d updated, %d errors\n\n", skipped, profileCount, errorCount)
}

// parseFiscalYear extracts numeric year from Nepali fiscal year format (e.g., "2080/81" -> 2080).
func ensureCompanyExists(ctx context.Context, queries *sqlc.Queries, symbol, name string) error {
	err := queries.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
		Symbol:      symbol,
		Name:        name,
		Sector:      13, // Others
		Description: "",
		LogoUrl:     "",
	})
	return err
}

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
