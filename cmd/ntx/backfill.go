package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

	// Fetch companies (has sector, filters out mutual funds, bonds, etc.)
	fmt.Println("Fetching company list...")
	companies, err := client.Companies(ctx)
	if err != nil {
		return fmt.Errorf("fetch companies: %w", err)
	}
	fmt.Printf("Found %d companies\n\n", len(companies))

	// Always sync companies first (required for foreign keys)
	syncCompanies(ctx, queries, companies)

	if c.Prices {
		backfillPrices(ctx, db, client, queries, companies)
	}

	if c.Reports {
		backfillReports(ctx, client, queries, companies)
	}

	if c.Dividends {
		backfillDividends(ctx, client, queries, companies)
	}

	if c.Profiles {
		backfillProfiles(ctx, client, queries, companies)
	}

	fmt.Println("\nBackfill complete!")
	return nil
}

func syncCompanies(ctx context.Context, queries *sqlc.Queries, companies []nepse.Company) {
	fmt.Println("=== Syncing Companies ===")

	for _, c := range companies {
		err := queries.UpsertCompanyBasic(ctx, sqlc.UpsertCompanyBasicParams{
			Symbol: c.Symbol,
			Name:   c.Name,
			Sector: nepse.SectorToInt(c.Sector),
		})
		if err != nil {
			slog.Error("failed to upsert company", "symbol", c.Symbol, "error", err)
		}
	}

	fmt.Printf("Synced %d companies\n\n", len(companies))
}

// worker runs tasks from a channel with rate limiting
type worker struct {
	workers     int
	rateLimit   time.Duration
	progressLog bool
}

func (w *worker) run(total int, task func(idx int) string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, w.workers)
	progressChan := make(chan string, total)

	progressDone := make(chan struct{})
	go func() {
		for msg := range progressChan {
			if w.progressLog {
				fmt.Println(msg)
			}
		}
		close(progressDone)
	}()

	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			msg := task(idx)
			if msg != "" {
				progressChan <- msg
			}
		}(i)
	}

	wg.Wait()
	close(progressChan)
	<-progressDone
}

func backfillPrices(ctx context.Context, db *sql.DB, client *nepse.Client, queries *sqlc.Queries, companies []nepse.Company) {
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

	total := len(companies)
	var priceCount, errorCount, skipped atomic.Int64

	w := &worker{workers: 5, rateLimit: 200 * time.Millisecond, progressLog: true}
	w.run(total, func(idx int) string {
		c := companies[idx]
		symbol := c.Symbol
		from := defaultFrom

		if latest, ok := latestMap[symbol]; ok {
			if latest >= to {
				skipped.Add(1)
				return fmt.Sprintf("[%d/%d] %s skipped (up to date)", idx+1, total, symbol)
			}
			t, err := time.Parse("2006-01-02", latest)
			if err == nil {
				from = t.AddDate(0, 0, 1).Format("2006-01-02")
			}
		}

		history, err := client.PriceHistory(ctx, symbol, from, to)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			if strings.Contains(err.Error(), "not found") {
				return fmt.Sprintf("[%d/%d] %s not found", idx+1, total, symbol)
			}
			errorCount.Add(1)
			return fmt.Sprintf("[%d/%d] %s error: %v", idx+1, total, symbol, err)
		}

		if len(history) == 0 {
			time.Sleep(200 * time.Millisecond)
			return fmt.Sprintf("[%d/%d] %s no new data", idx+1, total, symbol)
		}

		// Batch insert with transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			errorCount.Add(1)
			return fmt.Sprintf("[%d/%d] %s tx error: %v", idx+1, total, symbol, err)
		}

		qtx := queries.WithTx(tx)
		localCount := int64(0)
		for _, candle := range history {
			err := qtx.UpsertPrice(ctx, sqlc.UpsertPriceParams{
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
				Week52High:    sql.NullFloat64{Valid: false},
				Week52Low:     sql.NullFloat64{Valid: false},
			})
			if err != nil {
				slog.Error("failed to upsert price", "symbol", symbol, "date", candle.Date, "error", err)
				continue
			}
			localCount++
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			errorCount.Add(1)
			return fmt.Sprintf("[%d/%d] %s commit error: %v", idx+1, total, symbol, err)
		}

		priceCount.Add(localCount)
		time.Sleep(200 * time.Millisecond)
		return fmt.Sprintf("[%d/%d] %s done (%d records)", idx+1, total, symbol, len(history))
	})

	fmt.Printf("\nPrices: %d new records, %d skipped, %d errors\n\n", priceCount.Load(), skipped.Load(), errorCount.Load())
}

func backfillReports(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, companies []nepse.Company) {
	fmt.Println("=== Backfilling Reports (parallel, 5 workers) ===")

	total := len(companies)
	var reportCount, errorCount atomic.Int64

	w := &worker{workers: 5, rateLimit: 200 * time.Millisecond, progressLog: true}
	w.run(total, func(idx int) string {
		c := companies[idx]
		symbol := c.Symbol

		reports, err := client.Reports(ctx, symbol)
		if err != nil {
			errorCount.Add(1)
			time.Sleep(200 * time.Millisecond)
			return fmt.Sprintf("[%d/%d] %s error: %v", idx+1, total, symbol, err)
		}

		localCount := int64(0)
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
			localCount++
		}

		reportCount.Add(localCount)
		time.Sleep(200 * time.Millisecond)
		return fmt.Sprintf("[%d/%d] %s done (%d reports)", idx+1, total, symbol, len(reports))
	})

	fmt.Printf("\nReports: %d records, %d errors\n\n", reportCount.Load(), errorCount.Load())
}

func backfillDividends(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, companies []nepse.Company) {
	fmt.Println("=== Backfilling Dividends (parallel, 5 workers) ===")

	total := len(companies)
	var divCount, errorCount atomic.Int64

	w := &worker{workers: 5, rateLimit: 200 * time.Millisecond, progressLog: true}
	w.run(total, func(idx int) string {
		c := companies[idx]
		symbol := c.Symbol

		dividends, err := client.Dividends(ctx, symbol)
		if err != nil {
			errorCount.Add(1)
			time.Sleep(200 * time.Millisecond)
			return fmt.Sprintf("[%d/%d] %s error: %v", idx+1, total, symbol, err)
		}

		localCount := int64(0)
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
			localCount++
		}

		divCount.Add(localCount)
		time.Sleep(200 * time.Millisecond)
		return fmt.Sprintf("[%d/%d] %s done (%d dividends)", idx+1, total, symbol, len(dividends))
	})

	fmt.Printf("\nDividends: %d records, %d errors\n\n", divCount.Load(), errorCount.Load())
}

func backfillProfiles(ctx context.Context, client *nepse.Client, queries *sqlc.Queries, companies []nepse.Company) {
	fmt.Println("=== Backfilling Company Profiles (parallel, 5 workers) ===")

	total := len(companies)
	var profileCount, errorCount, skipped atomic.Int64

	w := &worker{workers: 5, rateLimit: 200 * time.Millisecond, progressLog: true}
	w.run(total, func(idx int) string {
		c := companies[idx]
		symbol := c.Symbol

		company, err := queries.GetCompany(ctx, symbol)
		if err == nil && company.Description != "" {
			skipped.Add(1)
			return fmt.Sprintf("[%d/%d] %s skipped (has description)", idx+1, total, symbol)
		}

		profile, err := client.CompanyProfile(ctx, symbol)
		if err != nil {
			errorCount.Add(1)
			time.Sleep(200 * time.Millisecond)
			return fmt.Sprintf("[%d/%d] %s error: %v", idx+1, total, symbol, err)
		}

		if profile.Profile == "" {
			time.Sleep(200 * time.Millisecond)
			return fmt.Sprintf("[%d/%d] %s empty", idx+1, total, symbol)
		}

		err = queries.UpdateCompanyDescription(ctx, sqlc.UpdateCompanyDescriptionParams{
			Description: profile.Profile,
			Symbol:      symbol,
		})
		if err != nil {
			slog.Error("failed to update description", "symbol", symbol, "error", err)
			time.Sleep(200 * time.Millisecond)
			return fmt.Sprintf("[%d/%d] %s db error: %v", idx+1, total, symbol, err)
		}

		profileCount.Add(1)
		time.Sleep(200 * time.Millisecond)
		return fmt.Sprintf("[%d/%d] %s done", idx+1, total, symbol)
	})

	fmt.Printf("\nProfiles: %d skipped, %d updated, %d errors\n\n", skipped.Load(), profileCount.Load(), errorCount.Load())
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
