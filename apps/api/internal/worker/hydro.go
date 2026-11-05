package worker

import (
	"context"
	"time"

	"github.com/voidarchive/ntx/internal/hydro"
)

// SyncWeather stores today's rain and discharge for every hydro plant; the
// plant's first run backfills its history (see hydro.SyncReadings).
func (w *Worker) SyncWeather(ctx context.Context) error {
	return hydro.SyncReadings(ctx, w.queries, hydro.NewWeather(), time.Now().In(nepalTime))
}
