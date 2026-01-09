package price

import (
	"context"
	"database/sql"
	"errors"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func (s *PriceService) GetPrice(
	ctx context.Context,
	req *connect.Request[ntxv1.GetPriceRequest],
) (*connect.Response[ntxv1.GetPriceResponse], error) {
	if req.Msg.Symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}

	company, err := s.queries.GetCompany(ctx, req.Msg.Symbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("company not found"))
	}

	price, err := s.queries.GetLatestPrice(ctx, company.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return connect.NewResponse(&ntxv1.GetPriceResponse{}), nil
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&ntxv1.GetPriceResponse{
		Price: priceToProto(price),
	}), nil
}

func (s *PriceService) GetPriceHistory(
	ctx context.Context,
	req *connect.Request[ntxv1.GetPriceHistoryRequest],
) (*connect.Response[ntxv1.GetPriceHistoryResponse], error) {
	if req.Msg.Symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}

	company, err := s.queries.GetCompany(ctx, req.Msg.Symbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("company not found"))
	}

	days := int32(365)
	if req.Msg.Days != nil && *req.Msg.Days > 0 {
		days = *req.Msg.Days
	}

	prices, err := s.queries.ListPricesByCompany(ctx, sqlc.ListPricesByCompanyParams{
		CompanyID: company.ID,
		Limit:     int64(days),
		Offset:    0,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&ntxv1.GetPriceHistoryResponse{
		Prices: pricesToProto(prices),
	}), nil
}
