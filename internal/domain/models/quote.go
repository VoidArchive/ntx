// Package models holds business-level data structures.
// It must NOT import infrastructure packages.
package models

// Quote represents a single NEPSE stock quote.
type Quote struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	LTP       float64 // Last Traded Price
	Volume    float64
	PrevClose float64
}

// PercentageChange calculates percentage change from previous close
func (q *Quote) PercentageChange() float64 {
	if q.PrevClose == 0 {
		return 0
	}
	return ((q.LTP - q.PrevClose) / q.PrevClose) * 100
}

// IsPositive returns true if the stock is trading higher than previous close
func (q *Quote) IsPositive() bool {
	return q.LTP >= q.PrevClose
}
