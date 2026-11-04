package nepse

import (
	"context"
	"fmt"
	"log/slog"
)

// Dividend represents dividend data from NEPSE.
// Uses the Dividends API which has complete data (bonus + cash).
type Dividend struct {
	CompanyID       int64
	FiscalYear      string  // AD format: "2024-2025"
	BonusPercentage float64 // Bonus share percentage
	RightPercentage *float64
	CashDividend    *float64
	ModifiedDate    string // When dividend was declared
	NoticeHeadline  string // Verbatim from the dividend notice
	NoticeBody      string
	NoticeURL       string // Filing path as NEPSE gives it; "" when no single filing matches
}

// Dividends fetches dividend history for a company using the Dividends API.
// This replaces CorporateActions as it provides complete data including cash dividends.
func (c *Client) Dividends(ctx context.Context, companyID int32) ([]Dividend, error) {
	dividends, err := c.api.Dividends(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("fetch dividends: %w", err)
	}

	// The dividend feed has no filing path; only corporate actions carry one.
	// Losing the path is better than losing the dividend, so a failure here is not fatal.
	paths := map[string]string{}
	actions, err := c.api.CorporateActions(ctx, companyID)
	if err != nil {
		slog.Warn("corporate actions unavailable, notice paths left empty", "company_id", companyID, "err", err)
	}
	for _, a := range actions {
		if a.FilePath == "" {
			continue
		}
		if _, seen := paths[a.FiscalYear]; seen {
			// Two filings in one year: we cannot tell which is the dividend notice, so record none.
			paths[a.FiscalYear] = ""
			continue
		}
		paths[a.FiscalYear] = a.FilePath
	}

	var result []Dividend
	for _, d := range dividends {
		// Skip entries without dividend notice
		if d.CompanyNews == nil || d.CompanyNews.DividendsNotice == nil {
			continue
		}

		dn := d.CompanyNews.DividendsNotice
		fiscalYear, fiscalYearNepali := "", ""
		if dn.FinancialYear != nil {
			fiscalYear = dn.FinancialYear.FYName
			fiscalYearNepali = dn.FinancialYear.FYNameNepali
		}

		// Skip entries without fiscal year
		if fiscalYear == "" {
			continue
		}

		div := Dividend{
			CompanyID:       int64(companyID),
			FiscalYear:      fiscalYear,
			BonusPercentage: dn.BonusShare,
			ModifiedDate:    d.ModifiedDate,
			NoticeHeadline:  d.CompanyNews.NewsHeadline,
			NoticeBody:      d.CompanyNews.NewsBody,
			// Corporate actions label the year in BS ("2080-2081"), matching FYNameNepali.
			NoticeURL: paths[fiscalYearNepali],
		}

		if dn.RightShare > 0 {
			div.RightPercentage = &dn.RightShare
		}
		if dn.CashDividend > 0 {
			div.CashDividend = &dn.CashDividend
		}

		result = append(result, div)
	}
	return result, nil
}
