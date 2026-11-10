// Package portfolio serves the owner's holdings behind one token (ADR 0006).
package portfolio

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// TokenHeader carries the owner token on every personal call.
const TokenHeader = "Ntx-Owner-Token" //nolint:gosec // a header name, not a credential

type Service struct {
	ntxv1connect.UnimplementedPortfolioServiceHandler
	queries *sqlc.Queries
}

// Handler mounts the service behind the token; an empty token refuses everyone.
func Handler(queries *sqlc.Queries, token string) (string, http.Handler) {
	return ntxv1connect.NewPortfolioServiceHandler(
		&Service{queries: queries},
		connect.WithInterceptors(ownerOnly(token)),
	)
}

// ownerOnly answers a missing and a wrong token identically, with no detail.
func ownerOnly(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			got := req.Header().Get(TokenHeader)
			if token == "" || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}
			return next(ctx, req)
		}
	}
}

func statusOf(s ntxv1.HoldingStatus) string {
	if s == ntxv1.HoldingStatus_HOLDING_STATUS_UNSPECIFIED {
		s = ntxv1.HoldingStatus_HOLDING_STATUS_HOLD
	}
	return strings.ToLower(strings.TrimPrefix(s.String(), "HOLDING_STATUS_"))
}

func statusFrom(s string) ntxv1.HoldingStatus {
	return ntxv1.HoldingStatus(ntxv1.HoldingStatus_value["HOLDING_STATUS_"+strings.ToUpper(s)])
}

func (s *Service) GetPortfolio(
	ctx context.Context,
	req *connect.Request[ntxv1.GetPortfolioRequest],
) (*connect.Response[ntxv1.GetPortfolioResponse], error) {
	holdings, err := s.load(ctx, req.Msg.GetFundNavs())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := &ntxv1.GetPortfolioResponse{Holdings: holdings}
	for _, h := range holdings {
		res.TotalValue += h.GetValue()
		if h.GetQuantity() == 0 {
			continue
		}
		if h.GetIsFund() {
			res.Funds++
			continue
		}
		res.Positions++
	}
	for _, h := range holdings {
		if res.TotalValue > 0 {
			h.WeightPercent = 100 * h.GetValue() / res.TotalValue
		}
	}
	res.Opening = opening(res)
	return connect.NewResponse(res), nil
}

func opening(p *ntxv1.GetPortfolioResponse) string {
	if len(p.GetHoldings()) == 0 {
		return "No holdings recorded."
	}
	across := plural(int(p.GetPositions()), "position")
	if p.GetFunds() > 0 {
		across += " and " + plural(int(p.GetFunds()), "fund")
	}
	return fmt.Sprintf("Rs %s across %s.", rupees(p.GetTotalValue()), across)
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// rupees prints lakh above one lakh and a grouped integer below, as the owner's note does.
func rupees(rs float64) string {
	if rs >= 1e5 {
		return fmt.Sprintf("%.2f lakh", rs/1e5)
	}
	whole := fmt.Sprintf("%.0f", rs)
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + "," + whole[i:]
	}
	return whole
}

// load builds every holding with its derived position, price and rows. Funds are
// priced from the NAVs the caller sends; weights are the caller's to fill once the total is known.
func (s *Service) load(ctx context.Context, navs map[string]float64) ([]*ntxv1.Holding, error) {
	rows, err := s.queries.ListHoldings(ctx)
	if err != nil {
		return nil, err
	}
	txs, err := s.queries.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	notes, err := s.queries.ListNotes(ctx)
	if err != nil {
		return nil, err
	}
	byHolding := map[int64][]sqlc.Transaction{}
	for _, tx := range txs {
		byHolding[tx.HoldingID] = append(byHolding[tx.HoldingID], tx)
	}
	notesBy := map[int64][]sqlc.HoldingNote{}
	for _, n := range notes {
		notesBy[n.HoldingID] = append(notesBy[n.HoldingID], n)
	}

	out := make([]*ntxv1.Holding, 0, len(rows))
	for _, r := range rows {
		qty, wac := position(byHolding[r.ID])
		h := &ntxv1.Holding{
			Symbol:            r.Symbol,
			Name:              r.CompanyName.String,
			Sector:            r.Sector.String,
			Status:            statusFrom(r.Status),
			TargetShares:      r.TargetShares,
			AllocationPercent: r.AllocationPercent,
			CarryRs:           r.CarryRs,
			DrawdownGate:      r.DrawdownGate != 0,
			Thesis:            r.Thesis,
			FrozenReason:      r.FrozenReason,
			ReviewDate:        r.ReviewDate,
			Quantity:          qty,
			Wac:               wac,
			IsFund:            !r.CompanyID.Valid,
		}
		if h.IsFund {
			h.Name, h.Sector = r.Symbol, "Mutual Fund"
			if nav, ok := navs[r.Symbol]; ok {
				h.Close = &nav
			}
		} else if r.ClosePrice.Valid {
			h.Close, h.CloseDay = &r.ClosePrice.Float64, r.CloseDay
		}
		if h.Close != nil {
			h.Value = float64(qty) * *h.Close
		}
		for _, tx := range byHolding[r.ID] {
			h.Transactions = append(h.Transactions, &ntxv1.Transaction{
				Id: tx.ID, Day: tx.Day, Kind: kindFrom(tx.Kind), Quantity: tx.Quantity, Price: tx.Price,
			})
		}
		for _, n := range notesBy[r.ID] {
			h.Notes = append(h.Notes, &ntxv1.Note{Id: n.ID, Day: n.Day, Text: n.Text})
		}
		out = append(out, h)
	}
	return out, nil
}

func (s *Service) UpsertHolding(
	ctx context.Context,
	req *connect.Request[ntxv1.UpsertHoldingRequest],
) (*connect.Response[ntxv1.UpsertHoldingResponse], error) {
	m := req.Msg
	symbol := strings.ToUpper(strings.TrimSpace(m.GetSymbol()))
	if symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}
	if m.GetReviewDate() != "" {
		if _, err := time.Parse(time.DateOnly, m.GetReviewDate()); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}
	gate := int64(0)
	if m.GetDrawdownGate() {
		gate = 1
	}
	if _, err := s.queries.UpsertHolding(ctx, sqlc.UpsertHoldingParams{
		Symbol:            symbol,
		Status:            statusOf(m.GetStatus()),
		TargetShares:      m.GetTargetShares(),
		AllocationPercent: m.GetAllocationPercent(),
		CarryRs:           m.GetCarryRs(),
		DrawdownGate:      gate,
		Thesis:            m.GetThesis(),
		FrozenReason:      m.GetFrozenReason(),
		ReviewDate:        m.GetReviewDate(),
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	holdings, err := s.load(ctx, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	for _, h := range holdings {
		if h.GetSymbol() == symbol {
			return connect.NewResponse(&ntxv1.UpsertHoldingResponse{Holding: h}), nil
		}
	}
	return nil, connect.NewError(connect.CodeInternal, errors.New("holding vanished after upsert"))
}

func (s *Service) DeleteHolding(
	ctx context.Context,
	req *connect.Request[ntxv1.DeleteHoldingRequest],
) (*connect.Response[ntxv1.DeleteHoldingResponse], error) {
	if err := s.queries.DeleteHolding(ctx, strings.ToUpper(strings.TrimSpace(req.Msg.GetSymbol()))); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.DeleteHoldingResponse{}), nil
}

func (s *Service) holding(ctx context.Context, symbol string) (sqlc.Holding, error) {
	h, err := s.queries.GetHolding(ctx, strings.ToUpper(strings.TrimSpace(symbol)))
	if errors.Is(err, sql.ErrNoRows) {
		return h, connect.NewError(connect.CodeNotFound, fmt.Errorf("no holding %q", symbol))
	}
	if err != nil {
		return h, connect.NewError(connect.CodeInternal, err)
	}
	return h, nil
}

func (s *Service) AddTransaction(
	ctx context.Context,
	req *connect.Request[ntxv1.AddTransactionRequest],
) (*connect.Response[ntxv1.AddTransactionResponse], error) {
	m := req.Msg
	if _, err := time.Parse(time.DateOnly, m.GetDay()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if m.GetKind() == ntxv1.TransactionKind_TRANSACTION_KIND_UNSPECIFIED || m.GetQuantity() <= 0 || m.GetPrice() < 0 {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("kind, a positive quantity and a price are required"),
		)
	}
	h, err := s.holding(ctx, m.GetSymbol())
	if err != nil {
		return nil, err
	}
	tx, err := s.queries.InsertTransaction(ctx, sqlc.InsertTransactionParams{
		HoldingID: h.ID, Day: m.GetDay(), Kind: kindOf(m.GetKind()), Quantity: m.GetQuantity(), Price: m.GetPrice(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.AddTransactionResponse{Transaction: &ntxv1.Transaction{
		Id: tx.ID, Day: tx.Day, Kind: kindFrom(tx.Kind), Quantity: tx.Quantity, Price: tx.Price,
	}}), nil
}

func (s *Service) AddNote(
	ctx context.Context,
	req *connect.Request[ntxv1.AddNoteRequest],
) (*connect.Response[ntxv1.AddNoteResponse], error) {
	m := req.Msg
	if _, err := time.Parse(time.DateOnly, m.GetDay()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if strings.TrimSpace(m.GetText()) == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("text is required"))
	}
	h, err := s.holding(ctx, m.GetSymbol())
	if err != nil {
		return nil, err
	}
	n, err := s.queries.InsertNote(
		ctx,
		sqlc.InsertNoteParams{HoldingID: h.ID, Day: m.GetDay(), Text: strings.TrimSpace(m.GetText())},
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.AddNoteResponse{Note: &ntxv1.Note{Id: n.ID, Day: n.Day, Text: n.Text}}), nil
}
