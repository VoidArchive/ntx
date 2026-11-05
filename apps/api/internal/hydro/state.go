package hydro

import (
	"time"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// HeavyRainMM is the Department of Hydrology and Meteorology's 24-hour "heavy" category.
const HeavyRainMM = 50.0

// StaleAfter is how old the latest reading may be before the page labels it stale.
const StaleAfter = 2 * 24 * time.Hour

// Rules are printed on the page beside the states they produce, as the checks are.
var Rules = []*ntxv1.Rule{
	{
		Name: "Heavy rain",
		Text: "The day's rain at the plant is 50 mm or more, the Department of Hydrology and Meteorology's " +
			"24-hour \"heavy\" category (DHM forecast terminology).",
	},
	{
		Name: "High discharge",
		Text: "The day's modelled discharge at the plant is above the 95th percentile of the plant's own " +
			"stored daily discharge over the last ten years. Discharge is modelled by GloFAS, not gauged.",
	},
}

// PlantState applies the two rules to the latest reading. High discharge is the
// worse state and wins when both hold; a plant without a percentile yet cannot be
// high discharge.
func PlantState(r sqlc.HydroReading, p95 float64, hasP95 bool) ntxv1.PlantState {
	if hasP95 && r.DischargeM3s.Valid && r.DischargeM3s.Float64 > p95 {
		return ntxv1.PlantState_PLANT_STATE_HIGH_DISCHARGE
	}
	if r.RainMm.Valid && r.RainMm.Float64 >= HeavyRainMM {
		return ntxv1.PlantState_PLANT_STATE_HEAVY_RAIN
	}
	return ntxv1.PlantState_PLANT_STATE_NORMAL
}

// Stale reports whether a reading day (YYYY-MM-DD) is more than two days before today.
func Stale(day string, today time.Time) bool {
	d, err := time.Parse(time.DateOnly, day)
	if err != nil {
		return true
	}
	midnight := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	return midnight.Sub(d) > StaleAfter
}

// BasinState is the worst state among a basin's plants; the enum orders states by severity.
func BasinState(states []ntxv1.PlantState) ntxv1.PlantState {
	worst := ntxv1.PlantState_PLANT_STATE_UNSPECIFIED
	for _, s := range states {
		worst = max(worst, s)
	}
	return worst
}
