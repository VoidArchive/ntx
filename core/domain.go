package core

import (
	"time"
)

// TransactionType defines the type of a stock transaction.
type TransactionType string

const (
	TransactionTypeBuy   TransactionType = "BUY"
	TransactionTypeSell  TransactionType = "SELL"
	TransactionTypeBonus TransactionType = "BONUS"
	TransactionTypeRight TransactionType = "RIGHT"
)

// Transaction represents a single stock transaction record.
type Transaction struct {
	ID         string          `json:"id"`
	Date       time.Time       `json:"date"`
	Symbol     string          `json:"symbol"`
	Type       TransactionType `json:"type"`
	Quantity   float64         `json:"quantity"`
	Rate       float64         `json:"rate"`
	Amount     float64         `json:"amount"` // Total amount after charges
	Commission float64         `json:"commission"`
	Tax        float64         `json:"tax"`
}

// Holding represents the aggregate state of a specific stock in a portfolio.
type Holding struct {
	Symbol         string  `json:"symbol"`
	TotalQuantity  float64 `json:"total_quantity"`
	WAC            float64 `json:"wac"` // Weighted Average Cost
	CurrentPrice   float64 `json:"current_price"`
	MarketValue    float64 `json:"market_value"`
	UnrealizedGain float64 `json:"unrealized_gain"`
	RealizedGain   float64 `json:"realized_gain"`
}

// Portfolio is a collection of holdings and transactions.
type Portfolio struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Holdings     map[string]*Holding `json:"holdings"`
	Transactions []Transaction       `json:"transactions"`
}
