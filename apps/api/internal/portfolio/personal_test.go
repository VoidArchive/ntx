package portfolio

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
)

func export(t *testing.T, c ntxv1connect.PortfolioServiceClient) []byte {
	t.Helper()
	res, err := c.ExportPersonal(context.Background(), connect.NewRequest(&ntxv1.ExportPersonalRequest{}))
	must(t, err)
	return res.Msg.GetJson()
}

func importFile(c ntxv1connect.PortfolioServiceClient, file []byte) error {
	_, err := c.ImportPersonal(context.Background(), connect.NewRequest(&ntxv1.ImportPersonalRequest{Json: file}))
	return err
}

func readBook(t *testing.T, c ntxv1connect.PortfolioServiceClient) *ntxv1.GetPortfolioResponse {
	t.Helper()
	res, err := c.GetPortfolio(context.Background(), connect.NewRequest(&ntxv1.GetPortfolioRequest{}))
	must(t, err)
	return res.Msg
}

// Every personal table filled, exported, wiped and imported reads back row for row.
func TestExportThenImportRoundTrips(t *testing.T) {
	c, _ := clientAndDB(t, testToken)
	book(t, c)
	ctx := context.Background()
	gate(t, c, &ntxv1.UpsertGateRequest{
		Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_METRIC, Metric: "npl_percent",
		Direction: ntxv1.GateDirection_GATE_DIRECTION_ABOVE, Threshold: 5, Periods: 2,
	})
	gate(t, c, &ntxv1.UpsertGateRequest{
		Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_EVENT, Text: "Plant damage reported", CheckedOn: "2026-10-16",
	})
	must(t, bankRow(t, c, &ntxv1.BankMetric{
		Symbol: "MEN", FiscalYear: "2082/83", Quarter: "Q4", NplPercent: ptr(3.1),
		PublishedDate: "2026-08-20", SourceUrl: "https://men.com.np/q4.pdf",
	}))
	_, err := c.UpdateRules(ctx, connect.NewRequest(&ntxv1.UpdateRulesRequest{Rules: &ntxv1.Rules{
		SoftCapPercent: 22, HardCapPercent: 27, SectorCapPercent: 45, MaxPositions: 9, MinEntryRs: 20000,
		CandidateMinRating: "C", MonthlyBudgetRs: 15000, MultiplierBelowMedian: 1.25, MultiplierBottomQuarter: 1.5,
	}}))
	must(t, err)
	before, file := readBook(t, c), export(t, c)
	for _, table := range []string{`"holdings"`, `"transactions"`, `"holding_notes"`, `"gates"`, `"rules"`, `"bank_metrics"`} {
		if !bytes.Contains(file, []byte(table)) {
			t.Errorf("export lacks %s", table)
		}
	}
	if bytes.Contains(file, []byte(`"Valid"`)) || !bytes.Contains(file, []byte(`"npl_percent": 3.1`)) {
		t.Errorf("bank figures are not plain numbers:\n%s", file)
	}

	_, err = c.DeleteHolding(ctx, connect.NewRequest(&ntxv1.DeleteHoldingRequest{Symbol: "MEN"}))
	must(t, err)
	must(t, importFile(c, file))
	if after := readBook(t, c); !proto.Equal(before, after) {
		t.Errorf("after import:\n%v\nbefore:\n%v", after, before)
	}
	if again := export(t, c); !bytes.Equal(again, file) {
		t.Errorf("second export differs:\n%s\n%s", again, file)
	}
}

// A file the import cannot take leaves every table as it was.
func TestBadImportChangesNothing(t *testing.T) {
	c, _ := clientAndDB(t, testToken)
	book(t, c)
	file := export(t, c)
	for name, bad := range map[string]string{
		"not json":       "not json",
		"wrong version":  `{"version": 7}`,
		"unknown field":  `{"version": 1, "watchlist": []}`,
		"orphan row":     `{"version": 1, "holdings": [], "transactions": [{"id": 1, "holding_id": 99, "day": "2026-01-01", "kind": "buy", "quantity": 10, "price": 100}]}`,
		"unknown bank":   `{"version": 1, "bank_metrics": [{"id": 1, "symbol": "NOPE", "fiscal_year": "2082/83", "quarter": "Q4", "published_date": "2026-08-20", "source_url": "https://x.com/a"}]}`,
		"double holding": strings.Replace(string(file), `"holdings": [`, `"holdings": [{"id": 9, "symbol": "MEN"},`, 1),
	} {
		if err := importFile(c, []byte(bad)); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: got %v", name, err)
		}
	}
	if again := export(t, c); !bytes.Equal(again, file) {
		t.Errorf("a refused import changed the tables:\n%s\n%s", again, file)
	}
}
