package portfolio

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// horizon is the furthest the projection looks; past it the sentence says "beyond ten years".
const horizon = 120

// flow is one dated movement of the owner's money: paid in positive, taken out negative.
type flow struct {
	day time.Time
	rs  float64
}

// goal measures the book against the owner's goal (GLOSSARY: Goal): the value, what
// was paid in, the money-weighted return since the first transaction, the month the
// goal is reached at flat prices and at that return, the value by month for the chart,
// and the sentence. closes are each held company's month-end closes by symbol.
func goal(
	today time.Time,
	r sqlc.Rule,
	p *ntxv1.GetPortfolioResponse,
	closes map[string][]sqlc.ListMonthEndClosesRow,
) *ntxv1.Goal {
	g := &ntxv1.Goal{GoalRs: r.GoalRs, ValueRs: p.GetTotalValue(), MonthlyRs: r.MonthlyBudgetRs + r.FundSipRs}
	flows := cashFlows(p.GetHoldings())
	for _, f := range flows {
		g.ContributedRs += f.rs
	}
	if len(flows) > 0 {
		g.Since = flows[0].day.Format(time.DateOnly)
		if rate, ok := irr(flows, today, g.GetValueRs()); ok {
			pct := rate * 100
			g.ReturnPercent = &pct
		}
		g.Series = valueByMonth(flows[0].day, today, p.GetHoldings(), closes, g.GetValueRs())
	}
	g.FlatMonth, g.ReturnMonth, g.Band = project(today, g)
	g.Opening = goalOpening(g, r)
	return g
}

// cashFlows is every purchase and sale across the book, oldest first; a bonus lot is
// free and moves no money.
func cashFlows(holdings []*ntxv1.Holding) []flow {
	var out []flow
	for _, h := range holdings {
		for _, tx := range h.GetTransactions() {
			day, err := time.Parse(time.DateOnly, tx.GetDay())
			if err != nil {
				continue
			}
			rs := float64(tx.GetQuantity()) * tx.GetPrice()
			switch tx.GetKind() {
			case ntxv1.TransactionKind_TRANSACTION_KIND_SELL:
				rs = -rs
			case ntxv1.TransactionKind_TRANSACTION_KIND_BONUS:
				continue
			}
			out = append(out, flow{day, rs})
		}
	}
	slices.SortStableFunc(out, func(a, b flow) int { return a.day.Compare(b.day) })
	return out
}

// irr is the annual rate at which the dated flows grow into today's value, by
// bisection; false when no rate between -99% and +1000% a year fits.
func irr(flows []flow, today time.Time, value float64) (float64, bool) {
	if value <= 0 {
		return 0, false
	}
	npv := func(rate float64) float64 {
		total := 0.0
		for _, f := range flows {
			years := today.Sub(f.day).Hours() / 24 / 365.25
			total -= f.rs * math.Pow(1+rate, years)
		}
		return total + value
	}
	lo, hi := -0.99, 10.0
	if npv(lo) > 0 == (npv(hi) > 0) {
		return 0, false
	}
	for range 200 {
		mid := (lo + hi) / 2
		if npv(mid) > 0 == (npv(lo) > 0) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2, true
}

// valueByMonth is the book's value and what was paid in at each month end from the first
// transaction to today: shares held at the latest close that month, fund units at
// what they cost (NAV history is not stored), and the latest month at today's value.
func valueByMonth(
	since, today time.Time,
	holdings []*ntxv1.Holding,
	closes map[string][]sqlc.ListMonthEndClosesRow,
	value float64,
) []*ntxv1.GoalPoint {
	var out []*ntxv1.GoalPoint
	last := today.Format("2006-01")
	for m := time.Date(since.Year(), since.Month(), 1, 0, 0, 0, 0, time.UTC); ; m = m.AddDate(0, 1, 0) {
		month := m.Format("2006-01")
		end := m.AddDate(0, 1, -1).Format(time.DateOnly)
		pt := &ntxv1.GoalPoint{Month: month}
		for _, h := range holdings {
			qty, cost := heldAt(h.GetTransactions(), end)
			pt.ContributedRs += cost
			if h.GetIsFund() {
				pt.ValueRs += cost
				continue
			}
			if close, ok := closeAt(closes[h.GetSymbol()], month); ok {
				pt.ValueRs += float64(qty) * close
			}
		}
		if month == last {
			pt.ValueRs = value
			out = append(out, pt)
			return out
		}
		out = append(out, pt)
	}
}

// heldAt is the shares held and the net money paid in on or before a day.
func heldAt(txs []*ntxv1.Transaction, day string) (int64, float64) {
	var qty int64
	cost := 0.0
	for _, tx := range txs {
		if tx.GetDay() > day {
			continue
		}
		switch tx.GetKind() {
		case ntxv1.TransactionKind_TRANSACTION_KIND_SELL:
			qty -= tx.GetQuantity()
			cost -= float64(tx.GetQuantity()) * tx.GetPrice()
		case ntxv1.TransactionKind_TRANSACTION_KIND_BONUS:
			qty += tx.GetQuantity()
		default:
			qty += tx.GetQuantity()
			cost += float64(tx.GetQuantity()) * tx.GetPrice()
		}
	}
	return qty, cost
}

// closeAt is the last month-end close at or before a month; rows are oldest first.
func closeAt(rows []sqlc.ListMonthEndClosesRow, month string) (float64, bool) {
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i].Month <= month && rows[i].ClosePrice.Valid {
			return rows[i].ClosePrice.Float64, true
		}
	}
	return 0, false
}

// project walks forward a month at a time until each path reaches the goal or the
// horizon: the flat path adds the monthly figure, the return path compounds too.
// The band carries both paths for the chart; a month name is empty past the horizon.
func project(today time.Time, g *ntxv1.Goal) (flatMonth, returnMonth string, band []*ntxv1.GoalPoint) {
	if g.GetValueRs() >= g.GetGoalRs() || g.GetMonthlyRs() <= 0 {
		return "", "", nil
	}
	monthly := 1.0
	if g.ReturnPercent != nil {
		monthly = math.Pow(1+g.GetReturnPercent()/100, 1.0/12)
	}
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	flat, grown := g.GetValueRs(), g.GetValueRs()
	band = append(band, &ntxv1.GoalPoint{Month: start.Format("2006-01"), FlatRs: flat, WithReturnRs: grown})
	for k := 1; k <= horizon; k++ {
		m := start.AddDate(0, k, 0)
		flat += g.GetMonthlyRs()
		grown = grown*monthly + g.GetMonthlyRs()
		band = append(band, &ntxv1.GoalPoint{Month: m.Format("2006-01"), FlatRs: flat, WithReturnRs: grown})
		if flatMonth == "" && flat >= g.GetGoalRs() {
			flatMonth = m.Format("January 2006")
		}
		if returnMonth == "" && grown >= g.GetGoalRs() {
			returnMonth = m.Format("January 2006")
		}
		if flatMonth != "" && (returnMonth != "" || monthly <= 1) {
			break
		}
	}
	return flatMonth, returnMonth, band
}

// goalOpening is the goal in one sentence: value of goal, the monthly figure and its
// parts, the two months, the return and its span. Arithmetic, never a forecast.
func goalOpening(g *ntxv1.Goal, r sqlc.Rule) string {
	head := fmt.Sprintf("Rs %s of the Rs %s goal", rupees(g.GetValueRs()), rupees(g.GetGoalRs()))
	switch {
	case g.GetValueRs() >= g.GetGoalRs():
		return fmt.Sprintf("Rs %s, past the Rs %s goal.", rupees(g.GetValueRs()), rupees(g.GetGoalRs()))
	case g.GetMonthlyRs() <= 0:
		return head + "; with no monthly figure in the rules there is no date."
	}
	monthly := fmt.Sprintf("At Rs %s a month", rupees(g.GetMonthlyRs()))
	if r.FundSipRs > 0 {
		monthly += fmt.Sprintf(
			" (Rs %s of shares, Rs %s of fund units)",
			rupees(r.MonthlyBudgetRs),
			rupees(r.FundSipRs),
		)
	}
	flat := "the book is more than ten years from it at flat prices"
	if g.GetFlatMonth() != "" {
		flat = fmt.Sprintf("the book reaches it around %s at flat prices", g.GetFlatMonth())
	}
	parts := []string{head + ".", monthly + " " + flat}
	tail := "."
	switch {
	case g.ReturnPercent == nil:
	case g.GetReturnPercent() < 0:
		tail = fmt.Sprintf(
			", or later if the %.1f%% a year it has returned since %s continues.",
			g.GetReturnPercent(),
			g.GetSince(),
		)
	case g.GetReturnMonth() != "":
		tail = fmt.Sprintf(", or around %s if it keeps the %.1f%% a year it has returned since %s.",
			g.GetReturnMonth(), g.GetReturnPercent(), g.GetSince())
	}
	parts[1] += tail
	return strings.Join(append(parts, "Arithmetic on the latest close, not a forecast."), " ")
}
