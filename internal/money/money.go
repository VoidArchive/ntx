package money

import (
	"fmt"
	"strconv"
	"strings"
)

// Money represents Nepali Rupees with precise decimal handling.
// Internally stores paisa (1 NPR = 100 paisa) to avoid floating point errors.
type Money int64

const (
	// PaisaPerRupee defines the subdivision of NPR (100 paisa = 1 rupee)
	PaisaPerRupee = 100
)

// NewMoney creates a Money value from rupees (supports fractional rupees).
func NewMoney(rupees float64) Money {
	return Money(rupees * PaisaPerRupee)
}

// NewMoneyFromPaisa creates a Money value directly from paisa.
func NewMoneyFromPaisa(paisa int64) Money {
	return Money(paisa)
}

// ParseMoney parses a string representation of money.
// Accepts formats like "123.45", "123", "NPR 123.45"
func ParseMoney(s string) (Money, error) {
	// Clean the string
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimPrefix(s, "NPR")
	s = strings.TrimPrefix(s, "Rs.")
	s = strings.TrimSpace(s)

	if s == "" || s == "-" {
		return Money(0), nil
	}

	rupees, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Money(0), fmt.Errorf("invalid money format %q: %w", s, err)
	}

	return NewMoney(rupees), nil
}

// Rupees returns the monetary value as rupees (float64).
// Use this for calculations that require floating point precision.
func (m Money) Rupees() float64 {
	return float64(m) / PaisaPerRupee
}

// Paisa returns the monetary value in paisa (int64).
func (m Money) Paisa() int64 {
	return int64(m)
}

// String returns a formatted string representation of the money.
// Format: "NPR 1,234.56"
func (m Money) String() string {
	rupees := m.Rupees()
	
	// Handle negative values
	if rupees < 0 {
		return fmt.Sprintf("NPR -%.2f", -rupees)
	}
	
	return fmt.Sprintf("NPR %.2f", rupees)
}

// Add returns the sum of two Money values.
func (m Money) Add(other Money) Money {
	return Money(int64(m) + int64(other))
}

// Subtract returns the difference of two Money values.
func (m Money) Subtract(other Money) Money {
	return Money(int64(m) - int64(other))
}

// Multiply returns the Money value multiplied by a factor.
// Useful for calculating total value (price * quantity).
func (m Money) Multiply(factor float64) Money {
	return Money(float64(m) * factor)
}

// MultiplyInt returns the Money value multiplied by an integer.
// More precise than Multiply for whole number factors.
func (m Money) MultiplyInt(factor int) Money {
	return Money(int64(m) * int64(factor))
}

// Divide returns the Money value divided by a factor.
// Useful for calculating average prices.
func (m Money) Divide(factor float64) Money {
	if factor == 0 {
		return Money(0)
	}
	return Money(float64(m) / factor)
}

// DivideInt returns the Money value divided by an integer.
// More precise than Divide for whole number factors.
func (m Money) DivideInt(factor int) Money {
	if factor == 0 {
		return Money(0)
	}
	return Money(int64(m) / int64(factor))
}

// IsZero returns true if the money value is zero.
func (m Money) IsZero() bool {
	return m == 0
}

// IsPositive returns true if the money value is greater than zero.
func (m Money) IsPositive() bool {
	return m > 0
}

// IsNegative returns true if the money value is less than zero.
func (m Money) IsNegative() bool {
	return m < 0
}

// Compare compares two Money values.
// Returns -1 if m < other, 0 if m == other, 1 if m > other.
func (m Money) Compare(other Money) int {
	if m < other {
		return -1
	}
	if m > other {
		return 1
	}
	return 0
}

// Equal returns true if two Money values are equal.
func (m Money) Equal(other Money) bool {
	return m == other
}

// LessThan returns true if m is less than other.
func (m Money) LessThan(other Money) bool {
	return m < other
}

// GreaterThan returns true if m is greater than other.
func (m Money) GreaterThan(other Money) bool {
	return m > other
}