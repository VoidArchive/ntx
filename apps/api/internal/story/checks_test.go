package story

import (
	"maps"
	"testing"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
)

func verdict(s Story, name string) string {
	for _, c := range s.Checks {
		if c.Name == name {
			return c.Verdict
		}
	}
	return "missing"
}

// nabilFull is NABIL as of 2026-11-03 with its notices.
func nabilFull() Input {
	return Input{
		Company:      nabil,
		Fundamentals: fundamentals(131, nabilRows),
		Prices:       []sqlc.Price{price("2026-11-03", 552)},
		Sector:       bankSector,
		Actions:      nabilNotices,
		Today:        "2026-11-03",
	}
}

func TestChecks(t *testing.T) {
	// Only FY 2082/83 filed: no earlier years, quarters or sector to compare with.
	short := Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows[18:22])}
	notRecorded := nabilFull()
	notRecorded.Actions = nil
	cheap := nabilFull()
	cheap.Prices = []sqlc.Price{price("2026-11-03", 300)}
	lowCash := nabilFull()
	lowCash.Actions = []sqlc.CorporateAction{
		notice("2025-2026", 0, 5, 0, ""), notice("2024-2025", 0, 11, 0, ""), notice("2023-2024", 0, 9.5, 0, ""),
	}
	gap := nabilFull()
	gap.Actions = []sqlc.CorporateAction{
		notice("2025-2026", 10, 12.5, 0, ""), notice("2024-2025", 0, 0, 0, ""), notice("2023-2024", 0, 9.5, 0, ""),
	}
	// History rows are changes, so the newest row is the last move or the first recording.
	steady := nabilFull()
	steady.Ownership = []sqlc.OwnershipHistory{observation("2025-10-01", 51, 49)}
	moved := nabilFull()
	moved.Ownership = []sqlc.OwnershipHistory{observation("2025-10-01", 60, 40), observation("2026-05-01", 51, 49)}
	oldMove := nabilFull()
	oldMove.Ownership = []sqlc.OwnershipHistory{observation("2024-01-01", 60, 40), observation("2025-06-01", 51, 49)}
	fresh := nabilFull()
	fresh.Ownership = []sqlc.OwnershipHistory{observation("2026-11-01", 51, 49)}
	dear := nabilFull()
	dear.Prices = []sqlc.Price{price("2026-11-03", 1200)}
	// The latest filing with EPS 40 on the same book value: 17.1% return on equity.
	rich := nabilFull()
	richRows := append([]fixtureRow{}, nabilRows[:22]...)
	richRows = append(richRows, fixtureRow{"2026-2027", "First Quarter", 40, 13.8, 233.9, 27056996729.0, 2555105000.0})
	rich.Fundamentals = fundamentals(131, richRows)
	noBook := nabilFull()
	noBookRows := append([]fixtureRow{}, nabilRows[:22]...)
	noBookRows = append(
		noBookRows,
		fixtureRow{"2026-2027", "First Quarter", 33.89, 16.14, 0, 27056996729.0, 2555105000.0},
	)
	noBook.Fundamentals = fundamentals(131, noBookRows)

	tests := []struct {
		check string
		in    Input
		want  string
	}{
		// EPS 18.64, 23.67, 22.9, 21.89, 28.36: two rises in four.
		{"EPS growth", nabilFull(), Watch},
		{"EPS growth", short, NeedsHistory},
		// Q1 2083/84 Rs 255.51 crore against Rs 175.70 crore.
		{"Quarterly profit", nabilFull(), Holds},
		{"Quarterly profit", short, NeedsHistory},
		// 17.16x against a three-year median of 18.18x, both on trailing EPS.
		{"PE against its own history", nabilFull(), Holds},
		{"PE against its own history", short, NeedsHistory},
		// 2.36x against a sector median of 1.56x.
		{"PB against the sector", nabilFull(), Watch},
		{"PB against the sector", cheap, Holds},
		{"PB against the sector", short, NeedsHistory},
		// 17.16x against a sector median of 18.52x; at Rs 1,200 it is 37.3x.
		{"PE against the sector", nabilFull(), Holds},
		{"PE against the sector", dear, Watch},
		{"PE against the sector", short, NeedsHistory},
		// EPS 33.89 on book value 233.90 is 14.5%, under the 15% floor.
		{"Return on equity", nabilFull(), Watch},
		{"Return on equity", rich, Holds},
		{"Return on equity", noBook, NeedsHistory},
		{"Dividend record", nabilFull(), Holds},
		{"Dividend record", gap, Watch},
		{"Dividend record", notRecorded, NeedsHistory},
		// 42.78% of FY 2082/83 profit.
		{"Cash payout", nabilFull(), Watch},
		{"Cash payout", lowCash, Holds},
		{"Cash payout", notRecorded, NeedsHistory},
		// Today is 2026-11-03: a year without a change holds, a move within the year is watch,
		// and a split first recorded days ago cannot be judged yet.
		{"Promoter holding", steady, Holds},
		{"Promoter holding", oldMove, Holds},
		{"Promoter holding", moved, Watch},
		{"Promoter holding", fresh, NeedsHistory},
		// 247.28 against 210.00 three years earlier.
		{"Book value growth", nabilFull(), Holds},
		{"Book value growth", short, NeedsHistory},
	}
	for _, tc := range tests {
		if got := verdict(Build(tc.in), tc.check); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.check, got, tc.want)
		}
	}
}

// A loss that shrank four years running rises every year and still fails.
func TestEPSGrowthLoss(t *testing.T) {
	var s Story
	for i, eps := range []float64{-1, -2, -3, -4, -5} {
		s.FiscalYears = append(s.FiscalYears, FiscalYear{Year: fiscal.Year(2082 - i), EPS: ptr(eps)})
	}
	if got := epsGrowth(s, Input{}); got != Watch {
		t.Errorf("shrinking loss = %q, want %q", got, Watch)
	}
}

// Nine holds with one check unevaluated is a B: an A means the whole record was seen.
func TestRatingNeverAWithMissingHistory(t *testing.T) {
	checks := []Check{{Name: "Promoter holding", Verdict: NeedsHistory}}
	for range 9 {
		checks = append(checks, Check{Verdict: Holds})
	}
	if got := rating(checks); got != (Rating{Letter: "B", Holds: 9, Evaluated: 9, Total: 10}) {
		t.Errorf("gated rating = %+v", got)
	}
}

func TestChecksAllPublished(t *testing.T) {
	s := Build(nabilFull())
	if len(s.Checks) != 10 {
		t.Fatalf("checks = %d, want 10", len(s.Checks))
	}
	axes := map[string]int{}
	for _, c := range s.Checks {
		if c.Name == "" || c.Rule == "" {
			t.Errorf("unnamed check %+v", c)
		}
		axes[c.Axis]++
	}
	// Every check sits on one snowflake axis; the counts are the page's section chips.
	if want := map[string]int{
		"value":     3,
		"past":      3,
		"health":    1,
		"dividend":  2,
		"ownership": 1,
	}; !maps.Equal(
		axes,
		want,
	) {
		t.Errorf("axes = %v, want %v", axes, want)
	}
}

func TestRating(t *testing.T) {
	// NABIL on 2026-11-03 with its notices: 5 hold, 4 watch, promoter holding needs history.
	got := Build(nabilFull()).Rating
	if got != (Rating{Letter: "C", Holds: 5, Evaluated: 9, Total: 10}) {
		t.Errorf("nabil rating = %+v", got)
	}
	short := Build(Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows[18:22])}).Rating
	if short.Letter != "" || short.Evaluated >= minEvaluated {
		t.Errorf("one filed year should be unrated, got %+v", short)
	}
	// API with filings only: two of four checks hold (12.6% return on equity is under the floor),
	// and without a close the valuation checks cannot evaluate, so two of ten is a D.
	if got := Build(
		Input{Company: api, Fundamentals: fundamentals(697, apiRows)},
	).Rating; got != (Rating{Letter: "D", Holds: 2, Evaluated: 4, Total: 10}) {
		t.Errorf("api rating = %+v", got)
	}

	// Letters on counts alone: holds and evaluated out of ten checks. A check
	// without history counts as not holding, so 5 of 5 evaluated is a C, not an A.
	for _, tc := range []struct {
		holds, evaluated int
		want             string
	}{
		{10, 10, "A"},
		{9, 10, "A"},
		{8, 10, "B"},
		{6, 10, "B"},
		{5, 10, "C"},
		{4, 10, "C"},
		{3, 10, "D"},
		{0, 10, "D"},
		{9, 9, "B"},
		{5, 5, "C"},
		{4, 4, "C"},
		{3, 4, "D"},
		{3, 3, ""},
	} {
		checks := make([]Check, 0, 10)
		for i := range tc.evaluated {
			checks = append(checks, Check{Verdict: verdictOf(i < tc.holds)})
		}
		for len(checks) < 10 {
			checks = append(checks, Check{Verdict: NeedsHistory})
		}
		if got := rating(checks).Letter; got != tc.want {
			t.Errorf("%d of %d: got %q, want %q", tc.holds, tc.evaluated, got, tc.want)
		}
	}
}
