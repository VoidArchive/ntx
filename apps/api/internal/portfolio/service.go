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
	"github.com/voidarchive/ntx/internal/story"
)

// now is the sheet's and calendar's clock; tests pin it.
var now = time.Now

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

func normSymbol(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }

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
	res, err := s.portfolio(ctx, req.Msg.GetFundNavs())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

// portfolio is the whole page: holdings with weights, gate verdicts, caps and the lede.
func (s *Service) portfolio(ctx context.Context, navs map[string]float64) (*ntxv1.GetPortfolioResponse, error) {
	holdings, err := s.load(ctx, navs)
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
	rules, err := s.queries.GetRules(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res.Rules = rulesToProto(rules)
	if err := s.judge(ctx, holdings); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	plants, err := s.queries.ListAllHydroPlants(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res.Caps = caps(res, rules, plants)
	index, err := s.indexYear(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	today := now()
	res.OrderSheet = sheet(month(today), rules, index, holdings)
	res.Calendar = calendar(today, holdings)
	res.CalendarOpening = calendarOpening(res.Calendar)
	if res.Candidates, err = s.candidates(ctx, res, rules, plants); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res.CandidatesOpening = candidatesOpening(len(res.Candidates), res, rules)
	res.Opening = opening(res)
	return res, nil
}

// candidates prices the screener as the screener page does and filters it by the rules.
// shortcut: every company's quarterly rows are parsed per request, as ListScreener does; store the trailing PE beside the rating when the page feels slow.
func (s *Service) candidates(
	ctx context.Context,
	res *ntxv1.GetPortfolioResponse,
	rules sqlc.Rule,
	plants []sqlc.HydroPlant,
) ([]*ntxv1.Candidate, error) {
	rows, err := s.queries.ListScreener(ctx, story.RatingVersion)
	if err != nil {
		return nil, err
	}
	earningsRows, err := s.queries.ListEarningsRows(ctx, nil)
	if err != nil {
		return nil, err
	}
	quarterly := map[int64][]sqlc.Fundamental{}
	for _, e := range earningsRows {
		quarterly[e.Fundamental.CompanyID] = append(quarterly[e.Fundamental.CompanyID], e.Fundamental)
	}
	pes := map[int64]float64{}
	for _, row := range rows {
		if e, ok := story.LatestEarnings(quarterly[row.ID]); ok && e.EPS > 0 && row.ClosePrice.Valid {
			pes[row.ID] = row.ClosePrice.Float64 / e.EPS
		}
	}
	return candidates(rows, pes, res, rules, plants), nil
}

// indexYear is the index's trailing 52 weeks of closes, oldest first; nil without a market day.
func (s *Service) indexYear(ctx context.Context) ([]sqlc.ListIndexClosesRow, error) {
	latest, err := s.queries.GetLatestMarketDay(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	day, err := time.Parse(time.DateOnly, latest.BusinessDate)
	if err != nil {
		return nil, err
	}
	return s.queries.ListIndexCloses(ctx, day.AddDate(0, 0, -364).Format(time.DateOnly))
}

// opening is the lede: the value at its close, then what the owner's rules say.
func opening(p *ntxv1.GetPortfolioResponse) string {
	if len(p.GetHoldings()) == 0 {
		return "No holdings recorded."
	}
	parts := []string{value(p)}
	var triggered []string
	needs := 0
	for _, h := range p.GetHoldings() {
		for _, g := range h.GetGates() {
			switch g.GetVerdict() {
			case Triggered:
				triggered = append(triggered, h.GetSymbol()+" "+g.GetLabel())
			case NeedsData:
				needs++
			}
		}
	}
	if len(triggered) > 0 {
		parts = append(
			parts,
			fmt.Sprintf("%s triggered: %s.", plural(len(triggered), "gate"), strings.Join(triggered, ", ")),
		)
	}
	if needs == 1 {
		parts = append(parts, "1 gate needs data.")
	} else if needs > 1 {
		parts = append(parts, fmt.Sprintf("%d gates need data.", needs))
	}
	var over []string
	for _, c := range p.GetCaps() {
		if c.GetVerdict() != Holds {
			over = append(over, c.GetDetail())
		}
	}
	if len(over) > 0 {
		parts = append(parts, "Over cap: "+strings.Join(over, ", ")+".")
	}
	parts = append(parts, p.GetOrderSheet().GetOpening())
	return strings.Join(parts, " ")
}

// value names the newest close the total rests on; fund rows carry their own NAV date.
func value(p *ntxv1.GetPortfolioResponse) string {
	across := plural(int(p.GetPositions()), "position")
	if p.GetFunds() > 0 {
		across += " and " + plural(int(p.GetFunds()), "fund")
	}
	day := ""
	for _, h := range p.GetHoldings() {
		day = max(day, h.GetCloseDay())
	}
	if day == "" {
		return fmt.Sprintf("Rs %s across %s, at the latest NAV.", rupees(p.GetTotalValue()), across)
	}
	return fmt.Sprintf("Rs %s across %s, at the %s close.", rupees(p.GetTotalValue()), across, day)
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
	gates, err := s.queries.ListGates(ctx)
	if err != nil {
		return nil, err
	}
	gatesBy := map[int64][]sqlc.Gate{}
	for _, g := range gates {
		gatesBy[g.HoldingID] = append(gatesBy[g.HoldingID], g)
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
		h.Gates = gatesToProto(r.Symbol, gatesBy[r.ID])
		out = append(out, h)
	}
	return out, nil
}

func (s *Service) UpsertHolding(
	ctx context.Context,
	req *connect.Request[ntxv1.UpsertHoldingRequest],
) (*connect.Response[ntxv1.UpsertHoldingResponse], error) {
	m := req.Msg
	symbol := normSymbol(m.GetSymbol())
	if symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}
	if m.GetReviewDate() != "" {
		if _, err := time.Parse(time.DateOnly, m.GetReviewDate()); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}
	gate, err := s.drawdownDefault(ctx, symbol, m.DrawdownGate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
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
	p, err := s.portfolio(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, h := range p.GetHoldings() {
		if h.GetSymbol() == symbol {
			return connect.NewResponse(&ntxv1.UpsertHoldingResponse{Holding: h}), nil
		}
	}
	return nil, connect.NewError(connect.CodeInternal, errors.New("holding vanished after upsert"))
}

// drawdownDefault keeps an explicit choice, else an existing row's, else the
// sector's: on for hydropower, whose plants carry the flood and GLOF risk.
func (s *Service) drawdownDefault(ctx context.Context, symbol string, chosen *bool) (int64, error) {
	if chosen != nil {
		return flag(*chosen), nil
	}
	if h, err := s.queries.GetHolding(ctx, symbol); err == nil {
		return h.DrawdownGate, nil
	}
	c, err := s.queries.GetCompany(ctx, symbol)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return flag(isHydro(c.Sector)), nil
}

func (s *Service) DeleteHolding(
	ctx context.Context,
	req *connect.Request[ntxv1.DeleteHoldingRequest],
) (*connect.Response[ntxv1.DeleteHoldingResponse], error) {
	if err := s.queries.DeleteHolding(ctx, normSymbol(req.Msg.GetSymbol())); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.DeleteHoldingResponse{}), nil
}

func (s *Service) holding(ctx context.Context, symbol string) (sqlc.Holding, error) {
	h, err := s.queries.GetHolding(ctx, normSymbol(symbol))
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
	if m.GetKind() == ntxv1.TransactionKind_TRANSACTION_KIND_SELL {
		txs, err := s.queries.ListTransactionsByHolding(ctx, h.ID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		if held, _ := position(txs); m.GetQuantity() > held {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("only %d shares held", held))
		}
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

// gatesToProto carries the stored rule; judge fills the verdict.
func gatesToProto(symbol string, gates []sqlc.Gate) []*ntxv1.Gate {
	out := make([]*ntxv1.Gate, 0, len(gates))
	for _, g := range gates {
		pg := &ntxv1.Gate{
			Id:        g.ID,
			Symbol:    symbol,
			Kind:      gateKindFrom(g.Kind),
			Metric:    g.Metric,
			Direction: directionFrom(g.Direction),
			Threshold: g.Threshold,
			Periods:   int32(min(g.Periods, 1<<31-1)), //nolint:gosec // clamped
			Text:      g.Text,
			CheckedOn: g.CheckedOn,
			Label:     g.Text,
		}
		if pg.Kind == ntxv1.GateKind_GATE_KIND_METRIC {
			pg.Label, pg.Unit = label(g.Metric, g.Direction, g.Threshold, int(g.Periods)), metrics[g.Metric].unit
		}
		out = append(out, pg)
	}
	return out
}

func (s *Service) MarkMonthExecuted(
	ctx context.Context,
	req *connect.Request[ntxv1.MarkMonthExecutedRequest],
) (*connect.Response[ntxv1.MarkMonthExecutedResponse], error) {
	for symbol, carry := range req.Msg.GetCarries() {
		if carry < 0 {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				fmt.Errorf("%s: a carry cannot be negative", symbol),
			)
		}
		if _, err := s.holding(ctx, symbol); err != nil {
			return nil, err
		}
		if err := s.queries.UpdateHoldingCarry(
			ctx,
			sqlc.UpdateHoldingCarryParams{CarryRs: carry, Symbol: normSymbol(symbol)},
		); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	return connect.NewResponse(&ntxv1.MarkMonthExecutedResponse{}), nil
}
