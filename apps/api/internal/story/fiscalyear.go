package story

import (
	"database/sql"
	"maps"
	"math"
	"slices"
	"time"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
)

// boundaryDays is how far from a fiscal-year boundary a close may sit and
// still stand for it; further and the return would cover part of the year.
const boundaryDays = 10

// FiscalYear is one row of the fiscal-year strip (GLOSSARY: Payout, Bonus-adjusted return).
type FiscalYear struct {
	Year                   fiscal.Year
	EPS, Profit, BookValue *float64 // from the annual report, else the Q4 filing
	Shares                 *float64 // paid-up capital / par
	Bonus, Cash, Rights    *float64 // % of par, nil when no notice
	Payout                 string   // paid | none | not recorded
	Return                 *float64 // bonus-adjusted, %; rights excluded
	Source                 string   // notice URL
}

// Growth is compounded annual growth in %, nil where a year is missing.
type Growth struct {
	Metric           string // profit | eps | book_value | price
	One, Three, Five *float64
}

// fiscalYears has a row for every year with a full-year filing or a notice, newest first.
func fiscalYears(in Input) []FiscalYear {
	quarters, annual := parse(in.Fundamentals)
	full := map[fiscal.Year]sqlc.Fundamental{}
	for _, f := range quarters {
		if f.quarter == fiscal.Q4 {
			full[f.year] = f.row
		}
	}
	for y, r := range annual {
		full[y] = r
	}
	notices := noticesByYear(in.Actions)

	years := map[fiscal.Year]bool{}
	for y := range full {
		years[y] = true
	}
	for y := range notices {
		years[y] = true
	}
	var out []FiscalYear
	for _, y := range slices.Sorted(maps.Keys(years)) {
		row := FiscalYear{Year: y, Payout: payoutStatus(y, notices, annual)}
		if r, ok := full[y]; ok {
			row.EPS, row.Profit, row.BookValue = val(r.Eps), val(r.ProfitAmount), val(r.BookValue)
			if r.PaidUpCapital.Valid {
				row.Shares = ptr(r.PaidUpCapital.Float64 / parValue)
			}
		}
		if n, ok := notices[y]; ok {
			row.Bonus, row.Cash, row.Rights = val(n.BonusPercentage), val(n.CashDividend), val(n.RightPercentage)
			row.Source = n.NoticeUrl.String
		}
		row.Return = bonusAdjustedReturn(y, in.Prices, row.Bonus, row.Cash)
		out = append(out, row)
	}
	slices.Reverse(out)
	return out
}

// noticesByYear keeps the newest notice per fiscal year; actions arrive newest first.
func noticesByYear(actions []sqlc.CorporateAction) map[fiscal.Year]sqlc.CorporateAction {
	out := map[fiscal.Year]sqlc.CorporateAction{}
	for _, a := range actions {
		y, err := fiscal.ParseYear(a.FiscalYear)
		if err != nil {
			continue
		}
		if _, ok := out[y]; !ok {
			out[y] = a
		}
	}
	return out
}

// bonusAdjustedReturn is (close_end x (1 + bonus) + cash per share) / close_start - 1, in %.
// Missing bonus or cash counts as none; the row's payout status says whether that is known.
func bonusAdjustedReturn(y fiscal.Year, prices []sqlc.Price, bonus, cash *float64) *float64 {
	start, err := y.Start()
	if err != nil {
		return nil
	}
	end, err := y.End()
	if err != nil {
		return nil
	}
	first, last := day(start), day(end)
	firstLimit, lastLimit := day(start.AddDate(0, 0, boundaryDays)), day(end.AddDate(0, 0, -boundaryDays))
	var open, close *float64
	// Prices are newest first: the last match on or after start is the earliest.
	for _, p := range prices {
		if !p.ClosePrice.Valid {
			continue
		}
		d := p.BusinessDate
		if close == nil && d <= last && d >= lastLimit {
			close = ptr(p.ClosePrice.Float64)
		}
		if d >= first && d <= firstLimit {
			open = ptr(p.ClosePrice.Float64)
		}
	}
	if open == nil || close == nil || *open == 0 {
		return nil
	}
	r := (*close*(1+deref(bonus)/100)+deref(cash)/100*parValue) / *open - 1
	return ptr(r * 100)
}

func growth(rows []FiscalYear) []Growth {
	out := []Growth{}
	for _, m := range []struct {
		metric string
		get    func(FiscalYear) *float64
	}{
		{"profit", func(r FiscalYear) *float64 { return r.Profit }},
		{"eps", func(r FiscalYear) *float64 { return r.EPS }},
		{"book_value", func(r FiscalYear) *float64 { return r.BookValue }},
	} {
		values := map[fiscal.Year]float64{}
		for _, r := range rows {
			if v := m.get(r); v != nil {
				values[r.Year] = *v
			}
		}
		g := Growth{Metric: m.metric}
		g.One, g.Three, g.Five = compound(values, 1), compound(values, 3), compound(values, 5)
		out = append(out, g)
	}
	returns := map[fiscal.Year]float64{}
	for _, r := range rows {
		if r.Return != nil {
			returns[r.Year] = *r.Return
		}
	}
	g := Growth{Metric: "price"}
	g.One, g.Three, g.Five = chained(returns, 1), chained(returns, 3), chained(returns, 5)
	return append(out, g)
}

// compound is the annual growth from n years before the newest value; growth
// from or to a loss has no meaning, so non-positive ends give nil.
func compound(values map[fiscal.Year]float64, n int) *float64 {
	if len(values) == 0 {
		return nil
	}
	newest := slices.Max(slices.Collect(maps.Keys(values)))
	v, base := values[newest], values[newest-fiscal.Year(n)]
	if v <= 0 || base <= 0 {
		return nil
	}
	return ptr((math.Pow(v/base, 1/float64(n)) - 1) * 100)
}

// chained compounds the newest n consecutive yearly returns into an annual rate.
func chained(returns map[fiscal.Year]float64, n int) *float64 {
	if len(returns) == 0 {
		return nil
	}
	newest := slices.Max(slices.Collect(maps.Keys(returns)))
	total := 1.0
	for y := newest; y > newest-fiscal.Year(n); y-- {
		r, ok := returns[y]
		if !ok {
			return nil
		}
		total *= 1 + r/100
	}
	return ptr((math.Pow(total, 1/float64(n)) - 1) * 100)
}

func day(t time.Time) string { return t.Format(time.DateOnly) }

func val(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	return ptr(n.Float64)
}

func deref(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
