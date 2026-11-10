package portfolio

import (
	"math"
	"testing"
	"time"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func build(symbol string, allocation, carry, close float64) *ntxv1.Holding {
	return &ntxv1.Holding{
		Symbol: symbol, Status: ntxv1.HoldingStatus_HOLDING_STATUS_BUILD, AllocationPercent: allocation,
		CarryRs: carry, Close: &close, CloseDay: "2026-11-09", Quantity: 10, CapVerdict: Holds,
	}
}

var defaults = sqlc.Rule{
	SoftCapPercent: 20, HardCapPercent: 25, SectorCapPercent: 40, MaxPositions: 8, MinEntryRs: 15000,
	MonthlyBudgetRs: 12000, MultiplierBelowMedian: 1.5, MultiplierBottomQuarter: 2,
}

// From the note: Rs 6,000 a month each to API at 343 and NTC at 910. API fills a lot
// at once; NTC cannot in month one, carries the Rs 6,000 and fills one lot in month two.
func TestOrderSheetBuysWholeLotsAndCarriesTheRest(t *testing.T) {
	one := sheet("Mangsir", defaults, nil, []*ntxv1.Holding{build("API", 50, 0, 343), build("NTC", 50, 0, 910)})
	two := sheet("Poush", defaults, nil, []*ntxv1.Holding{build("NTC", 50, 6000, 910)})
	for name, got := range map[string][2]float64{
		"API lots":          {float64(one.Lines[0].GetLots()), 1},
		"API cost":          {one.Lines[0].GetCostRs(), 3430},
		"API new carry":     {one.Lines[0].GetNewCarryRs(), 2570},
		"NTC lots month 1":  {float64(one.Lines[1].GetLots()), 0},
		"NTC carry month 1": {one.Lines[1].GetNewCarryRs(), 6000},
		"NTC lots month 2":  {float64(two.Lines[0].GetLots()), 1},
		"NTC shares":        {float64(two.Lines[0].GetShares()), 10},
		"NTC cost month 2":  {two.Lines[0].GetCostRs(), 9100},
		"NTC carry month 2": {two.Lines[0].GetNewCarryRs(), 2900},
		"budget":            {one.GetBudgetRs(), 12000},
		"multiplier":        {one.GetMultiplier(), 1},
	} {
		if math.Abs(got[0]-got[1]) > 0.001 {
			t.Errorf("%s = %.2f, want %.2f", name, got[0], got[1])
		}
	}
	if one.GetOpening() != "Order sheet for Mangsir: Rs 12,000; no index close recorded, so the budget stands." {
		t.Errorf("opening = %q", one.GetOpening())
	}
}

func TestOrderSheetListsFrozenAndCappedHoldingsWithoutSpend(t *testing.T) {
	frozen := &ntxv1.Holding{
		Symbol: "MNBBL", Status: ntxv1.HoldingStatus_HOLDING_STATUS_FROZEN, AllocationPercent: 25,
		FrozenReason: "NPL one bad quarter from 5%", ReviewDate: "2027-01-15", Quantity: 102,
	}
	capped := build("MEN", 50, 0, 664.80)
	capped.CapVerdict = Watch
	hold := build("EBL", 0, 0, 700)
	hold.Status = ntxv1.HoldingStatus_HOLDING_STATUS_HOLD
	noClose := build("SAHAS", 25, 0, 0)
	noClose.Close = nil
	s := sheet("Mangsir", defaults, nil, []*ntxv1.Holding{frozen, capped, hold, noClose})
	var got []string
	for _, l := range s.GetLines() {
		got = append(got, l.GetSymbol()+": "+l.GetNote())
	}
	want := []string{
		"MNBBL: frozen, NPL one bad quarter from 5%, review 2027-01-15",
		"MEN: over the soft cap",
		"SAHAS: no close recorded",
	}
	if len(got) != len(want) {
		t.Fatalf("lines %q", got)
	}
	for i := range want {
		if got[i] != want[i] || s.Lines[i].GetAvailableRs() != 0 || s.Lines[i].GetLots() != 0 {
			t.Errorf("line %d = %q (%v), want %q with nothing to spend", i, got[i], s.Lines[i], want[i])
		}
	}
}

func closes(latest float64) []sqlc.ListIndexClosesRow {
	var rows []sqlc.ListIndexClosesRow
	for i := 0; i <= 10; i++ {
		rows = append(rows, sqlc.ListIndexClosesRow{BusinessDate: "2026-01-01", IndexClose: 2000 + 100*float64(i)})
	}
	return append(rows, sqlc.ListIndexClosesRow{BusinessDate: "2026-11-09", IndexClose: latest})
}

// The index at 2,000 to 3,000 over the year: bottom quarter up to 2,250, median 2,500.
func TestMultiplierFollowsTheIndexInItsYear(t *testing.T) {
	for _, tc := range []struct {
		latest, want float64
		opening      string
	}{
		{2100, 2, "Order sheet for Mangsir: Rs 24,000, 2x the Rs 12,000 budget with the index at 2100.00 on 2026-11-09, " +
			"in the bottom quarter of its 52-week range of 2000.00 to 3000.00."},
		{2400, 1.5, "Order sheet for Mangsir: Rs 18,000, 1.5x the Rs 12,000 budget with the index at 2400.00 on 2026-11-09, " +
			"in the lower half of its 52-week range of 2000.00 to 3000.00."},
		{2800, 1, "Order sheet for Mangsir: Rs 12,000, the index at 2800.00 on 2026-11-09 " +
			"in the upper half of its 52-week range of 2000.00 to 3000.00."},
	} {
		s := sheet("Mangsir", defaults, closes(tc.latest), nil)
		if s.GetMultiplier() != tc.want || s.GetOpening() != tc.opening {
			t.Errorf("at %.0f: %gx, %q", tc.latest, s.GetMultiplier(), s.GetOpening())
		}
	}
}

func TestCalendarListsFilingsReviewsAndTicksSoonestFirst(t *testing.T) {
	men := &ntxv1.Holding{Symbol: "MEN", Quantity: 144, Gates: []*ntxv1.Gate{
		{Kind: ntxv1.GateKind_GATE_KIND_EVENT, Text: "Plant damage reported", CheckedOn: "2026-10-16"},
	}}
	mnbbl := &ntxv1.Holding{Symbol: "MNBBL", Quantity: 102, ReviewDate: "2027-01-15"}
	fund := &ntxv1.Holding{Symbol: "NIBLSF", Quantity: 5027, IsFund: true}
	watched := &ntxv1.Holding{Symbol: "NTC"}
	today := time.Date(2026, time.November, 10, 0, 0, 0, 0, time.UTC)
	var got []string
	for _, e := range calendar(today, []*ntxv1.Holding{men, mnbbl, fund, watched}) {
		got = append(got, e.GetDay()+" "+e.GetSymbol()+" "+e.GetWhat())
	}
	want := []string{
		"2026-11-16 MEN Q1 2083/84 filing due",
		"2026-11-16 MNBBL Q1 2083/84 filing due",
		"2027-01-15 MNBBL review",
		"2027-01-16 MEN re-check: Plant damage reported",
	}
	if len(got) != len(want) {
		t.Fatalf("calendar %q", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %q, want %q", i, got[i], want[i])
		}
	}
	// Past the Q1 due date, the next quarter's filing is the one on the calendar.
	later := calendar(
		time.Date(2026, time.November, 20, 0, 0, 0, 0, time.UTC),
		[]*ntxv1.Holding{{Symbol: "NTC", Quantity: 10}},
	)
	if len(later) != 1 || later[0].GetDay() != "2027-02-13" || later[0].GetWhat() != "Q2 2083/84 filing due" {
		t.Errorf("later = %v", later)
	}
}
