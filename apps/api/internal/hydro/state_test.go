package hydro

import (
	"database/sql"
	"testing"
	"time"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func reading(rain, discharge float64) sqlc.HydroReading {
	r := sqlc.HydroReading{Day: "2026-11-05"}
	if rain >= 0 {
		r.RainMm = sql.NullFloat64{Float64: rain, Valid: true}
	}
	if discharge >= 0 {
		r.DischargeM3s = sql.NullFloat64{Float64: discharge, Valid: true}
	}
	return r
}

func TestPlantState(t *testing.T) {
	const p95 = 120.0
	cases := []struct {
		name            string
		rain, discharge float64
		hasP95          bool
		want            ntxv1.PlantState
	}{
		{"normal", 12, 80, true, ntxv1.PlantState_PLANT_STATE_NORMAL},
		{"heavy rain at the threshold", 50, 80, true, ntxv1.PlantState_PLANT_STATE_HEAVY_RAIN},
		{"high discharge above p95", 0, 120.5, true, ntxv1.PlantState_PLANT_STATE_HIGH_DISCHARGE},
		{"at p95 is not high", 0, 120, true, ntxv1.PlantState_PLANT_STATE_NORMAL},
		{"both: discharge wins", 90, 200, true, ntxv1.PlantState_PLANT_STATE_HIGH_DISCHARGE},
		{"no percentile yet", 0, 200, false, ntxv1.PlantState_PLANT_STATE_NORMAL},
		{"no readings", -1, -1, true, ntxv1.PlantState_PLANT_STATE_NORMAL},
	}
	for _, c := range cases {
		if got := PlantState(reading(c.rain, c.discharge), p95, c.hasP95); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestStale(t *testing.T) {
	today := time.Date(2026, 11, 5, 9, 30, 0, 0, time.UTC)
	for day, want := range map[string]bool{
		"2026-11-05": false, "2026-11-04": false, "2026-11-03": false, "2026-11-02": true, "bad": true,
	} {
		if got := Stale(day, today); got != want {
			t.Errorf("%s: stale %v, want %v", day, got, want)
		}
	}
}

func TestBasinStateIsWorstOf(t *testing.T) {
	got := BasinState([]ntxv1.PlantState{
		ntxv1.PlantState_PLANT_STATE_NORMAL,
		ntxv1.PlantState_PLANT_STATE_HIGH_DISCHARGE,
		ntxv1.PlantState_PLANT_STATE_HEAVY_RAIN,
	})
	if got != ntxv1.PlantState_PLANT_STATE_HIGH_DISCHARGE {
		t.Fatalf("got %v", got)
	}
	if BasinState(nil) != ntxv1.PlantState_PLANT_STATE_UNSPECIFIED {
		t.Fatal("an empty basin has no state")
	}
}
