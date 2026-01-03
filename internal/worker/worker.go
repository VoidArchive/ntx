// Package worker provides background sync jobs for NEPSE data.
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/market"
	"github.com/voidarchive/ntx/internal/nepse"
)

// Worker syncs data from NEPSE to the local database.
type Worker struct {
	nepse   *nepse.Client
	queries *sqlc.Queries
	market  *market.Market

	// Daily state (reset each trading day)
	currentDate   string
	companySynced bool
	snapshotDone  bool
	lastPriceSync time.Time
}

// New creates a new Worker.
func New(client *nepse.Client, queries *sqlc.Queries, mkt *market.Market) *Worker {
	return &Worker{
		nepse:   client,
		queries: queries,
		market:  mkt,
	}
}

// Run starts the worker loop. Blocks until context is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	slog.Info("worker started")

	for {
		if err := w.tick(ctx); err != nil {
			return err
		}

		sleep := w.nextSleep(ctx)
		slog.Debug("worker sleeping", "duration", sleep)

		if err := sleepCtx(ctx, sleep); err != nil {
			return err
		}
	}
}

// tick performs one iteration of the worker loop.
func (w *Worker) tick(ctx context.Context) error {
	now := time.Now().In(market.NPT)
	today := now.Format(market.DateFormat)

	// Reset daily state on new day
	if today != w.currentDate {
		w.currentDate = today
		w.companySynced = false
		w.snapshotDone = false
		slog.Info("new trading day", "date", today)
	}

	// Company sync once per day (do it early, before market opens)
	if !w.companySynced {
		if err := w.syncCompanies(ctx); err != nil {
			slog.Error("company sync failed", "error", err)
			// Continue - don't block other jobs
		} else {
			w.companySynced = true
		}
	}

	// Price sync during market hours
	if w.market.IsOpen(ctx) {
		if err := w.syncPrices(ctx); err != nil {
			slog.Error("price sync failed", "error", err)
		} else {
			w.lastPriceSync = now
		}
		return nil
	}

	// Final snapshot after market close (15:00+)
	hour := now.Hour()
	if hour >= market.CloseHour && !w.snapshotDone && w.market.IsTradingDay(ctx, now) {
		if err := w.finalSnapshot(ctx); err != nil {
			slog.Error("final snapshot failed", "error", err)
		} else {
			w.snapshotDone = true
		}
	}

	return nil
}

// nextSleep calculates how long to sleep until the next action.
func (w *Worker) nextSleep(ctx context.Context) time.Duration {
	now := time.Now().In(market.NPT)

	// During market hours: sync every minute
	if w.market.IsOpen(ctx) {
		return time.Minute
	}

	hour := now.Hour()

	// After market close but before snapshot done: short sleep to retry
	if hour >= market.CloseHour && !w.snapshotDone && w.market.IsTradingDay(ctx, now) {
		return 30 * time.Second
	}

	// Before market open: sleep until open
	if hour < market.OpenHour {
		open := time.Date(now.Year(), now.Month(), now.Day(), market.OpenHour, 0, 0, 0, market.NPT)
		return time.Until(open)
	}

	// After everything done: sleep until next trading day
	nextOpen := w.market.NextOpenFrom(ctx, now)
	return time.Until(nextOpen)
}

// sleepCtx sleeps for the given duration, returning early if context is cancelled.
func sleepCtx(ctx context.Context, d time.Duration) error {
	// Cap sleep to 1 hour to allow periodic health checks
	if d > time.Hour {
		d = time.Hour
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
