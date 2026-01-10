package nepse

import (
	"context"
	"fmt"
)

type Ownership struct {
	CompanyID       int64
	ListedShares    int64
	PublicShares    int64
	PublicPercent   float64
	PromoterShares  int64
	PromoterPercent float64
}

func (c *Client) SecurityDetail(ctx context.Context, companyID int32) (*Ownership, error) {
	detail, err := c.api.SecurityDetail(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("fetch security detail: %w", err)
	}

	return &Ownership{
		CompanyID:       int64(detail.ID),
		ListedShares:    detail.ListedShares,
		PublicShares:    detail.PublicShares,
		PublicPercent:   detail.PublicPercent,
		PromoterShares:  detail.PromoterShares,
		PromoterPercent: detail.PromoterPercent,
	}, nil
}
