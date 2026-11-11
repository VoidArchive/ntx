package portfolio

import (
	"database/sql"
	"slices"
	"testing"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func listed(id int64, symbol, sector, rating, dividend string, close float64) sqlc.ListScreenerRow {
	return sqlc.ListScreenerRow{
		ID: id, Symbol: symbol, Name: symbol + " Ltd", Sector: sector,
		ClosePrice:     sql.NullFloat64{Float64: close, Valid: close > 0},
		BusinessDate:   sql.NullString{String: "2026-11-09", Valid: true},
		Rating:         sql.NullString{String: rating, Valid: true},
		DividendRecord: sql.NullString{String: dividend, Valid: true},
	}
}

func plant(symbol, basin string) sqlc.HydroPlant {
	return sqlc.HydroPlant{Symbol: symbol, Basin: basin}
}

// The owner holds MEN (Gandaki) and NIBLSF; the screener offers hydros in and out of
// that basin, one with no plant row, a bank without the dividend record and a C.
// ownersBook is the fixture: MEN and NIBLSF held, a screener with hydros in and out
// of MEN's basin, one with no plant row, a bank without the dividend record and a C.
func ownersBook() (*ntxv1.GetPortfolioResponse, []sqlc.ListScreenerRow, map[int64]priced, []sqlc.HydroPlant, sqlc.Rule) {
	held := &ntxv1.GetPortfolioResponse{
		TotalValue: 150000, Positions: 1,
		Holdings: []*ntxv1.Holding{
			{Symbol: "MEN", Sector: "Hydro Power", Quantity: 144, Value: 95000, WeightPercent: 63.3},
			{Symbol: "NIBLSF", Sector: "Mutual Fund", Quantity: 5027, Value: 55000, IsFund: true},
		},
	}
	rows := []sqlc.ListScreenerRow{
		listed(1, "MEN", "Hydro Power", "A", "holds", 664.80),
		listed(2, "SAHAS", "Hydro Power", "A", "holds", 685),
		listed(3, "API", "Hydro Power", "B", "holds", 343),
		listed(4, "UPPER", "Hydro Power", "B", "holds", 250),
		listed(5, "NTC", "Others", "A", "holds", 910),
		listed(6, "EBL", "Commercial Banks", "B", "watch", 700),
		listed(7, "NABIL", "Commercial Banks", "C", "holds", 500),
		listed(8, "SCB", "Commercial Banks", "B", "holds", 0),
		listed(9, "KBL", "Commercial Banks", "", "", 200),
	}
	pes := map[int64]priced{3: {21.6, "4Q to Q1 2083/84"}, 4: {30, "Q1 2083/84"}, 5: {17.9, "4Q to Q1 2083/84"}}
	plants := []sqlc.HydroPlant{plant("MEN", "Gandaki"), plant("SAHAS", "Gandaki"), plant("API", "Mahakali")}
	rules := defaults
	rules.OneHydroPerBasin, rules.CandidateDividendRecord, rules.CandidateMinRating = 1, 1, "B"
	rules.SectorCapPercent = 70

	return held, rows, pes, plants, rules
}

func TestCandidatesFollowTheOwnersRules(t *testing.T) {
	held, rows, pes, plants, rules := ownersBook()
	got := candidates(rows, pes, held, rules, plants)
	var names []string
	for _, c := range got {
		names = append(names, c.GetSymbol())
	}
	// Rating first, then PE against the sector: NTC is the lone A; API's PE sits below
	// the hydro median of the two priced hydros, UPPER's above it.
	if want := []string{"NTC", "API", "UPPER"}; !slices.Equal(names, want) {
		t.Fatalf("candidates %v, want %v", names, want)
	}
	api, upper := got[1], got[2]
	// Rs 15,000 at 343 is 43.7 shares: the next whole lot is 50, Rs 17,150, 1.4 months of Rs 12,000.
	if api.GetShares() != 50 || api.GetCostRs() != 17150 || api.GetMonths() < 1.42 || api.GetMonths() > 1.43 {
		t.Errorf("API entry %v", api)
	}
	if api.GetBasin() != "Mahakali" || api.GetNote() != "" || upper.GetNote() != "basin not recorded" {
		t.Errorf("basins: API %q %q, UPPER %q", api.GetBasin(), api.GetNote(), upper.GetNote())
	}
	if api.GetPe() != 21.6 || api.GetPePeriod() != "4Q to Q1 2083/84" || api.GetSectorPeMedian() != 25.8 ||
		api.GetCloseDay() != "2026-11-09" {
		t.Errorf("API pricing %v", api)
	}
}

// Hydropower already at 63% of the book: at a 40% cap no hydro has headroom. The
// check uses the lot's real cost; NTC's Rs 18,200 lot leaves Others well under 40%.
func TestCandidatesNeedSectorHeadroom(t *testing.T) {
	held, rows, pes, plants, rules := ownersBook()
	rules.SectorCapPercent = 40
	var names []string
	for _, c := range candidates(rows, pes, held, rules, plants) {
		names = append(names, c.GetSymbol())
	}
	if want := []string{"NTC"}; !slices.Equal(names, want) {
		t.Errorf("under the sector cap: %v, want %v", names, want)
	}
}

// A watched hydro (no shares) is no candidate itself but does not occupy its basin.
func TestWatchedHydroDoesNotBlockItsBasin(t *testing.T) {
	watched := &ntxv1.GetPortfolioResponse{Holdings: []*ntxv1.Holding{{Symbol: "SAHAS", Sector: "Hydro Power"}}}
	rows := []sqlc.ListScreenerRow{
		listed(2, "SAHAS", "Hydro Power", "A", "holds", 685),
		listed(1, "MEN", "Hydro Power", "A", "holds", 664.80),
	}
	plants := []sqlc.HydroPlant{plant("MEN", "Gandaki"), plant("SAHAS", "Gandaki")}
	rules := defaults
	rules.OneHydroPerBasin = 1
	got := candidates(rows, nil, watched, rules, plants)
	if len(got) != 1 || got[0].GetSymbol() != "MEN" || got[0].GetBasin() != "Gandaki" {
		t.Errorf("candidates %v", got)
	}
}

func TestCandidatesOpeningStatesTheStandingRule(t *testing.T) {
	r := defaults
	full := &ntxv1.GetPortfolioResponse{Positions: 8}
	for want, p := range map[string]struct {
		res *ntxv1.GetPortfolioResponse
		n   int
	}{
		"Positions stand at 2 of 8; a new scrip needs a sale first. 3 companies pass the rules.": {&ntxv1.GetPortfolioResponse{Positions: 2}, 3},
		"Positions stand at 8 of 8; a new scrip needs a sale first. 1 company passes the rules.": {full, 1},
		"Positions stand at 0 of 8; a new scrip needs a sale first. None passes the rules.":      {&ntxv1.GetPortfolioResponse{}, 0},
	} {
		if got := candidatesOpening(p.n, p.res, r); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
