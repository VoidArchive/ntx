package story

import (
	"fmt"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// Observation is one promoter/public split, as of a date.
type Observation struct {
	AsOf             string
	Promoter, Public *float64 // %
}

// Funds is the web's match of mutual fund holdings for this company.
type Funds struct {
	Count               int
	Total, LargestValue float64 // Rs
	LargestName         string
	AsOf                string // NAV report month
}

func observations(rows []sqlc.OwnershipHistory) []Observation {
	out := make([]Observation, 0, len(rows))
	for _, r := range rows {
		out = append(out, Observation{AsOf: r.AsOf, Promoter: val(r.PromoterPercent), Public: val(r.PublicPercent)})
	}
	return out
}

func ownership(obs []Observation) []Claim {
	if len(obs) == 0 {
		return nil
	}
	cur := obs[len(obs)-1]
	if cur.Promoter == nil || cur.Public == nil {
		return nil
	}
	claim := func(text string, value float64) Claim {
		return Claim{Kind: "ownership", Text: text, Value: value, Unit: "%", Period: cur.AsOf, Anchor: "ownership"}
	}
	out := []Claim{
		claim(fmt.Sprintf("Promoters hold %.2f%% and the public %.2f%%.", *cur.Promoter, *cur.Public), *cur.Promoter),
	}
	// Rows are changes, so the row before the newest is the split it moved from.
	if len(obs) == 1 {
		return append(
			out,
			claim(fmt.Sprintf("No change since NTX first recorded the split on %s.", cur.AsOf), *cur.Promoter),
		)
	}
	prev := obs[len(obs)-2]
	if prev.Promoter == nil {
		return out
	}
	return append(out, claim(
		fmt.Sprintf("The promoter share moved from %.2f%% to %.2f%% on %s.", *prev.Promoter, *cur.Promoter, cur.AsOf),
		*prev.Promoter,
	))
}

// funds says nothing when no fund matched: the NAV file covers open-ended funds only.
func funds(f *Funds) []Claim {
	if f == nil || f.Count == 0 {
		return nil
	}
	text := fmt.Sprintf(
		"%d mutual funds hold Rs %s crore of its shares; the largest is %s with Rs %s crore.",
		f.Count, crore(f.Total), f.LargestName, crore(f.LargestValue),
	)
	if f.Count == 1 {
		text = fmt.Sprintf("One mutual fund, %s, holds Rs %s crore of its shares.", f.LargestName, crore(f.Total))
	}
	return []Claim{{Kind: "funds", Text: text, Value: f.Total, Unit: "Rs", Period: f.AsOf, Anchor: "funds"}}
}
