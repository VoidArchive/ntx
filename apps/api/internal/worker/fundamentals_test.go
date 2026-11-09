package worker

import (
	"testing"

	"github.com/voidarchive/ntx/internal/nepse"
)

// Rows are GBIME's filings as NEPSE returned them on 2026-11-09, newest first.
func gbime(fy, q string, profit, paidUp, eps float64) nepse.Fundamental {
	return nepse.Fundamental{FiscalYear: fy, Quarter: q, ProfitAmount: profit, PaidUpCapital: paidUp, EPS: eps}
}

func TestNormalizeFundamentalsRescalesUnitsAndKeepsNewestFiling(t *testing.T) {
	in := []nepse.Fundamental{
		gbime("2026-2027", "First Quarter", 2313800000, 39640486720, 23.35), // rupees
		gbime("2025-2026", "", 6227156000, 38115852615.86, 16.34),           // corrected re-filing, rupees
		gbime("2025-2026", "", 6227156, 38115853, 16.34),                    // first filing, thousands
		gbime("2025-2026", "First Quarter", 1858650, 38115852615, 19.51),    // profit in thousands
		gbime("2022-2023", "Fourth Quarter", 7253455, 3577059, 20.72),       // both in thousands
		gbime("2022-2023", "Third Quarter", 4.36, 357.71, 17.08),            // both in crore
		gbime("2022-2023", "Second Quarter", 2782379221, 35771059655, 18.07),
	}
	got := normalizeFundamentals(in)
	want := []nepse.Fundamental{
		gbime("2026-2027", "First Quarter", 2313800000, 39640486720, 23.35),
		gbime("2025-2026", "", 6227156000, 38115852615.86, 16.34),
		gbime("2025-2026", "First Quarter", 1858650000, 38115852615, 19.51),
		gbime("2022-2023", "Fourth Quarter", 7253455000, 35770590000, 20.72),
		gbime("2022-2023", "Third Quarter", 4360000000, 35771000000, 17.08),
		gbime("2022-2023", "Second Quarter", 2782379221, 35771059655, 18.07),
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if !near(got[i].ProfitAmount, want[i].ProfitAmount) || !near(got[i].PaidUpCapital, want[i].PaidUpCapital) ||
			got[i].FiscalYear != want[i].FiscalYear ||
			got[i].Quarter != want[i].Quarter {
			t.Errorf("row %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestNormalizeFundamentalsLeavesDoubtfulRowsAlone(t *testing.T) {
	const cap = 36e9 // 360M shares
	in := []nepse.Fundamental{
		gbime("2025-2026", "First Quarter", 500e6, cap, 20),  // 3.6x under EPS: annualisation noise, kept
		gbime("2024-2025", "First Quarter", -100e6, cap, 20), // sign disagrees with EPS, kept
		gbime("2023-2024", "First Quarter", 100e6, cap, 20),  // 18x off: neither noise nor a unit, dropped
		gbime("2022-2023", "First Quarter", 1800e6, cap, 0),  // no EPS, kept
	}
	got := normalizeFundamentals(in)
	wantProfit := []float64{500e6, -100e6, 0, 1800e6}
	for i, w := range wantProfit {
		if !near(got[i].ProfitAmount, w) {
			t.Errorf("row %d profit = %v, want %v", i, got[i].ProfitAmount, w)
		}
	}
}

func TestNormalizeFundamentalsNeedsARupeeAnchor(t *testing.T) {
	in := []nepse.Fundamental{gbime("2025-2026", "First Quarter", 1858650, 38115853, 19.51)}
	if got := normalizeFundamentals(in); got[0].ProfitAmount != 1858650 || got[0].PaidUpCapital != 38115853 {
		t.Errorf("rescaled without an anchor: %+v", got[0])
	}
}

func near(a, b float64) bool { return a == b || (b != 0 && a/b > 0.999999 && a/b < 1.000001) }
