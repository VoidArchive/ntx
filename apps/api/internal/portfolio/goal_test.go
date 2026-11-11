package portfolio

import (
	"database/sql"
	"testing"
	"time"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// The rough check from the brief: Rs 3.74 lakh on 2026-10-16 at Rs 16,000 a month
// reaches Rs 5 lakh around June 2027 at flat prices. One buy of Rs 3.5 lakh a year
// earlier makes the realised return 6.9% a year, which brings the month to May.
func goalBook(buy float64) (time.Time, sqlc.Rule, *ntxv1.GetPortfolioResponse) {
	today := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	r := defaults
	r.GoalRs, r.FundSipRs = 500000, 4000
	p := &ntxv1.GetPortfolioResponse{TotalValue: 374000}
	if buy > 0 {
		p.Holdings = []*ntxv1.Holding{{Symbol: "MEN", Transactions: []*ntxv1.Transaction{
			{Day: "2025-10-16", Kind: ntxv1.TransactionKind_TRANSACTION_KIND_BUY, Quantity: 500, Price: buy / 500},
			{Day: "2026-03-01", Kind: ntxv1.TransactionKind_TRANSACTION_KIND_BONUS, Quantity: 50},
		}}}
	}
	return today, r, p
}

func TestGoalStatesARangeNotAPoint(t *testing.T) {
	today, r, p := goalBook(350000)
	closes := map[string][]sqlc.ListMonthEndClosesRow{"MEN": {
		{Month: "2025-10", ClosePrice: sql.NullFloat64{Float64: 700, Valid: true}},
		{Month: "2026-03", ClosePrice: sql.NullFloat64{Float64: 720, Valid: true}},
	}}
	g := goal(today, r, p, closes)
	want := "Rs 3.74 lakh of the Rs 5.00 lakh goal. At Rs 16,000 a month (Rs 12,000 of shares, Rs 4,000 of fund units) " +
		"the book reaches it around June 2027 at flat prices, or around May 2027 if it keeps the 6.9% a year it has " +
		"returned since 2025-10-16. Arithmetic on the latest close, not a forecast."
	if g.GetOpening() != want {
		t.Errorf("got  %q\nwant %q", g.GetOpening(), want)
	}
	if g.GetContributedRs() != 350000 || g.GetMonthlyRs() != 16000 || g.GetFlatMonth() != "June 2027" ||
		g.GetReturnMonth() != "May 2027" {
		t.Errorf("goal %v", g)
	}
}

// Thirteen month ends from the first buy to today; the bonus lifts the shares priced
// at the carried-forward close, and the last point is today's value. The band runs
// from this month to the flat month with both paths at every point.
func TestGoalSeriesAndBand(t *testing.T) {
	today, r, p := goalBook(350000)
	closes := map[string][]sqlc.ListMonthEndClosesRow{"MEN": {
		{Month: "2025-10", ClosePrice: sql.NullFloat64{Float64: 700, Valid: true}},
		{Month: "2026-03", ClosePrice: sql.NullFloat64{Float64: 720, Valid: true}},
	}}
	g := goal(today, r, p, closes)
	// Thirteen month ends from the first buy to today; the bonus lifts the shares
	// priced at the carried-forward close, and the last point is today's value.
	s := g.GetSeries()
	if len(s) != 13 || s[0].GetMonth() != "2025-10" || s[0].GetValueRs() != 350000 ||
		s[0].GetContributedRs() != 350000 {
		t.Fatalf("series %v", s)
	}
	if s[2].GetValueRs() != 350000 || s[5].GetValueRs() != 550*720 || s[12].GetValueRs() != 374000 ||
		s[12].GetMonth() != "2026-10" {
		t.Errorf("series %v %v %v", s[2], s[5], s[12])
	}
	// The band runs from this month to the flat month, both paths at every point.
	b := g.GetBand()
	if len(b) != 9 || b[0].GetMonth() != "2026-10" || b[0].GetFlatRs() != 374000 ||
		b[8].GetFlatRs() != 374000+8*16000 ||
		b[7].GetWithReturnRs() < 500000 ||
		b[6].GetWithReturnRs() >= 500000 {
		t.Errorf("band %v", b)
	}
}

func TestGoalSentencesForTheEdges(t *testing.T) {
	today, r, p := goalBook(0)
	flatOnly := "Rs 3.74 lakh of the Rs 5.00 lakh goal. At Rs 16,000 a month (Rs 12,000 of shares, Rs 4,000 of fund units) " +
		"the book reaches it around June 2027 at flat prices. Arithmetic on the latest close, not a forecast."
	if got := goal(today, r, p, nil).GetOpening(); got != flatOnly {
		t.Errorf("no transactions: %q", got)
	}

	past := &ntxv1.GetPortfolioResponse{TotalValue: 521000}
	if got, want := goal(today, r, past, nil).GetOpening(), "Rs 5.21 lakh, past the Rs 5.00 lakh goal."; got != want {
		t.Errorf("past: %q", got)
	}

	none := r
	none.MonthlyBudgetRs, none.FundSipRs = 0, 0
	if got, want := goal(
		today,
		none,
		p,
		nil,
	).GetOpening(),
		"Rs 3.74 lakh of the Rs 5.00 lakh goal; with no monthly figure in the rules there is no date."; got != want {
		t.Errorf("no monthly: %q", got)
	}

	slow := r
	slow.MonthlyBudgetRs, slow.FundSipRs = 100, 0
	if got, want := goal(today, slow, p, nil).GetOpening(), "Rs 3.74 lakh of the Rs 5.00 lakh goal. At Rs 100 a month "+
		"the book is more than ten years from it at flat prices. Arithmetic on the latest close, not a forecast."; got != want {
		t.Errorf("slow: %q", got)
	}

	_, _, lost := goalBook(400000)
	if got, want := goal(
		today,
		r,
		lost,
		nil,
	).GetOpening(),
		"Rs 3.74 lakh of the Rs 5.00 lakh goal. At Rs 16,000 a month "+
			"(Rs 12,000 of shares, Rs 4,000 of fund units) the book reaches it around June 2027 at flat prices, or later if the "+
			"-6.5% a year it has returned since 2025-10-16 continues. Arithmetic on the latest close, not a forecast."; got != want {
		t.Errorf("negative: %q", got)
	}
}
