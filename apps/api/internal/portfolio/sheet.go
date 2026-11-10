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

// minCloses is the fewest closes that can place the index in a year: a lone close
// sits at both ends of its own range.
const minCloses = 4

// filingGrace is the days after a quarter end a company has to publish its results.
const filingGrace = 30

// sheet proposes the month's purchases (GLOSSARY: Order sheet): the budget scaled
// by where the index sits in its own year, split across Build holdings under their
// caps in whole lots at the latest close, the remainder carried. Rows are the index
// closes of the trailing 52 weeks, oldest first, the newest last.
func sheet(month string, r sqlc.Rule, index []sqlc.ListIndexClosesRow, holdings []*ntxv1.Holding) *ntxv1.OrderSheet {
	s := &ntxv1.OrderSheet{Month: month, Multiplier: 1}
	band := ""
	if len(index) > 0 {
		latest := index[len(index)-1]
		s.IndexClose, s.IndexDay = &latest.IndexClose, latest.BusinessDate
		values := make([]float64, 0, len(index))
		for _, row := range index {
			values = append(values, row.IndexClose)
		}
		slices.Sort(values)
		s.RangeLow, s.RangeHigh, s.RangeMedian = values[0], values[len(values)-1], median(values)
		band, s.Multiplier = place(latest.IndexClose, values, s.RangeMedian, r)
	}
	s.BudgetRs = r.MonthlyBudgetRs * s.Multiplier
	s.Opening = sheetOpening(s, r, band, len(index))
	for _, h := range holdings {
		if l := line(h, s.BudgetRs); l != nil {
			s.Lines = append(s.Lines, l)
		}
	}
	return s
}

// place reads the latest close against the year's quartiles: the bottom quarter and
// below the median scale the budget up; too few closes leave it alone.
func place(latest float64, sorted []float64, med float64, r sqlc.Rule) (string, float64) {
	if len(sorted) < minCloses {
		return "", 1
	}
	switch {
	case latest <= sorted[len(sorted)/4]:
		return "in the bottom quarter", r.MultiplierBottomQuarter
	case latest < med:
		return "below the median", r.MultiplierBelowMedian
	}
	return "above the median", 1
}

func median(sorted []float64) float64 {
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// line is one holding's row. Hold holdings are not on the sheet, and nor are funds:
// a fund SIP is units at NAV, with no lot to fill.
func line(h *ntxv1.Holding, budget float64) *ntxv1.OrderLine {
	if h.GetIsFund() {
		return nil
	}
	l := &ntxv1.OrderLine{
		Symbol: h.GetSymbol(), Status: h.GetStatus(), AllocationPercent: h.GetAllocationPercent(),
		CarryRs: h.GetCarryRs(), Close: h.Close, CloseDay: h.GetCloseDay(), ReviewDate: h.GetReviewDate(),
		NewCarryRs: h.GetCarryRs(),
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
func sheetOpening(s *ntxv1.OrderSheet, r sqlc.Rule, band string, closes int) string {
	head := fmt.Sprintf("Order sheet for %s: Rs %s", s.GetMonth(), rupees(s.GetBudgetRs()))
	if s.IndexClose == nil {
		return head + "; no index close recorded, so the budget stands."
	}
	if band == "" {
		return fmt.Sprintf(
			"%s; only %s in the last 52 weeks, so the budget stands.",
			head,
			plural(closes, "index close"),
		)
	}
	rng := fmt.Sprintf("%s of its 52-week closes, %.2f to %.2f.", band, s.GetRangeLow(), s.GetRangeHigh())
	if s.Multiplier == 1 {
		return fmt.Sprintf("%s, the index at %.2f on %s %s", head, s.GetIndexClose(), s.GetIndexDay(), rng)
	}
	return fmt.Sprintf("%s, %gx the Rs %s budget with the index at %.2f on %s, %s",
		head, s.Multiplier, rupees(r.MonthlyBudgetRs), s.GetIndexClose(), s.GetIndexDay(), rng)
}

// calendar lists what is dated: the next filing due for each held company (quarter
// end plus thirty days), every review date, and each event gate's tick plus a quarter.
// Dates compare by calendar day, so a due date counts until it is over.
func calendar(today time.Time, holdings []*ntxv1.Holding) []*ntxv1.CalendarEntry {
	y, m, d := today.Date()
	today = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	filingDay, filing, hasFiling := nextFiling(today)
	var out []*ntxv1.CalendarEntry
	for _, h := range holdings {
		if hasFiling && h.GetQuantity() > 0 && !h.GetIsFund() {
			out = append(out, &ntxv1.CalendarEntry{Day: filingDay, Symbol: h.GetSymbol(), What: filing})
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
	day := today.Format(time.DateOnly)
	for _, e := range out {
		e.Overdue = e.GetDay() < day
	}
	slices.SortStableFunc(out, func(a, b *ntxv1.CalendarEntry) int { return cmp.Compare(a.GetDay(), b.GetDay()) })
	return out
}

// calendarOpening names the next date and counts the ones that have passed.
func calendarOpening(entries []*ntxv1.CalendarEntry) string {
	passed := 0
	var next *ntxv1.CalendarEntry
	for _, e := range entries {
		if e.GetOverdue() {
			passed++
		} else if next == nil {
			next = e
		}
	}
	parts := []string{}
	if next != nil {
		parts = append(parts, fmt.Sprintf("Next: %s, %s %s.", next.GetDay(), next.GetSymbol(), next.GetWhat()))
	}
	if passed > 0 {
		parts = append(parts, fmt.Sprintf("%s passed without a note.", plural(passed, "date")))
	}
	if len(parts) == 0 {
		return "Nothing dated."
	}
	return strings.Join(parts, " ")
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

func month(d time.Time) string {
	bs, err := fiscal.BSDate(d)
	if err != nil {
		return d.Format("January")
	}
	return strings.Fields(bs)[0]
}
