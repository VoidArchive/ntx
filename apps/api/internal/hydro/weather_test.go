package hydro

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

var today = time.Date(2026, 11, 5, 6, 0, 0, 0, nepalTime)

func day(offsetDays int) string { return today.AddDate(0, 0, offsetDays).Format(time.DateOnly) }

// fakeOpenMeteo answers both endpoints and records each call's start_date.
func fakeOpenMeteo(t *testing.T, starts *[]string) *Weather {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*starts = append(*starts, r.URL.Path+" "+r.URL.Query().Get("start_date"))
		var body string
		switch {
		case strings.HasSuffix(r.URL.Path, "/forecast"):
			if r.URL.Query().Get("models") != "ecmwf_ifs025" {
				t.Errorf("rain must name the model, got %q", r.URL.Query().Get("models"))
			}
			// Yesterday, today and tomorrow's forecast; one day without a figure.
			body = fmt.Sprintf(`{"daily":{"time":["%s","%s","%s","%s"],"precipitation_sum":[12.5,null,61,3]}}`,
				day(-2), day(-1), day(0), day(1))
		case strings.HasSuffix(r.URL.Path, "/flood"):
			old := today.AddDate(-10, 0, -1).Format(time.DateOnly)
			body = fmt.Sprintf(`{"daily":{"time":["%s","%s","%s","%s"],"river_discharge":[500,80,130,900]}}`,
				old, day(-1), day(0), day(1))
		default:
			body = `{"error":true,"reason":"unknown endpoint"}`
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &Weather{HTTP: srv.Client(), ForecastURL: srv.URL + "/forecast", FloodURL: srv.URL + "/flood"}
}

func seedPlant(t *testing.T, q *sqlc.Queries) {
	t.Helper()
	if err := q.UpsertHydroPlant(context.Background(), PlantRow{
		Symbol: "CHCL", Plant: "Chilime", River: "Chilime", Basin: "Gandaki", District: "Rasuwa",
		Mw: 22, Lat: 28.1698, Lon: 85.3196, Status: "operating", SourceUrl: "https://doed.gov.np",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSyncReadingsStoresOnlyPastDaysInsideTheWindow(t *testing.T) {
	ctx := context.Background()
	q := database.OpenTestDB(t)
	seedPlant(t, q)
	var starts []string
	w := fakeOpenMeteo(t, &starts)

	if err := SyncReadings(ctx, q, w, today); err != nil {
		t.Fatal(err)
	}
	rows, err := q.ListRecentReadings(ctx, sqlc.ListRecentReadingsParams{PlantID: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(rows))
	for _, r := range rows {
		b, _ := json.Marshal(map[string]any{
			"day":       r.Day,
			"rain":      r.RainMm,
			"discharge": r.DischargeM3s,
			"rs":        r.RainSource.String,
			"ds":        r.DischargeSource.String,
		})
		got = append(got, string(b))
	}
	// Tomorrow is a forecast and the eleven-year-old row is outside the window: neither is stored.
	// Yesterday's rain was null, so the row has discharge only.
	want := []string{
		fmt.Sprintf(
			`{"day":"%s","discharge":{"Float64":130,"Valid":true},"ds":"Open-Meteo, GloFAS v4","rain":{"Float64":61,"Valid":true},"rs":"Open-Meteo, ECMWF IFS"}`,
			day(0),
		),
		fmt.Sprintf(
			`{"day":"%s","discharge":{"Float64":80,"Valid":true},"ds":"Open-Meteo, GloFAS v4","rain":{"Float64":0,"Valid":false},"rs":""}`,
			day(-1),
		),
		fmt.Sprintf(
			`{"day":"%s","discharge":{"Float64":0,"Valid":false},"ds":"","rain":{"Float64":12.5,"Valid":true},"rs":"Open-Meteo, ECMWF IFS"}`,
			day(-2),
		),
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	// First run backfilled 92 days of rain and ten years of discharge; the second reads a week.
	if err := SyncReadings(ctx, q, w, today); err != nil {
		t.Fatal(err)
	}
	wantStarts := []string{
		"/forecast " + day(-92), "/flood " + today.AddDate(-10, 0, 0).Format(time.DateOnly),
		"/forecast " + day(-7), "/flood " + day(-7),
	}
	if strings.Join(starts, ",") != strings.Join(wantStarts, ",") {
		t.Fatalf("calls %v, want %v", starts, wantStarts)
	}
}

func TestSyncReadingsSkipsAFailedPlant(t *testing.T) {
	ctx := context.Background()
	q := database.OpenTestDB(t)
	seedPlant(t, q)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":true,"reason":"limit"}`))
	}))
	defer srv.Close()
	w := &Weather{HTTP: srv.Client(), ForecastURL: srv.URL + "/forecast", FloodURL: srv.URL + "/flood"}
	if err := SyncReadings(ctx, q, w, today); err != nil {
		t.Fatalf("one failed plant must not fail the run: %v", err)
	}
	n, err := q.CountDischargeReadings(ctx, 1)
	if err != nil || n != 0 {
		t.Fatalf("stored %d rows (err %v)", n, err)
	}
}
