// Package market serves the session-wide market day.
package market

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
)

type MarketService struct {
	ntxv1connect.UnimplementedMarketServiceHandler
	queries *sqlc.Queries
}

func NewMarketService(queries *sqlc.Queries) *MarketService {
	return &MarketService{queries: queries}
}

func (s *MarketService) GetMarketDay(
	ctx context.Context,
	_ *connect.Request[ntxv1.GetMarketDayRequest],
) (*connect.Response[ntxv1.GetMarketDayResponse], error) {
	day, err := s.queries.GetLatestMarketDay(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return connect.NewResponse(&ntxv1.GetMarketDayResponse{}), nil
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	date, err := time.Parse(time.DateOnly, day.BusinessDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	bs, err := fiscal.BSDate(date)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&ntxv1.GetMarketDayResponse{MarketDay: &ntxv1.MarketDay{
		BusinessDate:       day.BusinessDate,
		BusinessDateBs:     bs,
		IndexClose:         day.IndexClose,
		IndexChange:        day.IndexChange,
		IndexChangePercent: day.IndexChangePercent,
		Turnover:           day.Turnover,
		TradedShares:       day.TradedShares,
		Transactions:       day.Transactions,
		ScripsTraded:       int32(day.ScripsTraded), //nolint:gosec // a few hundred scrips
		MarketCap:          day.MarketCap,
		Advances:           int32(day.Advances),  //nolint:gosec // bounded by company count
		Declines:           int32(day.Declines),  //nolint:gosec // bounded by company count
		Unchanged:          int32(day.Unchanged), //nolint:gosec // bounded by company count
	}}), nil
}
