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
	Price       float64
	Cost        float64
	Description string
	Note        string
}

type Lot struct {
	Quantity int
	Price    float64
	Date     time.Time
}

type RealizedGain struct {
	StockSymbol string
	SaleDate    time.Time
	Quantity    int
	SalePrice   float64
	CostBasis   float64
	GainLoss    float64
	HoldingDays int
	IsLongTerm  bool
}

type SaleResult struct {
	RealizedGains []RealizedGain
	TotalGainLoss float64
	SharesSold    int
	TotalProceeds
}
