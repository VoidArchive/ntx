package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	c      *cron.Cron
	worker *Worker
}

func NewScheduler(worker *Worker) (*Scheduler, error) {
	loc, err := time.LoadLocation("Asia/Kathmandu")
	if err != nil {
		return nil, err
	}

	c := cron.New(
		cron.WithLocation(loc),
		cron.WithSeconds(),
	)
	return &Scheduler{c: c, worker: worker}, nil
}

func (s *Scheduler) Start(ctx context.Context) error {
	spec := "0 5 15 * * 0-4"

	_, err := s.c.AddFunc(spec, func() {
		jobCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()

		start := time.Now()
		slog.Info("companies sync started", slog.Time("start", start))
		if err := s.worker.SyncCompanies(jobCtx); err != nil {
			slog.Error("companies sync failed", slog.Any("err", err))
			return
		}
		slog.Info("companies sync finished", slog.Duration("took", time.Since(start)))

		// Sync fundamentals after companies
		start = time.Now()
		slog.Info("fundamentals sync started", slog.Time("start", start))
		if err := s.worker.SyncFundamentals(jobCtx); err != nil {
			slog.Error("fundamentals sync failed", slog.Any("err", err))
			return
		}
		slog.Info("fundamentals sync finished", slog.Duration("took", time.Since(start)))

		// Sync prices
		start = time.Now()
		loc, _ := time.LoadLocation("Asia/Kathmandu")
		businessDate := time.Now().In(loc).Format("2006-01-02")
		slog.Info("prices sync started", slog.Time("start", start), slog.String("date", businessDate))
		if err := s.worker.SyncPrices(jobCtx, businessDate); err != nil {
			slog.Error("prices sync failed", slog.Any("err", err))
			return
		}
		slog.Info("prices sync finished", slog.Duration("took", time.Since(start)))
	})
	if err != nil {
		return err
	}
	s.c.Start()
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	stopCtx := s.c.Stop()
	select {
	case <-stopCtx.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
