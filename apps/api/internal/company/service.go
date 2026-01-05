// Package company service
package company

import (
	"context"
	"database/sql"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

type CompanyService struct {
	ntxv1connect.UnimplementedCompanyServiceHandler
	queries *sqlc.Queries
}

func NewCompanyService(queries *sqlc.Queries) *CompanyService {
	return &CompanyService{queries: queries}
}

func pointerTo(s string) *string {
	return &s
}

func (s *CompanyService) ListCompanies(
	ctx context.Context,
	req *connect.Request[ntxv1.ListCompaniesRequest],
) (*connect.Response[ntxv1.ListCompaniesResponse], error) {
	hardcodedCompanies := []*ntxv1.Company{
		{
			Id:             1,
			Name:           "Nabil Bank Limited",
			Symbol:         "NABIL",
			Status:         ntxv1.CompanyStatus_COMPANY_STATUS_ACTIVE,
			Email:          pointerTo("info@nabilbank.com"),
			Website:        pointerTo("https://nabilbank.com"),
			InstrumentType: ntxv1.InstrumentType_INSTRUMENT_TYPE_EQUITY,
		},
		{
			Id:             2,
			Name:           "Nepal Investment Mega Bank Limited",
			Symbol:         "NIMB",
			Status:         ntxv1.CompanyStatus_COMPANY_STATUS_ACTIVE,
			Email:          pointerTo("info@nimbbl.com"),
			Website:        pointerTo("https://nimbbl.com"),
			InstrumentType: ntxv1.InstrumentType_INSTRUMENT_TYPE_EQUITY,
		},
		{
			Id:             3,
			Name:           "Chilime Hydropower Company Limited",
			Symbol:         "CHCL",
			Status:         ntxv1.CompanyStatus_COMPANY_STATUS_ACTIVE,
			Email:          pointerTo("info@chilime.com"),
			Website:        pointerTo("https://chilime.com"),
			InstrumentType: ntxv1.InstrumentType_INSTRUMENT_TYPE_BOND,
		},
	}

	return connect.NewResponse(&ntxv1.ListCompaniesResponse{
		Companies: hardcodedCompanies,
	}), nil
}

/*
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
*/

func companiesToProto(companies []sqlc.Company) []*ntxv1.Company {
	out := make([]*ntxv1.Company, len(companies))
	for i, c := range companies {
		out[i] = companyToProto(c)
	}
	return out
}

func companyToProto(c sqlc.Company) *ntxv1.Company {
	return &ntxv1.Company{
		Id:             c.ID,
		Name:           c.Name,
		Symbol:         c.Symbol,
		Status:         statusFromDB(c.Status),
		Email:          nullString(c.Email),
		Website:        nullString(c.Website),
		InstrumentType: instrumentFromDB(c.InstrumentType),
	}
}

func nullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func statusFromDB(s string) ntxv1.CompanyStatus {
	switch s {
	case "A":
		return ntxv1.CompanyStatus_COMPANY_STATUS_ACTIVE
	case "S":
		return ntxv1.CompanyStatus_COMPANY_STATUS_SUSPENDED
	case "D":
		return ntxv1.CompanyStatus_COMPANY_STATUS_DELISTED
	default:
		return ntxv1.CompanyStatus_COMPANY_STATUS_UNSPECIFIED
	}
}

func sectorEnumToDB(sector ntxv1.Sector) (string, bool) {
	switch sector {
	case ntxv1.Sector_SECTOR_COMMERCIAL_BANK:
		return "Commercial Banks", true
	case ntxv1.Sector_SECTOR_DEVELOPMENT_BANK:
		return "Development Banks", true
	case ntxv1.Sector_SECTOR_FINANCE:
		return "Finance", true
	case ntxv1.Sector_SECTOR_HOTEL:
		return "Hotel", true
	case ntxv1.Sector_SECTOR_HYDROPOWER:
		return "Hydropower", true
	case ntxv1.Sector_SECTOR_INVESTMENT:
		return "Investment", true
	case ntxv1.Sector_SECTOR_LIFE_INSURANCE:
		return "Life insurance", true
	case ntxv1.Sector_SECTOR_MANUFACTURING:
		return "Manufacturing", true
	case ntxv1.Sector_SECTOR_MICROFINANCE:
		return "Microfinance", true
	case ntxv1.Sector_SECTOR_NON_LIFE_INSURANCE:
		return "Non Life Insurance", true
	case ntxv1.Sector_SECTOR_TRADING:
		return "Trading", true
	case ntxv1.Sector_SECTOR_OTHERS:
		return "Other", true
	default:
		return "", false
	}
}

func instrumentFromDB(s string) ntxv1.InstrumentType {
	switch s {
	case "EQUITY":
		return ntxv1.InstrumentType_INSTRUMENT_TYPE_EQUITY
	case "BOND":
		return ntxv1.InstrumentType_INSTRUMENT_TYPE_BOND
	default:
		return ntxv1.InstrumentType_INSTRUMENT_TYPE_UNSPECIFIED
	}
}

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
