package worker

import (
	"math"
	"slices"

	"github.com/voidarchive/ntx/internal/fiscal"
	"github.com/voidarchive/ntx/internal/nepse"
)

// parValue is the face value of a NEPSE share in rupees.
const parValue = 100

// normalizeFundamentals keeps one filing per period and puts profit and
// paid-up capital in rupees. NEPSE sends no unit: companies type their figures
// in rupees, thousands, lakh or crore, and the feed lists re-filings newest
// first. Paid-up is anchored on the company's median rupee-looking paid-up;
// profit on annualized EPS x shares x quarter/4. A decade shift is a unit
// error, a 2-3x gap is NEPSE's loose annualization, and the gap between is
// unresolvable, so that profit is dropped rather than guessed.
func normalizeFundamentals(rows []nepse.Fundamental) []nepse.Fundamental {
	type key struct{ fy, q string }
	seen := map[key]bool{}
	var out []nepse.Fundamental
	var caps []float64
	for _, r := range rows {
		k := key{r.FiscalYear, r.Quarter}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, r)
		if r.PaidUpCapital >= 1e8 {
			caps = append(caps, r.PaidUpCapital)
		}
	}
	if len(caps) == 0 {
		return out
	}
	slices.Sort(caps)
	anchor := caps[len(caps)/2]

	for i := range out {
		r := &out[i]
		if r.EPS == 0 || r.PaidUpCapital <= 0 || r.ProfitAmount == 0 {
			continue
		}
		r.PaidUpCapital = rescale(r.PaidUpCapital, anchor, 10)
		q, err := fiscal.ParseQuarter(r.Quarter)
		if err != nil {
			continue
		}
		if q == fiscal.Annual {
			q = fiscal.Q4
		}
		expect := r.EPS * (r.PaidUpCapital / parValue) * float64(q) / 4
		if (r.ProfitAmount < 0) != (expect < 0) {
			continue
		}
		off := ratio(math.Abs(r.ProfitAmount), math.Abs(expect))
		switch {
		case off <= 10:
		case off < math.Sqrt(1000):
			r.ProfitAmount = 0
		default:
			r.ProfitAmount = rescale(r.ProfitAmount, expect, 1)
		}
	}
	return out
}

// rescale multiplies v by the power of ten that lands it nearest ref, when it
// is more than tolerance away.
func rescale(v, ref, tolerance float64) float64 {
	if ratio(math.Abs(v), math.Abs(ref)) <= tolerance {
		return v
	}
	return v * math.Pow(10, math.Round(math.Log10(math.Abs(ref)/math.Abs(v))))
}

func ratio(a, b float64) float64 { return math.Max(a/b, b/a) }
