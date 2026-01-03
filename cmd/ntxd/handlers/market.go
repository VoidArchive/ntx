package handlers

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/market"
	"github.com/voidarchive/ntx/internal/nepse"
)

// Sector names by ID.
var sectorNames = map[int64]string{
	1:  "Commercial Banks",
	2:  "Development Banks",
	3:  "Finance",
	4:  "Microfinance",
	5:  "Life Insurance",
	6:  "Non Life Insurance",
	7:  "Hydro Power",
	8:  "Manufacturing And Processing",
	9:  "Hotels And Tourism",
	10: "Trading",
	11: "Investment",
	12: "Mutual Fund",
	13: "Others",
}

// MarketService implements the MarketService RPC handlers.
type MarketService struct {
	ntxv1connect.UnimplementedMarketServiceHandler
	queries *sqlc.Queries
	market  *market.Market
	nepse   *nepse.Client
}

// NewMarketService creates a new MarketService.
func NewMarketService(queries *sqlc.Queries, mkt *market.Market, client *nepse.Client) *MarketService {
	return &MarketService{
		queries: queries,
		market:  mkt,
		nepse:   client,
	}
}

// GetStatus returns the current market status.
func (s *MarketService) GetStatus(
	ctx context.Context,
	req *connect.Request[ntxv1.GetStatusRequest],
) (*connect.Response[ntxv1.GetStatusResponse], error) {
	isOpen := s.market.IsOpen(ctx)
	now := time.Now().In(market.NPT)

	state := "closed"
	if isOpen {
		state = "open"
	} else if now.Hour() == market.OpenHour-1 {
		state = "pre-open"
	}

	return connect.NewResponse(&ntxv1.GetStatusResponse{
		Status: &ntxv1.MarketStatus{
			IsOpen: isOpen,
			State:  state,
			AsOf:   timestamppb.New(now),
		},
	}), nil
}

// ListIndices returns market indices.
func (s *MarketService) ListIndices(
	ctx context.Context,
	req *connect.Request[ntxv1.ListIndicesRequest],
) (*connect.Response[ntxv1.ListIndicesResponse], error) {
	now := time.Now()

	// Get main NEPSE index
	nepseIdx, err := s.nepse.NepseIndex(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	indices := []*ntxv1.Index{
		{
			Name:          nepseIdx.Name,
			Value:         nepseIdx.Value,
			Change:        nepseIdx.Change,
			PercentChange: nepseIdx.ChangePercent,
			Timestamp:     timestamppb.New(now),
		},
	}

	// Get sub-indices
	subIndices, err := s.nepse.SubIndices(ctx)
	if err != nil {
		// Log but don't fail - main index is enough
		return connect.NewResponse(&ntxv1.ListIndicesResponse{
			Indices: indices,
		}), nil
	}

	for _, idx := range subIndices {
		indices = append(indices, &ntxv1.Index{
			Name:          idx.Name,
			Value:         idx.Value,
			Change:        idx.Change,
			PercentChange: idx.ChangePercent,
			Timestamp:     timestamppb.New(now),
		})
	}

	return connect.NewResponse(&ntxv1.ListIndicesResponse{
		Indices: indices,
	}), nil
}

// ListSectors returns sector summaries.
func (s *MarketService) ListSectors(
	ctx context.Context,
	req *connect.Request[ntxv1.ListSectorsRequest],
) (*connect.Response[ntxv1.ListSectorsResponse], error) {
	rows, err := s.queries.GetSectorSummary(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	sectors := make([]*ntxv1.SectorSummary, len(rows))
	for i, row := range rows {
		name := sectorNames[row.Sector]
		if name == "" {
			name = "Unknown"
		}

		turnover := int64(0)
		switch t := row.Turnover.(type) {
		case int64:
			turnover = t
		case float64:
			turnover = int64(t)
		}

		sectors[i] = &ntxv1.SectorSummary{
			Sector:     ntxv1.Sector(safeInt32(row.Sector)),
			Name:       name,
			StockCount: safeInt32(row.StockCount),
			Turnover:   turnover,
		}
	}

	return connect.NewResponse(&ntxv1.ListSectorsResponse{
		Sectors: sectors,
	}), nil
}
