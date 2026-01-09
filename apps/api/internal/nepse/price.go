package nepse

import (
	"context"
	"fmt"
)

type Price struct {
	SecurityID    int32
	Symbol        string
	BusinessDate  string
	Open          float64
	High          float64
	Low           float64
	Close         float64
	LTP           float64
	PreviousClose float64
	Change        float64
	ChangePercent float64
	Volume        int64
	Turnover      float64
	Trades        int32
}

func (c *Client) TodaysPrices(ctx context.Context, businessDate string) ([]Price, error) {
	prices, err := c.api.TodaysPrices(ctx, businessDate)
	if err != nil {
		return nil, fmt.Errorf("fetch today prices: %w", err)
	}

	result := make([]Price, 0, len(prices))
	for _, p := range prices {
		result = append(result, Price{
			SecurityID:    p.SecurityID,
			Symbol:        p.Symbol,
			BusinessDate:  p.BusinessDate,
			Open:          p.OpenPrice,
			High:          p.HighPrice,
			Low:           p.LowPrice,
			Close:         p.ClosePrice,
			LTP:           p.LastTradedPrice,
			PreviousClose: p.PreviousClose,
			Change:        p.DifferenceRs,
			ChangePercent: p.PercentageChange,
			Volume:        p.TotalTradedQuantity,
			Turnover:      p.TotalTradedValue,
			Trades:        p.TotalTrades,
		})
	}
	return result, nil
}
