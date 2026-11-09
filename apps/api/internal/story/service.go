package story

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// Service serves GetCompanyStory: it loads rows, Build computes.
type Service struct {
	ntxv1connect.UnimplementedStoryServiceHandler
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) GetCompanyStory(
	ctx context.Context,
	req *connect.Request[ntxv1.GetCompanyStoryRequest],
) (*connect.Response[ntxv1.GetCompanyStoryResponse], error) {
	if req.Msg.Symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}
	in, err := s.load(ctx, req.Msg.Symbol)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if f := req.Msg.GetFunds(); f.GetCount() > 0 {
		in.Funds = &Funds{
			Count:        int(f.GetCount()),
			Total:        f.GetTotalValue(),
			LargestName:  f.GetLargestName(),
			LargestValue: f.GetLargestValue(),
			AsOf:         f.GetAsOf(),
		}
	}
	res := toProto(Build(in))
	addQuarters(res, in)
	addFooter(res, in)
	rows, err := s.queries.ListSectorPeers(ctx, in.Company.Sector)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("sector peers: %w", err))
	}
	res.Peers = peers(rows, in.Quarterly)
	return connect.NewResponse(res), nil
}

// peers ratios the latest close against the latest quarterly row, pricing
// earnings as the strips do (trailing where four consecutive quarters exist).
func peers(rows []sqlc.ListSectorPeersRow, quarterly map[int64][]sqlc.Fundamental) []*ntxv1.Peer {
	out := make([]*ntxv1.Peer, 0, len(rows))
	for _, r := range rows {
		p := &ntxv1.Peer{Symbol: r.Symbol, Name: r.Name}
		if e, ok := LatestEarnings(quarterly[r.ID]); ok && e.EPS > 0 {
			p.Pe = proto.Float64(r.Close / e.EPS)
		}
		if r.BookValue.Float64 > 0 {
			p.Pb = proto.Float64(r.Close / r.BookValue.Float64)
		}
		out = append(out, p)
	}
	slices.SortStableFunc(out, func(a, b *ntxv1.Peer) int {
		if a.Pe == nil || b.Pe == nil {
			return cmp.Compare(boolInt(a.Pe == nil), boolInt(b.Pe == nil))
		}
		return cmp.Compare(*a.Pe, *b.Pe)
	})
	return out
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// addFooter dates the data: the newest stored price, and the newest filing by the
// exchange's date when any row has one, else by when NTX first stored it.
func addFooter(res *ntxv1.GetCompanyStoryResponse, in Input) {
	if len(in.Prices) > 0 {
		res.SyncedAt = in.Prices[0].CreatedAt.UTC().Format("2006-01-02 15:04")
	}
	seen := ""
	for _, f := range in.Fundamentals {
		res.FiledAt = max(res.FiledAt, f.PublishedAt.String)
		seen = max(seen, f.CreatedAt.UTC().Format(time.DateOnly))
	}
	if res.FiledAt == "" && len(in.Fundamentals) > 0 {
		res.FiledAt, res.FiledAtSeen = seen, true
	}
}

func (s *Service) load(ctx context.Context, symbol string) (Input, error) {
	c, err := s.queries.GetCompany(ctx, symbol)
	if err != nil {
		return Input{}, err
	}
	in := Input{Company: c}
	if in.Fundamentals, err = s.queries.ListFundamentalsByCompany(ctx, c.ID); err != nil {
		return Input{}, fmt.Errorf("fundamentals: %w", err)
	}
	if in.Prices, err = s.queries.ListPricesByCompany(
		ctx,
		// Every close, for the fiscal-year boundaries; about 250 rows a year.
		sqlc.ListPricesByCompanyParams{CompanyID: c.ID, Limit: 10_000},
	); err != nil {
		return Input{}, fmt.Errorf("prices: %w", err)
	}
	if in.Actions, err = s.queries.ListCorporateActionsByCompany(ctx, c.ID); err != nil {
		return Input{}, fmt.Errorf("corporate actions: %w", err)
	}
	ca, err := s.queries.GetLatestCorporateActionByCompany(ctx, c.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Input{}, fmt.Errorf("corporate action: %w", err)
	}
	if err == nil {
		// A recorded action without cash is a zero cash dividend, as in ListSectorYield.
		in.CashDividend = &ca.CashDividend.Float64
	}
	if in.Ownership, err = s.queries.ListOwnershipHistory(ctx, c.ID); err != nil {
		return Input{}, fmt.Errorf("ownership history: %w", err)
	}
	earningsRows, err := s.queries.ListEarningsRows(ctx, c.Sector)
	if err != nil {
		return Input{}, fmt.Errorf("sector earnings: %w", err)
	}
	in.Sector.PE = SectorPE(earningsRows)
	in.Quarterly, _, _ = Quarterly(earningsRows)
	for _, d := range []struct {
		dst  *[]float64
		list func(context.Context, string) ([]float64, error)
	}{
		{&in.Sector.PB, s.queries.ListSectorPB},
		{&in.Sector.EPS, s.queries.ListSectorEPS},
		{&in.Sector.Yield, s.queries.ListSectorYield},
	} {
		if *d.dst, err = d.list(ctx, c.Sector); err != nil {
			return Input{}, fmt.Errorf("sector distribution: %w", err)
		}
	}
	return in, nil
}

func toProto(st Story) *ntxv1.GetCompanyStoryResponse {
	res := &ntxv1.GetCompanyStoryResponse{LastFiling: st.LastFiling}
	for _, c := range st.Claims {
		res.Claims = append(res.Claims, &ntxv1.Claim{
			Kind:   c.Kind,
			Text:   c.Text,
			Value:  c.Value,
			Unit:   c.Unit,
			Period: c.Period,
			Source: c.Source,
			Anchor: c.Anchor,
		})
	}
	for _, sp := range st.Strips {
		res.Strips = append(
			res.Strips,
			&ntxv1.Strip{Ratio: sp.Ratio, Value: sp.Value, Sector: rangeToProto(sp.Sector), Own: rangeToProto(sp.Own)},
		)
	}
	if q := st.Quote; q != nil {
		res.Quote = &ntxv1.Quote{
			Close:          q.Close,
			Change:         q.Change,
			ChangePercent:  q.ChangePercent,
			BusinessDate:   q.Date,
			BusinessDateBs: q.DateBS,
		}
	}
	for _, fy := range st.FiscalYears {
		res.FiscalYears = append(res.FiscalYears, &ntxv1.FiscalYear{
			Year:              fy.Year.String(),
			Eps:               fy.EPS,
			Profit:            fy.Profit,
			BookValue:         fy.BookValue,
			SharesOutstanding: fy.Shares,
			Bonus:             fy.Bonus,
			Cash:              fy.Cash,
			Rights:            fy.Rights,
			Payout:            fy.Payout,
			Return:            fy.Return,
			Source:            fy.Source,
		})
	}
	for _, g := range st.Growth {
		res.Growth = append(res.Growth, &ntxv1.Growth{Metric: g.Metric, One: g.One, Three: g.Three, Five: g.Five})
	}
	for _, o := range st.Ownership {
		res.Ownership = append(
			res.Ownership,
			&ntxv1.OwnershipObservation{AsOf: o.AsOf, Promoter: o.Promoter, Public: o.Public},
		)
	}
	for _, c := range st.Checks {
		res.Checks = append(res.Checks, &ntxv1.Check{Name: c.Name, Rule: c.Rule, Verdict: c.Verdict, Axis: c.Axis})
	}
	r := st.Rating
	res.Rating = &ntxv1.Rating{Letter: r.Letter, Holds: r.Holds, Evaluated: r.Evaluated, Total: r.Total}
	return res
}

func rangeToProto(r Range) *ntxv1.Range {
	if r.Count == 0 {
		return nil
	}
	count := int32(min(r.Count, 1<<31-1)) //nolint:gosec // clamped to int32 above
	return &ntxv1.Range{Low: r.Low, Median: r.Median, High: r.High, Count: count}
}
