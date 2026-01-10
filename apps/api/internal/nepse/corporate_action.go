package nepse

import (
	"context"
	"fmt"
)

type CorporateAction struct {
	CompanyID       int64
	FiscalYear      string
	BonusPercentage float64
	RightPercentage *float64
	CashDividend    *float64
	SubmittedDate   string
}

func (c *Client) CorporateActions(ctx context.Context, companyID int32) ([]CorporateAction, error) {
	actions, err := c.api.CorporateActions(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("fetch corporate actions: %w", err)
	}

	var result []CorporateAction
	for _, a := range actions {
		ca := CorporateAction{
			CompanyID:       int64(companyID),
			FiscalYear:      a.FiscalYear,
			BonusPercentage: a.BonusPercentage,
			SubmittedDate:   a.SubmittedDate,
		}
		if a.RightPercentage != nil {
			ca.RightPercentage = a.RightPercentage
		}
		if a.CashDividend != nil {
			ca.CashDividend = a.CashDividend
		}
		result = append(result, ca)
	}
	return result, nil
}
