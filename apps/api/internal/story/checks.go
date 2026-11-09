package story

import "time"

// Verdicts of a Check (GLOSSARY: Check, Holds / Watch). Neither is a recommendation.
const (
	Holds        = "holds"
	Watch        = "watch"
	NeedsHistory = "needs more history"
)

// Axes of the snowflake; each check sits on one, and each is a company page section.
const (
	AxisValue     = "value"
	AxisPast      = "past"
	AxisHealth    = "health"
	AxisDividend  = "dividend"
	AxisOwnership = "ownership"
)

// Check is one published rule and its verdict for this company.
type Check struct {
	Name, Rule, Verdict, Axis string
}

// cashCeiling is the cash payout rule's threshold, % of the fiscal year's profit.
const cashCeiling = 30

// roeFloor is the return on equity rule's threshold, in percent.
const roeFloor = 15

var rules = []struct {
	name, rule, axis string
	eval             func(Story, Input) string
}{
	{"EPS growth", "EPS grew in at least 3 of the last 4 fiscal years and is positive.", AxisPast, epsGrowth},
	{
		"Quarterly profit",
		"Net profit for the latest quarter is above the same quarter last year.",
		AxisPast,
		quarterlyProfit,
	},
	{"PE against its own history", "PE is below its own three-year median.", AxisValue, peOwn},
	{"PB against the sector", "PB is below the sector median.", AxisValue, pbSector},
	{
		"Dividend record",
		"A cash or bonus dividend was paid in each of the last 3 fiscal years.",
		AxisDividend,
		dividendRecord,
	},
	{"Cash payout", "The cash dividend is below 30% of the last fiscal year's profit.", AxisDividend, cashPayout},
	{
		"Promoter holding",
		"The promoter share has not changed in the last 12 months.",
		AxisOwnership,
		promoterHolding,
	},
	{"Book value growth", "Book value per share grew over the last 3 fiscal years.", AxisHealth, bookValueGrowth},
	{"PE against the sector", "PE is below the sector median among profitable companies.", AxisValue, peSector},
	{
		"Return on equity",
		"Annualized EPS is at least 15% of book value per share, both from the latest filing.",
		AxisPast,
		returnOnEquity,
	},
}

// Checks evaluates every published rule on the story's own figures.
func Checks(s Story, in Input) []Check {
	out := make([]Check, 0, len(rules))
	for _, r := range rules {
		out = append(out, Check{Name: r.name, Rule: r.rule, Verdict: r.eval(s, in), Axis: r.axis})
	}
	return out
}

// Rating is one letter for the checks (GLOSSARY: Rating). Holds are counted
// against every published check, so a check without history lowers the letter
// instead of vanishing from it (ADR 0005).
type Rating struct {
	Letter                  string // "" when fewer than minEvaluated checks evaluate
	Holds, Evaluated, Total int32
}

const minEvaluated = 4

// letterFloors are the share of all checks that hold, in percent, for each letter.
var letterFloors = []struct {
	letter string
	floor  int
}{{"A", 85}, {"B", 60}, {"C", 35}}

func rating(checks []Check) Rating {
	r := Rating{Total: int32(len(checks))} //nolint:gosec // ten checks
	for _, c := range checks {
		if c.Verdict == NeedsHistory {
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
	for _, g := range letterFloors {
		// An A means the whole record was seen.
		if g.letter == "A" && r.Evaluated < r.Total {
			continue
		}
		if int(r.Holds)*100 >= g.floor*int(r.Total) {
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

func peSector(s Story, _ Input) string {
	pe := s.Strips[0]
	if pe.Value == nil || pe.Sector.Count == 0 {
		return NeedsHistory
	}
	return verdictOf(*pe.Value < pe.Sector.Median)
}

// returnOnEquity divides two per-share figures from the same filing (GLOSSARY: ROE).
func returnOnEquity(_ Story, in Input) string {
	quarters, _ := parse(in.Fundamentals)
	if len(quarters) == 0 {
		return NeedsHistory
	}
	row := quarters[len(quarters)-1].row
	if !row.Eps.Valid || !row.BookValue.Valid || row.BookValue.Float64 <= 0 {
		return NeedsHistory
	}
	return verdictOf(row.Eps.Float64*100/row.BookValue.Float64 >= roeFloor)
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

// promoterHolding reads the history as the sync writes it: one row per change,
// so the newest row is the last move, or the first recording when there is one row.
func promoterHolding(s Story, in Input) string {
	if len(s.Ownership) == 0 {
		return NeedsHistory
	}
	newest, err := time.Parse(time.DateOnly, s.Ownership[len(s.Ownership)-1].AsOf)
	if err != nil {
		return NeedsHistory
	}
	if !newest.After(in.today().AddDate(-1, 0, 0)) {
		return Holds
	}
	if len(s.Ownership) >= 2 {
		return Watch
	}
	// Recorded for under a year with no change yet: nobody can know.
	return NeedsHistory
}

func bookValueGrowth(s Story, _ Input) string {
	for _, g := range s.Growth {
		if g.Metric == "book_value" && g.Three != nil {
			return verdictOf(*g.Three > 0)
		}
	}
	return NeedsHistory
}
