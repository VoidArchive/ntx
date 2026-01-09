package company

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func (s *CompanyService) GetCompany(
	ctx context.Context,
	req *connect.Request[ntxv1.GetCompanyRequest],
) (*connect.Response[ntxv1.GetCompanyResponse], error) {
	if req.Msg.Symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}

	company, err := s.queries.GetCompany(ctx, req.Msg.Symbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&ntxv1.GetCompanyResponse{
		Company: companyToProto(company),
	}), nil
}

func (s *CompanyService) ListCompanies(
	ctx context.Context,
	req *connect.Request[ntxv1.ListCompaniesRequest],
) (*connect.Response[ntxv1.ListCompaniesResponse], error) {
	sector := req.Msg.GetSector()
	query := req.Msg.GetQuery()

	const (
		defaultLimit int64 = 100
		maxLimit     int64 = 500
		maxQueryLen  int   = 100
	)

	if len(query) > maxQueryLen {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("query too long"))
	}

	limit := defaultLimit
	if req.Msg.Limit != nil {
		if *req.Msg.Limit <= 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("limit must be positive"))
		}
		limit = min(int64(*req.Msg.Limit), maxLimit)
	}

	var offset int64
	if req.Msg.Offset != nil {
		if *req.Msg.Offset < 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("offset must be non-negative"))
		}
		offset = int64(*req.Msg.Offset)
	}

	pattern := "%"
	if query != "" {
		pattern = "%" + query + "%"
	}

	// Sector filter (with optional query)
	if sector != ntxv1.Sector_SECTOR_UNSPECIFIED {
		sectorStr, ok := sectorEnumToDB(sector)
		if !ok {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid sector"))
		}
		companies, err := s.queries.ListCompaniesBySector(ctx, sqlc.ListCompaniesBySectorParams{
			Sector: sectorStr,
			Symbol: pattern,
			Name:   pattern,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		return connect.NewResponse(&ntxv1.ListCompaniesResponse{
			Companies: companiesToProto(companies),
		}), nil
	}

	// No sector: query-only search, otherwise list all
	if query != "" {
		companies, err := s.queries.SearchCompanies(ctx, sqlc.SearchCompaniesParams{
			Symbol: pattern,
			Name:   pattern,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		return connect.NewResponse(&ntxv1.ListCompaniesResponse{
			Companies: companiesToProto(companies),
		}), nil
	}

	companies, err := s.queries.ListCompanies(ctx, sqlc.ListCompaniesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&ntxv1.ListCompaniesResponse{
		Companies: companiesToProto(companies),
	}), nil
}

func (s *CompanyService) GetFundamentals(
	ctx context.Context,
	req *connect.Request[ntxv1.GetFundamentalsRequest],
) (*connect.Response[ntxv1.GetFundamentalsResponse], error) {
	if req.Msg.Symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}

	company, err := s.queries.GetCompany(ctx, req.Msg.Symbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("company not found"))
	}

	fundamentals, err := s.queries.ListFundamentalsByCompany(ctx, company.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if len(fundamentals) == 0 {
		return connect.NewResponse(&ntxv1.GetFundamentalsResponse{}), nil
	}

	return connect.NewResponse(&ntxv1.GetFundamentalsResponse{
		Latest:  fundamentalToProto(fundamentals[0]),
		History: fundamentalsToProto(fundamentals),
	}), nil
}

func (s *CompanyService) GetSectorStats(
	ctx context.Context,
	req *connect.Request[ntxv1.GetSectorStatsRequest],
) (*connect.Response[ntxv1.GetSectorStatsResponse], error) {
	sector := req.Msg.GetSector()
	if sector == ntxv1.Sector_SECTOR_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("sector is required"))
	}

	sectorStr, ok := sectorEnumToDB(sector)
	if !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid sector"))
	}

	stats, err := s.queries.GetSectorStats(ctx, sectorStr)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&ntxv1.GetSectorStatsResponse{
		Stats: &ntxv1.SectorStats{
			Sector:       sector,
			CompanyCount: int32(stats.CompanyCount),
			AvgEps:       nullFloat64(stats.AvgEps),
			AvgPeRatio:   nullFloat64(stats.AvgPeRatio),
			AvgBookValue: nullFloat64(stats.AvgBookValue),
		},
	}), nil
}
