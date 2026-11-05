// Package worker syncs NEPSE data into the database. It is the single sync
// implementation behind both the scheduled job and `ntx backfill`.
package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
	"github.com/voidarchive/ntx/internal/story"
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

// Company is one row of the companies list that the per-company sync methods take.
type Company = sqlc.ListCompaniesRow

// EachCompany runs step for every listed company, at most concurrency at a time.
// A company whose step fails is logged and skipped: one bad feed answer must not
// stop the run for the other ~280 companies.
func (w *Worker) EachCompany(
	ctx context.Context,
	concurrency int,
	name string,
	step func(context.Context, Company) error,
) error {
	companies, err := w.queries.ListCompanies(ctx, sqlc.ListCompaniesParams{Limit: 1000, Offset: 0})
	if err != nil {
		return fmt.Errorf("list companies: %w", err)
	}

	start := time.Now()
	slog.Info(name+" sync started", "companies", len(companies), "concurrency", concurrency)

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	for _, c := range companies {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := step(ctx, c); err != nil {
				slog.Warn("skip "+name, "symbol", c.Symbol, "err", err)
			}
		}()
	}
	wg.Wait()

	slog.Info(name+" sync finished", "took", time.Since(start))
	// A deadline hit mid-run left companies unsynced; report it so the step is logged as failed.
	return ctx.Err()
}

func (w *Worker) SyncCompanies(ctx context.Context) error {
	companies, err := w.nepse.Companies(ctx)
	if err != nil {
		return fmt.Errorf("nepse companies: %w", err)
	}
	return w.storeCompanies(ctx, companies)
}

// storeCompanies upserts the fetched set by exchange id, so a ticker rename updates
// the row instead of colliding with it, then delists every company that did not arrive.
func (w *Worker) storeCompanies(ctx context.Context, companies []nepse.Company) error {
	if len(companies) == 0 {
		return errors.New("nepse returned no companies; refusing to delist everything")
	}
	ids := make([]int64, 0, len(companies))
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
		ids = append(ids, c.ID)
	}
	if err := w.queries.MarkCompaniesDelistedExcept(ctx, ids); err != nil {
		return fmt.Errorf("delist missing companies: %w", err)
	}
	return nil
}

func (w *Worker) SyncFundamentals(ctx context.Context, c Company) error {
	fundamentals, err := w.nepse.Fundamentals(ctx, safeInt32(c.ID))
	if err != nil {
		return err
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
			PublishedAt:   nullString(f.PublishedAt),
			SourceUrl:     nullString(f.SourceURL),
			FiscalYearNp:  nullString(f.FiscalYearNp),
		}
		if err := w.queries.UpsertFundamental(ctx, params); err != nil {
			return fmt.Errorf("upsert fundamental: %w", err)
		}
	}
	return nil
}

func (w *Worker) SyncOwnership(ctx context.Context, c Company) error {
	ownership, err := w.nepse.SecurityDetail(ctx, safeInt32(c.ID))
	if err != nil {
		return err
	}
	return w.storeOwnership(ctx, c.ID, *ownership)
}

// storeOwnership keeps the latest values in ownership and appends to
// ownership_history only when they differ from the last recorded row, so the
// history is a list of real changes rather than one row per day.
func (w *Worker) storeOwnership(ctx context.Context, companyID int64, o nepse.Ownership) error {
	params := sqlc.UpsertOwnershipParams{
		CompanyID:       companyID,
		ListedShares:    nullInt64(o.ListedShares),
		PublicShares:    nullInt64(o.PublicShares),
		PublicPercent:   nullFloat64(o.PublicPercent),
		PromoterShares:  nullInt64(o.PromoterShares),
		PromoterPercent: nullFloat64(o.PromoterPercent),
	}
	if err := w.queries.UpsertOwnership(ctx, params); err != nil {
		return fmt.Errorf("upsert ownership: %w", err)
	}

	last, err := w.queries.GetLatestOwnershipHistory(ctx, companyID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("latest ownership history: %w", err)
	}
	if err == nil &&
		last.ListedShares == params.ListedShares &&
		last.PublicShares == params.PublicShares &&
		last.PublicPercent == params.PublicPercent &&
		last.PromoterShares == params.PromoterShares &&
		last.PromoterPercent == params.PromoterPercent {
		return nil
	}

	asOf := o.AsOf
	if asOf == "" {
		asOf = time.Now().In(nepalTime).Format("2006-01-02")
	}
	return w.queries.UpsertOwnershipHistory(ctx, sqlc.UpsertOwnershipHistoryParams{
		CompanyID:       companyID,
		AsOf:            asOf,
		ListedShares:    params.ListedShares,
		PublicShares:    params.PublicShares,
		PublicPercent:   params.PublicPercent,
		PromoterShares:  params.PromoterShares,
		PromoterPercent: params.PromoterPercent,
	})
}

func (w *Worker) SyncDividends(ctx context.Context, c Company) error {
	dividends, err := w.nepse.Dividends(ctx, safeInt32(c.ID))
	if err != nil {
		return err
	}

	for _, d := range dividends {
		params := sqlc.UpsertCorporateActionParams{
			CompanyID:       c.ID,
			FiscalYear:      d.FiscalYear,
			BonusPercentage: nullFloat64(d.BonusPercentage),
			RightPercentage: nullFloat64Ptr(d.RightPercentage),
			CashDividend:    nullFloat64Ptr(d.CashDividend),
			SubmittedDate:   nullString(d.ModifiedDate),
			NoticeHeadline:  nullString(d.NoticeHeadline),
			NoticeBody:      nullString(d.NoticeBody),
			NoticeUrl:       nullString(d.NoticeURL),
		}
		if err := w.queries.UpsertCorporateAction(ctx, params); err != nil {
			return fmt.Errorf("upsert dividend: %w", err)
		}
	}
	return nil
}

// SyncPriceHistory stores one company's daily OHLCV between startDate and endDate (YYYY-MM-DD).
func (w *Worker) SyncPriceHistory(ctx context.Context, c Company, startDate, endDate string) error {
	history, err := w.nepse.PriceHistory(ctx, safeInt32(c.ID), startDate, endDate)
	if err != nil {
		return err
	}

	for _, h := range history {
		params := sqlc.UpsertPriceParams{
			CompanyID:    c.ID,
			BusinessDate: h.BusinessDate,
			HighPrice:    nullFloat64(h.High),
			LowPrice:     nullFloat64(h.Low),
			ClosePrice:   nullFloat64(h.Close),
			Volume:       nullInt64(h.Volume),
			Turnover:     nullFloat64(h.Turnover),
			Trades:       nullInt64(int64(h.Trades)),
		}
		if err := w.queries.UpsertPrice(ctx, params); err != nil {
			return fmt.Errorf("upsert price: %w", err)
		}
	}
	return nil
}

// SyncPrices stores today's live-market snapshot for every known symbol under businessDate.
func (w *Worker) SyncPrices(ctx context.Context, businessDate string) error {
	status, err := w.nepse.MarketStatus(ctx)
	if err != nil {
		return fmt.Errorf("market status: %w", err)
	}
	if !tradedOn(status.AsOf, businessDate) {
		slog.Info("prices sync skipped: market did not trade", "date", businessDate, "asOf", status.AsOf)
		return nil
	}

	companies, err := w.queries.ListCompanies(ctx, sqlc.ListCompaniesParams{Limit: 1000, Offset: 0})
	if err != nil {
		return fmt.Errorf("list companies: %w", err)
	}

	symbolToID := make(map[string]int64, len(companies))
	for _, c := range companies {
		symbolToID[c.Symbol] = c.ID
	}

	prices, err := w.nepse.LiveMarket(ctx)
	if err != nil {
		return fmt.Errorf("fetch prices: %w", err)
	}

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

// SyncRatings stores the screener's rating per company from the story engine.
func (w *Worker) SyncRatings(ctx context.Context) error {
	return story.SyncRatings(ctx, w.queries)
}
