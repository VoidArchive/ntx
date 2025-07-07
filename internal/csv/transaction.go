package csv

import (
	"fmt"
	"time"
	"ntx/internal/money"
)

// Transaction represents a single transaction from the Meroshare CSV file.
// It contains all necessary information to track share purchases, sales,
// and corporate actions for portfolio management and tax calculations.
type Transaction struct {
	Scrip           string          // Stock symbol (e.g., "API", "NMB")
	Date            time.Time       // Transaction date
	Quantity        int             // Positive for buy/credit, negative for sell/debit
	Price           money.Money     // Price per share (zero if not entered yet)
	TransactionType TransactionType // Type: IPO, BONUS, RIGHTS, MERGER, REARRANGEMENT, REGULAR
	Description     string          // Original description from Meroshare CSV
	BalanceAfter    int             // Share balance after this transaction (whole shares)
}

// IsBuy returns true if the transaction represents a purchase or credit
// (positive quantity). This includes regular buys, IPOs, bonus shares, etc.
func (t Transaction) IsBuy() bool {
	return t.Quantity > 0
}

// IsSell returns true if the transaction represents a sale or debit
// (negative quantity).
func (t Transaction) IsSell() bool {
	return t.Quantity < 0
}

// AbsQuantity returns the absolute value of the transaction quantity.
// Useful for calculations that need the magnitude regardless of buy/sell direction.
func (t Transaction) AbsQuantity() int {
	if t.Quantity < 0 {
		return -t.Quantity
	}
	return t.Quantity
}

// NeedsPrice returns true if the transaction requires price input from the user.
// Bonus shares, rights, mergers, and rearrangements typically don't need market prices.
// IPOs and regular transactions need prices for proper portfolio calculations.
func (t Transaction) NeedsPrice() bool {
	// IPO, Bonus, Rights, Merger, and Rearrangement don't need market prices
	switch t.TransactionType {
	case TransactionTypeBonus, TransactionTypeRights, TransactionTypeMerger, TransactionTypeRearrangement:
		return false
	case TransactionTypeIPO:
		return t.Price.IsZero() // IPO needs price if not already set
	default:
		return t.Price.IsZero() // Regular transactions need price
	}
}

// Validate returns an error if the transaction has invalid field values.
func (t Transaction) Validate() error {
	if t.Scrip == "" {
		return fmt.Errorf("%w: scrip", ErrRequiredField)
	}
	if t.Date.IsZero() {
		return fmt.Errorf("%w: date", ErrRequiredField)
	}
	if t.Quantity == 0 {
		return ErrInvalidQuantity
	}
	if t.Price.IsNegative() {
		return fmt.Errorf("price cannot be negative: %s", t.Price)
	}
	return nil
}

// IsValidTransactionType returns true if the transaction type is recognized.
func (t Transaction) IsValidTransactionType() bool {
	switch t.TransactionType {
	case TransactionTypeIPO, TransactionTypeBonus, TransactionTypeRights,
		TransactionTypeMerger, TransactionTypeRearrangement, TransactionTypeRegular:
		return true
	default:
		return false
	}
}