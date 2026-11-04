// Package sector serves the home page's sector overview and filed-recently list.
package sector

import (
	"context"
	"fmt"
	"slices"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/company"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/story"
)

// shortcut: the list is capped while the first backfill makes every company "seen" today; page it once filings carry real dates.
const filedRecentlyLimit = 20

const moverCount = 5

type Service struct {
	ntxv1connect.UnimplementedSectorServiceHandler
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) GetSectorOverview(
	ctx context.Context,
	_ *connect.Request[ntxv1.GetSectorOverviewRequest],
) (*connect.Response[ntxv1.GetSectorOverviewResponse], error) {
	res := &ntxv1.GetSectorOverviewResponse{}
	counts, err := s.queries.ListSectorCounts(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	changes, err := s.queries.ListDayChanges(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	daily := map[string][]float64{}
	for _, c := range changes {
		daily[c.Sector] = append(daily[c.Sector], c.ChangePercent)
	}
	for _, c := range counts {
		row, err := s.summary(ctx, c)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		row.DayChange = median(story.Median(daily[c.Sector]))
		res.Sectors = append(res.Sectors, row)
	}
	res.Gainers, res.Losers = movers(changes)
	if res.FiledRecently, err = s.filedRecently(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(res), nil
}

func (s *Service) summary(ctx context.Context, c sqlc.ListSectorCountsRow) (*ntxv1.SectorSummary, error) {
	row := &ntxv1.SectorSummary{
		Name:         c.Sector,
		Sector:       company.SectorFromDB(c.Sector),
		CompanyCount: int32(c.CompanyCount), //nolint:gosec // a few hundred companies
	}
	pe, err := s.queries.ListSectorPE(ctx, c.Sector)
	if err != nil {
		return nil, fmt.Errorf("sector pe: %w", err)
	}
	r := story.Median(story.Profitable(pe))
	row.MedianPe, row.PeCount = median(r), int32(r.Count) //nolint:gosec // bounded by company count

	pb, err := s.queries.ListSectorPB(ctx, c.Sector)
	if err != nil {
		return nil, fmt.Errorf("sector pb: %w", err)
	}
	row.MedianPb = median(story.Median(pb))

	eps, err := s.queries.ListSectorEPS(ctx, c.Sector)
	if err != nil {
		return nil, fmt.Errorf("sector eps: %w", err)
	}
	row.QuarterCount = int32(len(eps)) //nolint:gosec // bounded by company count
	if len(eps) > 0 {
		share := float64(len(story.Profitable(eps))) / float64(len(eps))
		row.ProfitableShare = &share
	}

	yoy, err := s.queries.GetSectorProfitYoY(ctx, c.Sector)
	if err != nil {
		return nil, fmt.Errorf("sector profit: %w", err)
	}
	if yoy.PreviousProfit != 0 {
		change := (yoy.CurrentProfit - yoy.PreviousProfit) / abs(yoy.PreviousProfit) * 100
		row.ProfitChangeYoy = &change
	}
	return row, nil
}

func (s *Service) filedRecently(ctx context.Context) ([]*ntxv1.RecentFiling, error) {
	rows, err := s.queries.ListFiledRecently(ctx, filedRecentlyLimit)
	if err != nil {
		return nil, fmt.Errorf("filed recently: %w", err)
	}
	out := make([]*ntxv1.RecentFiling, 0, len(rows))
	for _, r := range rows {
		fundamentals, err := s.queries.ListFundamentalsByCompany(ctx, r.ID)
		if err != nil {
			return nil, fmt.Errorf("fundamentals %s: %w", r.Symbol, err)
		}
		// Earnings claims read only the company and its filings, so no prices or sector rows are loaded.
		st := story.Build(story.Input{
			Company:      sqlc.Company{ID: r.ID, Symbol: r.Symbol, Name: r.Name, Sector: r.Sector},
			Fundamentals: fundamentals,
		})
		f := &ntxv1.RecentFiling{Symbol: r.Symbol, Name: r.Name, Date: r.FiledOn, Published: r.Published == 1}
		for _, c := range st.Claims {
			if c.Kind == "earnings" {
				f.Claim = c.Text
				break
			}
		}
		out = append(out, f)
	}
	return out, nil
}

// movers takes changes sorted best first; an unchanged company is neither a gainer nor a loser.
func movers(changes []sqlc.ListDayChangesRow) (gainers, losers []*ntxv1.Mover) {
	for _, c := range changes {
		if c.ChangePercent > 0 && len(gainers) < moverCount {
			gainers = append(gainers, mover(c))
		}
	}
	for _, c := range slices.Backward(changes) {
		if c.ChangePercent < 0 && len(losers) < moverCount {
			losers = append(losers, mover(c))
		}
	}
	return gainers, losers
}

func mover(c sqlc.ListDayChangesRow) *ntxv1.Mover {
	return &ntxv1.Mover{Symbol: c.Symbol, Name: c.Name, Close: c.Close, ChangePercent: c.ChangePercent}
}

func median(r story.Range) *float64 {
	if r.Count == 0 {
		return nil
	}
	return &r.Median
}

func abs(v float64) float64 { return max(v, -v) }
