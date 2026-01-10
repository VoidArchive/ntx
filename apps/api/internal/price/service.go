// Package price provides the price service implementation
package price

import (
	"database/sql"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

type PriceService struct {
	ntxv1connect.UnimplementedPriceServiceHandler
	queries *sqlc.Queries
}

func NewPriceService(queries *sqlc.Queries) *PriceService {
	return &PriceService{queries: queries}
}

func priceToProto(p sqlc.Price) *ntxv1.Price {
	return &ntxv1.Price{
		Id:            p.ID,
		CompanyId:     p.CompanyID,
		BusinessDate:  p.BusinessDate,
		Open:          nullFloat64(p.OpenPrice),
		High:          nullFloat64(p.HighPrice),
		Low:           nullFloat64(p.LowPrice),
		Close:         nullFloat64(p.ClosePrice),
		Ltp:           nullFloat64(p.LastTradedPrice),
		PreviousClose: nullFloat64(p.PreviousClose),
		Change:        nullFloat64(p.ChangeAmount),
		ChangePercent: nullFloat64(p.ChangePercent),
		Volume:        nullInt64(p.Volume),
		Turnover:      nullFloat64(p.Turnover),
		Trades:        nullInt32(p.Trades),
	}
}

func pricesToProto(prices []sqlc.Price) []*ntxv1.Price {
	out := make([]*ntxv1.Price, len(prices))
	for i, p := range prices {
		out[i] = priceToProto(p)
	}
	return out
}

func nullFloat64(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

func nullInt64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

func nullInt32(ni sql.NullInt64) *int32 {
	if !ni.Valid {
		return nil
	}
	v := safeInt32(ni.Int64)
	return &v
}

func safeInt32(v int64) int32 {
	const maxInt32 = 1<<31 - 1
	if v > maxInt32 {
		return maxInt32
	}
	return int32(v) //nolint:gosec // bounds checked above
}
