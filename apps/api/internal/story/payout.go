package story

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
)

// payoutStatus never treats a silent feed as a fact: a year without a notice is
// "none" only when a later year has a notice, so the feed covers this company,
// or the year's annual report is filed, so its payout was decided.
// shortcut: "a fundamentals row for that year" (spec) is read as the annual report;
// a quarterly row predates the payout decision.
func payoutStatus(
	y fiscal.Year,
	notices map[fiscal.Year]sqlc.CorporateAction,
	annual map[fiscal.Year]sqlc.Fundamental,
) string {
	if n, ok := notices[y]; ok {
		if n.BonusPercentage.Float64 > 0 || n.CashDividend.Float64 > 0 || n.RightPercentage.Float64 > 0 {
			return "paid"
		}
		return "none"
	}
	if _, ok := annual[y]; ok {
		return "none"
	}
	for later := range notices {
		if later > y {
			return "none"
		}
	}
	return "not recorded"
}

// payout speaks about the newest fiscal-year row: what it paid, the run of years
// with any payout, and cash as a share of that year's profit.
// With no notices at all the feed has not covered the company, so the prose says nothing.
func payout(rows []FiscalYear, actions []sqlc.CorporateAction) []Claim {
	if len(rows) == 0 || len(actions) == 0 {
		return nil
	}
	last := rows[0]
	fy := period{last.Year, fiscal.Annual}.String()
	claim := func(text string, value float64, unit string) Claim {
		return Claim{
			Kind:   "payout",
			Text:   text,
			Value:  value,
			Unit:   unit,
			Period: fy,
			Source: last.Source,
			Anchor: "fiscal-years",
		}
	}
	if last.Payout == "not recorded" {
		return []Claim{claim(fmt.Sprintf("Payout for %s is not recorded.", fy), 0, "")}
	}
	if last.Payout == "none" {
		return []Claim{claim(fmt.Sprintf("No payout for %s.", fy), 0, "%")}
	}

	out := []Claim{claim(fmt.Sprintf("For %s the company paid %s.", fy, paid(last)), deref(last.Cash), "%")}

	run := paidRun(last.Year, noticesByYear(actions))
	if run >= 2 {
		out = append(out, claim(fmt.Sprintf("It paid out in each of the last %d fiscal years.", run), float64(run), ""))
	}

	if last.Cash != nil && *last.Cash > 0 && last.Shares != nil && last.Profit != nil && *last.Profit > 0 {
		share := *last.Cash / 100 * parValue * *last.Shares / *last.Profit * 100
		out = append(
			out,
			claim(fmt.Sprintf("The cash dividend was %.2f%% of %s profit.", share, fy), share, "%"),
		)
	}
	return out
}

// paid lists the year's non-zero parts, e.g. "10% bonus shares and 12.5% cash".
func paid(row FiscalYear) string {
	var parts []string
	for _, p := range []struct {
		v    *float64
		what string
	}{{row.Bonus, "bonus shares"}, {row.Cash, "cash"}, {row.Rights, "rights"}} {
		if p.v != nil && *p.v > 0 {
			parts = append(parts, percent(*p.v)+" "+p.what)
		}
	}
	n := len(parts)
	if n < 2 {
		return strings.Join(parts, "")
	}
	return strings.Join(parts[:n-1], ", ") + " and " + parts[n-1]
}

// paidRun counts consecutive fiscal years with any payout, ending at last.
func paidRun(last fiscal.Year, notices map[fiscal.Year]sqlc.CorporateAction) int {
	run := 0
	for y := last; ; y-- {
		n, ok := notices[y]
		if !ok || n.BonusPercentage.Float64+n.CashDividend.Float64+n.RightPercentage.Float64 <= 0 {
			return run
		}
		run++
	}
}

func percent(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) + "%" }
