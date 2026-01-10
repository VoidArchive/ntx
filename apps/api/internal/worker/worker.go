// Package worker
package worker

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
)

type Worker struct {
	nepse   *nepse.Client
	queries *sqlc.Queries
}

func New(client *nepse.Client, queries *sqlc.Queries) *Worker {
	return &Worker{
		nepse:   client,
		queries: queries,
	}
}

func (w *Worker) SyncCompanies(ctx context.Context) error {
	companies, err := w.nepse.Companies(ctx)
	if err != nil {
		return fmt.Errorf("nepse companies: %w", err)
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
		if err := w.queries.UpsertCompany(ctx, params); err != nil {
			return fmt.Errorf("upsert company %q: %w", c.Symbol, err)
		}
	}
	return nil
}

func (w *Worker) SyncFundamentals(ctx context.Context) error {
	companies, err := w.queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("list companies: %w", err)
	}

	for _, c := range companies {
		fundamentals, err := w.nepse.Fundamentals(ctx, safeInt32(c.ID))
		if err != nil {
			// Log and continue - don't fail entire sync for one company
			fmt.Printf("skip fundamentals for %s: %v\n", c.Symbol, err)
			continue
		}

		for _, f := range fundamentals {
			params := sqlc.UpsertFundamentalParams{
				CompanyID:     c.ID,
				FiscalYear:    f.FiscalYear,
				Quarter:       f.Quarter, // Empty string for annual data
				Eps:           nullFloat64(f.EPS),
				PeRatio:       nullFloat64(f.PERatio),
				BookValue:     nullFloat64(f.BookValue),
				PaidUpCapital: nullFloat64(f.PaidUpCapital),
				ProfitAmount:  nullFloat64(f.ProfitAmount),
			}
			if err := w.queries.UpsertFundamental(ctx, params); err != nil {
				return fmt.Errorf("upsert fundamental for %s: %w", c.Symbol, err)
			}
		}
	}
	return nil
}

func (w *Worker) SyncPrices(ctx context.Context, businessDate string) error {
	// Build symbol -> company ID map
	companies, err := w.queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("list companies: %w", err)
	}

	symbolToID := make(map[string]int64, len(companies))
	for _, c := range companies {
		symbolToID[c.Symbol] = c.ID
	}

	// Use LiveMarket - TodaysPrices requires auth that go-nepse doesn't support
	prices, err := w.nepse.LiveMarket(ctx)
	if err != nil {
		return fmt.Errorf("fetch prices: %w", err)
	}

	// Upsert prices
	for _, p := range prices {
		companyID, ok := symbolToID[p.Symbol]
		if !ok {
			continue // Skip unknown symbols
		}

		params := sqlc.UpsertPriceParams{
			CompanyID:       companyID,
			BusinessDate:    businessDate,
			OpenPrice:       nullFloat64(p.Open),
			HighPrice:       nullFloat64(p.High),
			LowPrice:        nullFloat64(p.Low),
			ClosePrice:      nullFloat64(p.LTP),
			LastTradedPrice: nullFloat64(p.LTP),
			PreviousClose:   nullFloat64(p.PreviousClose),
			ChangeAmount:    nullFloat64(p.LTP - p.PreviousClose),
			ChangePercent:   nullFloat64(p.ChangePercent),
			Volume:          nullInt64(p.Volume),
			Turnover:        nullFloat64(p.Turnover),
			Trades:          nullInt64(int64(p.Trades)),
		}
		if err := w.queries.UpsertPrice(ctx, params); err != nil {
			return fmt.Errorf("upsert price for %s: %w", p.Symbol, err)
		}
	}
	return nil
}

func (w *Worker) SyncOwnership(ctx context.Context) error {
	companies, err := w.queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("list companies: %w", err)
	}

	for _, c := range companies {
		ownership, err := w.nepse.SecurityDetail(ctx, safeInt32(c.ID))
		if err != nil {
			fmt.Printf("skip ownership for %s: %v\n", c.Symbol, err)
			continue
		}

		params := sqlc.UpsertOwnershipParams{
			CompanyID:       c.ID,
			ListedShares:    nullInt64(ownership.ListedShares),
			PublicShares:    nullInt64(ownership.PublicShares),
			PublicPercent:   nullFloat64(ownership.PublicPercent),
			PromoterShares:  nullInt64(ownership.PromoterShares),
			PromoterPercent: nullFloat64(ownership.PromoterPercent),
		}
		if err := w.queries.UpsertOwnership(ctx, params); err != nil {
			return fmt.Errorf("upsert ownership for %s: %w", c.Symbol, err)
		}
	}
	return nil
}

func (w *Worker) SyncDividends(ctx context.Context) error {
	companies, err := w.queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("list companies: %w", err)
	}

	for _, c := range companies {
		dividends, err := w.nepse.Dividends(ctx, safeInt32(c.ID))
		if err != nil {
			fmt.Printf("skip dividends for %s: %v\n", c.Symbol, err)
			continue
		}

		for _, d := range dividends {
			params := sqlc.UpsertCorporateActionParams{
				CompanyID:       c.ID,
				FiscalYear:      d.FiscalYear,
				BonusPercentage: nullFloat64(d.BonusPercentage),
				RightPercentage: nullFloat64Ptr(d.RightPercentage),
				CashDividend:    nullFloat64Ptr(d.CashDividend),
				SubmittedDate:   nullString(d.ModifiedDate),
			}
			if err := w.queries.UpsertCorporateAction(ctx, params); err != nil {
				return fmt.Errorf("upsert dividend for %s: %w", c.Symbol, err)
			}
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

func nullFloat64Ptr(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func nullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

func safeInt32(v int64) int32 {
	const maxInt32 = 1<<31 - 1
	if v > maxInt32 {
		return maxInt32
	}
	return int32(v) //nolint:gosec // bounds checked above
}
