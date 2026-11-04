package story

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
)

// Quarter is one quarter's own result, for the earnings chart.
type Quarter struct {
	Period        string // "Q1 2083/84"
	FiscalYear    string // "2083/84"
	Profit        *float64
	EPS           *float64 // quarter-only
	AnnualizedEPS *float64 // as filed
	Label         string   // "Best Q1 in five years", "3rd rise in a row", or ""
	PublishedAt   string
	SourceURL     string
	Monsoon       bool // a hydro Q1, whose annualized EPS overstates the year
}

// HydroEPS states a hydro company's latest EPS three ways (GLOSSARY: monsoon quarter).
type HydroEPS struct {
	Quarter, Annualized, FullYear float64
	FullYearPeriod                string // "FY 2081/82"
}

// EventMark is a dated filing or notice for the price chart's time axis.
type EventMark struct {
	Date  string // "2026-02-10"
	Kind  string // result | notice
	Label string // "Q2", "FY", "10% cash, 5% bonus"
}

// Quarters lists quarters with a quarter-only profit, oldest first, and the
// hydro EPS trio for hydro companies.
func Quarters(in Input) ([]Quarter, *HydroEPS) {
	filings, annual := parse(in.Fundamentals)
	hydro := in.Company.Sector == hydroSector
	var out []Quarter
	for _, f := range filings {
		if f.profit == nil {
			continue
		}
		q := Quarter{
			Period:      f.String(),
			FiscalYear:  f.year.String(),
			Profit:      f.profit,
			EPS:         f.eps,
			Label:       label(f, filings),
			PublishedAt: f.row.PublishedAt.String,
			SourceURL:   f.row.SourceUrl.String,
			Monsoon:     hydro && f.quarter == fiscal.Q1,
		}
		if f.row.Eps.Valid {
			q.AnnualizedEPS = &f.row.Eps.Float64
		}
		out = append(out, q)
	}
	if !hydro || len(filings) == 0 {
		return out, nil
	}
	cur := filings[len(filings)-1]
	fy, fyEPS, ok := lastFullYear(cur.year, filings, annual)
	if cur.eps == nil || !cur.row.Eps.Valid || !ok {
		return out, nil
	}
	return out, &HydroEPS{
		Quarter:        *cur.eps,
		Annualized:     cur.row.Eps.Float64,
		FullYear:       fyEPS,
		FullYearPeriod: period{fy, fiscal.Annual}.String(),
	}
}

// label reuses the record claims for one bar; a record beats a streak.
func label(f filing, filings []filing) string {
	same := sameQuarter(f.quarter, filings)
	noop := func(kind, text string, value float64, unit string) Claim {
		return Claim{Kind: kind, Text: text, Value: value, Unit: unit}
	}
	if c := best(f, same, noop); len(c) > 0 {
		return strings.TrimSuffix(c[0].Text, ".")
	}
	if c := streak(f, same, noop); len(c) > 0 {
		return ordinal(int(c[0].Value)) + " rise in a row"
	}
	return ""
}

func ordinal(n int) string {
	if n%100 >= 11 && n%100 <= 13 {
		return fmt.Sprintf("%dth", n)
	}
	suffix := [...]string{"th", "st", "nd", "rd", "th", "th", "th", "th", "th", "th"}[n%10]
	return fmt.Sprintf("%d%s", n, suffix)
}

// Events marks result publication and notice dates, oldest first. Undated rows are skipped.
func Events(rows []sqlc.Fundamental, actions []sqlc.CorporateAction) []EventMark {
	var out []EventMark
	for _, r := range rows {
		d, ok := day(r.PublishedAt.String)
		if !ok {
			continue
		}
		q, err := fiscal.ParseQuarter(r.Quarter)
		if err != nil {
			continue
		}
		out = append(out, EventMark{Date: d, Kind: "result", Label: q.String()})
	}
	for _, a := range actions {
		d, ok := day(a.SubmittedDate.String)
		if !ok {
			continue
		}
		out = append(out, EventMark{Date: d, Kind: "notice", Label: noticeLabel(a)})
	}
	slices.SortStableFunc(out, func(a, b EventMark) int { return cmp.Compare(a.Date, b.Date) })
	return out
}

func noticeLabel(a sqlc.CorporateAction) string {
	var parts []string
	for _, p := range []struct {
		v    float64
		name string
	}{
		{a.CashDividend.Float64, "cash"},
		{a.BonusPercentage.Float64, "bonus"},
		{a.RightPercentage.Float64, "right"},
	} {
		if p.v > 0 {
			parts = append(parts, fmt.Sprintf("%g%% %s", p.v, p.name))
		}
	}
	if len(parts) == 0 {
		return "Notice"
	}
	return strings.Join(parts, ", ")
}

// day keeps the calendar day of a NEPSE timestamp ("2026-02-10" or "2026-02-10T12:30:00").
func day(s string) (string, bool) {
	if len(s) < len(time.DateOnly) {
		return "", false
	}
	d := s[:len(time.DateOnly)]
	if _, err := time.Parse(time.DateOnly, d); err != nil {
		return "", false
	}
	return d, true
}

// addQuarters fills the response's quarters, hydro EPS and event marks.
func addQuarters(res *ntxv1.GetCompanyStoryResponse, in Input) {
	qs, hydro := Quarters(in)
	for _, q := range qs {
		res.Quarters = append(res.Quarters, &ntxv1.Quarter{
			Period:        q.Period,
			FiscalYear:    q.FiscalYear,
			Profit:        *q.Profit,
			Eps:           q.EPS,
			AnnualizedEps: q.AnnualizedEPS,
			Label:         q.Label,
			PublishedAt:   q.PublishedAt,
			SourceUrl:     q.SourceURL,
			IsMonsoon:     q.Monsoon,
		})
	}
	if hydro != nil {
		res.HydroEps = &ntxv1.HydroEps{
			Quarter:        hydro.Quarter,
			Annualized:     hydro.Annualized,
			FullYear:       hydro.FullYear,
			FullYearPeriod: hydro.FullYearPeriod,
		}
	}
	for _, e := range Events(in.Fundamentals, in.Actions) {
		res.Events = append(res.Events, &ntxv1.EventMark{Date: e.Date, Kind: e.Kind, Label: e.Label})
	}
}
