package handlers

import (
	"context"
	"math"
	"sort"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// ScreenerService implements the ScreenerService RPC handlers.
type ScreenerService struct {
	ntxv1connect.UnimplementedScreenerServiceHandler
	queries *sqlc.Queries
}

// NewScreenerService creates a new ScreenerService.
func NewScreenerService(queries *sqlc.Queries) *ScreenerService {
	return &ScreenerService{queries: queries}
}

// Screen filters stocks based on criteria.
func (s *ScreenerService) Screen(
	ctx context.Context,
	req *connect.Request[ntxv1.ScreenRequest],
) (*connect.Response[ntxv1.ScreenResponse], error) {
	sector := req.Msg.GetSector()

	var data []screenerRow

	if sector != ntxv1.Sector_SECTOR_UNSPECIFIED {
		rows, err := s.queries.GetScreenerDataBySector(ctx, int64(sector))
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		data = convertBySectorRows(rows)
	} else {
		rows, err := s.queries.GetScreenerData(ctx)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		data = convertRows(rows)
	}

	// Apply filters
	filtered := filterData(data, req.Msg)

	// Sort
	sortData(filtered, req.Msg.GetSortBy(), req.Msg.GetSortOrder())

	total := len(filtered)

	// Apply pagination
	offset := int(req.Msg.GetOffset())
	limit := int(req.Msg.GetLimit())
	if limit <= 0 {
		limit = 50
	}

	if offset > len(filtered) {
		filtered = nil
	} else {
		end := offset + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[offset:end]
	}

	// Convert to proto
	results := make([]*ntxv1.ScreenResult, len(filtered))
	for i, row := range filtered {
		results[i] = rowToScreenResult(row)
	}

	return connect.NewResponse(&ntxv1.ScreenResponse{
		Results: results,
		Total:   int32(total),
	}), nil
}

// ListTopGainers returns top gaining stocks.
func (s *ScreenerService) ListTopGainers(
	ctx context.Context,
	req *connect.Request[ntxv1.ListTopGainersRequest],
) (*connect.Response[ntxv1.ListTopGainersResponse], error) {
	limit := int64(req.Msg.GetLimit())
	if limit <= 0 {
		limit = 10
	}

	sector := req.Msg.GetSector()

	var prices []sqlc.Price
	var err error

	if sector != ntxv1.Sector_SECTOR_UNSPECIFIED {
		prices, err = s.queries.GetTopGainersBySector(ctx, sqlc.GetTopGainersBySectorParams{
			Sector: int64(sector),
			Limit:  limit,
		})
	} else {
		prices, err = s.queries.GetTopGainers(ctx, limit)
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	stocks := make([]*ntxv1.Price, len(prices))
	for i, p := range prices {
		stocks[i] = priceToProto(p)
	}

	return connect.NewResponse(&ntxv1.ListTopGainersResponse{
		Stocks: stocks,
	}), nil
}

// ListTopLosers returns top losing stocks.
func (s *ScreenerService) ListTopLosers(
	ctx context.Context,
	req *connect.Request[ntxv1.ListTopLosersRequest],
) (*connect.Response[ntxv1.ListTopLosersResponse], error) {
	limit := int64(req.Msg.GetLimit())
	if limit <= 0 {
		limit = 10
	}

	sector := req.Msg.GetSector()

	var prices []sqlc.Price
	var err error

	if sector != ntxv1.Sector_SECTOR_UNSPECIFIED {
		prices, err = s.queries.GetTopLosersBySector(ctx, sqlc.GetTopLosersBySectorParams{
			Sector: int64(sector),
			Limit:  limit,
		})
	} else {
		prices, err = s.queries.GetTopLosers(ctx, limit)
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	stocks := make([]*ntxv1.Price, len(prices))
	for i, p := range prices {
		stocks[i] = priceToProto(p)
	}

	return connect.NewResponse(&ntxv1.ListTopLosersResponse{
		Stocks: stocks,
	}), nil
}

// screenerRow is a unified type for screener data.
type screenerRow struct {
	Symbol            string
	Name              string
	Sector            int64
	Description       string
	LogoUrl           string
	PriceDate         string
	Open              float64
	High              float64
	Low               float64
	Close             float64
	PreviousClose     float64
	Volume            int64
	Turnover          int64
	Week52High        float64
	Week52Low         float64
	PE                float64
	PB                float64
	EPS               float64
	BookValue         float64
	MarketCap         float64
	DividendYield     float64
	ROE               float64
	SharesOutstanding int64
}

func (r screenerRow) change() float64 {
	if r.PreviousClose > 0 {
		return r.Close - r.PreviousClose
	}
	return 0
}

func (r screenerRow) percentChange() float64 {
	if r.PreviousClose > 0 {
		return ((r.Close - r.PreviousClose) / r.PreviousClose) * 100
	}
	return 0
}

func convertRows(rows []sqlc.GetScreenerDataRow) []screenerRow {
	out := make([]screenerRow, len(rows))
	for i, r := range rows {
		out[i] = screenerRow{
			Symbol:            r.Symbol,
			Name:              r.Name,
			Sector:            r.Sector,
			Description:       r.Description,
			LogoUrl:           r.LogoUrl,
			PriceDate:         r.PriceDate,
			Open:              r.Open,
			High:              r.High,
			Low:               r.Low,
			Close:             r.Close,
			PreviousClose:     nullFloat64(r.PreviousClose),
			Volume:            r.Volume,
			Turnover:          nullInt64(r.Turnover),
			Week52High:        nullFloat64(r.Week52High),
			Week52Low:         nullFloat64(r.Week52Low),
			PE:                nullFloat64(r.Pe),
			PB:                nullFloat64(r.Pb),
			EPS:               nullFloat64(r.Eps),
			BookValue:         nullFloat64(r.BookValue),
			MarketCap:         nullFloat64(r.MarketCap),
			DividendYield:     nullFloat64(r.DividendYield),
			ROE:               nullFloat64(r.Roe),
			SharesOutstanding: nullInt64(r.SharesOutstanding),
		}
	}
	return out
}

func convertBySectorRows(rows []sqlc.GetScreenerDataBySectorRow) []screenerRow {
	out := make([]screenerRow, len(rows))
	for i, r := range rows {
		out[i] = screenerRow{
			Symbol:            r.Symbol,
			Name:              r.Name,
			Sector:            r.Sector,
			Description:       r.Description,
			LogoUrl:           r.LogoUrl,
			PriceDate:         r.PriceDate,
			Open:              r.Open,
			High:              r.High,
			Low:               r.Low,
			Close:             r.Close,
			PreviousClose:     nullFloat64(r.PreviousClose),
			Volume:            r.Volume,
			Turnover:          nullInt64(r.Turnover),
			Week52High:        nullFloat64(r.Week52High),
			Week52Low:         nullFloat64(r.Week52Low),
			PE:                nullFloat64(r.Pe),
			PB:                nullFloat64(r.Pb),
			EPS:               nullFloat64(r.Eps),
			BookValue:         nullFloat64(r.BookValue),
			MarketCap:         nullFloat64(r.MarketCap),
			DividendYield:     nullFloat64(r.DividendYield),
			ROE:               nullFloat64(r.Roe),
			SharesOutstanding: nullInt64(r.SharesOutstanding),
		}
	}
	return out
}

func filterData(data []screenerRow, req *ntxv1.ScreenRequest) []screenerRow {
	var filtered []screenerRow

	for _, row := range data {
		// Skip rows without price data
		if row.Close == 0 {
			continue
		}

		// Price filters
		if req.MinPrice != nil && row.Close < *req.MinPrice {
			continue
		}
		if req.MaxPrice != nil && row.Close > *req.MaxPrice {
			continue
		}

		// PE filters
		if req.MinPe != nil && row.PE < *req.MinPe {
			continue
		}
		if req.MaxPe != nil && row.PE > *req.MaxPe {
			continue
		}

		// PB filters
		if req.MinPb != nil && row.PB < *req.MinPb {
			continue
		}
		if req.MaxPb != nil && row.PB > *req.MaxPb {
			continue
		}

		// Change filters
		pctChange := row.percentChange()
		if req.MinChange != nil && pctChange < *req.MinChange {
			continue
		}
		if req.MaxChange != nil && pctChange > *req.MaxChange {
			continue
		}

		// Market cap filters
		if req.MinMarketCap != nil && row.MarketCap < *req.MinMarketCap {
			continue
		}
		if req.MaxMarketCap != nil && row.MarketCap > *req.MaxMarketCap {
			continue
		}

		// Volume filter
		if req.MinVolume != nil && row.Volume < *req.MinVolume {
			continue
		}

		// 52-week high/low proximity filters
		if req.Near_52WHigh && row.Week52High > 0 {
			threshold := row.Week52High * 0.95
			if row.Close < threshold {
				continue
			}
		}
		if req.Near_52WLow && row.Week52Low > 0 {
			threshold := row.Week52Low * 1.05
			if row.Close > threshold {
				continue
			}
		}

		filtered = append(filtered, row)
	}

	return filtered
}

func sortData(data []screenerRow, sortBy ntxv1.SortBy, order ntxv1.SortOrder) {
	desc := order == ntxv1.SortOrder_SORT_ORDER_DESC

	sort.Slice(data, func(i, j int) bool {
		var less bool
		switch sortBy {
		case ntxv1.SortBy_SORT_BY_SYMBOL:
			less = data[i].Symbol < data[j].Symbol
		case ntxv1.SortBy_SORT_BY_PRICE:
			less = data[i].Close < data[j].Close
		case ntxv1.SortBy_SORT_BY_CHANGE:
			less = data[i].percentChange() < data[j].percentChange()
		case ntxv1.SortBy_SORT_BY_VOLUME:
			less = data[i].Volume < data[j].Volume
		case ntxv1.SortBy_SORT_BY_TURNOVER:
			less = data[i].Turnover < data[j].Turnover
		case ntxv1.SortBy_SORT_BY_MARKET_CAP:
			less = data[i].MarketCap < data[j].MarketCap
		case ntxv1.SortBy_SORT_BY_PE:
			// Handle 0/NaN PE values
			iPE := data[i].PE
			jPE := data[j].PE
			if iPE == 0 || math.IsNaN(iPE) {
				less = false
			} else if jPE == 0 || math.IsNaN(jPE) {
				less = true
			} else {
				less = iPE < jPE
			}
		default:
			less = data[i].Symbol < data[j].Symbol
		}
		if desc {
			return !less
		}
		return less
	})
}

func rowToScreenResult(row screenerRow) *ntxv1.ScreenResult {
	date, _ := time.Parse("2006-01-02", row.PriceDate)

	return &ntxv1.ScreenResult{
		Company: &ntxv1.Company{
			Symbol:      row.Symbol,
			Name:        row.Name,
			Sector:      ntxv1.Sector(row.Sector),
			Description: row.Description,
			LogoUrl:     row.LogoUrl,
		},
		Price: &ntxv1.Price{
			Symbol:        row.Symbol,
			Ltp:           row.Close,
			Change:        row.change(),
			PercentChange: row.percentChange(),
			Open:          row.Open,
			High:          row.High,
			Low:           row.Low,
			PreviousClose: row.PreviousClose,
			Volume:        row.Volume,
			Turnover:      row.Turnover,
			Week_52High:   row.Week52High,
			Week_52Low:    row.Week52Low,
			Timestamp:     timestamppb.New(date),
		},
		Fundamentals: &ntxv1.Fundamentals{
			Symbol:            row.Symbol,
			Pe:                row.PE,
			Pb:                row.PB,
			Eps:               row.EPS,
			BookValue:         row.BookValue,
			MarketCap:         row.MarketCap,
			DividendYield:     row.DividendYield,
			Roe:               row.ROE,
			SharesOutstanding: row.SharesOutstanding,
		},
	}
}
