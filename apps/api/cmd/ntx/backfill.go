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

func runBackfill(ctx context.Context, queries *sqlc.Queries, client *nepse.Client, opts backfillOptions) error {
	start := time.Now()

	if opts.companies {
		slog.Info("syncing companies...")
		if err := syncCompanies(ctx, queries, client); err != nil {
			return fmt.Errorf("sync companies: %w", err)
		}
		slog.Info("companies synced")
	}

	if opts.fundamentals {
		slog.Info("syncing fundamentals...", "concurrency", maxConcurrency)
		if err := syncFundamentalsConcurrent(ctx, queries, client); err != nil {
			return fmt.Errorf("sync fundamentals: %w", err)
		}
		slog.Info("fundamentals synced")
	}

	if opts.prices {
		slog.Info("syncing price history...", "concurrency", maxConcurrency)
		if err := syncPriceHistoryConcurrent(ctx, queries, client); err != nil {
			return fmt.Errorf("sync price history: %w", err)
		}
		slog.Info("price history synced")
	}

	if opts.ownership {
		slog.Info("syncing ownership...", "concurrency", maxConcurrency)
		if err := syncOwnershipConcurrent(ctx, queries, client); err != nil {
			return fmt.Errorf("sync ownership: %w", err)
		}
		slog.Info("ownership synced")
	}

	if opts.corporateActions {
		slog.Info("syncing dividends...", "concurrency", maxConcurrency)
		if err := syncDividendsConcurrent(ctx, queries, client); err != nil {
			return fmt.Errorf("sync dividends: %w", err)
		}
		slog.Info("dividends synced")
	}

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

	for _, r := range companies {
		wg.Add(1)
		go func(row sqlc.ListCompaniesRow) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			company := rowToCompany(row)
			if err := syncCompanyFundamentals(ctx, queries, client, company); err != nil {
				slog.Warn("skip fundamentals", "symbol", company.Symbol, "error", err)
			}
		}(r)
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

func syncCompanyFundamentals(
	ctx context.Context,
	queries *sqlc.Queries,
	client *nepse.Client,
	company sqlc.Company,
) error {
	fundamentals, err := client.Fundamentals(ctx, safeInt32(company.ID))
	if err != nil {
		return err
	}

	for _, f := range fundamentals {
		params := sqlc.UpsertFundamentalParams{
			CompanyID:     company.ID,
			FiscalYear:    f.FiscalYear,
			Quarter:       f.Quarter, // Empty string for annual data
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

func syncOwnershipConcurrent(ctx context.Context, queries *sqlc.Queries, client *nepse.Client) error {
	companies, err := queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)

	for _, r := range companies {
		wg.Add(1)
		go func(row sqlc.ListCompaniesRow) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			company := rowToCompany(row)
			if err := syncCompanyOwnership(ctx, queries, client, company); err != nil {
				slog.Warn("skip ownership", "symbol", company.Symbol, "error", err)
			}
		}(r)
	}

	wg.Wait()
	return nil
}

func syncCompanyOwnership(
	ctx context.Context,
	queries *sqlc.Queries,
	client *nepse.Client,
	company sqlc.Company,
) error {
	ownership, err := client.SecurityDetail(ctx, safeInt32(company.ID))
	if err != nil {
		return err
	}

	params := sqlc.UpsertOwnershipParams{
		CompanyID:       company.ID,
		ListedShares:    nullInt64(ownership.ListedShares),
		PublicShares:    nullInt64(ownership.PublicShares),
		PublicPercent:   nullFloat64(ownership.PublicPercent),
		PromoterShares:  nullInt64(ownership.PromoterShares),
		PromoterPercent: nullFloat64(ownership.PromoterPercent),
	}
	return queries.UpsertOwnership(ctx, params)
}

func syncDividendsConcurrent(ctx context.Context, queries *sqlc.Queries, client *nepse.Client) error {
	companies, err := queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)

	for _, r := range companies {
		wg.Add(1)
		go func(row sqlc.ListCompaniesRow) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			company := rowToCompany(row)
			if err := syncCompanyDividends(ctx, queries, client, company); err != nil {
				slog.Warn("skip dividends", "symbol", company.Symbol, "error", err)
			}
		}(r)
	}

	wg.Wait()
	return nil
}

func syncCompanyDividends(
	ctx context.Context,
	queries *sqlc.Queries,
	client *nepse.Client,
	company sqlc.Company,
) error {
	dividends, err := client.Dividends(ctx, safeInt32(company.ID))
	if err != nil {
		return err
	}

	for _, d := range dividends {
		params := sqlc.UpsertCorporateActionParams{
			CompanyID:       company.ID,
			FiscalYear:      d.FiscalYear,
			BonusPercentage: nullFloat64(d.BonusPercentage),
			RightPercentage: nullFloat64Ptr(d.RightPercentage),
			CashDividend:    nullFloat64Ptr(d.CashDividend),
			SubmittedDate:   nullString(d.ModifiedDate),
		}
		if err := queries.UpsertCorporateAction(ctx, params); err != nil {
			return fmt.Errorf("upsert dividend: %w", err)
		}
	}
	return nil
}

func syncPriceHistoryConcurrent(ctx context.Context, queries *sqlc.Queries, client *nepse.Client) error {
	companies, err := queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return err
	}

	// Calculate 1 year date range
	loc, _ := time.LoadLocation("Asia/Kathmandu")
	endDate := time.Now().In(loc)
	startDate := endDate.AddDate(-1, 0, 0)

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)

	for _, r := range companies {
		wg.Add(1)
		go func(row sqlc.ListCompaniesRow) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			company := rowToCompany(row)
			if err := syncCompanyPriceHistory(ctx, queries, client, company, startStr, endStr); err != nil {
				slog.Warn("skip price history", "symbol", company.Symbol, "error", err)
			}
		}(r)
	}

	wg.Wait()
	return nil
}

func syncCompanyPriceHistory(
	ctx context.Context,
	queries *sqlc.Queries,
	client *nepse.Client,
	company sqlc.Company,
	startDate, endDate string,
) error {
	history, err := client.PriceHistory(ctx, safeInt32(company.ID), startDate, endDate)
	if err != nil {
		return err
	}

	slog.Debug("price history received", "symbol", company.Symbol, "count", len(history))

	for _, h := range history {
		params := sqlc.UpsertPriceParams{
			CompanyID:    company.ID,
			BusinessDate: h.BusinessDate,
			HighPrice:    nullFloat64(h.High),
			LowPrice:     nullFloat64(h.Low),
			ClosePrice:   nullFloat64(h.Close),
			Volume:       nullInt64(h.Volume),
			Turnover:     nullFloat64(h.Turnover),
			Trades:       nullInt64(int64(h.Trades)),
		}
		if err := queries.UpsertPrice(ctx, params); err != nil {
			return fmt.Errorf("upsert price: %w", err)
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

func nullFloat64Ptr(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func safeInt32(v int64) int32 {
	const maxInt32 = 1<<31 - 1
	if v > maxInt32 {
		return maxInt32
	}
	return int32(v) //nolint:gosec // bounds checked above
}

func rowToCompany(r sqlc.ListCompaniesRow) sqlc.Company {
	return sqlc.Company{
		ID:             r.ID,
		Name:           r.Name,
		Symbol:         r.Symbol,
		Status:         r.Status,
		Email:          r.Email,
		Website:        r.Website,
		Sector:         r.Sector,
		InstrumentType: r.InstrumentType,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}
