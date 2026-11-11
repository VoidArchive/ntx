package portfolio

import (
	"testing"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
)

// Fixtures from the owner's note: MNBBL's NPL printed 4.77% at Q4 2082/83 and
// 4.73% at Q1 2083/84; NMB's printed 4.91% at Q4 2082/83.
func TestMetricGatesOnTheOwnersFixtures(t *testing.T) {
	above := directionOf(ntxv1.GateDirection_GATE_DIRECTION_ABOVE)
	ratchet := directionOf(ntxv1.GateDirection_GATE_DIRECTION_NOT_BELOW_PRIOR)
	q4 := point{Value: 4.77, Period: "Q4 2082/83", Source: "https://nepalstock.com/mnbbl-q4"}
	q1 := point{Value: 4.73, Period: "Q1 2083/84", Source: "https://nepalstock.com/mnbbl-q1"}
	for _, tc := range []struct {
		name, direction string
		threshold       float64
		periods         int
		series          []point
		want, period    string
	}{
		{"MNBBL 4.77 against 5 holds", above, 5, 1, []point{q4}, Holds, "Q4 2082/83"},
		{"NMB 4.91 against 4.5 triggers", above, 4.5, 1, []point{{Value: 4.91, Period: "Q4 2082/83", Source: "x"}}, Triggered, "Q4 2082/83"},
		{"MNBBL 4.73 at Q1 against 4.77 holds", above, 4.77, 1, []point{q4, q1}, Holds, "Q1 2083/84"},
		{"ratchet holds at Q1 against Q4", ratchet, 0, 1, []point{q4, q1}, Holds, "Q1 2083/84"},
		{"ratchet triggers at Q2 on 4.73", ratchet, 0, 1, []point{q4, q1, {Value: 4.73, Period: "Q2 2083/84", Source: "x"}}, Triggered, "Q2 2083/84"},
		{"ratchet holds at Q2 on 4.72", ratchet, 0, 1, []point{q4, q1, {Value: 4.72, Period: "Q2 2083/84", Source: "x"}}, Holds, "Q2 2083/84"},
		{"ratchet with no prior needs data", ratchet, 0, 1, []point{q4}, NeedsData, "Q4 2082/83"},
		{
			"ratchet with a missing prior needs data", ratchet, 0, 1,
			[]point{{Period: "Q4 2082/83", Missing: true}, q1},
			NeedsData, "Q1 2083/84",
		},
		{"two periods, one good quarter holds", above, 5, 2, []point{{Value: 4.9, Period: "Q1", Source: "x"}, {Value: 5.2, Period: "Q2", Source: "x"}}, Holds, "Q2"},
		{"two periods, both over triggers", above, 5, 2, []point{{Value: 5.1, Period: "Q1", Source: "x"}, {Value: 5.2, Period: "Q2", Source: "x"}}, Triggered, "Q2"},
		{"fewer rows than periods needs data", above, 5, 2, []point{{Value: 5.1, Period: "Q1", Source: "x"}}, NeedsData, "Q1"},
		{"no rows needs data", above, 5, 1, nil, NeedsData, ""},
		{"a row without a period needs data", above, 5, 1, []point{{Value: 5.1, Period: "", Source: ""}}, NeedsData, ""},
	} {
		got := evaluate(tc.direction, tc.threshold, tc.periods, tc.series)
		if got.Verdict != tc.want || got.Period != tc.period {
			t.Errorf("%s: got %s at %q, want %s at %q", tc.name, got.Verdict, got.Period, tc.want, tc.period)
		}
	}
}

func TestRatchetReportsThePriorFigure(t *testing.T) {
	ratchet := directionOf(ntxv1.GateDirection_GATE_DIRECTION_NOT_BELOW_PRIOR)
	got := evaluate(
		ratchet,
		0,
		1,
		[]point{{Value: 4.77, Period: "Q4 2082/83", Source: "x"}, {Value: 4.73, Period: "Q1 2083/84", Source: "x"}},
	)
	if got.Prior == nil || *got.Prior != 4.77 || got.Value == nil || *got.Value != 4.73 || got.Source != "x" {
		t.Errorf("got %+v", got)
	}
}

func TestGateLabels(t *testing.T) {
	for want, g := range map[string]struct {
		metric    string
		direction ntxv1.GateDirection
		threshold float64
		periods   int
	}{
		"NPL above 5%":                                 {"npl_percent", ntxv1.GateDirection_GATE_DIRECTION_ABOVE, 5, 1},
		"EPS below Rs 15 for 2 quarters":               {"eps", ntxv1.GateDirection_GATE_DIRECTION_BELOW, 15, 2},
		"NPL not below the prior quarter":              {"npl_percent", ntxv1.GateDirection_GATE_DIRECTION_NOT_BELOW_PRIOR, 0, 1},
		"Cash dividend below 5% for 2 fiscal years":    {"cash_dividend_percent", ntxv1.GateDirection_GATE_DIRECTION_BELOW, 5, 2},
		"Drawdown above 15%":                           {"drawdown_percent", ntxv1.GateDirection_GATE_DIRECTION_ABOVE, 15, 1},
		"PE above 30x":                                 {"pe", ntxv1.GateDirection_GATE_DIRECTION_ABOVE, 30, 1},
		"Promoter holding not above the prior reading": {"promoter_percent", ntxv1.GateDirection_GATE_DIRECTION_NOT_ABOVE_PRIOR, 0, 1},
	} {
		if got := label(g.metric, directionOf(g.direction), g.threshold, g.periods); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
