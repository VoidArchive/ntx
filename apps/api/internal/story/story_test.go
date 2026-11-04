package story

import (
	"math"
	"slices"
	"testing"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func texts(s Story) []string {
	out := make([]string, 0, len(s.Claims))
	for _, c := range s.Claims {
		out = append(out, c.Text)
	}
	return out
}

func claimWith(t *testing.T, s Story, prefix string) Claim {
	t.Helper()
	for _, c := range s.Claims {
		if len(c.Text) >= len(prefix) && c.Text[:len(prefix)] == prefix {
			return c
		}
	}
	t.Fatalf("no claim starting %q in %q", prefix, texts(s))
	return Claim{}
}

func TestNabilStory(t *testing.T) {
	s := Build(Input{
		Company:      nabil,
		Fundamentals: fundamentals(131, nabilRows),
		Prices:       []sqlc.Price{price("2026-11-03", 552.0), price("2026-11-02", 551.0)},
		Sector:       bankSector,
	})
	want := []string{
		"Net profit for Q1 2083/84 was Rs 255.51 crore.",
		"Up 45.4% on Q1 2082/83.",
		"Best Q1 in five years.",
		"EPS was Rs 9.44 for the quarter and Rs 33.89 annualized.",
		"PE is 16.29x at Rs 552.00, against a sector median of 18.52x (19 profitable).",
		// 19.705 in binary floating point sits just below the half.
		"Its three-year median PE is 19.70x.",
		"PB is 2.36x, against a sector median of 1.56x.",
		"Its three-year median PB is 2.21x.",
	}
	if got := texts(s); !slices.Equal(got, want) {
		t.Fatalf("claims:\n got %q\nwant %q", got, want)
	}
	first := s.Claims[0]
	if first.Kind != "earnings" || first.Value != 2555105000 || first.Unit != "Rs" || first.Period != "Q1 2083/84" ||
		first.Anchor == "" {
		t.Errorf("profit claim = %+v", first)
	}
	if s.LastFiling != "Q1 2083/84" {
		t.Errorf("last filing = %q", s.LastFiling)
	}
	if s.Quote == nil || s.Quote.Close != 552 || s.Quote.Change == nil || *s.Quote.Change != 1 ||
		s.Quote.DateBS != "Kartik 17, 2083" {
		t.Errorf("quote = %+v", s.Quote)
	}
}

// The exchange publishes year-to-date profit; each quarter's own profit is the difference.
func TestQuarterOnlyFromYearToDate(t *testing.T) {
	tests := []struct {
		quarter     string
		profit      float64
		epsSentence string
	}{
		{"First Quarter", 1756991000, "EPS was Rs 6.49 for the quarter and Rs 25.49 annualized."},
		{"Second Quarter", 3002650000, "EPS was Rs 11.10 for the quarter and Rs 35.04 annualized."},
		{"Third Quarter", 2005005000, "EPS was Rs 7.41 for the quarter and Rs 33.02 annualized."},
		{"Fourth Quarter", 1141150000, "EPS was Rs 4.22 for the quarter and Rs 28.36 annualized."},
	}
	for _, tc := range tests {
		t.Run(tc.quarter, func(t *testing.T) {
			s := Build(Input{Company: nabil, Fundamentals: fundamentals(131, asOf(nabilRows, "2025-2026", tc.quarter))})
			if got := claimWith(t, s, "Net profit").Value; math.Abs(got-tc.profit) > 0.5 {
				t.Errorf("quarter-only profit = %.0f, want %.0f", got, tc.profit)
			}
			if got := claimWith(t, s, "EPS was").Text; got != tc.epsSentence {
				t.Errorf("eps = %q, want %q", got, tc.epsSentence)
			}
		})
	}
}

func TestHydroStatesEPSThreeWays(t *testing.T) {
	tests := []struct {
		name    string
		rows    []fixtureRow
		eps     string
		monsoon bool
	}{
		{
			"monsoon quarter is flagged",
			asOf(apiRows, "2025-2026", "First Quarter"),
			"EPS was Rs 4.32 for the quarter, Rs 17.30 annualized and Rs 7.04 for FY 2081/82.",
			true,
		},
		{
			"year end needs no flag",
			apiRows,
			"EPS was Rs 4.73 for the quarter, Rs 15.06 annualized and Rs 15.06 for FY 2082/83.",
			false,
		},
	}
	const flag = "Q1 is the monsoon quarter, so the annualized EPS overstates the year."
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := Build(Input{Company: api, Fundamentals: fundamentals(697, tc.rows)})
			if got := claimWith(t, s, "EPS was").Text; got != tc.eps {
				t.Errorf("eps = %q, want %q", got, tc.eps)
			}
			if got := slices.Contains(texts(s), flag); got != tc.monsoon {
				t.Errorf("monsoon flag = %v, want %v in %q", got, tc.monsoon, texts(s))
			}
		})
	}
}

func TestLossIsStatedAsLoss(t *testing.T) {
	s := Build(Input{Company: api, Fundamentals: fundamentals(697, asOf(apiRows, "2023-2024", "First Quarter"))})
	if got := s.Claims[0].Text; got != "Net loss for Q1 2080/81 was Rs 0.40 crore." {
		t.Errorf("first claim = %q", got)
	}
}

func TestSectorPEExcludesLossMakers(t *testing.T) {
	s := Build(Input{
		Company:      nabil,
		Fundamentals: fundamentals(131, nabilRows),
		Prices:       []sqlc.Price{price("2026-11-03", 552.0)},
		Sector:       bankSector,
	})
	pe, pb := s.Strips[0], s.Strips[1]
	if pe.Ratio != "pe" || pe.Sector.Count != 19 || math.Abs(pe.Sector.Median-18.5185) > 1e-4 || pe.Sector.Low <= 0 {
		t.Errorf("pe strip = %+v", pe)
	}
	if pb.Ratio != "pb" || pb.Sector.Count != 19 {
		t.Errorf("pb strip includes everyone: %+v", pb)
	}
	if pe.Value == nil || math.Abs(*pe.Value-16.288) > 1e-3 || pe.Own.Count != 12 {
		t.Errorf("pe company/own = %v %+v", pe.Value, pe.Own)
	}
}

func TestTooFewFilingsSaysNothing(t *testing.T) {
	s := Build(Input{Company: nabil, Fundamentals: fundamentals(131, nabilRows[:1])})
	if len(s.Claims) != 0 || s.LastFiling != "FY 2076/77" {
		t.Errorf("claims %q, last filing %q", texts(s), s.LastFiling)
	}
}

// No fixture company has a long streak or a loss after a profitable run, so these rows are made up.
func TestStreakAndFirstLoss(t *testing.T) {
	const paid = 1e9
	tests := []struct {
		name string
		rows []fixtureRow
		want string
	}{
		{
			"streak",
			[]fixtureRow{
				{"2021-2022", "First Quarter", 1, 10, 100, paid, 1e7},
				{"2022-2023", "First Quarter", 1, 10, 100, paid, 2e7},
				{"2023-2024", "First Quarter", 1, 10, 100, paid, 3e7},
				{"2024-2025", "First Quarter", 1, 10, 100, paid, 4e7},
			},
			"Third consecutive year-on-year increase in Q1.",
		},
		{
			"first loss",
			[]fixtureRow{
				{"2024-2025", "First Quarter", 1, 10, 100, paid, 1e7},
				{"2024-2025", "Second Quarter", 1, 10, 100, paid, 2e7},
				{"2024-2025", "Third Quarter", 1, 10, 100, paid, 3e7},
				{"2024-2025", "Fourth Quarter", 1, 10, 100, paid, 4e7},
				{"2025-2026", "First Quarter", -1, -10, 100, paid, -1e7},
			},
			"First loss in 4 quarters.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := Build(Input{Company: nabil, Fundamentals: fundamentals(131, tc.rows)})
			if !slices.Contains(texts(s), tc.want) {
				t.Errorf("want %q in %q", tc.want, texts(s))
			}
		})
	}
}
