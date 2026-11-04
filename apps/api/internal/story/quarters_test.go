package story

import (
	"database/sql"
	"math"
	"slices"
	"testing"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func TestNabilQuarters(t *testing.T) {
	qs, hydro := Quarters(Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows)})
	if hydro != nil {
		t.Errorf("a bank has no hydro EPS trio: %+v", hydro)
	}
	// 2078/79 Q4 is dropped: its Q3 is missing, so it has no quarter-only profit.
	if len(qs) != 17 || qs[0].Period != "Q1 2079/80" || qs[16].Period != "Q1 2083/84" {
		t.Fatalf("quarters oldest to newest, got %d from %q to %q", len(qs), qs[0].Period, qs[len(qs)-1].Period)
	}
	last := qs[16]
	if last.FiscalYear != "2083/84" || math.Abs(*last.Profit-2555105000) > 0.5 ||
		math.Abs(*last.EPS-9.44) > 0.005 || *last.AnnualizedEPS != 33.89 || last.Monsoon {
		t.Errorf("latest quarter = %+v", last)
	}
	labels := map[string]string{}
	for _, q := range qs {
		if q.Label != "" {
			labels[q.Period] = q.Label
		}
	}
	want := map[string]string{
		"Q1 2081/82": "Best Q1 in three years",
		"Q3 2081/82": "Best Q3 in three years",
		// Also the second rise in a row; the record wins.
		"Q3 2082/83": "Best Q3 in four years",
		"Q1 2083/84": "Best Q1 in five years",
	}
	if len(labels) != len(want) {
		t.Errorf("labels = %q, want %q", labels, want)
	}
	for p, l := range want {
		if labels[p] != l {
			t.Errorf("label %s = %q, want %q", p, labels[p], l)
		}
	}
}

func TestStreakLabel(t *testing.T) {
	const paid = 1e9
	qs, _ := Quarters(Input{Company: nabil, Fundamentals: fundamentals(131, []fixtureRow{
		{"2020-2021", "First Quarter", 1, 10, 100, paid, 5e7},
		{"2021-2022", "First Quarter", 1, 10, 100, paid, 1e7},
		{"2022-2023", "First Quarter", 1, 10, 100, paid, 2e7},
		{"2023-2024", "First Quarter", 1, 10, 100, paid, 3e7},
		{"2024-2025", "First Quarter", 1, 10, 100, paid, 4e7},
	})})
	got := make([]string, 0, len(qs))
	for _, q := range qs {
		got = append(got, q.Label)
	}
	if want := []string{"", "", "", "2nd rise in a row", "3rd rise in a row"}; !slices.Equal(got, want) {
		t.Errorf("labels = %q, want %q", got, want)
	}
}

func TestHydroQuarters(t *testing.T) {
	qs, hydro := Quarters(
		Input{Company: api, Fundamentals: fundamentals(697, asOf(apiRows, "2025-2026", "First Quarter"))},
	)
	last := qs[len(qs)-1]
	if !last.Monsoon || qs[len(qs)-2].Monsoon {
		t.Errorf("only Q1 is the monsoon quarter: %+v, %+v", last, qs[len(qs)-2])
	}
	if hydro == nil || math.Abs(hydro.Quarter-4.32) > 0.005 || hydro.Annualized != 17.3 || hydro.FullYear != 7.04 ||
		hydro.FullYearPeriod != "FY 2081/82" {
		t.Errorf("hydro EPS = %+v", hydro)
	}
}

func TestEvents(t *testing.T) {
	str := func(s string) sql.NullString { return sql.NullString{String: s, Valid: s != ""} }
	rows := []sqlc.Fundamental{
		{FiscalYear: "2025-2026", Quarter: "Second Quarter", PublishedAt: str("2026-02-10T12:30:00")},
		{FiscalYear: "2025-2026", Quarter: "First Quarter", PublishedAt: str("2025-11-12")},
		{FiscalYear: "2024-2025", Quarter: "", PublishedAt: str("2025-12-01")},
		{FiscalYear: "2025-2026", Quarter: "Third Quarter"}, // not recorded: no mark
	}
	actions := []sqlc.CorporateAction{
		{FiscalYear: "2081/82", CashDividend: num(10), BonusPercentage: num(5), SubmittedDate: str("2025-12-20")},
		{FiscalYear: "2080/81", BonusPercentage: num(0), SubmittedDate: str("2024-12-20")},
		{FiscalYear: "2079/80", CashDividend: num(8)}, // no date: no mark
	}
	got := Events(rows, actions)
	want := []EventMark{
		{Date: "2024-12-20", Kind: "notice", Label: "Notice"},
		{Date: "2025-11-12", Kind: "result", Label: "Q1"},
		{Date: "2025-12-01", Kind: "result", Label: "FY"},
		{Date: "2025-12-20", Kind: "notice", Label: "10% cash, 5% bonus"},
		{Date: "2026-02-10", Kind: "result", Label: "Q2"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("events:\n got %+v\nwant %+v", got, want)
	}
}
