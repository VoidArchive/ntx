package portfolio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/story"
)

const testToken = "owner-secret"

// Every test reads the book on 2026-11-10, Kartik 24, 2083.
func init() {
	now = func() time.Time { return time.Date(2026, time.November, 10, 0, 0, 0, 0, time.UTC) }
}

func client(t *testing.T, header string) ntxv1connect.PortfolioServiceClient {
	t.Helper()
	c, _ := clientAndDB(t, header)
	return c
}

// clientAndDB serves the portfolio behind the token over HTTP, as the web reaches it,
// and returns a client that sends the given header value (empty sends none).
// MEN and SAHAS are listed with the Oct 16 closes; both have a plant in the Gandaki basin.
func clientAndDB(t *testing.T, header string) (ntxv1connect.PortfolioServiceClient, *sqlc.Queries) {
	t.Helper()
	ctx := context.Background()
	q := database.OpenTestDB(t)
	for _, c := range []struct {
		id            int64
		name, symbol  string
		close         float64
		plant, source string
	}{
		{1, "Mountain Energy Nepal", "MEN", 664.80, "Mistri Khola", "https://men.com.np"},
		{2, "Sahas Urja", "SAHAS", 685, "Solu Khola", "https://sahasurja.com"},
	} {
		must(t, q.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
			ID: c.id, Name: c.name, Symbol: c.symbol, Status: "A", Sector: "Hydropower", InstrumentType: "Equity",
		}))
		must(t, q.UpsertPrice(ctx, sqlc.UpsertPriceParams{
			CompanyID: c.id, BusinessDate: "2026-10-16", ClosePrice: sql.NullFloat64{Float64: c.close, Valid: true},
		}))
		must(t, q.UpsertHydroPlant(ctx, sqlc.UpsertHydroPlantParams{
			Symbol: c.symbol, Plant: c.plant, River: c.plant, Basin: "Gandaki", District: "Myagdi",
			Mw: 42, Lat: 28.4, Lon: 83.6, Status: "operating", SourceUrl: c.source,
		}))
	}

	mux := http.NewServeMux()
	mux.Handle(Handler(q, testToken))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	withHeader := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if header != "" {
				req.Header().Set(TokenHeader, header)
			}
			return next(ctx, req)
		}
	})
	return ntxv1connect.NewPortfolioServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(withHeader)), q
}

func TestEveryMethodRefusesWithoutTheToken(t *testing.T) {
	ctx := context.Background()
	for name, c := range map[string]ntxv1connect.PortfolioServiceClient{"missing": client(t, ""), "wrong": client(t, "guess")} {
		calls := map[string]func() error{
			"GetPortfolio": func() error {
				_, err := c.GetPortfolio(ctx, connect.NewRequest(&ntxv1.GetPortfolioRequest{}))
				return err
			},
			"UpsertHolding": func() error {
				_, err := c.UpsertHolding(ctx, connect.NewRequest(&ntxv1.UpsertHoldingRequest{Symbol: "MEN"}))
				return err
			},
			"DeleteHolding": func() error {
				_, err := c.DeleteHolding(ctx, connect.NewRequest(&ntxv1.DeleteHoldingRequest{Symbol: "MEN"}))
				return err
			},
			"AddTransaction": func() error {
				_, err := c.AddTransaction(ctx, connect.NewRequest(&ntxv1.AddTransactionRequest{Symbol: "MEN"}))
				return err
			},
			"AddNote": func() error {
				_, err := c.AddNote(ctx, connect.NewRequest(&ntxv1.AddNoteRequest{Symbol: "MEN"}))
				return err
			},
			"UpsertGate": func() error {
				_, err := c.UpsertGate(ctx, connect.NewRequest(&ntxv1.UpsertGateRequest{Symbol: "MEN"}))
				return err
			},
			"DeleteGate": func() error {
				_, err := c.DeleteGate(ctx, connect.NewRequest(&ntxv1.DeleteGateRequest{Id: 1}))
				return err
			},
			"UpdateRules": func() error {
				_, err := c.UpdateRules(ctx, connect.NewRequest(&ntxv1.UpdateRulesRequest{}))
				return err
			},
			"GetHoldingStory": func() error {
				_, err := c.GetHoldingStory(ctx, connect.NewRequest(&ntxv1.GetHoldingStoryRequest{Symbol: "MEN"}))
				return err
			},
			"MarkMonthExecuted": func() error {
				_, err := c.MarkMonthExecuted(ctx, connect.NewRequest(&ntxv1.MarkMonthExecutedRequest{}))
				return err
			},
			"UpsertBankMetric": func() error {
				_, err := c.UpsertBankMetric(ctx, connect.NewRequest(&ntxv1.UpsertBankMetricRequest{}))
				return err
			},
		}
		for method, call := range calls {
			err := call()
			var cerr *connect.Error
			if !errors.As(err, &cerr) || cerr.Code() != connect.CodeUnauthenticated || cerr.Message() != "" {
				t.Errorf("%s token, %s: got %v, want Unauthenticated with no detail", name, method, err)
			}
		}
	}
}

func boolf(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// book records the owner's Oct 16 position: MEN 120 at 556.54 plus the 24-share
// bonus and a note, and NIBLSF 5,027 units at par.
func book(t *testing.T, c ntxv1connect.PortfolioServiceClient) {
	t.Helper()
	ctx := context.Background()
	up, err := c.UpsertHolding(ctx, connect.NewRequest(&ntxv1.UpsertHoldingRequest{
		Symbol: "MEN", Status: ntxv1.HoldingStatus_HOLDING_STATUS_BUILD, TargetShares: 150,
		Thesis: "42 MW Mistri Khola, Kaligandaki.",
	}))
	must(t, err)
	// The drawdown gate was not chosen: a hydro holding gets it on by default.
	if h := up.Msg.GetHolding(); h.GetName() != "Mountain Energy Nepal" ||
		h.GetStatus() != ntxv1.HoldingStatus_HOLDING_STATUS_BUILD || !h.GetDrawdownGate() {
		t.Fatalf("upsert returned %v", h)
	}
	_, err = c.UpsertHolding(ctx, connect.NewRequest(&ntxv1.UpsertHoldingRequest{Symbol: "NIBLSF"}))
	must(t, err)
	for _, tx := range []*ntxv1.AddTransactionRequest{
		{Symbol: "MEN", Day: "2026-02-20", Kind: ntxv1.TransactionKind_TRANSACTION_KIND_BUY, Quantity: 120, Price: 556.54},
		{Symbol: "MEN", Day: "2026-11-08", Kind: ntxv1.TransactionKind_TRANSACTION_KIND_BONUS, Quantity: 24},
		{Symbol: "NIBLSF", Day: "2026-10-01", Kind: ntxv1.TransactionKind_TRANSACTION_KIND_BUY, Quantity: 5027, Price: 10},
	} {
		_, err := c.AddTransaction(ctx, connect.NewRequest(tx))
		must(t, err)
	}
	_, err = c.AddNote(ctx, connect.NewRequest(&ntxv1.AddNoteRequest{
		Symbol: "MEN", Day: "2026-10-16", Text: "Parked at 120; over the soft cap.",
	}))
	must(t, err)
}

// MEN is the one position; the fund, priced at NAV 10.29, is in the denominator but is not a position.
func TestOwnerReadsWeightsWithFundsInTheDenominator(t *testing.T) {
	c := client(t, testToken)
	book(t, c)

	res, err := c.GetPortfolio(context.Background(), connect.NewRequest(&ntxv1.GetPortfolioRequest{
		FundNavs: map[string]float64{"NIBLSF": 10.29},
	}))
	must(t, err)
	m := res.Msg
	if len(m.GetHoldings()) != 2 || m.GetPositions() != 1 || m.GetFunds() != 1 {
		t.Fatalf("holdings %d positions %d funds %d", len(m.GetHoldings()), m.GetPositions(), m.GetFunds())
	}
	men, fund := m.GetHoldings()[0], m.GetHoldings()[1]
	for name, got := range map[string][2]float64{
		"MEN quantity":   {float64(men.GetQuantity()), 144},
		"MEN wac":        {men.GetWac(), 463.78},
		"MEN close":      {men.GetClose(), 664.80},
		"MEN value":      {men.GetValue(), 95731.20},
		"MEN weight":     {men.GetWeightPercent(), 64.92},
		"fund value":     {fund.GetValue(), 51727.83},
		"MEN rows":       {float64(len(men.GetTransactions())*10 + len(men.GetNotes())), 21},
		"fund is a fund": {boolf(fund.GetIsFund()), 1},
	} {
		if math.Abs(got[0]-got[1]) > 0.01 {
			t.Errorf("%s = %.2f, want %.2f", name, got[0], got[1])
		}
	}
	if men.GetCloseDay() != "2026-10-16" || fund.GetSector() != "Mutual Fund" {
		t.Errorf("close day %q, fund sector %q", men.GetCloseDay(), fund.GetSector())
	}
	if m.GetOpening() != "Rs 1.47 lakh across 1 position and 1 fund, at the 2026-10-16 close. "+
		"Over cap: MEN 64.9% of the 25% hard cap, Hydropower 64.9% of the 40% sector cap."+
		" Order sheet for Kartik: Rs 12,000; no index close recorded, so the budget stands." {
		t.Errorf("opening = %q", m.GetOpening())
	}
}

func TestDeletingAHoldingDropsItsRows(t *testing.T) {
	c := client(t, testToken)
	book(t, c)
	ctx := context.Background()

	_, err := c.DeleteHolding(ctx, connect.NewRequest(&ntxv1.DeleteHoldingRequest{Symbol: "NIBLSF"}))
	must(t, err)
	res, err := c.GetPortfolio(ctx, connect.NewRequest(&ntxv1.GetPortfolioRequest{}))
	must(t, err)
	if len(res.Msg.GetHoldings()) != 1 ||
		res.Msg.GetOpening() != "Rs 95,731 across 1 position, at the 2026-10-16 close. "+
			"Over cap: MEN 100.0% of the 25% hard cap, Hydropower 100.0% of the 40% sector cap."+
			" Order sheet for Kartik: Rs 12,000; no index close recorded, so the budget stands." {
		t.Errorf("after delete: %d holdings, opening %q", len(res.Msg.GetHoldings()), res.Msg.GetOpening())
	}
	// Re-adding the symbol starts from nothing: its transactions went with it.
	again, err := c.UpsertHolding(ctx, connect.NewRequest(&ntxv1.UpsertHoldingRequest{Symbol: "NIBLSF"}))
	must(t, err)
	if again.Msg.GetHolding().GetQuantity() != 0 || len(again.Msg.GetHolding().GetTransactions()) != 0 {
		t.Errorf("rows survived the delete: %v", again.Msg.GetHolding())
	}
}

func TestSaleBeyondTheHoldingIsRefused(t *testing.T) {
	c := client(t, testToken)
	book(t, c)
	_, err := c.AddTransaction(context.Background(), connect.NewRequest(&ntxv1.AddTransactionRequest{
		Symbol: "MEN", Day: "2026-11-09", Kind: ntxv1.TransactionKind_TRANSACTION_KIND_SELL, Quantity: 145, Price: 700,
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("got %v, want InvalidArgument", err)
	}
}

func TestTransactionNeedsAHolding(t *testing.T) {
	c := client(t, testToken)
	_, err := c.AddTransaction(context.Background(), connect.NewRequest(&ntxv1.AddTransactionRequest{
		Symbol: "EBL", Day: "2026-09-24", Kind: ntxv1.TransactionKind_TRANSACTION_KIND_BUY, Quantity: 13, Price: 697,
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("got %v, want NotFound", err)
	}
}

// filings seeds MEN's last two quarters as the exchange filed them, and SAHAS's
// 10% bonus notice for FY 2082/83.
func seedFilings(t *testing.T, q *sqlc.Queries) {
	t.Helper()
	ctx := context.Background()
	for _, f := range []struct {
		fy, quarter string
		eps, bv     float64
	}{{"2025-2026", "Fourth Quarter", 20.10, 141.2}, {"2026-2027", "First Quarter", 25.52, 146.7}} {
		must(t, q.UpsertFundamental(ctx, sqlc.UpsertFundamentalParams{
			CompanyID: 1, FiscalYear: f.fy, Quarter: f.quarter,
			Eps:       sql.NullFloat64{Float64: f.eps, Valid: true},
			BookValue: sql.NullFloat64{Float64: f.bv, Valid: true},
			SourceUrl: sql.NullString{String: "https://nepalstock.com/men/" + f.quarter, Valid: true},
		}))
	}
	must(t, q.UpsertCorporateAction(ctx, sqlc.UpsertCorporateActionParams{
		CompanyID: 2, FiscalYear: "2082/83", BonusPercentage: sql.NullFloat64{Float64: 10, Valid: true},
		SubmittedDate: sql.NullString{String: "2026-10-20", Valid: true},
		NoticeUrl:     sql.NullString{String: "https://nepalstock.com/sahas-bonus", Valid: true},
	}))
}

func gate(t *testing.T, c ntxv1connect.PortfolioServiceClient, g *ntxv1.UpsertGateRequest) *ntxv1.Gate {
	t.Helper()
	res, err := c.UpsertGate(context.Background(), connect.NewRequest(g))
	must(t, err)
	return res.Msg.GetGate()
}

// MEN at 62% is over every cap, Hydropower is over the sector cap and both hydro
// holdings share the Gandaki basin; the bank metric has no owner row for Q1.
func TestGatesAndCapsOnTheBook(t *testing.T) {
	c, q := clientAndDB(t, testToken)
	seedFilings(t, q)
	book(t, c)
	ctx := context.Background()
	off := false
	_, err := c.UpsertHolding(ctx, connect.NewRequest(&ntxv1.UpsertHoldingRequest{Symbol: "SAHAS", DrawdownGate: &off}))
	must(t, err)
	_, err = c.AddTransaction(ctx, connect.NewRequest(&ntxv1.AddTransactionRequest{
		Symbol: "SAHAS", Day: "2026-10-01", Kind: ntxv1.TransactionKind_TRANSACTION_KIND_BUY, Quantity: 10, Price: 685,
	}))
	must(t, err)

	metric := ntxv1.GateKind_GATE_KIND_METRIC
	for _, g := range []*ntxv1.UpsertGateRequest{
		{Symbol: "MEN", Kind: metric, Metric: "eps", Direction: ntxv1.GateDirection_GATE_DIRECTION_BELOW, Threshold: 15, Periods: 2},
		{Symbol: "MEN", Kind: metric, Metric: "weight_percent", Direction: ntxv1.GateDirection_GATE_DIRECTION_ABOVE, Threshold: 20},
		{Symbol: "MEN", Kind: metric, Metric: "drawdown_percent", Direction: ntxv1.GateDirection_GATE_DIRECTION_ABOVE, Threshold: 15},
		{Symbol: "MEN", Kind: metric, Metric: "npl_percent", Direction: ntxv1.GateDirection_GATE_DIRECTION_ABOVE, Threshold: 5},
		{Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_EVENT, Text: "Plant damage reported", CheckedOn: "2026-10-16"},
		{Symbol: "SAHAS", Kind: metric, Metric: "drawdown_percent", Direction: ntxv1.GateDirection_GATE_DIRECTION_ABOVE, Threshold: 15},
	} {
		gate(t, c, g)
	}

	res, err := c.GetPortfolio(
		ctx,
		connect.NewRequest(&ntxv1.GetPortfolioRequest{FundNavs: map[string]float64{"NIBLSF": 10.29}}),
	)
	must(t, err)
	m := res.Msg
	men := m.GetHoldings()[0]
	if men.GetSymbol() != "MEN" || len(men.GetGates()) != 5 {
		t.Fatalf("MEN has %d gates", len(men.GetGates()))
	}
	assertGates(t, men.GetGates(), []string{
		"holds Q1 2083/84 EPS below Rs 15 for 2 quarters",
		"triggered 2026-10-16 Weight above 20%",
		"holds 2026-10-16 Drawdown above 15%",
		"needs data Q1 2083/84 NPL above 5%",
		"  Plant damage reported",
	})
	// SAHAS switched the drawdown gate off, so its gate is not run rather than short of data.
	assertGates(t, m.GetHoldings()[2].GetGates(), []string{"  Drawdown above 15%"})
	if eps := men.GetGates()[0]; eps.GetValue() != 25.52 ||
		eps.GetSource() != "https://nepalstock.com/men/First Quarter" {
		t.Errorf("eps gate carries %v from %q", eps.GetValue(), eps.GetSource())
	}
	if men.GetCapVerdict() != Triggered || m.GetHoldings()[1].GetCapVerdict() != "" ||
		m.GetHoldings()[2].GetCapVerdict() != Holds {
		t.Errorf(
			"cap verdicts %q %q %q",
			men.GetCapVerdict(),
			m.GetHoldings()[1].GetCapVerdict(),
			m.GetHoldings()[2].GetCapVerdict(),
		)
	}
	assertCaps(t, m.GetCaps(), []string{
		"Single holding: triggered, MEN 62.0% of the 25% hard cap",
		"Sector cap: triggered, Hydropower 66.5% of the 40% sector cap",
		"Positions: holds, 2 of 8",
		"One hydro per basin: triggered, Gandaki basin: MEN and SAHAS",
	})
	if m.GetOpening() != "Rs 1.54 lakh across 2 positions and 1 fund, at the 2026-10-16 close. "+
		"1 gate triggered: MEN Weight above 20%. 1 gate needs data. "+
		"Over cap: MEN 62.0% of the 25% hard cap, Hydropower 66.5% of the 40% sector cap, Gandaki basin: MEN and SAHAS."+
		" Order sheet for Kartik: Rs 12,000; no index close recorded, so the budget stands." {
		t.Errorf("opening = %q", m.GetOpening())
	}
	if m.GetRules().GetSoftCapPercent() != 20 || m.GetRules().GetMaxPositions() != 8 ||
		!m.GetRules().GetOneHydroPerBasin() {
		t.Errorf("default rules %v", m.GetRules())
	}
}

// MEN at 21.3% of the book, the owner's Oct 16 reading: over the soft cap, under the hard cap.
func TestOverTheSoftCapReadsWatch(t *testing.T) {
	c := client(t, testToken)
	book(t, c)
	ctx := context.Background()
	up, err := c.UpdateRules(ctx, connect.NewRequest(&ntxv1.UpdateRulesRequest{Rules: &ntxv1.Rules{
		SoftCapPercent: 20, HardCapPercent: 25, SectorCapPercent: 40, MaxPositions: 7, MinEntryRs: 15000,
		OneHydroPerBasin: true, CandidateMinRating: "B", CandidateDividendRecord: true, MonthlyBudgetRs: 12000,
		MultiplierBelowMedian: 1.5, MultiplierBottomQuarter: 2,
	}}))
	must(t, err)
	if up.Msg.GetRules().GetMaxPositions() != 7 {
		t.Fatalf("rules %v", up.Msg.GetRules())
	}
	res, err := c.GetPortfolio(
		ctx,
		connect.NewRequest(&ntxv1.GetPortfolioRequest{FundNavs: map[string]float64{"NIBLSF": 70.36}}),
	)
	must(t, err)
	if men := res.Msg.GetHoldings()[0]; men.GetCapVerdict() != Watch {
		t.Errorf("MEN cap %q", men.GetCapVerdict())
	}
	assertCaps(t, res.Msg.GetCaps(), []string{
		"Single holding: watch, MEN 21.3% of the 20% soft cap",
		"Sector cap: triggered, Mutual Fund 78.7% of the 40% sector cap",
		"Positions: holds, 1 of 7",
		"One hydro per basin: holds, no basin held twice",
	})
	if res.Msg.GetOpening() != "Rs 4.49 lakh across 1 position and 1 fund, at the 2026-10-16 close. "+
		"Over cap: MEN 21.3% of the 20% soft cap, Mutual Fund 78.7% of the 40% sector cap."+
		" Order sheet for Kartik: Rs 12,000; no index close recorded, so the budget stands." {
		t.Errorf("opening = %q", res.Msg.GetOpening())
	}
}

func TestGateEditsAndTheBonusHint(t *testing.T) {
	c, q := clientAndDB(t, testToken)
	seedFilings(t, q)
	book(t, c)
	ctx := context.Background()
	_, err := c.UpsertHolding(ctx, connect.NewRequest(&ntxv1.UpsertHoldingRequest{Symbol: "SAHAS"}))
	must(t, err)

	g := gate(t, c, &ntxv1.UpsertGateRequest{
		Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_METRIC, Metric: "eps",
		Direction: ntxv1.GateDirection_GATE_DIRECTION_BELOW, Threshold: 15,
	})
	if g.GetId() == 0 || g.GetVerdict() != Holds || g.GetPeriods() != 1 {
		t.Fatalf("inserted %v", g)
	}
	edited := gate(t, c, &ntxv1.UpsertGateRequest{
		Id: g.GetId(), Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_METRIC, Metric: "eps",
		Direction: ntxv1.GateDirection_GATE_DIRECTION_BELOW, Threshold: 30,
	})
	if edited.GetId() != g.GetId() || edited.GetVerdict() != Triggered {
		t.Errorf("edited %v", edited)
	}
	// A gate id belongs to one holding; another symbol cannot edit it.
	if _, err := c.UpsertGate(ctx, connect.NewRequest(&ntxv1.UpsertGateRequest{
		Id: g.GetId(), Symbol: "SAHAS", Kind: ntxv1.GateKind_GATE_KIND_METRIC, Metric: "eps",
		Direction: ntxv1.GateDirection_GATE_DIRECTION_BELOW, Threshold: 30,
	})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("cross-holding edit: %v", err)
	}

	story, err := c.GetHoldingStory(ctx, connect.NewRequest(&ntxv1.GetHoldingStoryRequest{Symbol: "men"}))
	must(t, err)
	if h := story.Msg.GetHolding(); len(h.GetGates()) != 1 || len(h.GetNotes()) != 1 || h.GetQuantity() != 144 {
		t.Errorf("story %v", h)
	}
	// MEN's bonus is in the book; SAHAS's notice is not.
	if story.Msg.GetBonusHint() != "" {
		t.Errorf("MEN hint %q", story.Msg.GetBonusHint())
	}
	sahas, err := c.GetHoldingStory(ctx, connect.NewRequest(&ntxv1.GetHoldingStoryRequest{Symbol: "SAHAS"}))
	must(t, err)
	if sahas.Msg.GetBonusHint() != "A 10% bonus for FY 2082/83, published 2026-10-20, is not yet credited in the book." {
		t.Errorf("SAHAS hint %q", sahas.Msg.GetBonusHint())
	}

	_, err = c.DeleteGate(ctx, connect.NewRequest(&ntxv1.DeleteGateRequest{Id: g.GetId()}))
	must(t, err)
	story, err = c.GetHoldingStory(ctx, connect.NewRequest(&ntxv1.GetHoldingStoryRequest{Symbol: "MEN"}))
	must(t, err)
	if len(story.Msg.GetHolding().GetGates()) != 0 {
		t.Errorf("gate survived the delete")
	}
}

func TestGateEntryRefusesWhatItCannotRead(t *testing.T) {
	c := client(t, testToken)
	book(t, c)
	ctx := context.Background()
	for name, bad := range map[string]*ntxv1.UpsertGateRequest{
		"unknown metric": {Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_METRIC, Metric: "npa", Direction: ntxv1.GateDirection_GATE_DIRECTION_ABOVE},
		"no direction":   {Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_METRIC, Metric: "eps"},
		"empty event":    {Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_EVENT},
		"fund":           {Symbol: "NIBLSF", Kind: ntxv1.GateKind_GATE_KIND_METRIC, Metric: "eps", Direction: ntxv1.GateDirection_GATE_DIRECTION_ABOVE},
	} {
		if _, err := c.UpsertGate(ctx, connect.NewRequest(bad)); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: got %v, want InvalidArgument", name, err)
		}
	}
	_, err := c.GetHoldingStory(ctx, connect.NewRequest(&ntxv1.GetHoldingStoryRequest{Symbol: "EBL"}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("story of a stranger: %v", err)
	}
}

func assertCaps(t *testing.T, caps []*ntxv1.Cap, want []string) {
	t.Helper()
	var got []string
	for _, cp := range caps {
		got = append(got, cp.GetRule()+": "+cp.GetVerdict()+", "+cp.GetDetail())
	}
	if len(got) != len(want) {
		t.Fatalf("caps %q", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("cap %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func assertGates(t *testing.T, gates []*ntxv1.Gate, want []string) {
	t.Helper()
	var got []string
	for _, g := range gates {
		got = append(got, g.GetVerdict()+" "+g.GetPeriod()+" "+g.GetLabel())
	}
	if !slices.Equal(got, want) {
		t.Errorf("gates %q, want %q", got, want)
	}
}

// With the index in the bottom quarter of its year the budget doubles; MEN, a Build
// holding over the hard cap, is listed and gets nothing; the carry the owner confirms is written.
func TestOrderSheetAndCalendarOnTheBook(t *testing.T) {
	c, q := clientAndDB(t, testToken)
	book(t, c)
	ctx := context.Background()
	for i, close := range []float64{2000, 2300, 2600, 2900, 3000, 2050} {
		must(t, q.UpsertMarketDay(ctx, sqlc.UpsertMarketDayParams{
			BusinessDate: time.Date(2026, time.June+time.Month(i), 1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly),
			IndexClose:   close,
		}))
	}
	gate(t, c, &ntxv1.UpsertGateRequest{
		Symbol: "MEN", Kind: ntxv1.GateKind_GATE_KIND_EVENT, Text: "Plant damage reported", CheckedOn: "2026-10-16",
	})
	res, err := c.GetPortfolio(
		ctx,
		connect.NewRequest(&ntxv1.GetPortfolioRequest{FundNavs: map[string]float64{"NIBLSF": 10.29}}),
	)
	must(t, err)
	sh := res.Msg.GetOrderSheet()
	if sh.GetMultiplier() != 2 || sh.GetBudgetRs() != 24000 || sh.GetMonth() != "Kartik" || len(sh.GetLines()) != 1 ||
		sh.GetLines()[0].GetNote() != "over the soft cap" {
		t.Errorf("sheet %v", sh)
	}
	want := "Order sheet for Kartik: Rs 24,000, 2x the Rs 12,000 budget with the index at 2050.00 on 2026-11-01, " +
		"in the bottom quarter of its 52-week closes, 2000.00 to 3000.00."
	if sh.GetOpening() != want || !strings.HasSuffix(res.Msg.GetOpening(), want) {
		t.Errorf("sheet opening %q\nlede %q", sh.GetOpening(), res.Msg.GetOpening())
	}
	var days []string
	for _, e := range res.Msg.GetCalendar() {
		days = append(days, e.GetDay()+" "+e.GetSymbol()+" "+e.GetWhat())
	}
	if !slices.Equal(
		days,
		[]string{"2026-11-16 MEN Q1 2083/84 filing due", "2027-01-16 MEN re-check: Plant damage reported"},
	) {
		t.Errorf("calendar %q", days)
	}

	_, err = c.MarkMonthExecuted(
		ctx,
		connect.NewRequest(&ntxv1.MarkMonthExecutedRequest{Carries: map[string]float64{"men": 2570}}),
	)
	must(t, err)
	story, err := c.GetHoldingStory(ctx, connect.NewRequest(&ntxv1.GetHoldingStoryRequest{Symbol: "MEN"}))
	must(t, err)
	if story.Msg.GetHolding().GetCarryRs() != 2570 {
		t.Errorf("carry %v", story.Msg.GetHolding().GetCarryRs())
	}
	for name, bad := range map[string]float64{"EBL": 10, "MEN": -1} {
		_, err := c.MarkMonthExecuted(
			ctx,
			connect.NewRequest(&ntxv1.MarkMonthExecutedRequest{Carries: map[string]float64{name: bad}}),
		)
		if err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

// SAHAS shares MEN's basin and is left out; API sits in the Mahakali basin and is in;
// NTC, an A outside hydro, leads. The sector cap is lifted so hydro has headroom.
func TestCandidatesOnTheBook(t *testing.T) {
	c, q := clientAndDB(t, testToken)
	book(t, c)
	ctx := context.Background()
	for _, co := range []struct {
		id           int64
		symbol, name string
		sector       string
		close        float64
		rating, dr   string
		plant, basin string
	}{
		{2, "SAHAS", "Sahas Urja", "Hydropower", 685, "A", "holds", "", ""},
		{3, "API", "Api Power", "Hydro Power", 343, "B", "holds", "Naugadh Gad", "Mahakali"},
		{4, "NTC", "Nepal Doorsanchar", "Others", 910, "A", "holds", "", ""},
		{5, "NABIL", "Nabil Bank", "Commercial Banks", 500, "C", "holds", "", ""},
	} {
		if co.id != 2 {
			must(t, q.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
				ID: co.id, Name: co.name, Symbol: co.symbol, Status: "A", Sector: co.sector, InstrumentType: "Equity",
			}))
			must(t, q.UpsertPrice(ctx, sqlc.UpsertPriceParams{
				CompanyID:    co.id,
				BusinessDate: "2026-10-16",
				ClosePrice:   sql.NullFloat64{Float64: co.close, Valid: true},
			}))
		}
		if co.plant != "" {
			must(t, q.UpsertHydroPlant(ctx, sqlc.UpsertHydroPlantParams{
				Symbol: co.symbol, Plant: co.plant, River: co.plant, Basin: co.basin, District: "Darchula",
				Mw: 8, Lat: 29.8, Lon: 80.6, Status: "operating", SourceUrl: "https://apipower.com.np",
			}))
		}
		must(t, q.UpsertCompanyRating(ctx, sqlc.UpsertCompanyRatingParams{
			CompanyID: co.id, Letter: co.rating, Holds: 7, Evaluated: 9, Total: 10, Version: story.RatingVersion,
			DividendRecord: co.dr,
		}))
	}
	rules := &ntxv1.Rules{
		SoftCapPercent: 20, HardCapPercent: 25, SectorCapPercent: 90, MaxPositions: 8, MinEntryRs: 15000,
		OneHydroPerBasin: true, CandidateMinRating: "B", CandidateDividendRecord: true, MonthlyBudgetRs: 12000,
		MultiplierBelowMedian: 1.5, MultiplierBottomQuarter: 2,
	}
	_, err := c.UpdateRules(ctx, connect.NewRequest(&ntxv1.UpdateRulesRequest{Rules: rules}))
	must(t, err)

	res, err := c.GetPortfolio(
		ctx,
		connect.NewRequest(&ntxv1.GetPortfolioRequest{FundNavs: map[string]float64{"NIBLSF": 10.29}}),
	)
	must(t, err)
	var got []string
	for _, cd := range res.Msg.GetCandidates() {
		got = append(
			got,
			fmt.Sprintf(
				"%s %s %d@%.0f %s",
				cd.GetSymbol(),
				cd.GetRating(),
				cd.GetShares(),
				cd.GetCostRs(),
				cd.GetBasin(),
			),
		)
	}
	if want := []string{"NTC A 20@18200 ", "API B 50@17150 Mahakali"}; !slices.Equal(got, want) {
		t.Errorf("candidates %q, want %q", got, want)
	}
	if res.Msg.GetCandidatesOpening() != "Positions stand at 1 of 8; a new scrip needs a sale first. 2 companies pass the rules." {
		t.Errorf("opening %q", res.Msg.GetCandidatesOpening())
	}
}
