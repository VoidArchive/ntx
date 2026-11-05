package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// Nepal has no daylight saving, so a fixed zone avoids depending on tzdata.
var nepalTime = time.FixedZone("NPT", 5*3600+45*60)

type Scheduler struct {
	c      *cron.Cron
	worker *Worker
}

func NewScheduler(worker *Worker) (*Scheduler, error) {
	c := cron.New(
		cron.WithLocation(nepalTime),
		cron.WithSeconds(),
	)
	return &Scheduler{c: c, worker: worker}, nil
}

// step is one unit of a scheduled job with its own time budget.
type step struct {
	name    string
	timeout time.Duration
	run     func(context.Context) error
}

// runSteps runs every step even when an earlier one fails or times out:
// each has its own deadline so a slow step cannot starve the ones after it.
func runSteps(ctx context.Context, steps []step) {
	for _, s := range steps {
		stepCtx, cancel := context.WithTimeout(ctx, s.timeout)
		start := time.Now()
		slog.Info("sync step started", "step", s.name, "timeout", s.timeout)
		err := s.run(stepCtx)
		cancel()
		if err != nil {
			slog.Error("sync step failed", "step", s.name, "took", time.Since(start), "err", err)
			continue
		}
		slog.Info("sync step finished", "step", s.name, "took", time.Since(start))
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	w := s.worker

	// NEPSE moved to a Mon-Fri week in April 2026; holidays are caught by SyncPrices via the market status.
	// Ownership and dividends change slowly, so they refresh before the 11:00 open.
	_, err := s.c.AddFunc("0 0 10 * * 1-5", func() {
		runSteps(ctx, []step{
			{"ownership", 30 * time.Minute, func(ctx context.Context) error {
				return w.EachCompany(ctx, 1, "ownership", w.SyncOwnership)
			}},
			{"dividends", 30 * time.Minute, func(ctx context.Context) error {
				return w.EachCompany(ctx, 1, "dividends", w.SyncDividends)
			}},
		})
	})
	if err != nil {
		return err
	}

	// Prices go first: they are the one thing that must land every trading day.
	_, err = s.c.AddFunc("0 5 15 * * 1-5", func() {
		businessDate := time.Now().In(nepalTime).Format("2006-01-02")
		runSteps(ctx, []step{
			{"prices", 5 * time.Minute, func(ctx context.Context) error {
				return w.SyncPrices(ctx, businessDate)
			}},
			{"market day", 2 * time.Minute, func(ctx context.Context) error {
				return w.SyncMarketDay(ctx, businessDate)
			}},
			{"companies", 5 * time.Minute, w.SyncCompanies},
			{"fundamentals", 30 * time.Minute, func(ctx context.Context) error {
				return w.EachCompany(ctx, 1, "fundamentals", w.SyncFundamentals)
			}},
			// Last, so the letters read the day's closes and any new filing.
			{"ratings", 10 * time.Minute, w.SyncRatings},
		})
	})
	if err != nil {
		return err
	}

	// Weather every day of the week, holidays included: rivers do not observe them.
	// A plant's first run writes ten years of discharge row by row, so the budget
	// covers a whole fresh table; the daily run takes about a minute.
	_, err = s.c.AddFunc("0 0 6 * * *", func() {
		runSteps(ctx, []step{{"weather", 3 * time.Hour, w.SyncWeather}})
	})
	if err != nil {
		return err
	}
	s.c.Start()
	return nil
}
