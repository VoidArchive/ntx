package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
)

const maxConcurrency = 5

func runBackfill(ctx context.Context, queries *sqlc.Queries, client *nepse.Client) error {
	start := time.Now()

	// 1. Sync companies first (required for fundamentals/prices)
	slog.Info("syncing companies...")
	if err := syncCompanies(ctx, queries, client); err != nil {
		return fmt.Errorf("sync companies: %w", err)
	}
	slog.Info("companies synced")

	// 2. Sync fundamentals with concurrency
	slog.Info("syncing fundamentals...", "concurrency", maxConcurrency)
	if err := syncFundamentalsConcurrent(ctx, queries, client); err != nil {
		return fmt.Errorf("sync fundamentals: %w", err)
	}
	slog.Info("fundamentals synced")

	// 3. Sync prices (already bulk - no concurrency needed)
	loc, _ := time.LoadLocation("Asia/Kathmandu")
	businessDate := time.Now().In(loc).Format("2006-01-02")
	slog.Info("syncing prices...", "date", businessDate)
	if err := syncPrices(ctx, queries, client, businessDate); err != nil {
		return fmt.Errorf("sync prices: %w", err)
	}
	slog.Info("prices synced")

	slog.Info("backfill complete", "duration", time.Since(start))
	return nil
}

func syncCompanies(ctx context.Context, queries *sqlc.Queries, client *nepse.Client) error {
	companies, err := client.Companies(ctx)
	if err != nil {
		return err
	}

	for _, c := range companies {
		params := sqlc.UpsertCompanyParams{
			ID:             c.ID,
			Name:           c.Name,
			Symbol:         c.Symbol,
			Status:         c.Status,
			Email:          nullString(c.Email),
			Website:        nullString(c.Website),
			Sector:         c.Sector,
			InstrumentType: c.InstrumentType,
		}
		if err := queries.UpsertCompany(ctx, params); err != nil {
			return fmt.Errorf("upsert %s: %w", c.Symbol, err)
		}
	}
	return nil
}

func syncFundamentalsConcurrent(ctx context.Context, queries *sqlc.Queries, client *nepse.Client) error {
	companies, err := queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)
	errChan := make(chan error, len(companies))

	for _, c := range companies {
		wg.Add(1)
		go func(company sqlc.Company) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			if err := syncCompanyFundamentals(ctx, queries, client, company); err != nil {
				slog.Warn("skip fundamentals", "symbol", company.Symbol, "error", err)
			}
		}(c)
	}

	wg.Wait()
	close(errChan)

	// Check for fatal errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}

func syncCompanyFundamentals(ctx context.Context, queries *sqlc.Queries, client *nepse.Client, company sqlc.Company) error {
	fundamentals, err := client.Fundamentals(ctx, int32(company.ID))
	if err != nil {
		return err
	}

	for _, f := range fundamentals {
		params := sqlc.UpsertFundamentalParams{
			CompanyID:     company.ID,
			FiscalYear:    f.FiscalYear,
			Quarter:       nullString(f.Quarter),
			Eps:           nullFloat64(f.EPS),
			PeRatio:       nullFloat64(f.PERatio),
			BookValue:     nullFloat64(f.BookValue),
			PaidUpCapital: nullFloat64(f.PaidUpCapital),
			ProfitAmount:  nullFloat64(f.ProfitAmount),
		}
		if err := queries.UpsertFundamental(ctx, params); err != nil {
			return fmt.Errorf("upsert fundamental: %w", err)
		}
	}
	return nil
}

func syncPrices(ctx context.Context, queries *sqlc.Queries, client *nepse.Client, businessDate string) error {
	// Build symbol -> company ID map
	companies, err := queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return err
	}

	symbolToID := make(map[string]int64, len(companies))
	for _, c := range companies {
		symbolToID[c.Symbol] = c.ID
	}

	// Fetch all prices in one call
	prices, err := client.TodaysPrices(ctx, businessDate)
	if err != nil {
		return err
	}

	for _, p := range prices {
		companyID, ok := symbolToID[p.Symbol]
		if !ok {
			continue
		}

		params := sqlc.UpsertPriceParams{
			CompanyID:       companyID,
			BusinessDate:    p.BusinessDate,
			OpenPrice:       nullFloat64(p.Open),
			HighPrice:       nullFloat64(p.High),
			LowPrice:        nullFloat64(p.Low),
			ClosePrice:      nullFloat64(p.Close),
			LastTradedPrice: nullFloat64(p.LTP),
			PreviousClose:   nullFloat64(p.PreviousClose),
			ChangeAmount:    nullFloat64(p.Change),
			ChangePercent:   nullFloat64(p.ChangePercent),
			Volume:          nullInt64(p.Volume),
			Turnover:        nullFloat64(p.Turnover),
			Trades:          nullInt64(int64(p.Trades)),
		}
		if err := queries.UpsertPrice(ctx, params); err != nil {
			return fmt.Errorf("upsert price %s: %w", p.Symbol, err)
		}
	}
	return nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullFloat64(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func nullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}
