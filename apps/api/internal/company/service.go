// Package company service
package company

import (
	"database/sql"

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

func companiesToProto(companies []sqlc.Company) []*ntxv1.Company {
	out := make([]*ntxv1.Company, len(companies))
	for i, c := range companies {
		out[i] = companyToProto(c)
	}
	return out
}

func listCompaniesRowsToProto(rows []sqlc.ListCompaniesRow) []*ntxv1.Company {
	out := make([]*ntxv1.Company, len(rows))
	for i, r := range rows {
		out[i] = &ntxv1.Company{
			Id:             r.ID,
			Name:           r.Name,
			Symbol:         r.Symbol,
			Status:         statusFromDB(r.Status),
			Email:          nullString(r.Email),
			Website:        nullString(r.Website),
			InstrumentType: instrumentFromDB(r.InstrumentType),
			Sector:         sectorFromDB(r.Sector),
			ListedShares:   nullInt64Ptr(r.ListedShares),
		}
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
		Sector:         sectorFromDB(c.Sector),
	}
}

func nullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

var statusMap = map[string]ntxv1.CompanyStatus{
	"A": ntxv1.CompanyStatus_COMPANY_STATUS_ACTIVE,
	"S": ntxv1.CompanyStatus_COMPANY_STATUS_SUSPENDED,
	"D": ntxv1.CompanyStatus_COMPANY_STATUS_DELISTED,
}

func statusFromDB(s string) ntxv1.CompanyStatus {
	if status, ok := statusMap[s]; ok {
		return status
	}
	return ntxv1.CompanyStatus_COMPANY_STATUS_UNSPECIFIED
}

var sectorMap = map[ntxv1.Sector]string{
	ntxv1.Sector_SECTOR_COMMERCIAL_BANK:    "Commercial Banks",
	ntxv1.Sector_SECTOR_DEVELOPMENT_BANK:   "Development Banks",
	ntxv1.Sector_SECTOR_FINANCE:            "Finance",
	ntxv1.Sector_SECTOR_HOTEL:              "Hotel",
	ntxv1.Sector_SECTOR_HYDROPOWER:         "Hydropower",
	ntxv1.Sector_SECTOR_INVESTMENT:         "Investment",
	ntxv1.Sector_SECTOR_LIFE_INSURANCE:     "Life Insurance",
	ntxv1.Sector_SECTOR_MANUFACTURING:      "Manufacturing",
	ntxv1.Sector_SECTOR_MICROFINANCE:       "Microfinance",
	ntxv1.Sector_SECTOR_NON_LIFE_INSURANCE: "Non Life Insurance",
	ntxv1.Sector_SECTOR_TRADING:            "Trading",
	ntxv1.Sector_SECTOR_MUTUAL_FUND:        "Mutual Funds",
	ntxv1.Sector_SECTOR_OTHERS:             "Others",
}

func sectorEnumToDB(sector ntxv1.Sector) (string, bool) {
	str, ok := sectorMap[sector]
	return str, ok
}

var sectorDBMap map[string]ntxv1.Sector

func init() {
	sectorDBMap = make(map[string]ntxv1.Sector, len(sectorMap))
	for k, v := range sectorMap {
		sectorDBMap[v] = k
	}

	// Add aliases for data mismatches
	sectorDBMap["Hydro Power"] = ntxv1.Sector_SECTOR_HYDROPOWER
	sectorDBMap["Hotels And Tourism"] = ntxv1.Sector_SECTOR_HOTEL
	sectorDBMap["Manufacturing And Processing"] = ntxv1.Sector_SECTOR_MANUFACTURING
	sectorDBMap["Tradings"] = ntxv1.Sector_SECTOR_TRADING
}

func sectorFromDB(s string) ntxv1.Sector {
	if sector, ok := sectorDBMap[s]; ok {
		return sector
	}
	return ntxv1.Sector_SECTOR_UNSPECIFIED
}

var instrumentMap = map[string]ntxv1.InstrumentType{
	"EQUITY": ntxv1.InstrumentType_INSTRUMENT_TYPE_EQUITY,
	"BOND":   ntxv1.InstrumentType_INSTRUMENT_TYPE_BOND,
}

func instrumentFromDB(s string) ntxv1.InstrumentType {
	if instrument, ok := instrumentMap[s]; ok {
		return instrument
	}
	return ntxv1.InstrumentType_INSTRUMENT_TYPE_UNSPECIFIED
}

func fundamentalsToProto(fundamentals []sqlc.Fundamental) []*ntxv1.Fundamental {
	out := make([]*ntxv1.Fundamental, len(fundamentals))
	for i, f := range fundamentals {
		out[i] = fundamentalToProto(f)
	}
	return out
}

func fundamentalToProto(f sqlc.Fundamental) *ntxv1.Fundamental {
	var quarter *string
	if f.Quarter != "" {
		quarter = &f.Quarter
	}
	return &ntxv1.Fundamental{
		Id:            f.ID,
		CompanyId:     f.CompanyID,
		FiscalYear:    f.FiscalYear,
		Quarter:       quarter,
		Eps:           nullFloat64(f.Eps),
		PeRatio:       nullFloat64(f.PeRatio),
		BookValue:     nullFloat64(f.BookValue),
		PaidUpCapital: nullFloat64(f.PaidUpCapital),
		ProfitAmount:  nullFloat64(f.ProfitAmount),
	}
}

func nullFloat64(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

func nullInt64Ptr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}
