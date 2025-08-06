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
