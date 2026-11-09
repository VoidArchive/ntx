package sector

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

type quarter struct {
	fy                 string
	eps, bv, profitYTD float64
}

type fixture struct {
	company   sqlc.Company
	published string // "" leaves published_at NULL, so the row is only "seen"
	quarters  []quarter
	closes    map[string]float64
}

func daysAgo(n int) string { return time.Now().UTC().AddDate(0, 0, -n).Format(time.DateOnly) }

func seed(t *testing.T, q *sqlc.Queries, fixtures []fixture) {
	t.Helper()
	ctx := context.Background()
	for _, f := range fixtures {
		c := f.company
		if err := q.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
			ID: c.ID, Name: c.Name, Symbol: c.Symbol, Status: "A", Sector: c.Sector, InstrumentType: "Equity",
		}); err != nil {
			t.Fatal(err)
		}
		for _, r := range f.quarters {
			if err := q.UpsertFundamental(ctx, sqlc.UpsertFundamentalParams{
				CompanyID: c.ID, FiscalYear: r.fy, Quarter: "First Quarter",
				Eps: num(r.eps), BookValue: num(r.bv), PaidUpCapital: num(1e10), ProfitAmount: num(r.profitYTD),
				PublishedAt: sql.NullString{String: f.published, Valid: f.published != ""},
				SourceUrl:   sql.NullString{String: "https://example.test/" + c.Symbol, Valid: true},
			}); err != nil {
				t.Fatal(err)
			}
		}
		for date, close := range f.closes {
			if err := q.UpsertPrice(
				ctx,
				sqlc.UpsertPriceParams{CompanyID: c.ID, BusinessDate: date, ClosePrice: num(close)},
			); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func num(v float64) sql.NullFloat64 { return sql.NullFloat64{Float64: v, Valid: true} }

func near(got *float64, want float64) bool { return got != nil && math.Abs(*got-want) < 1e-6 }

// overview seeds two sectors, three banks (one loss-making) and one hydro, and calls the handler.
func overview(t *testing.T) *ntxv1.GetSectorOverviewResponse {
	t.Helper()
	q := database.OpenTestDB(t)
	bank := func(id int64, symbol string) sqlc.Company {
		return sqlc.Company{ID: id, Name: symbol + " Bank", Symbol: symbol, Sector: "Commercial Banks"}
	}
	seed(t, q, []fixture{
		{bank(1, "AAA"), daysAgo(10), []quarter{
			{"2025-2026", 15, 200, 1e9}, {"2026-2027", 20, 200, 1.2e9},
		}, map[string]float64{"2026-11-02": 380, "2026-11-03": 400}},
		// Filed 40 days ago: in the medians, outside the filed-recently window.
		{bank(2, "BBB"), daysAgo(40), []quarter{
			{"2025-2026", 8, 200, 5e8}, {"2026-2027", 10, 200, 6e8},
		}, map[string]float64{"2026-11-02": 300, "2026-11-03": 300}},
		{bank(3, "LOSS"), "", []quarter{{"2026-2027", -5, 100, -1e8}}, map[string]float64{"2026-11-03": 100}},
		{
			sqlc.Company{ID: 4, Name: "River Hydro", Symbol: "HYD", Sector: "Hydro Power"},
			daysAgo(5),
			[]quarter{{"2026-2027", 4, 100, 4e7}},
			map[string]float64{"2026-11-02": 210, "2026-11-03": 200},
		},
	})

	res, err := NewService(
		q,
	).GetSectorOverview(context.Background(), connect.NewRequest(&ntxv1.GetSectorOverviewRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	return res.Msg
}

func TestSectorMedians(t *testing.T) {
	sectors := overview(t).GetSectors()
	if len(sectors) != 2 {
		t.Fatalf("sectors = %v", sectors)
	}
	banks, hydro := sectors[0], sectors[1]
	if banks.GetName() != "Commercial Banks" || banks.GetCompanyCount() != 3 {
		t.Errorf("largest sector first: %v", banks)
	}
	// PE 20 and 30; the loss-maker's negative PE is left out and the count says so.
	if !near(banks.MedianPe, 25) || banks.GetPeCount() != 2 {
		t.Errorf("median pe = %v over %d", banks.MedianPe, banks.GetPeCount())
	}
	// PB 2, 1.5 and 1: loss-makers count for PB.
	if !near(banks.MedianPb, 1.5) {
		t.Errorf("median pb = %v", banks.MedianPb)
	}
	if !near(banks.ProfitableShare, 2.0/3) || banks.GetQuarterCount() != 3 {
		t.Errorf("profitable share = %v of %d", banks.ProfitableShare, banks.GetQuarterCount())
	}
	// 1.8bn against 1.5bn a year earlier; LOSS has no prior year and drops out.
	if !near(banks.ProfitChangeYoy, 20) {
		t.Errorf("profit change yoy = %v", banks.ProfitChangeYoy)
	}
	// AAA +5.263%, BBB 0%; LOSS has no previous close.
	if !near(banks.DayChange, 50.0/19) {
		t.Errorf("day change = %v", banks.DayChange)
	}
	if hydro.GetSector() != ntxv1.Sector_SECTOR_HYDROPOWER || hydro.ProfitChangeYoy != nil {
		t.Errorf("hydro = %v", hydro)
	}
}

func TestFiledRecentlyWindow(t *testing.T) {
	filed := overview(t).GetFiledRecently()
	want := []struct {
		symbol, date string
		published    bool
	}{
		{"LOSS", time.Now().UTC().Format(time.DateOnly), false},
		{"HYD", daysAgo(5), true},
		{"AAA", daysAgo(10), true},
	}
	if len(filed) != len(want) {
		t.Fatalf("filed recently = %v", filed)
	}
	for i, w := range want {
		f := filed[i]
		if f.GetSymbol() != w.symbol || f.GetDate() != w.date || f.GetPublished() != w.published {
			t.Errorf("filed[%d] = %v, want %v", i, f, w)
		}
	}
	if got := filed[2].GetClaim(); got != "Net profit for Q1 2083/84 was Rs 120.00 crore." {
		t.Errorf("AAA claim = %q", got)
	}
	if got := filed[2].GetSource(); got != "https://example.test/AAA" {
		t.Errorf("AAA source = %q", got)
	}
}

func TestMovers(t *testing.T) {
	res := overview(t)
	// BBB is unchanged and LOSS has no previous close, so neither is a mover.
	g, l := res.GetGainers(), res.GetLosers()
	if len(g) != 1 || g[0].GetSymbol() != "AAA" || g[0].GetClose() != 400 {
		t.Errorf("gainers = %v", g)
	}
	if len(l) != 1 || l[0].GetSymbol() != "HYD" || math.Abs(l[0].GetChangePercent()+100.0/21) > 1e-9 {
		t.Errorf("losers = %v", l)
	}
}
