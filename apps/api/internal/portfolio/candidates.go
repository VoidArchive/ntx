package portfolio

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
	"github.com/voidarchive/ntx/internal/story"
)

// Verdicts of a Candidate (GLOSSARY: Candidates, Turned away); NeedsData is shared with gates.
const (
	Passes     = "passes"
	TurnedAway = "turned away"
)

var verdictRank = map[string]int{Passes: 0, NeedsData: 1, TurnedAway: 2}

// isHydro reads the sector string loosely: the exchange writes "Hydro Power", fixtures "Hydropower".
func isHydro(sector string) bool { return strings.Contains(strings.ToLower(sector), "hydro") }

// isLender is the four NEPSE classes the lender rules apply to: commercial and
// development banks, finance and microfinance (GLOSSARY: Lender).
func isLender(sector string) bool {
	s := strings.ToLower(sector)
	return strings.Contains(s, "bank") || strings.Contains(s, "finance")
}

// priced is a company's PE with the period its EPS names, as the screener prices it.
type priced struct {
	PE     float64
	Period string
}

// letterRank orders ratings; an unrated company ("") never qualifies.
var letterRank = map[string]int{"A": 0, "B": 1, "C": 2, "D": 3}

// rulebook is what the rules compare a company against: what is held, each sector's
// value, the basins the owner's plants sit in, the owner's bank rows and the newest
// quarter each sector has filed.
type rulebook struct {
	held        map[string]bool
	sectorValue map[string]float64
	heldBasins  map[string]string   // basin -> a held symbol with a plant there
	basins      map[string][]string // by symbol, from the plant table
	total       float64
	medians     map[string]float64 // sector PE medians
	pes         map[int64]priced
	bank        map[int64][]sqlc.ListBankMetricsRow
	newest      map[string]filing // by sector
	rules       sqlc.Rule
	minRank     int
}

func bookOf(
	rows []sqlc.ListScreenerRow,
	pes map[int64]priced,
	p *ntxv1.GetPortfolioResponse,
	r sqlc.Rule,
	plants []sqlc.HydroPlant,
	bank []sqlc.ListBankMetricsRow,
	today time.Time,
) rulebook {
	b := rulebook{
		held: map[string]bool{}, sectorValue: map[string]float64{}, heldBasins: map[string]string{},
		basins: map[string][]string{}, total: p.GetTotalValue(), medians: sectorMedians(rows, pes), pes: pes,
		bank: map[int64][]sqlc.ListBankMetricsRow{}, newest: newestFilings(rows, today), rules: r,
	}
	// A watched symbol is already tracked, so it is no candidate; only shares held
	// count toward a sector or occupy a basin.
	withShares := map[string]bool{}
	for _, h := range p.GetHoldings() {
		b.held[h.GetSymbol()] = true
		if h.GetQuantity() > 0 {
			b.sectorValue[h.GetSector()] += h.GetValue()
			withShares[h.GetSymbol()] = true
		}
	}
	for _, pl := range plants {
		if !slices.Contains(b.basins[pl.Symbol], pl.Basin) {
			b.basins[pl.Symbol] = append(b.basins[pl.Symbol], pl.Basin)
		}
		if _, taken := b.heldBasins[pl.Basin]; withShares[pl.Symbol] && !taken {
			b.heldBasins[pl.Basin] = pl.Symbol
		}
	}
	for _, m := range bank {
		b.bank[m.CompanyID] = append(b.bank[m.CompanyID], m)
	}
	rank, ok := letterRank[strings.ToUpper(r.CandidateMinRating)]
	if !ok {
		rank = letterRank["B"]
	}
	b.minRank = rank
	return b
}

// latestFiling is the quarter a screener row's figures come from; false without one.
func latestFiling(row sqlc.ListScreenerRow) (filing, bool) {
	y, err := fiscal.ParseFilingYear(row.FiscalYearNp.String, row.FiscalYear.String)
	if err != nil {
		return filing{}, false
	}
	q, err := fiscal.ParseQuarter(row.Quarter.String)
	if err != nil || q == fiscal.Annual {
		return filing{}, false
	}
	return filing{year: y, quarter: q}, true
}

// newestFilings is the newest quarter each sector has filed, skipping a quarter that
// has not ended yet: three companies carry a Q4 2083/84 row that would otherwise mark
// every peer as behind.
func newestFilings(rows []sqlc.ListScreenerRow, today time.Time) map[string]filing {
	out := map[string]filing{}
	for _, row := range rows {
		f, ok := latestFiling(row)
		if !ok {
			continue
		}
		if end, err := fiscal.QuarterEnd(f.year, f.quarter); err != nil || end.After(today) {
			continue
		}
		if cur, seen := out[row.Sector]; !seen || f.rank() > cur.rank() {
			out[row.Sector] = f
		}
	}
	return out
}

// candidates reads every listed company against the owner's rules (GLOSSARY:
// Candidates): each passes, needs data (a lender without its bank row) or is turned
// away with the rule named. pes are each company's PE by id, as the screener prices it.
func candidates(
	rows []sqlc.ListScreenerRow,
	pes map[int64]priced,
	p *ntxv1.GetPortfolioResponse,
	r sqlc.Rule,
	plants []sqlc.HydroPlant,
	bank []sqlc.ListBankMetricsRow,
	today time.Time,
) []*ntxv1.Candidate {
	b := bookOf(rows, pes, p, r, plants, bank, today)
	out := make([]*ntxv1.Candidate, 0, len(rows))
	for _, row := range rows {
		out = append(out, b.admit(row))
	}
	slices.SortStableFunc(out, func(x, y *ntxv1.Candidate) int {
		if d := cmp.Compare(verdictRank[x.GetVerdict()], verdictRank[y.GetVerdict()]); d != 0 {
			return d
		}
		if d := cmp.Compare(letterRank[x.GetRating()], letterRank[y.GetRating()]); d != 0 {
			return d
		}
		return cmp.Compare(againstSector(x), againstSector(y))
	})
	return out
}

// turnedAway applies the rules that need no entry size: held, unrated or rated
// too low, unpriced, or without the dividend record. Empty when none applies.
func (b rulebook) turnedAway(row sqlc.ListScreenerRow) string {
	rank, rated := letterRank[row.Rating.String]
	switch {
	case b.held[row.Symbol]:
		return "held"
	case !rated:
		return "unrated"
	case rank > b.minRank:
		return fmt.Sprintf("rated %s, below %s", row.Rating.String, b.rules.CandidateMinRating)
	case !row.ClosePrice.Valid || row.ClosePrice.Float64 <= 0:
		return "no close recorded"
	case b.rules.CandidateDividendRecord != 0 && row.DividendRecord.String != story.Holds:
		return "the Dividend record check does not hold"
	}
	return ""
}

// standing is the sector's share of the book once the entry's real cost is in it.
func (b rulebook) standing(sector string, cost float64) float64 {
	return (b.sectorValue[sector] + cost) * 100 / (b.total + cost)
}

// admit is one company against the rules, with its figures whatever the verdict so a
// turned-away company is explainable.
func (b rulebook) admit(row sqlc.ListScreenerRow) *ntxv1.Candidate {
	c := &ntxv1.Candidate{
		Symbol: row.Symbol, Name: row.Name, Sector: row.Sector, Rating: row.Rating.String,
		Close: row.ClosePrice.Float64, CloseDay: row.BusinessDate.String, Verdict: Passes,
	}
	if pe, ok := b.pes[row.ID]; ok {
		c.Pe, c.PePeriod = &pe.PE, pe.Period
	}
	if m, ok := b.medians[row.Sector]; ok {
		c.SectorPeMedian = &m
	}
	b.price(row, c)
	turn := func(reason string) *ntxv1.Candidate {
		c.Verdict, c.Reason = TurnedAway, reason
		return c
	}
	for _, rule := range []func() string{
		func() string { return b.turnedAway(row) },
		func() string { return b.basin(row, c) },
		func() string { return b.entry(row, c) },
	} {
		if reason := rule(); reason != "" {
			return turn(reason)
		}
	}
	if isLender(row.Sector) {
		b.lender(row, c)
	}
	return c
}

// basin applies the one-hydro-per-basin rule from the plant table, whatever the
// sector string says.
func (b rulebook) basin(row sqlc.ListScreenerRow, c *ntxv1.Candidate) string {
	basins := b.basins[row.Symbol]
	if len(basins) == 0 {
		if isHydro(row.Sector) {
			c.Note = "basin not recorded"
		}
		return ""
	}
	c.Basin = strings.Join(basins, ", ")
	for _, x := range basins {
		if held, taken := b.heldBasins[x]; b.rules.OneHydroPerBasin != 0 && taken {
			return fmt.Sprintf("a plant in the %s basin, which %s already occupies", x, held)
		}
	}
	return ""
}

// entry sizes the smallest whole lot that reaches the minimum entry and the months of
// budget it costs, then tests the sector cap with it and the PE against the sector.
func (b rulebook) entry(row sqlc.ListScreenerRow, c *ntxv1.Candidate) string {
	r := b.rules
	c.Shares = int64(math.Ceil(r.MinEntryRs/c.Close/lot)) * lot
	c.CostRs = float64(c.Shares) * c.Close
	// An empty book has no sector to cap.
	if b.total > 0 && b.standing(row.Sector, c.CostRs) >= r.SectorCapPercent {
		return fmt.Sprintf("%s would stand at %.1f%% after a Rs %s entry, over the %s%% cap",
			row.Sector, b.standing(row.Sector, c.CostRs), rupees(c.CostRs), num(r.SectorCapPercent))
	}
	if r.MonthlyBudgetRs > 0 {
		c.Months = c.CostRs / r.MonthlyBudgetRs
	}
	if ratio := againstSector(c); r.MaxPeToSectorMedian > 0 && !math.IsInf(ratio, 1) && ratio > r.MaxPeToSectorMedian {
		return fmt.Sprintf("PE %.1fx, %.1fx the %s median of %.1fx, above the %sx rule",
			c.GetPe(), ratio, row.Sector, c.GetSectorPeMedian(), num(r.MaxPeToSectorMedian))
	}
	return ""
}

// price adds the PE on the latest filing's annualized EPS beside the trailing one, and
// marks a company whose latest filing is older than the newest its sector has filed
// (GLOSSARY: Behind). Neither is a rule: a hydro's Q1 annualizes high (docs/data.md).
func (b rulebook) price(row sqlc.ListScreenerRow, c *ntxv1.Candidate) {
	f, ok := latestFiling(row)
	if !ok {
		return
	}
	if row.Eps.Valid && row.Eps.Float64 > 0 && row.ClosePrice.Valid && row.ClosePrice.Float64 > 0 {
		pe := row.ClosePrice.Float64 / row.Eps.Float64
		c.AnnualizedPe, c.AnnualizedPeriod = &pe, f.period()
		if isHydro(row.Sector) && f.quarter == fiscal.Q1 {
			c.AnnualizedPeriod += ", monsoon quarter"
		}
	}
	if newest, ok := b.newest[row.Sector]; ok && f.rank() < newest.rank() {
		c.Behind = fmt.Sprintf("priced on %s; the sector has filed %s", f.period(), newest.period())
	}
}

// lender reads the owner's bank row for the company's latest filed quarter (or a newer
// one the owner typed from the PDF before the sync caught up) against the ceilings. No
// row, or a row missing a ruled figure, is Needs data: an unchecked lender never passes.
func (b rulebook) lender(row sqlc.ListScreenerRow, c *ntxv1.Candidate) {
	r := b.rules
	period := b.bankRow(row, c)
	var missing, failed []string
	for _, k := range []struct {
		name  string
		on    bool
		value *float64
		wrong func(float64) string // the failure in words, "" when the figure is fine
	}{
		{"NPL", r.NplCeilingPercent > 0, c.NplPercent, func(v float64) string {
			if v > r.NplCeilingPercent {
				return fmt.Sprintf("NPL %s%% above the %s%% ceiling", num(v), num(r.NplCeilingPercent))
			}
			return ""
		}},
		{"CAR", r.CarFloorPercent > 0, c.CarPercent, func(v float64) string {
			if v < r.CarFloorPercent {
				return fmt.Sprintf("CAR %s%% below the %s%% floor", num(v), num(r.CarFloorPercent))
			}
			return ""
		}},
		{"distributable profit", r.DistributableAboveZero != 0, c.DistributableProfit, func(v float64) string {
			if v <= 0 {
				return fmt.Sprintf("distributable profit Rs %s, not above zero", money(v))
			}
			return ""
		}},
	} {
		switch {
		case !k.on:
		case k.value == nil:
			missing = append(missing, k.name)
		default:
			if w := k.wrong(*k.value); w != "" {
				failed = append(failed, w)
			}
		}
	}
	switch {
	case len(failed) > 0:
		c.Verdict, c.Reason = TurnedAway, fmt.Sprintf("%s, %s", list(failed), period)
	case len(missing) > 0:
		c.Verdict, c.Reason = NeedsData, fmt.Sprintf("needs %s %s", period, list(missing))
	}
}

// bankRow copies the owner's newest bank row at or after the company's latest filed
// quarter onto the candidate and names the quarter the lender rules speak of.
func (b rulebook) bankRow(row sqlc.ListScreenerRow, c *ntxv1.Candidate) string {
	latest, filed := latestFiling(row)
	var m *sqlc.ListBankMetricsRow
	for i := range b.bank[row.ID] {
		x := &b.bank[row.ID][i]
		f, ok := bankFiling(*x)
		if !ok || (filed && f.rank() < latest.rank()) {
			continue
		}
		if m == nil || f.rank() > mustFiling(*m).rank() {
			m = x
		}
	}
	switch {
	case m != nil:
		c.NplPercent, c.CarPercent = optional(m.NplPercent), optional(m.CarPercent)
		c.DistributableProfit = optional(m.DistributableProfit)
		c.MetricsPeriod, c.MetricsSource = mustFiling(*m).period(), m.SourceUrl
		return c.MetricsPeriod
	case filed:
		return latest.period()
	}
	return "its latest quarter"
}

func bankFiling(m sqlc.ListBankMetricsRow) (filing, bool) {
	y, err := fiscal.ParseYear(m.FiscalYear)
	if err != nil {
		return filing{}, false
	}
	q, err := fiscal.ParseQuarter(m.Quarter)
	if err != nil || q == fiscal.Annual {
		return filing{}, false
	}
	return filing{year: y, quarter: q}, true
}

// mustFiling is bankFiling for a row bankFiling already accepted.
func mustFiling(m sqlc.ListBankMetricsRow) filing {
	f, _ := bankFiling(m)
	return f
}

// money prints a signed rupee amount in crore above one crore, else as rupees does.
func money(rs float64) string {
	sign := ""
	if rs < 0 {
		sign, rs = "-", -rs
	}
	if rs >= 1e7 {
		return fmt.Sprintf("%s%.2f crore", sign, rs/1e7)
	}
	return sign + rupees(rs)
}

// againstSector is PE over the sector median; unpriced rows sort last.
func againstSector(c *ntxv1.Candidate) float64 {
	if c.Pe == nil || c.SectorPeMedian == nil || c.GetSectorPeMedian() <= 0 {
		return math.Inf(1)
	}
	return c.GetPe() / c.GetSectorPeMedian()
}

// sectorMedians is the median PE of each sector's priced, profitable companies.
func sectorMedians(rows []sqlc.ListScreenerRow, pes map[int64]priced) map[string]float64 {
	bySector := map[string][]float64{}
	for _, row := range rows {
		if pe, ok := pes[row.ID]; ok && pe.PE > 0 {
			bySector[row.Sector] = append(bySector[row.Sector], pe.PE)
		}
	}
	out := map[string]float64{}
	for sector, values := range bySector {
		slices.Sort(values)
		out[sector] = median(values)
	}
	return out
}

// candidatesOpening states the owner's standing rule from the note, then each verdict's count.
func candidatesOpening(cands []*ntxv1.Candidate, p *ntxv1.GetPortfolioResponse, r sqlc.Rule) string {
	head := fmt.Sprintf(
		"Positions stand at %d of %d; a new scrip needs a sale first.",
		p.GetPositions(),
		r.MaxPositions,
	)
	n := map[string]int{}
	for _, c := range cands {
		n[c.GetVerdict()]++
	}
	parts := []string{fmt.Sprintf("%d companies pass the rules", n[Passes])}
	switch n[Passes] {
	case 0:
		parts[0] = "None passes the rules"
	case 1:
		parts[0] = "1 company passes the rules"
	}
	if needs := n[NeedsData]; needs == 1 {
		parts = append(parts, "1 lender needs data")
	} else if needs > 1 {
		parts = append(parts, fmt.Sprintf("%d lenders need data", needs))
	}
	if away := n[TurnedAway]; away == 1 {
		parts = append(parts, "1 is turned away")
	} else if away > 1 {
		parts = append(parts, fmt.Sprintf("%d are turned away", away))
	}
	return head + " " + list(parts) + "."
}
