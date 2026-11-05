package story

import (
	"database/sql"
	"math"
	"slices"
	"testing"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// Hand-made: the local database has no notices yet (corporate_actions is empty).
var nabilNotices = []sqlc.CorporateAction{
	notice("2025-2026", 10, 12.5, 0, "https://example.test/nabil-2082-83.pdf"),
	notice("2024-2025", 0, 11, 0, ""),
	notice("2023-2024", 0, 9.5, 0, ""),
	notice("2022-2023", 0, 8.5, 0, ""),
}

func notice(fy string, bonus, cash, rights float64, url string) sqlc.CorporateAction {
	return sqlc.CorporateAction{
		CompanyID:       131,
		FiscalYear:      fy,
		BonusPercentage: num(bonus),
		CashDividend:    num(cash),
		RightPercentage: num(rights),
		NoticeUrl:       sql.NullString{String: url, Valid: url != ""},
	}
}

func claimsOf(s Story, kind string) []Claim {
	var out []Claim
	for _, c := range s.Claims {
		if c.Kind == kind {
			out = append(out, c)
		}
	}
	return out
}

func TestNabilPayout(t *testing.T) {
	s := Build(Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows), Actions: nabilNotices})
	got := claimsOf(s, "payout")
	want := []string{
		"For FY 2082/83 the company paid 10% bonus shares and 12.5% cash.",
		"It paid out in each of the last 4 fiscal years.",
		"The cash dividend was 42.78% of FY 2082/83 profit.",
	}
	if !slices.Equal(texts(Story{Claims: got}), want) {
		t.Fatalf("payout claims:\n got %q\nwant %q", texts(Story{Claims: got}), want)
	}
	first := got[0]
	if first.Period != "FY 2082/83" || first.Source != "https://example.test/nabil-2082-83.pdf" ||
		first.Anchor != "payout" || first.Value != 12.5 {
		t.Errorf("payout claim = %+v", first)
	}
}

// A year counts as "no payout" only with a later notice, or its annual report once any notice exists; silence is "not recorded".
func TestNoPayoutVersusNotRecorded(t *testing.T) {
	tests := []struct {
		name    string
		actions []sqlc.CorporateAction
		year    string
		want    string
	}{
		{"notice with a payout", nabilNotices, "2082/83", "paid"},
		{"notice recording nothing", []sqlc.CorporateAction{notice("2025-2026", 0, 0, 0, "")}, "2082/83", "none"},
		{
			"silent year before a later notice",
			[]sqlc.CorporateAction{notice("2025-2026", 0, 5, 0, "")},
			"2080/81",
			"none",
		},
		{
			"silent year with its annual report",
			[]sqlc.CorporateAction{notice("2019-2020", 0, 5, 0, "")},
			"2081/82",
			"none",
		},
		{"annual report but no notices at all", nil, "2081/82", "not recorded"},
		{"silent year with no annual report and no later notice", nil, "2082/83", "not recorded"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := Build(Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows), Actions: tc.actions})
			for _, fy := range s.FiscalYears {
				if fy.Year.String() == tc.year {
					if fy.Payout != tc.want {
						t.Errorf("FY %s payout = %q, want %q", tc.year, fy.Payout, tc.want)
					}
					return
				}
			}
			t.Fatalf("no row for FY %s", tc.year)
		})
	}
}

func TestPayoutNotRecordedWithoutNotices(t *testing.T) {
	s := Build(Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows)})
	if got := claimsOf(s, "payout"); len(got) != 0 {
		t.Errorf("payout claims without notices = %q", texts(Story{Claims: got}))
	}
	if row := s.FiscalYears[0]; row.Year.String() != "2082/83" || row.Payout != "not recorded" || row.Cash != nil {
		t.Errorf("newest row = %+v", row)
	}
}

func TestNoPayoutClaim(t *testing.T) {
	s := Build(Input{
		Company:      nabil,
		Fundamentals: fundamentals(131, nabilRows),
		Actions:      []sqlc.CorporateAction{notice("2025-2026", 0, 0, 0, "")},
	})
	got := texts(Story{Claims: claimsOf(s, "payout")})
	if want := []string{"No payout for FY 2082/83."}; !slices.Equal(got, want) {
		t.Errorf("payout claims = %q, want %q", got, want)
	}
}

func TestBonusAdjustedReturn(t *testing.T) {
	// FY 2082/83 runs 2025-07-17 to 2026-07-16.
	prices := []sqlc.Price{price("2026-07-16", 550), price("2025-07-17", 500)}
	s := Build(Input{
		Company:      nabil,
		Fundamentals: fundamentals(131, nabilRows),
		Prices:       prices,
		Actions:      []sqlc.CorporateAction{notice("2025-2026", 10, 5, 20, "")},
	})
	row := s.FiscalYears[0]
	if row.Year.String() != "2082/83" || row.Return == nil {
		t.Fatalf("newest row = %+v", row)
	}
	// (550 x 1.10 + 5) / 500 - 1; the 20% rights are left out.
	if math.Abs(*row.Return-22) > 1e-9 {
		t.Errorf("return = %v, want 22", *row.Return)
	}
	if row.Rights == nil || *row.Rights != 20 {
		t.Errorf("rights shown = %v", row.Rights)
	}
	for _, fy := range s.FiscalYears[1:] {
		if fy.Return != nil {
			t.Errorf("FY %s has no boundary closes but return %v", fy.Year, *fy.Return)
		}
	}
	price := growthOf(t, s, "price")
	if price.One == nil || math.Abs(*price.One-22) > 1e-9 || price.Three != nil {
		t.Errorf("price growth = %+v", price)
	}
}

// A close far from the boundary would make a part-year return, so it is not used.
func TestReturnNeedsClosesNearBoundaries(t *testing.T) {
	prices := []sqlc.Price{price("2026-07-16", 550), price("2025-11-04", 500)}
	s := Build(Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows), Prices: prices})
	if r := s.FiscalYears[0].Return; r != nil {
		t.Errorf("return from a November start = %v", *r)
	}
}

func growthOf(t *testing.T, s Story, metric string) Growth {
	t.Helper()
	for _, g := range s.Growth {
		if g.Metric == metric {
			return g
		}
	}
	t.Fatalf("no %s growth in %+v", metric, s.Growth)
	return Growth{}
}

func TestNabilFiscalYears(t *testing.T) {
	s := Build(Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows), Actions: nabilNotices})
	var years []string
	for _, fy := range s.FiscalYears {
		years = append(years, fy.Year.String())
	}
	// 2083/84 has only Q1 so far, and there are no 2077/78 rows.
	want := []string{"2082/83", "2081/82", "2080/81", "2079/80", "2078/79", "2076/77"}
	if !slices.Equal(years, want) {
		t.Fatalf("years = %q, want %q", years, want)
	}
	cur := s.FiscalYears[0]
	// No annual report for 2082/83 yet, so the Q4 row stands in.
	if *cur.EPS != 28.36 || *cur.Profit != 7905796000 || *cur.BookValue != 247.28 ||
		math.Abs(*cur.Shares-270569967.29) > 0.01 || *cur.Bonus != 10 || *cur.Cash != 12.5 {
		t.Errorf("2082/83 row = %+v", cur)
	}
	if prev := s.FiscalYears[1]; *prev.EPS != 21.89 {
		t.Errorf("2081/82 prefers the annual report: EPS %v", *prev.EPS)
	}

	eps := growthOf(t, s, "eps")
	if eps.One == nil || math.Abs(*eps.One-29.5569) > 1e-3 ||
		eps.Three == nil || math.Abs(*eps.Three-6.2110) > 1e-3 || eps.Five != nil {
		t.Errorf("eps growth = one %v three %v five %v", eps.One, eps.Three, eps.Five)
	}
}
