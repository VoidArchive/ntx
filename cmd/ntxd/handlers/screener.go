package handlers

import (
	"context"
	"math"
	"sort"

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
		data = convertScreenerRows(rows)
	} else {
		rows, err := s.queries.GetScreenerData(ctx)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		data = convertScreenerRows(rows)
	}

	// Apply filters
	filtered := filterData(data, req.Msg)

	// Sort
	sortData(filtered, req.Msg.GetSortBy(), req.Msg.GetSortOrder())

	total := len(filtered)

	// Apply pagination
	offset := int(req.Msg.GetOffset())
	limit := int(clampLimit(int64(req.Msg.GetLimit()), defaultLimit))

	if offset > len(filtered) {
		filtered = nil
	} else {
		end := min(offset+limit, len(filtered))
		filtered = filtered[offset:end]
	}

	// Convert to proto
	results := make([]*ntxv1.ScreenResult, len(filtered))
	for i, row := range filtered {
		results[i] = rowToScreenResult(row)
	}

	return connect.NewResponse(&ntxv1.ScreenResponse{
		Results: results,
		Total:   safeIntToInt32(total),
	}), nil
}

// topMoversParams holds the query functions for fetching top movers.
type topMoversParams struct {
	getAll      func(ctx context.Context, limit int64) ([]sqlc.Price, error)
	getBySector func(ctx context.Context, sector, limit int64) ([]sqlc.Price, error)
}

// fetchTopMovers is a shared helper for ListTopGainers and ListTopLosers.
func (s *ScreenerService) fetchTopMovers(
	ctx context.Context, sector ntxv1.Sector, limit int32, params topMoversParams,
) ([]*ntxv1.Price, error) {
	lim := clampLimit(int64(limit), 10)

	var prices []sqlc.Price
	var err error

	if sector != ntxv1.Sector_SECTOR_UNSPECIFIED {
		prices, err = params.getBySector(ctx, int64(sector), lim)
	} else {
		prices, err = params.getAll(ctx, lim)
	}

	if err != nil {
		return nil, err
	}

	return s.pricesToProto(prices), nil
}

// ListTopGainers returns top gaining stocks.
func (s *ScreenerService) ListTopGainers(
	ctx context.Context,
	req *connect.Request[ntxv1.ListTopGainersRequest],
) (*connect.Response[ntxv1.ListTopGainersResponse], error) {
	stocks, err := s.fetchTopMovers(ctx, req.Msg.GetSector(), req.Msg.GetLimit(), topMoversParams{
		getAll: s.queries.GetTopGainers,
		getBySector: func(ctx context.Context, sector, limit int64) ([]sqlc.Price, error) {
			return s.queries.GetTopGainersBySector(ctx, sqlc.GetTopGainersBySectorParams{Sector: sector, Limit: limit})
		},
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.ListTopGainersResponse{Stocks: stocks}), nil
}

// ListTopLosers returns top losing stocks.
func (s *ScreenerService) ListTopLosers(
	ctx context.Context,
	req *connect.Request[ntxv1.ListTopLosersRequest],
) (*connect.Response[ntxv1.ListTopLosersResponse], error) {
	stocks, err := s.fetchTopMovers(ctx, req.Msg.GetSector(), req.Msg.GetLimit(), topMoversParams{
		getAll: s.queries.GetTopLosers,
		getBySector: func(ctx context.Context, sector, limit int64) ([]sqlc.Price, error) {
			return s.queries.GetTopLosersBySector(ctx, sqlc.GetTopLosersBySectorParams{Sector: sector, Limit: limit})
		},
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.ListTopLosersResponse{Stocks: stocks}), nil
}

func (s *ScreenerService) pricesToProto(prices []sqlc.Price) []*ntxv1.Price {
	result := make([]*ntxv1.Price, len(prices))
	for i, p := range prices {
		result[i] = priceToProto(p)
	}
	return result
}

// screenerRow is a unified type for screener data.
type screenerRow struct {
	Symbol            string
	Name              string
	Sector            int64
	Description       string
	LogoURL           string
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

// screenerDataRow is an interface for sqlc row types with screener data.
type screenerDataRow interface {
	sqlc.GetScreenerDataRow | sqlc.GetScreenerDataBySectorRow
}

// toScreenerRow converts any screener data row type to a unified screenerRow.
func toScreenerRow[T screenerDataRow](r T) screenerRow {
	// Use type assertion via any to access fields (both types have identical fields)
	switch v := any(r).(type) {
	case sqlc.GetScreenerDataRow:
		return screenerRow{
			Symbol: v.Symbol, Name: v.Name, Sector: v.Sector, Description: v.Description,
			LogoURL: v.LogoUrl, PriceDate: v.PriceDate, Open: v.Open, High: v.High,
			Low: v.Low, Close: v.Close, PreviousClose: nullFloat64(v.PreviousClose),
			Volume: v.Volume, Turnover: nullInt64(v.Turnover),
			Week52High: nullFloat64(v.Week52High), Week52Low: nullFloat64(v.Week52Low),
			PE: nullFloat64(v.Pe), PB: nullFloat64(v.Pb), EPS: nullFloat64(v.Eps),
			BookValue: nullFloat64(v.BookValue), MarketCap: nullFloat64(v.MarketCap),
			DividendYield: nullFloat64(v.DividendYield), ROE: nullFloat64(v.Roe),
			SharesOutstanding: nullInt64(v.SharesOutstanding),
		}
	case sqlc.GetScreenerDataBySectorRow:
		return screenerRow{
			Symbol: v.Symbol, Name: v.Name, Sector: v.Sector, Description: v.Description,
			LogoURL: v.LogoUrl, PriceDate: v.PriceDate, Open: v.Open, High: v.High,
			Low: v.Low, Close: v.Close, PreviousClose: nullFloat64(v.PreviousClose),
			Volume: v.Volume, Turnover: nullInt64(v.Turnover),
			Week52High: nullFloat64(v.Week52High), Week52Low: nullFloat64(v.Week52Low),
			PE: nullFloat64(v.Pe), PB: nullFloat64(v.Pb), EPS: nullFloat64(v.Eps),
			BookValue: nullFloat64(v.BookValue), MarketCap: nullFloat64(v.MarketCap),
			DividendYield: nullFloat64(v.DividendYield), ROE: nullFloat64(v.Roe),
			SharesOutstanding: nullInt64(v.SharesOutstanding),
		}
	}
	return screenerRow{}
}

// convertScreenerRows converts a slice of screener data rows to unified screenerRows.
func convertScreenerRows[T screenerDataRow](rows []T) []screenerRow {
	out := make([]screenerRow, len(rows))
	for i, r := range rows {
		out[i] = toScreenerRow(r)
	}
	return out
}

type filterFunc func(row screenerRow) bool

// rangeFilter creates a filter that checks if value is within optional min/max bounds.
func rangeFilter(getValue func(screenerRow) float64, minVal, maxVal *float64) filterFunc {
	return func(r screenerRow) bool {
		v := getValue(r)
		return (minVal == nil || v >= *minVal) && (maxVal == nil || v <= *maxVal)
	}
}

// addRangeFilter appends a range filter if min or max is set.
func addRangeFilter(
	filters *[]filterFunc, getValue func(screenerRow) float64, minVal, maxVal *float64,
) {
	if minVal != nil || maxVal != nil {
		*filters = append(*filters, rangeFilter(getValue, minVal, maxVal))
	}
}

// buildFilters constructs all filter functions from a screen request.
func buildFilters(req *ntxv1.ScreenRequest) []filterFunc {
	var filters []filterFunc

	addRangeFilter(&filters, func(r screenerRow) float64 { return r.Close }, req.MinPrice, req.MaxPrice)
	addRangeFilter(&filters, func(r screenerRow) float64 { return r.PE }, req.MinPe, req.MaxPe)
	addRangeFilter(&filters, func(r screenerRow) float64 { return r.PB }, req.MinPb, req.MaxPb)
	addRangeFilter(&filters, func(r screenerRow) float64 { return r.percentChange() }, req.MinChange, req.MaxChange)
	addRangeFilter(&filters, func(r screenerRow) float64 { return r.MarketCap }, req.MinMarketCap, req.MaxMarketCap)

	if req.MinVolume != nil {
		filters = append(filters, func(r screenerRow) bool { return r.Volume >= *req.MinVolume })
	}
	if req.Near_52WHigh {
		filters = append(filters, func(r screenerRow) bool { return r.Week52High <= 0 || r.Close >= r.Week52High*0.95 })
	}
	if req.Near_52WLow {
		filters = append(filters, func(r screenerRow) bool { return r.Week52Low <= 0 || r.Close <= r.Week52Low*1.05 })
	}

	return filters
}

// passesFilters checks if a row passes all filter functions.
func passesFilters(row screenerRow, filters []filterFunc) bool {
	for _, f := range filters {
		if !f(row) {
			return false
		}
	}
	return true
}

func filterData(data []screenerRow, req *ntxv1.ScreenRequest) []screenerRow {
	filters := buildFilters(req)

	var filtered []screenerRow
	for _, row := range data {
		if row.Close == 0 {
			continue
		}
		if passesFilters(row, filters) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

// sortableRow pairs a screenerRow with pre-calculated values for sorting.
type sortableRow struct {
	row           screenerRow
	percentChange float64
}

func sortData(data []screenerRow, sortBy ntxv1.SortBy, order ntxv1.SortOrder) {
	desc := order == ntxv1.SortOrder_SORT_ORDER_DESC

	// Wrap rows with pre-calculated percent change
	rows := make([]sortableRow, len(data))
	for i := range data {
		rows[i] = sortableRow{
			row:           data[i],
			percentChange: data[i].percentChange(),
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		var less bool
		switch sortBy {
		case ntxv1.SortBy_SORT_BY_SYMBOL:
			less = rows[i].row.Symbol < rows[j].row.Symbol
		case ntxv1.SortBy_SORT_BY_PRICE:
			less = rows[i].row.Close < rows[j].row.Close
		case ntxv1.SortBy_SORT_BY_CHANGE:
			less = rows[i].percentChange < rows[j].percentChange
		case ntxv1.SortBy_SORT_BY_VOLUME:
			less = rows[i].row.Volume < rows[j].row.Volume
		case ntxv1.SortBy_SORT_BY_TURNOVER:
			less = rows[i].row.Turnover < rows[j].row.Turnover
		case ntxv1.SortBy_SORT_BY_MARKET_CAP:
			less = rows[i].row.MarketCap < rows[j].row.MarketCap
		case ntxv1.SortBy_SORT_BY_PE:
			// Handle 0/NaN PE values - push them to the end
			iPE := rows[i].row.PE
			jPE := rows[j].row.PE
			if iPE == 0 || math.IsNaN(iPE) {
				less = false
			} else if jPE == 0 || math.IsNaN(jPE) {
				less = true
			} else {
				less = iPE < jPE
			}
		default:
			less = rows[i].row.Symbol < rows[j].row.Symbol
		}
		if desc {
			return !less
		}
		return less
	})

	// Copy sorted rows back
	for i := range rows {
		data[i] = rows[i].row
	}
}

func rowToScreenResult(row screenerRow) *ntxv1.ScreenResult {
	date := parseDate(row.PriceDate)

	return &ntxv1.ScreenResult{
		Company: &ntxv1.Company{
			Symbol:      row.Symbol,
			Name:        row.Name,
			Sector:      ntxv1.Sector(safeInt32(row.Sector)),
			Description: row.Description,
			LogoUrl:     row.LogoURL,
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
