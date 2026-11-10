package portfolio

import (
	"fmt"
	"strconv"
	"strings"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
)

// Verdicts of a Gate (GLOSSARY: Triggered / Holds / Needs data) and of a cap.
const (
	Triggered = "triggered"
	Holds     = "holds"
	Watch     = "watch"
	NeedsData = "needs data"
)

// point is one figure of a metric's series with its provenance; series run oldest first.
type point struct {
	Value          float64
	Period, Source string
}

type verdict struct {
	Verdict        string
	Value, Prior   *float64
	Period, Source string
}

// metrics is the fixed list a metric gate may name. per is the row cadence the
// periods count runs over; "" is a current figure with one row.
var metrics = map[string]struct{ label, unit, per string }{
	"eps":                   {"EPS", "Rs", "quarter"},
	"pe":                    {"PE", "x", "quarter"},
	"roe":                   {"ROE", "%", "quarter"},
	"book_value":            {"Book value", "Rs", "fiscal year"},
	"cash_dividend_percent": {"Cash dividend", "%", "fiscal year"},
	"bonus_percent":         {"Bonus", "%", "fiscal year"},
	"promoter_percent":      {"Promoter holding", "%", "reading"},
	"weight_percent":        {"Weight", "%", ""},
	"drawdown_percent":      {"Drawdown", "%", ""},
	"npl_percent":           {"NPL", "%", "quarter"},
	"car_percent":           {"CAR", "%", "quarter"},
	"loan_growth_percent":   {"Loan growth", "%", "quarter"},
	"distributable_profit":  {"Distributable profit", "Rs", "quarter"},
}

func directionOf(d ntxv1.GateDirection) string {
	return strings.ToLower(strings.TrimPrefix(d.String(), "GATE_DIRECTION_"))
}

func directionFrom(s string) ntxv1.GateDirection {
	return ntxv1.GateDirection(ntxv1.GateDirection_value["GATE_DIRECTION_"+strings.ToUpper(s)])
}

func gateKindOf(k ntxv1.GateKind) string {
	return strings.ToLower(strings.TrimPrefix(k.String(), "GATE_KIND_"))
}

func gateKindFrom(s string) ntxv1.GateKind {
	return ntxv1.GateKind(ntxv1.GateKind_value["GATE_KIND_"+strings.ToUpper(s)])
}

func isPrior(direction string) bool { return strings.HasSuffix(direction, "_prior") }

// evaluate reads the newest `periods` rows. Triggered needs every one on the wrong
// side of the threshold, or for a prior direction every one failing to improve on
// the row before it; a row without a period has no provenance and cannot be judged.
func evaluate(direction string, threshold float64, periods int, series []point) verdict {
	periods = max(periods, 1)
	need := periods
	if isPrior(direction) {
		need++
	}
	v := verdict{Verdict: NeedsData}
	if len(series) == 0 {
		return v
	}
	last := series[len(series)-1]
	v.Value, v.Period, v.Source = &last.Value, last.Period, last.Source
	if len(series) >= 2 && isPrior(direction) {
		v.Prior = &series[len(series)-2].Value
	}
	if len(series) < need {
		return v
	}
	v.Verdict = Triggered
	for i := len(series) - periods; i < len(series); i++ {
		if series[i].Period == "" {
			return verdict{Verdict: NeedsData}
		}
		wrong := false
		switch direction {
		case directionOf(ntxv1.GateDirection_GATE_DIRECTION_ABOVE):
			wrong = series[i].Value > threshold
		case directionOf(ntxv1.GateDirection_GATE_DIRECTION_BELOW):
			wrong = series[i].Value < threshold
		case directionOf(ntxv1.GateDirection_GATE_DIRECTION_NOT_BELOW_PRIOR):
			wrong = series[i].Value >= series[i-1].Value
		case directionOf(ntxv1.GateDirection_GATE_DIRECTION_NOT_ABOVE_PRIOR):
			wrong = series[i].Value <= series[i-1].Value
		}
		if !wrong {
			v.Verdict = Holds
		}
	}
	return v
}

// label words a metric gate once, for the lede and the table: "NPL above 5% for 2 quarters".
func label(metric, direction string, threshold float64, periods int) string {
	m, ok := metrics[metric]
	if !ok {
		return metric
	}
	var out string
	if isPrior(direction) {
		per := m.per
		if per == "" {
			per = "reading"
		}
		out = fmt.Sprintf(
			"%s %s the prior %s",
			m.label,
			strings.ReplaceAll(strings.TrimSuffix(direction, "_prior"), "_", " "),
			per,
		)
	} else {
		n := strconv.FormatFloat(threshold, 'f', -1, 64)
		if m.unit == "Rs" {
			n = "Rs " + n
		} else {
			n += m.unit
		}
		out = fmt.Sprintf("%s %s %s", m.label, direction, n)
	}
	if periods > 1 && m.per != "" {
		out += fmt.Sprintf(" for %d %ss", periods, m.per)
	}
	return out
}
