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
		limit  int64 = 100
		offset int64 = 0
	)

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
