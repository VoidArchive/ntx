package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/voidarchive/ntx/internal/worker"
)

const maxConcurrency = 5

func runBackfill(ctx context.Context, w *worker.Worker, opts backfillOptions) error {
	start := time.Now()
	each := func(name string, step func(context.Context, worker.Company) error) func(context.Context) error {
		return func(ctx context.Context) error { return w.EachCompany(ctx, maxConcurrency, name, step) }
	}
	steps := []struct {
		name string
		on   bool
		run  func(context.Context) error
	}{
		{"companies", opts.companies, w.SyncCompanies},
		{"fundamentals", opts.fundamentals, each("fundamentals", w.SyncFundamentals)},
		{"price history", opts.prices, each("price history", lastYearOfPrices(w))},
		{"ownership", opts.ownership, each("ownership", w.SyncOwnership)},
		{"dividends", opts.corporateActions, each("dividends", w.SyncDividends)},
		{"ratings", opts.ratings, w.SyncRatings},
		{"weather", opts.weather, w.SyncWeather},
	}
	for _, s := range steps {
		if !s.on {
			continue
		}
		slog.Info(s.name + " sync started")
		if err := s.run(ctx); err != nil {
			return fmt.Errorf("sync %s: %w", s.name, err)
		}
	}
	slog.Info("backfill complete", "duration", time.Since(start))
	return nil
}

// lastYearOfPrices fixes the date range when the backfill starts, so every company gets the same window.
func lastYearOfPrices(w *worker.Worker) func(context.Context, worker.Company) error {
	end := time.Now().In(time.FixedZone("NPT", 5*3600+45*60))
	startDate, endDate := end.AddDate(-1, 0, 0).Format(time.DateOnly), end.Format(time.DateOnly)
	return func(ctx context.Context, c worker.Company) error {
		return w.SyncPriceHistory(ctx, c, startDate, endDate)
	}
}
