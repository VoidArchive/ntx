package story

import (
	"context"
	"errors"
	"slices"
	"testing"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func TestGetCompanyStory(t *testing.T) {
	ctx := context.Background()
	q := database.OpenTestDB(t)

	seed(t, q)

	res, err := NewService(q).GetCompanyStory(ctx, connect.NewRequest(&ntxv1.GetCompanyStoryRequest{Symbol: "NABIL"}))
	if err != nil {
		t.Fatal(err)
	}
	m := res.Msg
	if got := m.GetClaims()[0].GetText(); got != "Net profit for Q1 2083/84 was Rs 255.51 crore." {
		t.Errorf("first claim = %q", got)
	}
	if got := m.GetClaims()[4].GetText(); got != "PE is 16.29x at Rs 552.00, against a sector median of 16.29x (1 profitable)." {
		t.Errorf("pe claim = %q", got)
	}
	pb := m.GetStrips()[1]
	if pb.GetRatio() != "pb" || pb.GetSector().GetCount() != 2 {
		t.Errorf("pb strip counts both banks: %v", pb)
	}
	if m.GetStrips()[3].Sector != nil {
		t.Errorf("yield strip without dividends should have no sector range: %v", m.GetStrips()[3])
	}
	if qt := m.GetQuote(); qt.GetChange() != 1 || qt.GetBusinessDateBs() != "Kartik 17, 2083" {
		t.Errorf("quote = %v", qt)
	}
	if qs := m.GetQuarters(); len(qs) != 17 || qs[16].GetLabel() != "Best Q1 in five years" || m.GetHydroEps() != nil {
		t.Errorf("quarters = %d, latest %v, hydro %v", len(qs), qs[len(qs)-1], m.GetHydroEps())
	}
	if m.GetLastFiling() != "Q1 2083/84" {
		t.Errorf("last filing = %q", m.GetLastFiling())
	}

	if m.GetFiscalYears()[0].GetPayout() != "not recorded" {
		t.Errorf("no notices stored: %v", m.GetFiscalYears()[0])
	}
}

func TestGetCompanyStoryUnknownSymbol(t *testing.T) {
	q := database.OpenTestDB(t)
	seed(t, q)
	_, err := NewService(
		q,
	).GetCompanyStory(context.Background(), connect.NewRequest(&ntxv1.GetCompanyStoryRequest{Symbol: "NOPE"}))
	if ce := new(connect.Error); !errors.As(err, &ce) || ce.Code() != connect.CodeNotFound {
		t.Errorf("unknown symbol: %v", err)
	}
}

func TestGetCompanyStoryPayoutOwnershipFunds(t *testing.T) {
	ctx := context.Background()
	q := database.OpenTestDB(t)
	seed(t, q)
	n := nabilNotices[0]
	if err := q.UpsertCorporateAction(ctx, sqlc.UpsertCorporateActionParams{
		CompanyID: n.CompanyID, FiscalYear: n.FiscalYear, BonusPercentage: n.BonusPercentage,
		RightPercentage: n.RightPercentage, CashDividend: n.CashDividend, NoticeUrl: n.NoticeUrl,
	}); err != nil {
		t.Fatal(err)
	}
	o := observation("2026-11-01", 51, 49)
	if err := q.UpsertOwnershipHistory(ctx, sqlc.UpsertOwnershipHistoryParams{
		CompanyID: o.CompanyID, AsOf: o.AsOf, PromoterPercent: o.PromoterPercent, PublicPercent: o.PublicPercent,
	}); err != nil {
		t.Fatal(err)
	}

	res, err := NewService(q).GetCompanyStory(ctx, connect.NewRequest(&ntxv1.GetCompanyStoryRequest{
		Symbol: "NABIL",
		Funds: &ntxv1.FundHoldings{
			Count:        1,
			TotalValue:   319592955,
			LargestName:  "NIBL Sahabhagita Fund",
			LargestValue: 319592955,
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	m := res.Msg
	var kinds []string
	for _, c := range m.GetClaims() {
		kinds = append(kinds, c.GetKind())
	}
	if !slices.Contains(kinds, "payout") || !slices.Contains(kinds, "ownership") || !slices.Contains(kinds, "funds") {
		t.Errorf("claim kinds = %q", kinds)
	}
	if fy := m.GetFiscalYears()[0]; fy.GetYear() != "2082/83" || fy.GetPayout() != "paid" || fy.GetCash() != 12.5 {
		t.Errorf("newest fiscal year = %v", fy)
	}
	if len(m.GetGrowth()) != 4 || len(m.GetOwnership()) != 1 {
		t.Errorf("growth %v, ownership %v", m.GetGrowth(), m.GetOwnership())
	}
}

// seed stores NABIL, API and a loss-making bank in NABIL's sector.
func seed(t *testing.T, q *sqlc.Queries) {
	t.Helper()
	ctx := context.Background()
	loser := sqlc.Company{ID: 1, Name: "Loss Bank", Symbol: "LOSS", Sector: "Commercial Banks"}
	for _, c := range []sqlc.Company{nabil, loser, api} {
		if err := q.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
			ID: c.ID, Name: c.Name, Symbol: c.Symbol, Status: "A", Sector: c.Sector, InstrumentType: "Equity",
		}); err != nil {
			t.Fatal(err)
		}
	}
	seed := map[int64][]fixtureRow{
		nabil.ID: nabilRows,
		loser.ID: {{"2026-2027", "First Quarter", -2, -100, 150, 1e10, -5e7}},
		api.ID:   apiRows,
	}
	for id, rows := range seed {
		for _, f := range fundamentals(id, rows) {
			if err := q.UpsertFundamental(ctx, sqlc.UpsertFundamentalParams{
				CompanyID: f.CompanyID, FiscalYear: f.FiscalYear, Quarter: f.Quarter, Eps: f.Eps, PeRatio: f.PeRatio,
				BookValue: f.BookValue, PaidUpCapital: f.PaidUpCapital, ProfitAmount: f.ProfitAmount,
			}); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, p := range []sqlc.UpsertPriceParams{
		{CompanyID: nabil.ID, BusinessDate: "2026-11-02", ClosePrice: num(551)},
		{CompanyID: nabil.ID, BusinessDate: "2026-11-03", ClosePrice: num(552)},
		{CompanyID: loser.ID, BusinessDate: "2026-11-03", ClosePrice: num(200)},
		{CompanyID: api.ID, BusinessDate: "2026-11-03", ClosePrice: num(340)},
	} {
		if err := q.UpsertPrice(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
}
