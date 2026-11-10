package portfolio

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/story"
)

// isHydro reads the sector string loosely: the exchange writes "Hydro Power", fixtures "Hydropower".
func isHydro(sector string) bool { return strings.Contains(strings.ToLower(sector), "hydro") }

// letterRank orders ratings; an unrated company ("") never qualifies.
var letterRank = map[string]int{"A": 0, "B": 1, "C": 2, "D": 3}

// rulebook is what the rules compare a company against: what is held, each sector's
// value, and the basins the owner's plants sit in.
type rulebook struct {
	held        map[string]bool
	sectorValue map[string]float64
	heldBasins  map[string]bool
	basins      map[string][]string // by symbol, from the plant table
	total       float64
	medians     map[string]float64 // sector PE medians
	pes         map[int64]float64
	rules       sqlc.Rule
	minRank     int
}

func bookOf(
	rows []sqlc.ListScreenerRow,
	pes map[int64]float64,
	p *ntxv1.GetPortfolioResponse,
	r sqlc.Rule,
	plants []sqlc.HydroPlant,
) rulebook {
	b := rulebook{
		held: map[string]bool{}, sectorValue: map[string]float64{}, heldBasins: map[string]bool{},
		basins: map[string][]string{}, total: p.GetTotalValue(), medians: sectorMedians(rows, pes), pes: pes, rules: r,
	}
	for _, h := range p.GetHoldings() {
		b.held[h.GetSymbol()] = true
		if h.GetQuantity() > 0 {
			b.sectorValue[h.GetSector()] += h.GetValue()
		}
	}
	for _, pl := range plants {
		if !slices.Contains(b.basins[pl.Symbol], pl.Basin) {
			b.basins[pl.Symbol] = append(b.basins[pl.Symbol], pl.Basin)
		}
		if b.held[pl.Symbol] {
			b.heldBasins[pl.Basin] = true
		}
	}
	rank, ok := letterRank[strings.ToUpper(r.CandidateMinRating)]
	if !ok {
		rank = letterRank["B"]
	}
	b.minRank = rank
	return b
}

// candidates filters the screener by the owner's rules (GLOSSARY: Candidates): not
// held, rated at least the minimum, the Dividend record check holding when asked,
// sector headroom after a minimum entry, and for a hydro no plant in a basin the
// owner already holds. pes are each company's PE by id, as the screener prices it.
func candidates(
	rows []sqlc.ListScreenerRow,
	pes map[int64]float64,
	p *ntxv1.GetPortfolioResponse,
	r sqlc.Rule,
	plants []sqlc.HydroPlant,
) []*ntxv1.Candidate {
	b := bookOf(rows, pes, p, r, plants)
	var out []*ntxv1.Candidate
	for _, row := range rows {
		if c := b.admit(row); c != nil {
			out = append(out, c)
		}
	}
	slices.SortStableFunc(out, func(a, b *ntxv1.Candidate) int {
		if d := cmp.Compare(letterRank[a.GetRating()], letterRank[b.GetRating()]); d != 0 {
			return d
		}
		return cmp.Compare(againstSector(a), againstSector(b))
	})
	return out
}

// turnedAway applies the rules that need no figure of the company's own: held,
// unrated or rated too low, unpriced, no dividend record, no sector headroom.
func (b rulebook) turnedAway(row sqlc.ListScreenerRow) bool {
	r := b.rules
	rank, rated := letterRank[row.Rating.String]
	if b.held[row.Symbol] || !rated || rank > b.minRank || !row.ClosePrice.Valid || row.ClosePrice.Float64 <= 0 {
		return true
	}
	if r.CandidateDividendRecord != 0 && row.DividendRecord.String != story.Holds {
		return true
	}
	// An empty book has no sector to cap.
	return b.total > 0 && (b.sectorValue[row.Sector]+r.MinEntryRs)*100/(b.total+r.MinEntryRs) > r.SectorCapPercent
}

// admit is one company against the rules; nil when any rule turns it away.
func (b rulebook) admit(row sqlc.ListScreenerRow) *ntxv1.Candidate {
	if b.turnedAway(row) {
		return nil
	}
	r := b.rules
	c := &ntxv1.Candidate{
		Symbol: row.Symbol, Name: row.Name, Sector: row.Sector, Rating: row.Rating.String,
		Close: row.ClosePrice.Float64, CloseDay: row.BusinessDate.String,
	}
	// The basin rule reads the plant table, whatever the sector string says.
	if basins := b.basins[row.Symbol]; len(basins) > 0 {
		if r.OneHydroPerBasin != 0 && slices.ContainsFunc(basins, func(x string) bool { return b.heldBasins[x] }) {
			return nil
		}
		c.Basin = strings.Join(basins, ", ")
	} else if isHydro(row.Sector) {
		c.Note = "basin not recorded"
	}
	if pe, ok := b.pes[row.ID]; ok {
		c.Pe = &pe
	}
	if m, ok := b.medians[row.Sector]; ok {
		c.SectorPeMedian = &m
	}
	// The smallest whole lot that reaches the minimum entry, and the months of budget it costs.
	c.Shares = int64(math.Ceil(r.MinEntryRs/c.Close/lot)) * lot
	c.CostRs = float64(c.Shares) * c.Close
	if r.MonthlyBudgetRs > 0 {
		c.Months = c.CostRs / r.MonthlyBudgetRs
	}
	return c
}

// againstSector is PE over the sector median; unpriced rows sort last.
func againstSector(c *ntxv1.Candidate) float64 {
	if c.Pe == nil || c.SectorPeMedian == nil || c.GetSectorPeMedian() <= 0 {
		return math.Inf(1)
	}
	return c.GetPe() / c.GetSectorPeMedian()
}

// sectorMedians is the median PE of each sector's priced, profitable companies.
func sectorMedians(rows []sqlc.ListScreenerRow, pes map[int64]float64) map[string]float64 {
	bySector := map[string][]float64{}
	for _, row := range rows {
		if pe, ok := pes[row.ID]; ok && pe > 0 {
			bySector[row.Sector] = append(bySector[row.Sector], pe)
		}
	}
	out := map[string]float64{}
	for sector, values := range bySector {
		slices.Sort(values)
		out[sector] = median(values)
	}
	return out
}

// candidatesOpening states the owner's standing rule from the note, then the count.
func candidatesOpening(n int, p *ntxv1.GetPortfolioResponse, r sqlc.Rule) string {
	head := fmt.Sprintf(
		"Positions stand at %d of %d; a new scrip needs a sale first.",
		p.GetPositions(),
		r.MaxPositions,
	)
	switch n {
	case 0:
		return head + " None passes the rules."
	case 1:
		return head + " 1 company passes the rules."
	}
	return fmt.Sprintf("%s %d companies pass the rules.", head, n)
}
