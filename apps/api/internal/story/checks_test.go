package story

import (
	"testing"

	"github.com/voidarchive/ntx/internal/database/sqlc"
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
	steady := nabilFull()
	steady.Ownership = []sqlc.OwnershipHistory{
		observation("2026-02-01", 51, 49), observation("2026-05-01", 51, 49),
		observation("2026-08-01", 51, 49), observation("2026-11-01", 51, 49),
	}
	moved := nabilFull()
	moved.Ownership = append([]sqlc.OwnershipHistory{}, steady.Ownership...)
	moved.Ownership[0] = observation("2026-02-01", 60, 40)
	three := nabilFull()
	three.Ownership = steady.Ownership[1:]

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
		// 16.29x against a three-year median of 19.71x.
		{"PE against its own history", nabilFull(), Holds},
		{"PE against its own history", short, NeedsHistory},
		// 2.36x against a sector median of 1.56x.
		{"PB against the sector", nabilFull(), Watch},
		{"PB against the sector", cheap, Holds},
		{"PB against the sector", short, NeedsHistory},
		{"Dividend record", nabilFull(), Holds},
		{"Dividend record", gap, Watch},
		{"Dividend record", notRecorded, NeedsHistory},
		// 42.78% of FY 2082/83 profit.
		{"Cash payout", nabilFull(), Watch},
		{"Cash payout", lowCash, Holds},
		{"Cash payout", notRecorded, NeedsHistory},
		{"Promoter holding", steady, Holds},
		{"Promoter holding", moved, Watch},
		{"Promoter holding", three, NeedsHistory},
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

func TestChecksAllPublished(t *testing.T) {
	s := Build(nabilFull())
	if len(s.Checks) != 8 {
		t.Fatalf("checks = %d, want 8", len(s.Checks))
	}
	for _, c := range s.Checks {
		if c.Name == "" || c.Rule == "" {
			t.Errorf("unnamed check %+v", c)
		}
	}
}

func TestRating(t *testing.T) {
	// NABIL on 2026-11-03 with its notices: 4 hold, 3 watch, promoter holding needs history.
	got := Build(nabilFull()).Rating
	if got != (Rating{Letter: "C", Holds: 4, Evaluated: 7}) {
		t.Errorf("nabil rating = %+v", got)
	}
	short := Build(Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows[18:22])}).Rating
	if short.Letter != "" || short.Evaluated >= minEvaluated {
		t.Errorf("one filed year should be unrated, got %+v", short)
	}

	// Letters on counts alone: holds, evaluated.
	for _, tc := range []struct {
		holds, evaluated int
		want             string
	}{
		{8, 8, "A"},
		{7, 8, "A"},
		{6, 8, "B"},
		{5, 8, "B"},
		{4, 8, "C"},
		{3, 8, "C"},
		{2, 8, "D"},
		{0, 8, "D"},
		{4, 4, "A"},
		{3, 4, "B"},
		{2, 4, "C"},
		{1, 4, "D"},
		{5, 5, "A"},
		{3, 5, "C"},
		{3, 3, ""},
	} {
		checks := make([]Check, 0, len(rules))
		for i := range tc.evaluated {
			checks = append(checks, Check{Verdict: verdictOf(i < tc.holds)})
		}
		for len(checks) < len(rules) {
			checks = append(checks, Check{Verdict: NeedsHistory})
		}
		if got := rating(checks).Letter; got != tc.want {
			t.Errorf("%d of %d: got %q, want %q", tc.holds, tc.evaluated, got, tc.want)
		}
	}
}
