package nepse

import (
	"context"
	"fmt"
)

type Fundamental struct {
	CompanyID     int64
	FiscalYear    string
	Quarter       string // empty for annual reports
	EPS           float64
	PERatio       float64
	BookValue     float64
	PaidUpCapital float64
	ProfitAmount  float64
}

func (c *Client) Fundamentals(ctx context.Context, securityID int32) ([]Fundamental, error) {
	reports, err := c.api.Reports(ctx, securityID)
	if err != nil {
		return nil, fmt.Errorf("fetch reports: %w", err)
	}

	var fundamentals []Fundamental
	for _, r := range reports {
		if r.FiscalReport == nil {
			continue
		}
		fr := r.FiscalReport

		var fiscalYear string
		if fr.FinancialYear != nil {
			fiscalYear = fr.FinancialYear.FYName
		}

		fundamentals = append(fundamentals, Fundamental{
			CompanyID:     int64(r.ID),
			FiscalYear:    fiscalYear,
			Quarter:       r.QuarterName(),
			EPS:           fr.EPSValue,
			PERatio:       fr.PEValue,
			BookValue:     fr.NetWorthPerShare,
			PaidUpCapital: fr.PaidUpCapital,
			ProfitAmount:  fr.ProfitAmount,
		})
	}
	return fundamentals, nil
}
