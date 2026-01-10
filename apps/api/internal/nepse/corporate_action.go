package nepse

import (
	"context"
	"fmt"
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
}

// Dividends fetches dividend history for a company using the Dividends API.
// This replaces CorporateActions as it provides complete data including cash dividends.
func (c *Client) Dividends(ctx context.Context, companyID int32) ([]Dividend, error) {
	dividends, err := c.api.Dividends(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("fetch dividends: %w", err)
	}

	var result []Dividend
	for _, d := range dividends {
		// Skip entries without dividend notice
		if d.CompanyNews == nil || d.CompanyNews.DividendsNotice == nil {
			continue
		}

		dn := d.CompanyNews.DividendsNotice
		fiscalYear := ""
		if dn.FinancialYear != nil {
			fiscalYear = dn.FinancialYear.FYName
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
