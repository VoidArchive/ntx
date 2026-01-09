package price

import (
	"context"
	"database/sql"
	"errors"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
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
