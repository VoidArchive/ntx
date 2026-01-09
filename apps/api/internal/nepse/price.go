package nepse

import (
	"context"
	"fmt"
	"strconv"
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

// LiveMarket returns real-time price data for all securities.
// This is more reliable than TodaysPrices which may return empty results.
func (c *Client) LiveMarket(ctx context.Context) ([]Price, error) {
	entries, err := c.api.LiveMarket(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch live market: %w", err)
	}

	result := make([]Price, 0, len(entries))
	for _, e := range entries {
		securityID, _ := strconv.ParseInt(e.SecurityID, 10, 32)
		result = append(result, Price{
			SecurityID:    int32(securityID),
			Symbol:        e.Symbol,
			Open:          e.OpenPrice,
			High:          e.HighPrice,
			Low:           e.LowPrice,
			LTP:           e.LastTradedPrice,
			PreviousClose: e.PreviousClose,
			ChangePercent: e.PercentageChange,
			Volume:        e.TotalTradeQuantity,
			Turnover:      e.TotalTradeValue,
		})
	}
	return result, nil
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

// HistoricalPrice represents OHLCV data for a single day.
// Note: NEPSE API does not provide open price in historical data.
type HistoricalPrice struct {
	BusinessDate string
	High         float64
	Low          float64
	Close        float64
	Volume       int64
	Turnover     float64
	Trades       int32
}

// PriceHistory returns historical OHLCV data for a security within a date range.
// Date format: "YYYY-MM-DD"
func (c *Client) PriceHistory(
	ctx context.Context,
	securityID int32,
	startDate, endDate string,
) ([]HistoricalPrice, error) {
	history, err := c.api.PriceHistory(ctx, securityID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("fetch price history: %w", err)
	}

	result := make([]HistoricalPrice, 0, len(history))
	for _, h := range history {
		result = append(result, HistoricalPrice{
			BusinessDate: h.BusinessDate,
			High:         h.HighPrice,
			Low:          h.LowPrice,
			Close:        h.ClosePrice,
			Volume:       h.TotalTradedQuantity,
			Turnover:     h.TotalTradedValue,
			Trades:       h.TotalTrades,
		})
	}
	return result, nil
}
