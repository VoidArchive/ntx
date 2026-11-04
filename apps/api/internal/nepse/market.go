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

// MarketDay is the session-wide figures from the market summary and the NEPSE
// index. The feed has no business date; the caller supplies it.
type MarketDay struct {
	IndexClose         float64
	IndexChange        float64
	IndexChangePercent float64
	Turnover           float64
	TradedShares       int64
	Transactions       int64
	ScripsTraded       int64
	MarketCap          float64
}

func (c *Client) MarketDay(ctx context.Context) (MarketDay, error) {
	summary, err := c.api.MarketSummary(ctx)
	if err != nil {
		return MarketDay{}, fmt.Errorf("fetch market summary: %w", err)
	}
	index, err := c.api.NepseIndex(ctx)
	if err != nil {
		return MarketDay{}, fmt.Errorf("fetch nepse index: %w", err)
	}
	return MarketDay{
		IndexClose:         index.IndexValue,
		IndexChange:        index.PointChange,
		IndexChangePercent: index.PercentChange,
		Turnover:           summary.TotalTurnover,
		TradedShares:       int64(summary.TotalTradedShares),
		Transactions:       int64(summary.TotalTransactions),
		ScripsTraded:       int64(summary.TotalScripsTraded),
		MarketCap:          summary.TotalMarketCapitalization,
	}, nil
}
