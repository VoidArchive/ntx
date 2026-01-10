package company

import (
	"context"
	"database/sql"
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

func (s *CompanyService) GetOwnership(
	ctx context.Context,
	req *connect.Request[ntxv1.GetOwnershipRequest],
) (*connect.Response[ntxv1.GetOwnershipResponse], error) {
	if req.Msg.Symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}

	ownership, err := s.queries.GetOwnershipBySymbol(ctx, req.Msg.Symbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("ownership data not found"))
	}

	return connect.NewResponse(&ntxv1.GetOwnershipResponse{
		Ownership: ownershipToProto(ownership),
	}), nil
}

func (s *CompanyService) GetCorporateActions(
	ctx context.Context,
	req *connect.Request[ntxv1.GetCorporateActionsRequest],
) (*connect.Response[ntxv1.GetCorporateActionsResponse], error) {
	if req.Msg.Symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}

	actions, err := s.queries.GetCorporateActionsBySymbol(ctx, req.Msg.Symbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("corporate actions not found"))
	}

	protoActions := make([]*ntxv1.CorporateAction, 0, len(actions))
	for _, a := range actions {
		protoActions = append(protoActions, corporateActionToProto(a))
	}

	return connect.NewResponse(&ntxv1.GetCorporateActionsResponse{
		Actions: protoActions,
	}), nil
}

func ownershipToProto(o sqlc.Ownership) *ntxv1.Ownership {
	return &ntxv1.Ownership{
		CompanyId:       o.CompanyID,
		ListedShares:    nullInt64Val(o.ListedShares),
		PublicShares:    nullInt64Val(o.PublicShares),
		PublicPercent:   nullFloat64Val(o.PublicPercent),
		PromoterShares:  nullInt64Val(o.PromoterShares),
		PromoterPercent: nullFloat64Val(o.PromoterPercent),
		UpdatedAt:       o.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func corporateActionToProto(a sqlc.CorporateAction) *ntxv1.CorporateAction {
	return &ntxv1.CorporateAction{
		Id:              a.ID,
		CompanyId:       a.CompanyID,
		FiscalYear:      a.FiscalYear,
		BonusPercentage: nullFloat64Val(a.BonusPercentage),
		RightPercentage: nullFloat64Ptr(a.RightPercentage),
		CashDividend:    nullFloat64Ptr(a.CashDividend),
		SubmittedDate:   nullStringVal(a.SubmittedDate),
	}
}

func nullInt64Val(ni sql.NullInt64) int64 {
	if !ni.Valid {
		return 0
	}
	return ni.Int64
}

func nullFloat64Val(nf sql.NullFloat64) float64 {
	if !nf.Valid {
		return 0
	}
	return nf.Float64
}

func nullFloat64Ptr(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

func nullStringVal(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}
