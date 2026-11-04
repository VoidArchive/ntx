package story

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"connectrpc.com/connect"

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
	res := toProto(Build(in))
	addQuarters(res, in)
	return connect.NewResponse(res), nil
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
		sqlc.ListPricesByCompanyParams{CompanyID: c.ID, Limit: 2},
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
	for _, d := range []struct {
		dst  *[]float64
		list func(context.Context, string) ([]float64, error)
	}{
		{&in.Sector.PE, s.queries.ListSectorPE},
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
	return res
}

func rangeToProto(r Range) *ntxv1.Range {
	if r.Count == 0 {
		return nil
	}
	count := int32(min(r.Count, 1<<31-1)) //nolint:gosec // clamped to int32 above
	return &ntxv1.Range{Low: r.Low, Median: r.Median, High: r.High, Count: count}
}
