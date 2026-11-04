package nepse

import (
	"context"
	"fmt"
)

// MarketStatus is the exchange's open/close state. AsOf is the time of the
// last state change, e.g. "2026-11-04T15:00:00"; on a holiday it still points
// at the previous session.
type MarketStatus struct {
	IsOpen string
	AsOf   string
}

func (c *Client) MarketStatus(ctx context.Context) (MarketStatus, error) {
	status, err := c.api.MarketStatus(ctx)
	if err != nil {
		return MarketStatus{}, fmt.Errorf("fetch market status: %w", err)
	}
	return MarketStatus{IsOpen: status.IsOpen, AsOf: status.AsOf}, nil
}
