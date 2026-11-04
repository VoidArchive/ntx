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

	if opts.companies {
		slog.Info("companies sync started")
		if err := w.SyncCompanies(ctx); err != nil {
			return fmt.Errorf("sync companies: %w", err)
		}
		slog.Info("companies sync finished")
	}

	if opts.fundamentals {
		if err := w.EachCompany(ctx, maxConcurrency, "fundamentals", w.SyncFundamentals); err != nil {
			return fmt.Errorf("sync fundamentals: %w", err)
		}
	}

	if opts.prices {
		loc, err := time.LoadLocation("Asia/Kathmandu")
		if err != nil {
			return err
		}
		end := time.Now().In(loc)
		startDate, endDate := end.AddDate(-1, 0, 0).Format("2006-01-02"), end.Format("2006-01-02")
		history := func(ctx context.Context, c worker.Company) error {
			return w.SyncPriceHistory(ctx, c, startDate, endDate)
		}
		if err := w.EachCompany(ctx, maxConcurrency, "price history", history); err != nil {
			return fmt.Errorf("sync price history: %w", err)
		}
	}

	if opts.ownership {
		if err := w.EachCompany(ctx, maxConcurrency, "ownership", w.SyncOwnership); err != nil {
			return fmt.Errorf("sync ownership: %w", err)
		}
	}

	if opts.corporateActions {
		if err := w.EachCompany(ctx, maxConcurrency, "dividends", w.SyncDividends); err != nil {
			return fmt.Errorf("sync dividends: %w", err)
		}
	}

	slog.Info("backfill complete", "duration", time.Since(start))
	return nil
}
