// Package domain contains the core business logic and domain models for NTX.
//
// This package implements the domain layer of clean architecture, providing:
// - Core business entities (Transaction, Lot, Holding)
// - Value objects (Money for precise financial calculations)
// - Domain services (FIFO queue for cost basis calculations)
// - Business rules for NEPSE stock portfolio management
//
// The domain layer has no external dependencies and contains pure business logic
// that can be tested independently of infrastructure concerns.
package domain

import "time"

// TransactionType represents the type of stock transaction
type TransactionType string

const (
	TransactionBuy      TransactionType = "BUY"      // Regular purchase (ON-CR)
	TransactionSell     TransactionType = "SELL"     // Regular sale (ON-DR)
	TransactionBonus    TransactionType = "BONUS"    // Bonus shares (CA-BONUS)
	TransactionRights   TransactionType = "RIGHTS"   // Rights issue purchase
	TransactionSplit    TransactionType = "SPLIT"    // Stock split adjustment
	TransactionDividend TransactionType = "DIVIDEND" // Cash dividend (tracking)
	TransactionMerger   TransactionType = "MERGER"   // Merger/acquisition (CA-REARRANGEMENT)
	TransactionOther    TransactionType = "OTHER"    // Manual/unknown entries
)

type Transaction struct {
	ID          int
	StockSymbol string
	Date        time.Time
	Type        TransactionType
	Quantity    int
	Price       Money
	Cost        Money
	Description string
	Note        string
}

type Lot struct {
	Quantity int
	Price    Money
	Date     time.Time
}

type RealizedGain struct {
	StockSymbol string
	SaleDate    time.Time
	Quantity    int
	SalePrice   Money
	CostBasis   Money
	GainLoss    Money
	HoldingDays int
	IsLongTerm  bool
}

type SaleResult struct {
	RealizedGains  []RealizedGain
	SharesSold     int
	TotalGainLoss  Money
	TotalProceeds  Money
	TotalCostBasis Money
}

type Holding struct {
	StockSymbol        string
	TotalShares        int
	WeightedAvgCost    Money
	TotalCost          Money
	CurrentPrice       Money
	MarketValue        Money
	UnrealizedGainLoss Money
	UnrealizedGainPct  float64
	lastUpdated        time.Time
}

type PortfolioSummary struct {
	TotalInvested     Money
	TotalMarketValue  Money
	TotalUnrealizedPL Money
	TotalRealizedPL   Money
	UnrealizedPLPct   float64
	HoldingsCount     int
	LastUpdated       time.Time
}
