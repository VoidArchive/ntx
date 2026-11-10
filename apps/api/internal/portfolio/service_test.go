package portfolio

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

const testToken = "owner-secret"

// client serves the portfolio behind the token over HTTP, as the web reaches it,
// and returns a client that sends the given header value (empty sends none).
func client(t *testing.T, header string) ntxv1connect.PortfolioServiceClient {
	t.Helper()
	ctx := context.Background()
	q := database.OpenTestDB(t)
	if err := q.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
		ID:             1,
		Name:           "Mountain Energy Nepal",
		Symbol:         "MEN",
		Status:         "A",
		Sector:         "Hydropower",
		InstrumentType: "Equity",
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.UpsertPrice(ctx, sqlc.UpsertPriceParams{
		CompanyID: 1, BusinessDate: "2026-10-16", ClosePrice: sql.NullFloat64{Float64: 664.80, Valid: true},
	}); err != nil {
		t.Fatal(err)
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
	return ntxv1connect.NewPortfolioServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(withHeader))
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
		Symbol: "MEN", Status: ntxv1.HoldingStatus_HOLDING_STATUS_BUILD, TargetShares: 150, DrawdownGate: true,
		Thesis: "42 MW Mistri Khola, Kaligandaki.",
	}))
	must(t, err)
	if h := up.Msg.GetHolding(); h.GetName() != "Mountain Energy Nepal" ||
		h.GetStatus() != ntxv1.HoldingStatus_HOLDING_STATUS_BUILD {
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
	if men.GetQuantity() != 144 || math.Abs(men.GetWac()-463.78) > 0.01 || men.GetClose() != 664.80 {
		t.Errorf("MEN qty %d wac %.2f close %.2f", men.GetQuantity(), men.GetWac(), men.GetClose())
	}
	if math.Abs(men.GetValue()-95731.20) > 0.01 || math.Abs(men.GetWeightPercent()-64.92) > 0.01 {
		t.Errorf("MEN value %.2f weight %.2f, want 95731.20 and 64.92", men.GetValue(), men.GetWeightPercent())
	}
	if !fund.GetIsFund() || math.Abs(fund.GetValue()-51727.83) > 0.01 || fund.GetSector() != "Mutual Fund" {
		t.Errorf("fund %v", fund)
	}
	if len(men.GetTransactions()) != 2 || len(men.GetNotes()) != 1 {
		t.Errorf("MEN rows: %d transactions, %v", len(men.GetTransactions()), men.GetNotes())
	}
	if m.GetOpening() != "Rs 1.47 lakh across 1 position and 1 fund." {
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
	if len(res.Msg.GetHoldings()) != 1 || res.Msg.GetOpening() != "Rs 95,731 across 1 position." {
		t.Errorf("after delete: %d holdings, opening %q", len(res.Msg.GetHoldings()), res.Msg.GetOpening())
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
