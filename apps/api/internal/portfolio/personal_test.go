package portfolio

import (
	"bytes"
	"context"
	"regexp"
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
	_, err = c.UpdateRules(ctx, connect.NewRequest(&ntxv1.UpdateRulesRequest{Rules: &ntxv1.Rules{
		SoftCapPercent: 20, HardCapPercent: 25, MaxPositions: 8,
	}}))
	must(t, err)
	must(t, importFile(c, file))
	if after := readBook(t, c); !proto.Equal(before, after) {
		t.Errorf("after import:\n%v\nbefore:\n%v", after, before)
	}
	if again := export(t, c); !bytes.Equal(again, file) {
		t.Errorf("second export differs:\n%s\n%s", again, file)
	}
}

// A book with nothing in it exports tables as "[]", so the file imports back.
func TestEmptyBookRoundTrips(t *testing.T) {
	c, _ := clientAndDB(t, testToken)
	file := export(t, c)
	if bytes.Contains(file, []byte("null")) {
		t.Errorf("an empty table exported as null:\n%s", file)
	}
	must(t, importFile(c, file))
}

// A file written before the goal and lender rules existed keeps the row's current
// values for them; a file written after carries its own.
func TestOldFileKeepsTheRulesItDoesNotName(t *testing.T) {
	c, _ := clientAndDB(t, testToken)
	ctx := context.Background()
	rules := readBook(t, c).GetRules()
	rules.GoalRs, rules.CarFloorPercent = 600000, 11
	_, err := c.UpdateRules(ctx, connect.NewRequest(&ntxv1.UpdateRulesRequest{Rules: rules}))
	must(t, err)
	file := export(t, c)
	old := regexp.MustCompile(`,\n\s+"(goal_rs|fund_sip_rs|npl_ceiling_percent|car_floor_percent|distributable_above_zero|max_pe_to_sector_median)": [0-9.]+`).
		ReplaceAll(file, nil)
	if bytes.Equal(old, file) {
		t.Fatal("the new rule keys were not in the export")
	}
	must(t, importFile(c, old))
	if got := readBook(
		t,
		c,
	).GetRules(); got.GetGoalRs() != 600000 || got.GetCarFloorPercent() != 11 ||
		got.GetNplCeilingPercent() != 5 {
		t.Errorf("after an old file: %v", got)
	}
	must(t, importFile(c, file))
	if again := export(t, c); !bytes.Equal(again, file) {
		t.Errorf("the new fields did not round-trip:\n%s\n%s", again, file)
	}
}

// table is the exported file with one table's rows replaced.
func table(file []byte, name, rows string) string {
	re := regexp.MustCompile(`"` + name + `": \[[^\]]*\]`)
	return re.ReplaceAllString(string(file), `"`+name+`": [`+rows+`]`)
}

// A file the import cannot take leaves every table as it was.
func TestBadImportChangesNothing(t *testing.T) {
	c, _ := clientAndDB(t, testToken)
	book(t, c)
	file := export(t, c)
	for name, bad := range map[string]string{
		"not json":       "not json",
		"wrong version":  strings.Replace(string(file), `"version": 1`, `"version": 7`, 1),
		"unknown field":  strings.Replace(string(file), `"holdings": [`, `"watchlist": [], "holdings": [`, 1),
		"version only":   `{"version": 1}`,
		"no rules":       strings.Replace(string(file), `"rules": {`, `"rules": null, "x": {`, 1),
		"zero rules":     strings.Replace(string(file), `"soft_cap_percent": 20`, `"soft_cap_percent": 0`, 1),
		"orphan row":     table(file, "transactions", `{"id": 1, "holding_id": 99, "day": "2026-01-01", "kind": "buy", "quantity": 10, "price": 100}`),
		"unknown bank":   table(file, "bank_metrics", `{"id": 1, "symbol": "NOPE", "fiscal_year": "2082/83", "quarter": "Q4", "npl_percent": 3, "published_date": "2026-08-20", "source_url": "https://x.com/a"}`),
		"no provenance":  table(file, "bank_metrics", `{"id": 1, "symbol": "MEN", "fiscal_year": "2082/83", "quarter": "Q4", "npl_percent": 3, "published_date": "", "source_url": ""}`),
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
