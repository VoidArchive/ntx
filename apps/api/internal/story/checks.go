package story

import "slices"

// Verdicts of a Check (GLOSSARY: Check, Holds / Watch). Neither is a recommendation.
const (
	Holds        = "holds"
	Watch        = "watch"
	NeedsHistory = "needs more history"
)

// Check is one published rule and its verdict for this company.
type Check struct {
	Name, Rule, Verdict string
}

// cashCeiling is the cash payout rule's threshold, % of the fiscal year's profit.
const cashCeiling = 30

var rules = []struct {
	name, rule string
	eval       func(Story, Input) string
}{
	{"EPS growth", "EPS grew in at least 3 of the last 4 fiscal years and is positive.", epsGrowth},
	{"Quarterly profit", "Net profit for the latest quarter is above the same quarter last year.", quarterlyProfit},
	{"PE against its own history", "PE is below its own three-year median.", peOwn},
	{"PB against the sector", "PB is below the sector median.", pbSector},
	{"Dividend record", "A cash or bonus dividend was paid in each of the last 3 fiscal years.", dividendRecord},
	{"Cash payout", "The cash dividend is below 30% of the last fiscal year's profit.", cashPayout},
	{"Promoter holding", "The promoter share is unchanged over the last 4 observations.", promoterHolding},
	{"Book value growth", "Book value per share grew over the last 3 fiscal years.", bookValueGrowth},
}

// Checks evaluates every published rule on the story's own figures.
func Checks(s Story, in Input) []Check {
	out := make([]Check, 0, len(rules))
	for _, r := range rules {
		out = append(out, Check{Name: r.name, Rule: r.rule, Verdict: r.eval(s, in)})
	}
	return out
}

// Rating is one letter for the checks (GLOSSARY: Rating). Holds are scaled to
// eight so a company with unevaluated checks is rated on the ones it has.
type Rating struct {
	Letter           string // "" when fewer than minEvaluated checks evaluate
	Holds, Evaluated int32
}

const minEvaluated = 4

// aGates must evaluate before an A: EPS growth and quarterly profit measure
// direction, so a shrinking loss passes both; these two measure level.
var aGates = []string{"Quarterly profit", "PE against its own history"}

// ratingScale is the glossary's eight; the rule count may grow, the letter grid does not.
const ratingScale = 8

func rating(checks []Check) Rating {
	var r Rating
	gated := false
	for _, c := range checks {
		if c.Verdict == NeedsHistory {
			gated = gated || slices.Contains(aGates, c.Name)
			continue
		}
		r.Evaluated++
		if c.Verdict == Holds {
			r.Holds++
		}
	}
	if r.Evaluated < minEvaluated {
		return r
	}
	// holds/evaluated >= floor/8 in integers, so the share rounds down: A from 7 of 8, B from 5, C from 3.
	for _, g := range []struct {
		letter string
		floor  int
	}{{"A", 7}, {"B", 5}, {"C", 3}} {
		if g.letter == "A" && gated {
			continue
		}
		if int(r.Holds)*ratingScale >= g.floor*int(r.Evaluated) {
			r.Letter = g.letter
			return r
		}
	}
	r.Letter = "D"
	return r
}

func verdictOf(ok bool) string {
	if ok {
		return Holds
	}
	return Watch
}

// consecutive is the newest n fiscal-year rows, or nil when a year is missing between them.
func consecutive(rows []FiscalYear, n int) []FiscalYear {
	if len(rows) < n {
		return nil
	}
	for i := 1; i < n; i++ {
		if rows[i].Year != rows[i-1].Year-1 {
			return nil
		}
	}
	return rows[:n]
}

func epsGrowth(s Story, _ Input) string {
	rows := consecutive(s.FiscalYears, 5)
	if rows == nil {
		return NeedsHistory
	}
	rises := 0
	for i := range 4 {
		if rows[i].EPS == nil || rows[i+1].EPS == nil {
			return NeedsHistory
		}
		if *rows[i].EPS > *rows[i+1].EPS {
			rises++
		}
	}
	// A loss that shrank three years running is still a loss.
	return verdictOf(rises >= 3 && *rows[0].EPS > 0)
}

func quarterlyProfit(_ Story, in Input) string {
	quarters, _ := parse(in.Fundamentals)
	if len(quarters) == 0 || quarters[len(quarters)-1].profit == nil {
		return NeedsHistory
	}
	cur := quarters[len(quarters)-1]
	prev, ok := sameQuarter(cur.quarter, quarters)[cur.year-1]
	if !ok {
		return NeedsHistory
	}
	return verdictOf(*cur.profit > prev)
}

func peOwn(s Story, _ Input) string {
	pe := s.Strips[0]
	// A median over fewer quarters is not three years.
	if pe.Value == nil || pe.Own.Count < ownQuarters {
		return NeedsHistory
	}
	return verdictOf(*pe.Value < pe.Own.Median)
}

func pbSector(s Story, _ Input) string {
	pb := s.Strips[1]
	if pb.Value == nil || pb.Sector.Count == 0 {
		return NeedsHistory
	}
	return verdictOf(*pb.Value < pb.Sector.Median)
}

// dividendRecord never reads a "not recorded" year as no dividend.
func dividendRecord(s Story, _ Input) string {
	rows := consecutive(s.FiscalYears, 3)
	if rows == nil {
		return NeedsHistory
	}
	for _, r := range rows {
		if r.Payout == "not recorded" {
			return NeedsHistory
		}
	}
	for _, r := range rows {
		if deref(r.Cash) <= 0 && deref(r.Bonus) <= 0 {
			return Watch
		}
	}
	return Holds
}

func cashPayout(s Story, _ Input) string {
	if len(s.FiscalYears) == 0 || s.FiscalYears[0].Payout == "not recorded" {
		return NeedsHistory
	}
	share, ok := cashShare(s.FiscalYears[0])
	if !ok {
		return NeedsHistory
	}
	return verdictOf(share < cashCeiling)
}

func promoterHolding(s Story, _ Input) string {
	if len(s.Ownership) < ownershipWindow {
		return NeedsHistory
	}
	window := s.Ownership[len(s.Ownership)-ownershipWindow:]
	for _, o := range window {
		if o.Promoter == nil {
			return NeedsHistory
		}
	}
	for _, o := range window {
		if *o.Promoter != *window[0].Promoter {
			return Watch
		}
	}
	return Holds
}

func bookValueGrowth(s Story, _ Input) string {
	for _, g := range s.Growth {
		if g.Metric == "book_value" && g.Three != nil {
			return verdictOf(*g.Three > 0)
		}
	}
	return NeedsHistory
}
