package handlers

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// PriceService implements the PriceService RPC handlers.
type PriceService struct {
	ntxv1connect.UnimplementedPriceServiceHandler
	queries *sqlc.Queries
}

// NewPriceService creates a new PriceService.
func NewPriceService(queries *sqlc.Queries) *PriceService {
	return &PriceService{queries: queries}
}

// GetPrice returns the latest price for a symbol.
func (s *PriceService) GetPrice(
	ctx context.Context,
	req *connect.Request[ntxv1.GetPriceRequest],
) (*connect.Response[ntxv1.GetPriceResponse], error) {
	symbol := strings.ToUpper(req.Msg.GetSymbol())
	if symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errSymbolRequired)
	}

	price, err := s.queries.GetLatestPrice(ctx, symbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&ntxv1.GetPriceResponse{
		Price: priceToProto(price),
	}), nil
}

// ListCandles returns OHLCV candles for a date range.
func (s *PriceService) ListCandles(
	ctx context.Context,
	req *connect.Request[ntxv1.ListCandlesRequest],
) (*connect.Response[ntxv1.ListCandlesResponse], error) {
	symbol := strings.ToUpper(req.Msg.GetSymbol())
	if symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errSymbolRequired)
	}

	from := req.Msg.GetFrom()
	to := req.Msg.GetTo()

	// Default to last 30 days if not specified
	now := time.Now()
	fromDate := now.AddDate(0, -1, 0).Format("2006-01-02")
	toDate := now.Format("2006-01-02")

	if from != nil {
		fromDate = from.AsTime().Format("2006-01-02")
	}
	if to != nil {
		toDate = to.AsTime().Format("2006-01-02")
	}

	prices, err := s.queries.GetPriceHistory(ctx, sqlc.GetPriceHistoryParams{
		Symbol: symbol,
		Date:   fromDate,
		Date_2: toDate,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// TODO: Implement weekly/monthly aggregation based on TimeFrame
	// For now, we only support daily candles

	candles := make([]*ntxv1.OHLCV, len(prices))
	for i, p := range prices {
		candles[i] = ohlcvToProto(p)
	}

	return connect.NewResponse(&ntxv1.ListCandlesResponse{
		Candles: candles,
	}), nil
}

func priceToProto(p sqlc.Price) *ntxv1.Price {
	change := 0.0
	percentChange := 0.0
	prevClose := nullFloat64(p.PreviousClose)

	if prevClose > 0 {
		change = p.Close - prevClose
		percentChange = (change / prevClose) * 100
	}

	date, _ := time.Parse("2006-01-02", p.Date)

	return &ntxv1.Price{
		Symbol:        p.Symbol,
		Ltp:           p.Close,
		Change:        change,
		PercentChange: percentChange,
		Open:          p.Open,
		High:          p.High,
		Low:           p.Low,
		PreviousClose: prevClose,
		Volume:        p.Volume,
		Turnover:      nullInt64(p.Turnover),
		Week_52High:   nullFloat64(p.Week52High),
		Week_52Low:    nullFloat64(p.Week52Low),
		Timestamp:     timestamppb.New(date),
	}
}

func ohlcvToProto(p sqlc.Price) *ntxv1.OHLCV {
	date, _ := time.Parse("2006-01-02", p.Date)
	return &ntxv1.OHLCV{
		Date:     timestamppb.New(date),
		Open:     p.Open,
		High:     p.High,
		Low:      p.Low,
		Close:    p.Close,
		Volume:   p.Volume,
		Turnover: nullInt64(p.Turnover),
	}
}
