package story

import (
	"fmt"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
)

// trailingQuarters is the window of a trailing EPS: one fiscal year of quarters.
const trailingQuarters = 4

// Earnings is a company's latest per-share earnings on the basis the engine
// prices it: the sum of the last four quarter-only profits over today's shares
// when those quarters are consecutive, else the exchange's annualized figure.
type Earnings struct {
	EPS      float64
	Period   string // "4Q to Q1 2083/84" when trailing, else the filing, "Q1 2083/84"
	Trailing bool
}

// LatestEarnings prices a company's earnings from its stored quarterly rows.
// False when there is no quarterly row with an EPS.
func LatestEarnings(rows []sqlc.Fundamental) (Earnings, bool) {
	quarters, _ := parse(rows)
	if len(quarters) == 0 {
		return Earnings{}, false
	}
	cur := quarters[len(quarters)-1]
	if t := trailingAt(len(quarters)-1, quarters); t != nil {
		return Earnings{EPS: *t, Period: fmt.Sprintf("4Q to %s", cur), Trailing: true}, true
	}
	if !cur.row.Eps.Valid {
		return Earnings{}, false
	}
	return Earnings{EPS: cur.row.Eps.Float64, Period: cur.String()}, true
}

// trailingAt is the trailing four-quarter EPS as of quarters[i]: the four
// quarter-only profits ending there, consecutive, over the shares of that row.
// Nil when a quarter is missing or the share count is unknown.
func trailingAt(i int, quarters []filing) *float64 {
	if i+1 < trailingQuarters {
		return nil
	}
	cur := quarters[i]
	if !cur.row.PaidUpCapital.Valid || cur.row.PaidUpCapital.Float64 <= 0 {
		return nil
	}
	total := 0.0
	for k := range trailingQuarters {
		f := quarters[i-k]
		if f.profit == nil || f.period != quartersBack(cur.period, k) {
			return nil
		}
		total += *f.profit
	}
	return ptr(total / (cur.row.PaidUpCapital.Float64 / parValue))
}

// quartersBack is the period k quarters before p, crossing the fiscal-year boundary.
func quartersBack(p period, k int) period {
	q := int(p.quarter) - k
	if q < 1 {
		return period{p.year - 1, fiscal.Quarter(q + 4)}
	}
	return period{p.year, fiscal.Quarter(q)}
}

// SectorPE prices every company in the rows the way LatestEarnings does; a
// company whose earnings are not positive is left out, as Profitable would.
func SectorPE(rows []sqlc.ListEarningsRowsRow) []float64 {
	byCompany, closes, order := Quarterly(rows)
	var out []float64
	for _, id := range order {
		e, ok := LatestEarnings(byCompany[id])
		if ok && e.EPS > 0 && closes[id] > 0 {
			out = append(out, closes[id]/e.EPS)
		}
	}
	return out
}

// Quarterly groups earnings rows by company, with each company's latest close
// (0 without one) and the companies in row order.
// ByCompany groups the quarterly rows by company for pricing, as the screener and
// the owner's Candidates both do.
func ByCompany(rows []sqlc.ListEarningsRowsRow) map[int64][]sqlc.Fundamental {
	out := map[int64][]sqlc.Fundamental{}
	for _, e := range rows {
		out[e.Fundamental.CompanyID] = append(out[e.Fundamental.CompanyID], e.Fundamental)
	}
	return out
}

func Quarterly(rows []sqlc.ListEarningsRowsRow) (map[int64][]sqlc.Fundamental, map[int64]float64, []int64) {
	byCompany := map[int64][]sqlc.Fundamental{}
	closes := map[int64]float64{}
	var order []int64
	for _, r := range rows {
		id := r.Fundamental.CompanyID
		if _, seen := closes[id]; !seen {
			order = append(order, id)
		}
		byCompany[id] = append(byCompany[id], r.Fundamental)
		closes[id] = r.Close.Float64
	}
	return byCompany, closes, order
}
