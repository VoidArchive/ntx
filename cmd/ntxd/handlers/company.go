// Package handlers implements the RPC request and response from the proto file.
package handlers

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// CompanyService implements the CompanyService RPC handlers.
type CompanyService struct {
	ntxv1connect.UnimplementedCompanyServiceHandler
	queries *sqlc.Queries
}

// NewCompanyService creates a new CompanyService.
func NewCompanyService(queries *sqlc.Queries) *CompanyService {
	return &CompanyService{queries: queries}
}

// ListCompanies returns all companies, optionally filtered by sector or search query.
func (s *CompanyService) ListCompanies(
	ctx context.Context,
	req *connect.Request[ntxv1.ListCompaniesRequest],
) (*connect.Response[ntxv1.ListCompaniesResponse], error) {
	var companies []sqlc.Company
	var err error

	sector := req.Msg.GetSector()
	query := req.Msg.GetQuery()

	if query != "" {
		pattern := "%" + query + "%"
		companies, err = s.queries.SearchCompanies(ctx, sqlc.SearchCompaniesParams{
			Symbol: pattern,
			Name:   pattern,
			Limit:  100,
		})
	} else if sector != ntxv1.Sector_SECTOR_UNSPECIFIED {
		companies, err = s.queries.ListCompaniesBySector(ctx, int64(sector))
	} else {
		companies, err = s.queries.ListCompanies(ctx)
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	result := make([]*ntxv1.Company, len(companies))
	for i, c := range companies {
		result[i] = companyToProto(c)
	}

	return connect.NewResponse(&ntxv1.ListCompaniesResponse{
		Companies: result,
	}), nil
}

// GetCompany returns a single company by symbol.
func (s *CompanyService) GetCompany(
	ctx context.Context,
	req *connect.Request[ntxv1.GetCompanyRequest],
) (*connect.Response[ntxv1.GetCompanyResponse], error) {
	symbol := strings.ToUpper(req.Msg.GetSymbol())
	if symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errSymbolRequired)
	}

	company, err := s.queries.GetCompany(ctx, symbol)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&ntxv1.GetCompanyResponse{
		Company: companyToProto(company),
	}), nil
}

// GetFundamentals returns fundamentals for a company.
func (s *CompanyService) GetFundamentals(
	ctx context.Context,
	req *connect.Request[ntxv1.GetFundamentalsRequest],
) (*connect.Response[ntxv1.GetFundamentalsResponse], error) {
	symbol := strings.ToUpper(req.Msg.GetSymbol())
	if symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errSymbolRequired)
	}

	fundamentals, err := s.queries.GetFundamentals(ctx, symbol)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&ntxv1.GetFundamentalsResponse{
		Fundamentals: fundamentalsToProto(fundamentals),
	}), nil
}

// ListReports returns financial reports for a company.
func (s *CompanyService) ListReports(
	ctx context.Context,
	req *connect.Request[ntxv1.ListReportsRequest],
) (*connect.Response[ntxv1.ListReportsResponse], error) {
	symbol := strings.ToUpper(req.Msg.GetSymbol())
	if symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errSymbolRequired)
	}

	limit := clampLimit(int64(req.Msg.GetLimit()), 20)

	reportType := req.Msg.GetType()

	var reports []sqlc.Report
	var err error

	if reportType != ntxv1.ReportType_REPORT_TYPE_UNSPECIFIED {
		reports, err = s.queries.GetReportsByType(ctx, sqlc.GetReportsByTypeParams{
			Symbol: symbol,
			Type:   int64(reportType),
			Limit:  limit,
		})
	} else {
		reports, err = s.queries.GetReports(ctx, sqlc.GetReportsParams{
			Symbol: symbol,
			Limit:  limit,
		})
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	result := make([]*ntxv1.Report, len(reports))
	for i, r := range reports {
		result[i] = reportToProto(r)
	}

	return connect.NewResponse(&ntxv1.ListReportsResponse{
		Reports: result,
	}), nil
}

func companyToProto(c sqlc.Company) *ntxv1.Company {
	return &ntxv1.Company{
		Symbol:      c.Symbol,
		Name:        c.Name,
		Sector:      ntxv1.Sector(safeInt32(c.Sector)),
		Description: c.Description,
		LogoUrl:     c.LogoUrl,
	}
}

func fundamentalsToProto(f sqlc.Fundamental) *ntxv1.Fundamentals {
	return &ntxv1.Fundamentals{
		Symbol:            f.Symbol,
		Pe:                nullFloat64(f.Pe),
		Pb:                nullFloat64(f.Pb),
		Eps:               nullFloat64(f.Eps),
		BookValue:         nullFloat64(f.BookValue),
		MarketCap:         nullFloat64(f.MarketCap),
		DividendYield:     nullFloat64(f.DividendYield),
		Roe:               nullFloat64(f.Roe),
		SharesOutstanding: nullInt64(f.SharesOutstanding),
	}
}

func reportToProto(r sqlc.Report) *ntxv1.Report {
	return &ntxv1.Report{
		Symbol:     r.Symbol,
		Type:       ntxv1.ReportType(safeInt32(r.Type)),
		FiscalYear: safeInt32(r.FiscalYear),
		Quarter:    safeInt32(r.Quarter),
		Revenue:    nullFloat64(r.Revenue),
		NetIncome:  nullFloat64(r.NetIncome),
		Eps:        nullFloat64(r.Eps),
		BookValue:  nullFloat64(r.BookValue),
		NplRatio:   nullFloat64(r.NplRatio),
	}
}
