package portfolio

import (
	"context"
	"database/sql"
	"testing"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func ptr(v float64) *float64 { return &v }

func bankRow(t *testing.T, c ntxv1connect.PortfolioServiceClient, m *ntxv1.BankMetric) error {
	t.Helper()
	_, err := c.UpsertBankMetric(context.Background(), connect.NewRequest(&ntxv1.UpsertBankMetricRequest{Metric: m}))
	return err
}

func seedFiling(t *testing.T, q *sqlc.Queries, fy, quarter string) {
	t.Helper()
	must(t, q.UpsertFundamental(context.Background(), sqlc.UpsertFundamentalParams{
		CompanyID: 1, FiscalYear: fy, Quarter: quarter,
		Eps:       sql.NullFloat64{Float64: 20, Valid: true},
		SourceUrl: sql.NullString{String: "https://nepalstock.com/men/" + quarter, Valid: true},
	}))
}

// A bank row without its period, filing date or source is refused (ADR 0002).
func TestBankRowsNeedProvenance(t *testing.T) {
	c, _ := clientAndDB(t, testToken)
	book(t, c)
	full := func() *ntxv1.BankMetric {
		return &ntxv1.BankMetric{
			Symbol: "MEN", FiscalYear: "2082/83", Quarter: "Q4", NplPercent: ptr(3.1),
			PublishedDate: "2026-08-20", SourceUrl: "https://men.com.np/q4.pdf",
		}
	}
	for name, edit := range map[string]func(*ntxv1.BankMetric){
		"no source":       func(m *ntxv1.BankMetric) { m.SourceUrl = "" },
		"not a url":       func(m *ntxv1.BankMetric) { m.SourceUrl = "q4.pdf" },
		"no published":    func(m *ntxv1.BankMetric) { m.PublishedDate = "" },
		"no fiscal year":  func(m *ntxv1.BankMetric) { m.FiscalYear = "" },
		"annual":          func(m *ntxv1.BankMetric) { m.Quarter = "FY" },
		"no quarter":      func(m *ntxv1.BankMetric) { m.Quarter = "" },
		"no figure":       func(m *ntxv1.BankMetric) { m.NplPercent = nil },
		"unknown company": func(m *ntxv1.BankMetric) { m.Symbol = "NOPE" },
		"a fund":          func(m *ntxv1.BankMetric) { m.Symbol = "NIBLSF" },
	} {
		m := full()
		edit(m)
		err := bankRow(t, c, m)
		if connect.CodeOf(err) != connect.CodeInvalidArgument && connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("%s: got %v", name, err)
		}
	}
	must(t, bankRow(t, c, full()))
	res, err := c.GetHoldingStory(
		context.Background(),
		connect.NewRequest(&ntxv1.GetHoldingStoryRequest{Symbol: "MEN"}),
	)
	must(t, err)
	rows := res.Msg.GetBankMetrics()
	if len(rows) != 1 || rows[0].GetQuarter() != "Q4" || rows[0].GetFiscalYear() != "2082/83" ||
		rows[0].GetNplPercent() != 3.1 || rows[0].CarPercent != nil {
		t.Fatalf("bank rows %v", rows)
	}
}

// The gate holds on the Q4 row; when Q1 fundamentals land without the owner's
// figures it turns to Needs data for that quarter, and holds again once entered.
func TestBankGateFollowsTheQuarters(t *testing.T) {
	c, q := clientAndDB(t, testToken)
	book(t, c)
	seedFiling(t, q, "2025-2026", "Fourth Quarter")
	must(t, bankRow(t, c, &ntxv1.BankMetric{
		Symbol: "MEN", FiscalYear: "2082/83", Quarter: "Q4", NplPercent: ptr(3.1),
		PublishedDate: "2026-08-20", SourceUrl: "https://men.com.np/q4.pdf",
	}))
	gate(t, c, &ntxv1.UpsertGateRequest{
		Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_METRIC, Metric: "npl_percent",
		Direction: ntxv1.GateDirection_GATE_DIRECTION_ABOVE, Threshold: 5,
	})
	gate(t, c, &ntxv1.UpsertGateRequest{
		Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_METRIC, Metric: "npl_percent",
		Direction: ntxv1.GateDirection_GATE_DIRECTION_NOT_BELOW_PRIOR,
	})
	men := func() *ntxv1.Holding {
		res, err := c.GetHoldingStory(
			context.Background(),
			connect.NewRequest(&ntxv1.GetHoldingStoryRequest{Symbol: "MEN"}),
		)
		must(t, err)
		return res.Msg.GetHolding()
	}
	assertGates(t, men().GetGates(), []string{
		"holds Q4 2082/83 NPL above 5%", "needs data Q4 2082/83 NPL not below the prior quarter",
	})

	seedFiling(t, q, "2026-2027", "First Quarter")
	g := men().GetGates()
	assertGates(t, g, []string{
		"needs data Q1 2083/84 NPL above 5%", "needs data Q1 2083/84 NPL not below the prior quarter",
	})
	if g[0].Value != nil || g[1].GetPrior() != 3.1 {
		t.Errorf("missing quarter carries value %v, prior %v", g[0].Value, g[1].Prior)
	}

	must(t, bankRow(t, c, &ntxv1.BankMetric{
		Symbol: "MEN", FiscalYear: "2083/84", Quarter: "Q1", NplPercent: ptr(2.9),
		PublishedDate: "2026-11-05", SourceUrl: "https://men.com.np/q1.pdf",
	}))
	assertGates(t, men().GetGates(), []string{
		"holds Q1 2083/84 NPL above 5%", "holds Q1 2083/84 NPL not below the prior quarter",
	})
}
