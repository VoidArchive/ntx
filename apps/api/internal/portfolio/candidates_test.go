package portfolio

import (
	"database/sql"
	"slices"
	"testing"
	"time"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// day is the owner's check of 2026-11-11: Q1 2083/84 has ended, Q4 2083/84 has not.
var day = time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)

func listed(id int64, symbol, sector, rating, dividend string, close float64) sqlc.ListScreenerRow {
	return sqlc.ListScreenerRow{
		ID: id, Symbol: symbol, Name: symbol + " Ltd", Sector: sector,
		ClosePrice:     sql.NullFloat64{Float64: close, Valid: close > 0},
		BusinessDate:   sql.NullString{String: "2026-11-09", Valid: true},
		Rating:         sql.NullString{String: rating, Valid: true},
		DividendRecord: sql.NullString{String: dividend, Valid: true},
	}
}

// filed adds the latest filing's quarter and annualized EPS to a screener row.
func filed(row sqlc.ListScreenerRow, year, quarter string, eps float64) sqlc.ListScreenerRow {
	row.FiscalYearNp = sql.NullString{String: year, Valid: true}
	row.Quarter = sql.NullString{String: quarter, Valid: true}
	row.Eps = sql.NullFloat64{Float64: eps, Valid: true}
	return row
}

func plant(symbol, basin string) sqlc.HydroPlant {
	return sqlc.HydroPlant{Symbol: symbol, Basin: basin}
}

func bankRowFor(id int64, year, quarter string, npl, car, distributable float64) sqlc.ListBankMetricsRow {
	return sqlc.ListBankMetricsRow{
		CompanyID: id, FiscalYear: year, Quarter: quarter,
		NplPercent:          sql.NullFloat64{Float64: npl, Valid: true},
		CarPercent:          sql.NullFloat64{Float64: car, Valid: true},
		DistributableProfit: sql.NullFloat64{Float64: distributable, Valid: true},
		PublishedDate:       "2026-10-28", SourceUrl: "https://nepalstock.com.np/filing.pdf",
	}
}

// passing is the symbols of the companies that pass, in order.
func passing(cands []*ntxv1.Candidate) []string {
	var out []string
	for _, c := range cands {
		if c.GetVerdict() == Passes {
			out = append(out, c.GetSymbol())
		}
	}
	return out
}

func find(cands []*ntxv1.Candidate, symbol string) *ntxv1.Candidate {
	i := slices.IndexFunc(cands, func(c *ntxv1.Candidate) bool { return c.GetSymbol() == symbol })
	if i < 0 {
		return nil
	}
	return cands[i]
}

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
	got := candidates(rows, pes, held, rules, plants, nil, day)
	// Rating first, then PE against the sector: NTC is the lone A; API's PE sits below
	// the hydro median of the two priced hydros, UPPER's above it.
	if want := []string{"NTC", "API", "UPPER"}; !slices.Equal(passing(got), want) {
		t.Fatalf("candidates %v, want %v", passing(got), want)
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

// Every other company is turned away with the rule named, so an absence is explainable.
func TestTurnedAwayNamesTheRule(t *testing.T) {
	held, rows, pes, plants, rules := ownersBook()
	got := candidates(rows, pes, held, rules, plants, nil, day)
	for symbol, want := range map[string]string{
		"MEN":   "held",
		"SAHAS": "a plant in the Gandaki basin, which MEN already occupies",
		"EBL":   "the Dividend record check does not hold",
		"NABIL": "rated C, below B",
		"SCB":   "no close recorded",
		"KBL":   "unrated",
	} {
		if c := find(got, symbol); c.GetVerdict() != TurnedAway || c.GetReason() != want {
			t.Errorf("%s: %s %q, want turned away %q", symbol, c.GetVerdict(), c.GetReason(), want)
		}
	}
}

// Hydropower already at 63% of the book: at a 40% cap no hydro has headroom. The
// check uses the lot's real cost; NTC's Rs 18,200 lot leaves Others well under 40%.
func TestCandidatesNeedSectorHeadroom(t *testing.T) {
	held, rows, pes, plants, rules := ownersBook()
	rules.SectorCapPercent = 40
	got := candidates(rows, pes, held, rules, plants, nil, day)
	if want := []string{"NTC"}; !slices.Equal(passing(got), want) {
		t.Errorf("under the sector cap: %v, want %v", passing(got), want)
	}
	if want := "Hydro Power would stand at 67.1% after a Rs 17,150 entry, over the 40% cap"; find(
		got,
		"API",
	).GetReason() !=
		want {
		t.Errorf("API: %q, want %q", find(got, "API").GetReason(), want)
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
	got := candidates(rows, nil, watched, rules, plants, nil, day)
	if p := passing(got); len(p) != 1 || p[0] != "MEN" || got[0].GetBasin() != "Gandaki" {
		t.Errorf("candidates %v", got)
	}
}

// lenders is the eight companies the page admitted on 2026-11-11, priced as it priced
// them, with peers that set each sector's median and newest quarter: a C-rated
// development bank on Q1 2083/84, two microfinance peers (one with a Q4 2083/84 row
// for a quarter that has not ended), and a held commercial bank.
func lenders() (*ntxv1.GetPortfolioResponse, []sqlc.ListScreenerRow, map[int64]priced, sqlc.Rule) {
	held := &ntxv1.GetPortfolioResponse{
		TotalValue: 374000, Positions: 2,
		Holdings: []*ntxv1.Holding{
			{Symbol: "MEN", Sector: "Hydro Power", Quantity: 300, Value: 200000},
			{Symbol: "NMB", Sector: "Commercial Banks", Quantity: 200, Value: 74000},
			{Symbol: "NIBLSF", Sector: "Mutual Fund", Quantity: 9000, Value: 100000, IsFund: true},
		},
	}
	rows := []sqlc.ListScreenerRow{
		filed(listed(1, "SADBL", "Development Banks", "B", "holds", 395), "2082/83", "Q4", 23.83),
		filed(listed(2, "CZBIL", "Commercial Banks", "B", "holds", 224), "2083/84", "Q1", 2.68),
		filed(listed(3, "MLBL", "Development Banks", "B", "holds", 375.9), "2082/83", "Q4", 13.96),
		filed(listed(4, "GBIME", "Commercial Banks", "B", "holds", 272), "2083/84", "Q1", 23.35),
		filed(listed(5, "GLBSL", "Microfinance", "B", "holds", 1489.9), "2083/84", "Q1", 41.09),
		filed(listed(6, "SANIMA", "Commercial Banks", "B", "holds", 408), "2082/83", "Q4", 25.75),
		filed(listed(7, "ICFC", "Finance", "B", "holds", 619), "2083/84", "Q1", 23.59),
		filed(listed(8, "MFIL", "Finance", "B", "holds", 698.8), "2082/83", "Q4", 27.85),
		filed(listed(9, "JBBL", "Development Banks", "C", "holds", 300), "2083/84", "Q1", 10),
		filed(listed(10, "MERO", "Microfinance", "C", "holds", 800), "2083/84", "Q4", 25),
		filed(listed(11, "FOWAD", "Microfinance", "C", "holds", 900), "2083/84", "Q1", 45),
		filed(listed(12, "NMB", "Commercial Banks", "B", "holds", 370), "2083/84", "Q1", 20),
	}
	pes := map[int64]priced{
		1: {16.5, "4Q to Q4 2082/83"}, 2: {14.5, "4Q to Q1 2083/84"}, 3: {26.7, "4Q to Q4 2082/83"},
		4: {11.4, "4Q to Q1 2083/84"}, 5: {50.2, "4Q to Q1 2083/84"}, 6: {15.3, "4Q to Q4 2082/83"},
		7: {23.2, "4Q to Q1 2083/84"}, 8: {24.5, "4Q to Q4 2082/83"}, 9: {30, "4Q to Q1 2083/84"},
		10: {31.5, "4Q to Q1 2083/84"}, 11: {20, "4Q to Q1 2083/84"}, 12: {18.5, "4Q to Q1 2083/84"},
	}
	rules := defaults
	rules.CandidateMinRating, rules.CandidateDividendRecord = "B", 1
	rules.NplCeilingPercent, rules.CarFloorPercent, rules.DistributableAboveZero, rules.MaxPeToSectorMedian = 5, 12, 1, 1.25
	return held, rows, pes, rules
}

// The published Q1 2083/84 figures, typed by the owner; MFIL has only its Q3 2082/83 row.
func lenderRows() []sqlc.ListBankMetricsRow {
	return []sqlc.ListBankMetricsRow{
		bankRowFor(1, "2083/84", "Q1", 7.64, 12.5, -31200000),
		bankRowFor(2, "2083/84", "Q1", 5.88, 12.1, -50000000),
		bankRowFor(3, "2083/84", "Q1", 6.49, 13, 100000000),
		bankRowFor(4, "2083/84", "Q1", 4.79, 11.32, 1982000000),
		bankRowFor(6, "2083/84", "Q1", 1.08, 12.37, 500000000),
		bankRowFor(7, "2083/84", "Q1", 1.95, 13.03, 50000000),
		bankRowFor(8, "2082/83", "Q3", 3.99, 14, 100000000),
	}
}

// With the figures entered, SANIMA and ICFC pass; four are turned away by a named
// lender rule, GLBSL by price, and MFIL still needs its Q4 figures.
func TestLendersPassOnlyWithTheirFigures(t *testing.T) {
	held, rows, pes, rules := lenders()
	got := candidates(rows, pes, held, rules, nil, lenderRows(), day)
	// Both B; ICFC sits nearer its sector median than SANIMA does.
	if want := []string{"ICFC", "SANIMA"}; !slices.Equal(passing(got), want) {
		t.Fatalf("passing %v, want %v", passing(got), want)
	}
	for symbol, want := range map[string][2]string{
		"SADBL":  {TurnedAway, "NPL 7.64% above the 5% ceiling and distributable profit Rs -3.12 crore, not above zero, Q1 2083/84"},
		"CZBIL":  {TurnedAway, "NPL 5.88% above the 5% ceiling and distributable profit Rs -5.00 crore, not above zero, Q1 2083/84"},
		"MLBL":   {TurnedAway, "NPL 6.49% above the 5% ceiling, Q1 2083/84"},
		"GBIME":  {TurnedAway, "CAR 11.32% below the 12% floor, Q1 2083/84"},
		"GLBSL":  {TurnedAway, "PE 50.2x, 1.6x the Microfinance median of 31.5x, above the 1.25x rule"},
		"MFIL":   {NeedsData, "needs Q4 2082/83 NPL, CAR and distributable profit"},
		"SANIMA": {Passes, ""},
		"NMB":    {TurnedAway, "held"},
	} {
		c := find(got, symbol)
		if c.GetVerdict() != want[0] || c.GetReason() != want[1] {
			t.Errorf("%s: %s %q, want %s %q", symbol, c.GetVerdict(), c.GetReason(), want[0], want[1])
		}
	}
	sanima := find(got, "SANIMA")
	if sanima.GetNplPercent() != 1.08 || sanima.GetMetricsPeriod() != "Q1 2083/84" || sanima.GetMetricsSource() == "" {
		t.Errorf("SANIMA figures %v", sanima)
	}
}

// Without a single bank row no lender passes: every one needs data, and the price
// rule still turns GLBSL away.
func TestLendersWithoutFiguresNeedData(t *testing.T) {
	held, rows, pes, rules := lenders()
	got := candidates(rows, pes, held, rules, nil, nil, day)
	if p := passing(got); len(p) != 0 {
		t.Fatalf("passing %v, want none", p)
	}
	for _, symbol := range []string{"SADBL", "CZBIL", "MLBL", "GBIME", "SANIMA", "ICFC", "MFIL"} {
		if c := find(got, symbol); c.GetVerdict() != NeedsData {
			t.Errorf("%s: %s %q, want needs data", symbol, c.GetVerdict(), c.GetReason())
		}
	}
	if want := "needs Q1 2083/84 NPL, CAR and distributable profit"; find(got, "CZBIL").GetReason() != want {
		t.Errorf("CZBIL: %q", find(got, "CZBIL").GetReason())
	}
	if c := find(got, "GLBSL"); c.GetVerdict() != TurnedAway {
		t.Errorf("GLBSL: %s", c.GetVerdict())
	}
	if want := "Positions stand at 2 of 8; a new scrip needs a sale first. None passes the rules, 7 lenders need data and 5 are turned away."; candidatesOpening(
		got,
		held,
		rules,
	) != want {
		t.Errorf("opening %q", candidatesOpening(got, held, rules))
	}
}

// The trailing PE sits beside the PE on the latest annualized EPS, and a company whose
// latest filing is older than the sector's newest is marked; the Q4 2083/84 row for a
// quarter that has not ended does not count as the sector's newest.
func TestCandidatesShowTheLatestQuarterAndWhoIsBehind(t *testing.T) {
	held, rows, pes, rules := lenders()
	got := candidates(rows, pes, held, rules, nil, lenderRows(), day)
	czbil := find(got, "CZBIL")
	if pe := czbil.GetAnnualizedPe(); pe < 83.5 || pe > 83.6 || czbil.GetAnnualizedPeriod() != "Q1 2083/84" ||
		czbil.GetBehind() != "" {
		t.Errorf("CZBIL %v", czbil)
	}
	for symbol, want := range map[string]string{
		"SADBL": "priced on Q4 2082/83; the sector has filed Q1 2083/84",
		"MFIL":  "priced on Q4 2082/83; the sector has filed Q1 2083/84",
		"GLBSL": "",
		"FOWAD": "",
	} {
		if got := find(got, symbol).GetBehind(); got != want {
			t.Errorf("%s behind %q, want %q", symbol, got, want)
		}
	}
	hydro := candidates(
		[]sqlc.ListScreenerRow{filed(listed(1, "MEN", "Hydro Power", "A", "holds", 600), "2083/84", "Q1", 60)},
		nil,
		&ntxv1.GetPortfolioResponse{},
		defaults,
		nil,
		nil,
		day,
	)
	if hydro[0].GetAnnualizedPe() != 10 || hydro[0].GetAnnualizedPeriod() != "Q1 2083/84, monsoon quarter" {
		t.Errorf("hydro %v", hydro[0])
	}
}

func TestCandidatesOpeningStatesTheStandingRule(t *testing.T) {
	r := defaults
	full := &ntxv1.GetPortfolioResponse{Positions: 8}
	verdicts := func(v ...string) []*ntxv1.Candidate {
		var out []*ntxv1.Candidate
		for _, x := range v {
			out = append(out, &ntxv1.Candidate{Verdict: x})
		}
		return out
	}
	for want, p := range map[string]struct {
		res   *ntxv1.GetPortfolioResponse
		cands []*ntxv1.Candidate
	}{
		"Positions stand at 2 of 8; a new scrip needs a sale first. 3 companies pass the rules.":                                       {&ntxv1.GetPortfolioResponse{Positions: 2}, verdicts(Passes, Passes, Passes)},
		"Positions stand at 8 of 8; a new scrip needs a sale first. 1 company passes the rules and 1 is turned away.":                  {full, verdicts(Passes, TurnedAway)},
		"Positions stand at 0 of 8; a new scrip needs a sale first. None passes the rules.":                                            {&ntxv1.GetPortfolioResponse{}, nil},
		"Positions stand at 0 of 8; a new scrip needs a sale first. None passes the rules, 1 lender needs data and 2 are turned away.": {&ntxv1.GetPortfolioResponse{}, verdicts(NeedsData, TurnedAway, TurnedAway)},
	} {
		if got := candidatesOpening(p.cands, p.res, r); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
