package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
)

// SyncMarketDay records businessDate's session summary. It runs after
// SyncPrices, whose rows supply advances, declines and unchanged.
func (w *Worker) SyncMarketDay(ctx context.Context, businessDate string) error {
	day, err := w.nepse.MarketDay(ctx)
	if err != nil {
		return err
	}
	return w.storeMarketDay(ctx, businessDate, day)
}

// storeMarketDay writes nothing when businessDate has no price rows: that is a
// holiday, and the feed's figures still belong to the previous session.
func (w *Worker) storeMarketDay(ctx context.Context, businessDate string, day nepse.MarketDay) error {
	breadth, err := w.queries.CountBreadth(ctx, businessDate)
	if err != nil {
		return fmt.Errorf("count breadth: %w", err)
	}
	if breadth.Priced == 0 {
		slog.Info("market day skipped: no price rows", "date", businessDate)
		return nil
	}
	return w.queries.UpsertMarketDay(ctx, sqlc.UpsertMarketDayParams{
		BusinessDate:       businessDate,
		IndexClose:         day.IndexClose,
		IndexChange:        day.IndexChange,
		IndexChangePercent: day.IndexChangePercent,
		Turnover:           day.Turnover,
		TradedShares:       day.TradedShares,
		Transactions:       day.Transactions,
		ScripsTraded:       day.ScripsTraded,
		MarketCap:          day.MarketCap,
		Advances:           breadth.Advances,
		Declines:           breadth.Declines,
		Unchanged:          breadth.Unchanged,
	})
}
