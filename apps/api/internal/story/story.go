// Package story turns one company's stored rows into the sentences and
// strips its page opens with (ADR 0001). Build is pure: the handler loads
// rows and sector distributions, Build only computes.
//
// Stored quarterly profit is year-to-date and stored EPS is annualized
// (docs/data.md); every quarter comparison here uses derived quarter-only
// figures and leaves the rows untouched.
package story

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
)

// hydroSector is the database's sector string, which differs from the Go sector enum.
const hydroSector = "Hydro Power"

// parValue is NEPSE's face value per share, so shares outstanding = paid-up capital / 100.
const parValue = 100

// ownQuarters is the company's own three-year window, in quarterly rows.
const ownQuarters = 12

// Input is everything Build needs for one company.
type Input struct {
	Company      sqlc.Company
	Fundamentals []sqlc.Fundamental // any order
	Prices       []sqlc.Price       // newest first; the first two give the quote
	CashDividend *float64           // latest fiscal year's cash dividend, % of par
	Sector       Sector
	Actions      []sqlc.CorporateAction  // notices, newest submitted first
	Ownership    []sqlc.OwnershipHistory // oldest first
	Funds        *Funds                  // nil when the web matched none
}

// Sector holds one value per company in the sector, from each company's latest
// quarterly row at its latest close. PE is unfiltered; Build drops non-positive earnings.
type Sector struct {
	PE, PB, EPS, Yield []float64
}

// Story is the ordered output for one company.
type Story struct {
	Claims      []Claim
	Strips      []Strip
	LastFiling  string // period of the newest filing, "" when there is none
	Quote       *Quote
	FiscalYears []FiscalYear // newest first
	Growth      []Growth
	Ownership   []Observation // oldest first
}

// Claim is one sentence computed from the rows (GLOSSARY: Claim).
type Claim struct {
	Kind   string // earnings | valuation | payout | ownership | funds | record
	Text   string
	Value  float64
	Unit   string // "Rs", "%", "x", ""
	Period string // "Q1 2083/84" or "FY 2082/83"
	Source string // filing URL, "" when not recorded
	Anchor string // id of the evidence section on the page
}

// Strip places one ratio on the sector's distribution and on its own range.
type Strip struct {
	Ratio  string // pe | pb | eps | yield
	Value  *float64
	Sector Range
	Own    Range
}

// Range is low, median and high over Count values; Count 0 means no data.
type Range struct {
	Low, Median, High float64
	Count             int
}

// Quote is the latest close and its change on the previous trading day.
type Quote struct {
	Close         float64
	Change        *float64
	ChangePercent *float64
	Date, DateBS  string
}

type period struct {
	year    fiscal.Year
	quarter fiscal.Quarter
}

func (p period) String() string { return p.quarter.String() + " " + p.year.String() }

// rank orders an annual report after the year's Q4.
func (p period) rank() int {
	if p.quarter == fiscal.Annual {
		return int(p.year)*10 + 5
	}
	return int(p.year)*10 + int(p.quarter)
}

type filing struct {
	period
	row    sqlc.Fundamental
	profit *float64 // quarter-only; nil when the previous quarter is missing
	eps    *float64 // quarter-only
}

// Build computes the story. A claim whose inputs are missing is omitted.
// Claims run earnings, valuation, payout, ownership, funds.
func Build(in Input) Story {
	s := opening(in)
	s.FiscalYears = fiscalYears(in)
	s.Growth = growth(s.FiscalYears)
	s.Claims = append(s.Claims, payout(s.FiscalYears, in.Actions)...)
	s.Ownership = observations(in.Ownership)
	s.Claims = append(s.Claims, ownership(s.Ownership)...)
	s.Claims = append(s.Claims, funds(in.Funds)...)
	return s
}

func opening(in Input) Story {
	quarters, annual := parse(in.Fundamentals)
	s := Story{Quote: quote(in.Prices)}
	s.LastFiling = lastFiling(quarters, annual)
	if len(quarters) == 0 {
		s.Strips = strips(in, nil, nil)
		return s
	}
	cur := quarters[len(quarters)-1]
	s.Claims = append(s.Claims, earnings(in.Company, quarters, annual)...)
	s.Strips = strips(in, s.Quote, quarters)
	s.Claims = append(s.Claims, valuation(cur, s.Quote, s.Strips)...)
	return s
}

func parse(rows []sqlc.Fundamental) ([]filing, map[fiscal.Year]sqlc.Fundamental) {
	var quarters []filing
	annual := map[fiscal.Year]sqlc.Fundamental{}
	for _, r := range rows {
		y, err := fiscal.ParseYear(r.FiscalYear)
		if err != nil {
			continue
		}
		q, err := fiscal.ParseQuarter(r.Quarter)
		if err != nil {
			continue
		}
		if q == fiscal.Annual {
			annual[y] = r
			continue
		}
		quarters = append(quarters, filing{period: period{y, q}, row: r})
	}
	slices.SortFunc(quarters, func(a, b filing) int { return cmp.Compare(a.rank(), b.rank()) })

	ytd := map[period]float64{}
	for _, f := range quarters {
		if f.row.ProfitAmount.Valid {
			ytd[f.period] = f.row.ProfitAmount.Float64
		}
	}
	for i := range quarters {
		f := &quarters[i]
		this, ok := ytd[f.period]
		if !ok {
			continue
		}
		profit := this
		if f.quarter != fiscal.Q1 {
			prev, ok := ytd[period{f.year, f.quarter - 1}]
			if !ok {
				continue
			}
			profit = this - prev
		}
		f.profit = &profit
		if f.row.PaidUpCapital.Valid && f.row.PaidUpCapital.Float64 > 0 {
			eps := profit / (f.row.PaidUpCapital.Float64 / parValue)
			f.eps = &eps
		}
	}
	return quarters, annual
}

func lastFiling(quarters []filing, annual map[fiscal.Year]sqlc.Fundamental) string {
	var latest *period
	if len(quarters) > 0 {
		latest = &quarters[len(quarters)-1].period
	}
	for y := range annual {
		p := period{y, fiscal.Annual}
		if latest == nil || p.rank() > latest.rank() {
			latest = &p
		}
	}
	if latest == nil {
		return ""
	}
	return latest.String()
}

func quote(prices []sqlc.Price) *Quote {
	if len(prices) == 0 || !prices[0].ClosePrice.Valid {
		return nil
	}
	p := prices[0]
	q := &Quote{Close: p.ClosePrice.Float64, Date: p.BusinessDate}
	if d, err := time.Parse(time.DateOnly, p.BusinessDate); err == nil {
		q.DateBS, _ = fiscal.BSDate(d)
	}
	// Backfilled rows have no change columns, so fall back to the previous close.
	if p.ChangeAmount.Valid {
		q.Change = &p.ChangeAmount.Float64
	}
	if q.Change == nil && len(prices) > 1 && prices[1].ClosePrice.Valid {
		change := q.Close - prices[1].ClosePrice.Float64
		q.Change = &change
	}
	if q.Change != nil && q.Close != *q.Change {
		pct := *q.Change / (q.Close - *q.Change) * 100
		q.ChangePercent = &pct
	}
	return q
}

var numberWords = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}

var ordinals = []string{
	"",
	"first",
	"second",
	"third",
	"fourth",
	"fifth",
	"sixth",
	"seventh",
	"eighth",
	"ninth",
	"tenth",
}

func earnings(c sqlc.Company, quarters []filing, annual map[fiscal.Year]sqlc.Fundamental) []Claim {
	cur := quarters[len(quarters)-1]
	claim := func(kind, text string, value float64, unit string) Claim {
		return Claim{
			Kind:   kind,
			Text:   text,
			Value:  value,
			Unit:   unit,
			Period: cur.String(),
			Source: cur.row.SourceUrl.String,
			Anchor: "quarterly",
		}
	}
	var out []Claim

	if cur.profit != nil {
		word := "profit"
		if *cur.profit < 0 {
			word = "loss"
		}
		out = append(
			out,
			claim(
				"earnings",
				fmt.Sprintf("Net %s for %s was Rs %s crore.", word, cur, crore(*cur.profit)),
				*cur.profit,
				"Rs",
			),
		)
		out = append(out, records(cur, quarters, claim)...)
	}

	if cur.eps != nil && cur.row.Eps.Valid {
		annualized := cur.row.Eps.Float64
		text := fmt.Sprintf("EPS was Rs %.2f for the quarter and Rs %.2f annualized.", *cur.eps, annualized)
		fy, fyEPS, ok := lastFullYear(cur.year, quarters, annual)
		if c.Sector == hydroSector && ok {
			text = fmt.Sprintf(
				"EPS was Rs %.2f for the quarter, Rs %.2f annualized and Rs %.2f for FY %s.",
				*cur.eps,
				annualized,
				fyEPS,
				fy,
			)
		}
		out = append(out, claim("earnings", text, *cur.eps, "Rs"))
		if c.Sector == hydroSector && cur.quarter == fiscal.Q1 {
			out = append(
				out,
				claim(
					"earnings",
					"Q1 is the monsoon quarter, so the annualized EPS overstates the year.",
					annualized,
					"Rs",
				),
			)
		}
	}
	return out
}

// claimFn builds a claim stamped with the latest quarter's period, source and anchor.
type claimFn func(kind, text string, value float64, unit string) Claim

// records compares the latest quarter-only profit with the same quarter of earlier years.
func records(cur filing, quarters []filing, claim claimFn) []Claim {
	same := sameQuarter(cur.quarter, quarters)
	out := yearOnYear(cur, same, claim)
	out = append(out, best(cur, same, claim)...)
	out = append(out, streak(cur, same, claim)...)
	return append(out, firstLoss(cur, quarters, claim)...)
}

// sameQuarter is each year's quarter-only profit for one quarter.
func sameQuarter(q fiscal.Quarter, quarters []filing) map[fiscal.Year]float64 {
	same := map[fiscal.Year]float64{}
	for _, f := range quarters {
		if f.quarter == q && f.profit != nil {
			same[f.year] = *f.profit
		}
	}
	return same
}

func yearOnYear(cur filing, same map[fiscal.Year]float64, claim claimFn) []Claim {
	prev, ok := same[cur.year-1]
	if !ok || prev <= 0 {
		return nil
	}
	change := (*cur.profit - prev) / prev * 100
	dir := "Up"
	if change < 0 {
		dir = "Down"
	}
	text := fmt.Sprintf("%s %.1f%% on %s.", dir, abs(change), period{cur.year - 1, cur.quarter})
	return []Claim{claim("earnings", text, change, "%")}
}

func best(cur filing, same map[fiscal.Year]float64, claim claimFn) []Claim {
	years := 0
	for y := cur.year - 4; y <= cur.year; y++ {
		v, ok := same[y]
		if !ok {
			continue
		}
		if y != cur.year && v >= *cur.profit {
			return nil
		}
		years++
	}
	if years < 3 {
		return nil
	}
	text := fmt.Sprintf("Best %s in %s years.", cur.quarter, numberWords[years])
	return []Claim{claim("record", text, *cur.profit, "Rs")}
}

func streak(cur filing, same map[fiscal.Year]float64, claim claimFn) []Claim {
	n := 0
	for y := cur.year; ; y-- {
		v, ok := same[y]
		prev, okPrev := same[y-1]
		if !ok || !okPrev || v <= prev {
			break
		}
		n++
	}
	if n < 2 {
		return nil
	}
	ordinal := capitalize(ordinals[min(n, len(ordinals)-1)])
	text := fmt.Sprintf("%s consecutive year-on-year increase in %s.", ordinal, cur.quarter)
	return []Claim{claim("record", text, float64(n), "")}
}

func firstLoss(cur filing, quarters []filing, claim claimFn) []Claim {
	if *cur.profit >= 0 {
		return nil
	}
	profitable := 0
	for i := len(quarters) - 2; i >= 0 && quarters[i].profit != nil && *quarters[i].profit > 0; i-- {
		profitable++
	}
	if profitable < 4 {
		return nil
	}
	return []Claim{claim("record", fmt.Sprintf("First loss in %d quarters.", profitable), *cur.profit, "Rs")}
}

// lastFullYear prefers the audited annual EPS, falling back to the year's Q4 row.
func lastFullYear(
	upTo fiscal.Year,
	quarters []filing,
	annual map[fiscal.Year]sqlc.Fundamental,
) (fiscal.Year, float64, bool) {
	q4 := map[fiscal.Year]float64{}
	for _, f := range quarters {
		if f.quarter == fiscal.Q4 && f.row.Eps.Valid {
			q4[f.year] = f.row.Eps.Float64
		}
	}
	for y := upTo; y >= upTo-10; y-- {
		if r, ok := annual[y]; ok && r.Eps.Valid {
			return y, r.Eps.Float64, true
		}
		if v, ok := q4[y]; ok {
			return y, v, true
		}
	}
	return 0, 0, false
}

func strips(in Input, q *Quote, quarters []filing) []Strip {
	pe, pb, eps, yield := current(in, q, quarters)
	ownPE, ownPB, ownEPS := ownRanges(quarters[max(0, len(quarters)-ownQuarters):])
	return []Strip{
		{Ratio: "pe", Value: pe, Sector: rangeOf(positive(in.Sector.PE)), Own: rangeOf(ownPE)},
		{Ratio: "pb", Value: pb, Sector: rangeOf(in.Sector.PB), Own: rangeOf(ownPB)},
		{Ratio: "eps", Value: eps, Sector: rangeOf(in.Sector.EPS), Own: rangeOf(ownEPS)},
		// shortcut: no own yield range, it needs a price per fiscal year; add with the fiscal-year rows (#25).
		{Ratio: "yield", Value: yield, Sector: rangeOf(in.Sector.Yield)},
	}
}

// current is this company's PE, PB, EPS and yield at the latest close, computed
// the same way as the sector queries so the strips compare like with like.
func current(in Input, q *Quote, quarters []filing) (pe, pb, eps, yield *float64) {
	if len(quarters) == 0 {
		return nil, nil, nil, nil
	}
	cur := quarters[len(quarters)-1].row
	if cur.Eps.Valid {
		eps = &cur.Eps.Float64
	}
	if q == nil || q.Close == 0 {
		return nil, nil, eps, nil
	}
	if eps != nil && *eps > 0 {
		pe = ptr(q.Close / *eps)
	}
	if cur.BookValue.Valid && cur.BookValue.Float64 != 0 {
		pb = ptr(q.Close / cur.BookValue.Float64)
	}
	if in.CashDividend != nil {
		yield = ptr(*in.CashDividend * parValue / q.Close)
	}
	return pe, pb, eps, yield
}

func ownRanges(own []filing) (pe, pb, eps []float64) {
	for _, f := range own {
		r := f.row
		if r.PeRatio.Valid && r.PeRatio.Float64 > 0 {
			pe = append(pe, r.PeRatio.Float64)
		}
		// Prices go back one year only; the stored PE (price at filing / EPS) gives back the price.
		if r.PeRatio.Valid && r.Eps.Valid && r.BookValue.Valid && r.BookValue.Float64 != 0 {
			pb = append(pb, r.PeRatio.Float64*r.Eps.Float64/r.BookValue.Float64)
		}
		if r.Eps.Valid {
			eps = append(eps, r.Eps.Float64)
		}
	}
	return pe, pb, eps
}

func valuation(cur filing, q *Quote, strips []Strip) []Claim {
	claim := func(text string, value float64, anchor string) Claim {
		return Claim{
			Kind:   "valuation",
			Text:   text,
			Value:  value,
			Unit:   "x",
			Period: cur.String(),
			Source: cur.row.SourceUrl.String,
			Anchor: anchor,
		}
	}
	pe, pb := strips[0], strips[1]
	var out []Claim
	if pe.Value != nil && pe.Sector.Count > 0 {
		out = append(
			out,
			claim(
				fmt.Sprintf(
					"PE is %.2fx at Rs %.2f, against a sector median of %.2fx (%d profitable).",
					*pe.Value,
					q.Close,
					pe.Sector.Median,
					pe.Sector.Count,
				),
				*pe.Value,
				"strip-pe",
			),
		)
	}
	if pe.Value != nil && pe.Own.Count > 0 {
		out = append(
			out,
			claim(fmt.Sprintf("Its three-year median PE is %.2fx.", pe.Own.Median), pe.Own.Median, "strip-pe"),
		)
	}
	if pb.Value != nil && pb.Sector.Count > 0 {
		out = append(
			out,
			claim(
				fmt.Sprintf("PB is %.2fx, against a sector median of %.2fx.", *pb.Value, pb.Sector.Median),
				*pb.Value,
				"strip-pb",
			),
		)
	}
	if pb.Value != nil && pb.Own.Count > 0 {
		out = append(
			out,
			claim(fmt.Sprintf("Its three-year median PB is %.2fx.", pb.Own.Median), pb.Own.Median, "strip-pb"),
		)
	}
	return out
}

func rangeOf(values []float64) Range {
	if len(values) == 0 {
		return Range{}
	}
	v := slices.Sorted(slices.Values(values))
	n := len(v)
	median := v[n/2]
	if n%2 == 0 {
		median = (v[n/2-1] + v[n/2]) / 2
	}
	return Range{Low: v[0], Median: median, High: v[n-1], Count: n}
}

func positive(values []float64) []float64 {
	var out []float64
	for _, v := range values {
		if v > 0 {
			out = append(out, v)
		}
	}
	return out
}

func crore(rs float64) string { return fmt.Sprintf("%.2f", abs(rs)/1e7) }

func abs(v float64) float64 { return max(v, -v) }

func ptr(v float64) *float64 { return &v }

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-'a'+'A') + s[1:]
}
