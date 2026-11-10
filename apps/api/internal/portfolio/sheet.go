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
)

// lot is NEPSE's minimum purchase; below it an order is an odd lot that rarely fills.
const lot = 10

// filingGrace is the days after a quarter end a company has to publish its results.
const filingGrace = 30

// sheet proposes the month's purchases (GLOSSARY: Order sheet): the budget scaled
// by where the index sits in its own year, split across Build holdings under their
// caps in whole lots at the latest close, the remainder carried. Rows are the index
// closes of the trailing 52 weeks, oldest first, the newest last.
func sheet(month string, r sqlc.Rule, index []sqlc.ListIndexClosesRow, holdings []*ntxv1.Holding) *ntxv1.OrderSheet {
	s := &ntxv1.OrderSheet{Month: month, Multiplier: 1}
	if len(index) > 0 {
		latest := index[len(index)-1]
		s.IndexClose, s.IndexDay = &latest.IndexClose, latest.BusinessDate
		values := make([]float64, 0, len(index))
		for _, row := range index {
			values = append(values, row.IndexClose)
		}
		slices.Sort(values)
		s.RangeLow, s.RangeHigh, s.RangeMedian = values[0], values[len(values)-1], values[len(values)/2]
		switch {
		case latest.IndexClose <= s.RangeLow+(s.RangeHigh-s.RangeLow)/4:
			s.Multiplier = r.MultiplierBottomQuarter
		case latest.IndexClose < s.RangeMedian:
			s.Multiplier = r.MultiplierBelowMedian
		}
	}
	s.BudgetRs = r.MonthlyBudgetRs * s.Multiplier
	s.Opening = sheetOpening(s, r)
	for _, h := range holdings {
		if l := line(h, s.BudgetRs); l != nil {
			s.Lines = append(s.Lines, l)
		}
	}
	return s
}

// line is one holding's row; Hold holdings are not on the sheet.
func line(h *ntxv1.Holding, budget float64) *ntxv1.OrderLine {
	l := &ntxv1.OrderLine{
		Symbol: h.GetSymbol(), Status: h.GetStatus(), AllocationPercent: h.GetAllocationPercent(),
		CarryRs: h.GetCarryRs(), Close: h.Close, ReviewDate: h.GetReviewDate(), NewCarryRs: h.GetCarryRs(),
	}
	switch {
	case h.GetStatus() == ntxv1.HoldingStatus_HOLDING_STATUS_FROZEN:
		l.Note = "frozen"
		if h.GetFrozenReason() != "" {
			l.Note += ", " + h.GetFrozenReason()
		}
		if h.GetReviewDate() != "" {
			l.Note += ", review " + h.GetReviewDate()
		}
		return l
	case h.GetStatus() != ntxv1.HoldingStatus_HOLDING_STATUS_BUILD:
		return nil
	case h.GetCapVerdict() == Watch || h.GetCapVerdict() == Triggered:
		l.Note = "over the soft cap"
		return l
	case h.Close == nil || h.GetClose() <= 0:
		l.Note = "no close recorded"
		return l
	}
	l.AvailableRs = h.GetAllocationPercent()/100*budget + h.GetCarryRs()
	l.Lots = int32(math.Floor(l.AvailableRs / (lot * h.GetClose()))) //nolint:gosec // a handful of lots
	l.Shares = int64(l.Lots) * lot
	l.CostRs = float64(l.Shares) * h.GetClose()
	l.NewCarryRs = l.AvailableRs - l.CostRs
	return l
}

// sheetOpening states the budget, the index, its range and the multiplier (ADR 0002).
func sheetOpening(s *ntxv1.OrderSheet, r sqlc.Rule) string {
	if s.IndexClose == nil {
		return fmt.Sprintf(
			"Order sheet for %s: Rs %s; no index close recorded, so the budget stands.",
			s.GetMonth(),
			rupees(s.GetBudgetRs()),
		)
	}
	where := "upper half"
	switch {
	case s.Multiplier == r.MultiplierBottomQuarter && s.Multiplier != 1:
		where = "bottom quarter"
	case s.Multiplier == r.MultiplierBelowMedian && s.Multiplier != 1:
		where = "lower half"
	}
	rng := fmt.Sprintf("in the %s of its 52-week range of %.2f to %.2f.", where, s.GetRangeLow(), s.GetRangeHigh())
	if s.Multiplier == 1 {
		return fmt.Sprintf("Order sheet for %s: Rs %s, the index at %.2f on %s %s",
			s.GetMonth(), rupees(s.GetBudgetRs()), s.GetIndexClose(), s.GetIndexDay(), rng)
	}
	return fmt.Sprintf(
		"Order sheet for %s: Rs %s, %gx the Rs %s budget with the index at %.2f on %s, %s",
		s.GetMonth(),
		rupees(s.GetBudgetRs()),
		s.Multiplier,
		rupees(r.MonthlyBudgetRs),
		s.GetIndexClose(),
		s.GetIndexDay(),
		rng,
	)
}

// calendar lists what is dated: the next filing due for each held company (quarter
// end plus thirty days), every review date, and each event gate's tick plus a quarter.
func calendar(today time.Time, holdings []*ntxv1.Holding) []*ntxv1.CalendarEntry {
	var out []*ntxv1.CalendarEntry
	for _, h := range holdings {
		if h.GetQuantity() > 0 && !h.GetIsFund() {
			if day, what, ok := nextFiling(today); ok {
				out = append(out, &ntxv1.CalendarEntry{Day: day, Symbol: h.GetSymbol(), What: what})
			}
		}
		if h.GetReviewDate() != "" {
			out = append(out, &ntxv1.CalendarEntry{Day: h.GetReviewDate(), Symbol: h.GetSymbol(), What: "review"})
		}
		for _, g := range h.GetGates() {
			if g.GetKind() != ntxv1.GateKind_GATE_KIND_EVENT || g.GetCheckedOn() == "" {
				continue
			}
			checked, err := time.Parse(time.DateOnly, g.GetCheckedOn())
			if err != nil {
				continue
			}
			out = append(out, &ntxv1.CalendarEntry{
				Day: checked.AddDate(0, 3, 0).
					Format(time.DateOnly),
				Symbol: h.GetSymbol(),
				What:   "re-check: " + g.GetText(),
			})
		}
	}
	slices.SortStableFunc(out, func(a, b *ntxv1.CalendarEntry) int { return cmp.Compare(a.GetDay(), b.GetDay()) })
	return out
}

// nextFiling is the first quarter-end-plus-grace on or after today: the quarter
// just ended while its grace runs, else the quarter under way.
func nextFiling(today time.Time) (string, string, bool) {
	y, err := fiscal.YearOf(today)
	if err != nil {
		return "", "", false
	}
	q, err := fiscal.QuarterOf(today)
	if err != nil {
		return "", "", false
	}
	prevY, prevQ := y, q-1
	if prevQ < fiscal.Q1 {
		prevY, prevQ = y-1, fiscal.Q4
	}
	for _, p := range []struct {
		y fiscal.Year
		q fiscal.Quarter
	}{{prevY, prevQ}, {y, q}} {
		end, err := fiscal.QuarterEnd(p.y, p.q)
		if err != nil {
			continue
		}
		due := end.AddDate(0, 0, filingGrace)
		if !due.Before(today) {
			return due.Format(time.DateOnly), p.q.String() + " " + p.y.String() + " filing due", true
		}
	}
	return "", "", false
}

// month is the Bikram Sambat month name of a day, "Mangsir".
func month(d time.Time) string {
	bs, err := fiscal.BSDate(d)
	if err != nil {
		return d.Format("January")
	}
	return strings.Fields(bs)[0]
}
