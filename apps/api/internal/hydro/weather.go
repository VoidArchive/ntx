package hydro

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// Source strings printed beside every figure (ADR 0002).
const (
	RainSource      = "Open-Meteo, ECMWF IFS"
	DischargeSource = "Open-Meteo, GloFAS v4"
)

const (
	recentDays       = 7  // each run re-reads a week so a missed morning fills itself
	rainBackfillDays = 92 // the forecast endpoint's maximum reach into the past
	dischargeYears   = 10 // the percentile window; older rows are never stored
)

// Nepal has no daylight saving, so a fixed zone avoids depending on tzdata.
var nepalTime = time.FixedZone("NPT", 5*3600+45*60)

// Weather reads Open-Meteo. The URLs are fields so tests can point them at a local server.
type Weather struct {
	HTTP        *http.Client
	ForecastURL string
	FloodURL    string
}

func NewWeather() *Weather {
	return &Weather{
		HTTP:        &http.Client{Timeout: 30 * time.Second},
		ForecastURL: "https://api.open-meteo.com/v1/forecast",
		FloodURL:    "https://flood-api.open-meteo.com/v1/flood",
	}
}

// daily is the slice of an Open-Meteo answer both endpoints share; a value is
// null when the model has no figure for that day.
type daily struct {
	Days      []string   `json:"time"`
	Rain      []*float64 `json:"precipitation_sum"`
	Discharge []*float64 `json:"river_discharge"`
}

type openMeteo struct {
	Daily  daily  `json:"daily"`
	Error  bool   `json:"error"`
	Reason string `json:"reason"`
}

func (w *Weather) get(ctx context.Context, base string, q url.Values) (daily, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"?"+q.Encode(), nil)
	if err != nil {
		return daily{}, err
	}
	res, err := w.HTTP.Do(req)
	if err != nil {
		return daily{}, err
	}
	defer res.Body.Close()
	var body openMeteo
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return daily{}, fmt.Errorf("decode %s: %w", base, err)
	}
	if body.Error || res.StatusCode != http.StatusOK {
		return daily{}, fmt.Errorf("%s: %d %s", base, res.StatusCode, body.Reason)
	}
	return body.Daily, nil
}

func point(lat, lon float64, start, end string) url.Values {
	return url.Values{
		"latitude":   {strconv.FormatFloat(lat, 'f', 4, 64)},
		"longitude":  {strconv.FormatFloat(lon, 'f', 4, 64)},
		"start_date": {start},
		"end_date":   {end},
		"timezone":   {"Asia/Kathmandu"},
	}
}

// Rain asks the forecast endpoint for ECMWF IFS daily precipitation, model named so the source can be.
func (w *Weather) Rain(ctx context.Context, lat, lon float64, start, end string) (daily, error) {
	q := point(lat, lon, start, end)
	q.Set("daily", "precipitation_sum")
	q.Set("models", "ecmwf_ifs025")
	return w.get(ctx, w.ForecastURL, q)
}

// Discharge asks the flood endpoint for GloFAS v4 daily river discharge (modelled, 5 km cells).
func (w *Weather) Discharge(ctx context.Context, lat, lon float64, start, end string) (daily, error) {
	q := point(lat, lon, start, end)
	q.Set("daily", "river_discharge")
	q.Set("models", "seamless_v4")
	return w.get(ctx, w.FloodURL, q)
}

// SyncReadings stores today's and past readings for every plant. A plant whose
// calls fail is logged and skipped; the next morning retries it.
func SyncReadings(ctx context.Context, q *sqlc.Queries, w *Weather, today time.Time) error {
	plants, err := q.ListAllHydroPlants(ctx)
	if err != nil {
		return fmt.Errorf("list plants: %w", err)
	}
	start := time.Now()
	for _, p := range plants {
		if err := syncPlant(ctx, q, w, p, today); err != nil {
			slog.Warn("skip weather", "symbol", p.Symbol, "plant", p.Plant, "err", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	slog.Info("weather sync finished", "plants", len(plants), "took", time.Since(start))
	return nil
}

// syncPlant reads a week of both figures, or on a plant's first run 92 days of
// rain and ten years of discharge, and stores only days up to today.
func syncPlant(ctx context.Context, q *sqlc.Queries, w *Weather, p sqlc.HydroPlant, today time.Time) error {
	n, err := q.CountDischargeReadings(ctx, p.ID)
	if err != nil {
		return err
	}
	end := today.Format(time.DateOnly)
	rainStart := today.AddDate(0, 0, -recentDays)
	dischargeStart := rainStart
	if n == 0 {
		rainStart = today.AddDate(0, 0, -rainBackfillDays)
		dischargeStart = today.AddDate(-dischargeYears, 0, 0)
	}
	floor := today.AddDate(-dischargeYears, 0, 0).Format(time.DateOnly)

	rain, err := w.Rain(ctx, p.Lat, p.Lon, rainStart.Format(time.DateOnly), end)
	if err != nil {
		return err
	}
	err = store(ctx, rain.Days, rain.Rain, floor, end, func(day string, v float64) error {
		return q.UpsertRainReading(ctx, sqlc.UpsertRainReadingParams{
			PlantID: p.ID, Day: day, RainMm: nullFloat(v), RainSource: nullString(RainSource),
		})
	})
	if err != nil {
		return err
	}

	discharge, err := w.Discharge(ctx, p.Lat, p.Lon, dischargeStart.Format(time.DateOnly), end)
	if err != nil {
		return err
	}
	return store(ctx, discharge.Days, discharge.Discharge, floor, end, func(day string, v float64) error {
		return q.UpsertDischargeReading(ctx, sqlc.UpsertDischargeReadingParams{
			PlantID: p.ID, Day: day, DischargeM3s: nullFloat(v), DischargeSource: nullString(DischargeSource),
		})
	})
}

// store writes one value per day inside [floor, end]. A later day is a forecast
// and is discarded; an earlier one is outside the ten-year window.
func store(ctx context.Context, days []string, values []*float64, floor, end string,
	upsert func(day string, v float64) error,
) error {
	if len(days) != len(values) {
		return errors.New("open-meteo answer has mismatched day and value counts")
	}
	for i, day := range days {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if values[i] == nil || day > end || day < floor {
			continue
		}
		if err := upsert(day, *values[i]); err != nil {
			return fmt.Errorf("store %s: %w", day, err)
		}
	}
	return nil
}
