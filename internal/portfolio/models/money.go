package models

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Money represents currency amounts with paisa precision
// Stores values as integers to avoid floating-point precision errors
type Money struct {
	Paisa int64 `json:"paisa"`
}

// NewMoney creates Money from rupees (float64)
func NewMoney(rupees float64) Money {
	return Money{Paisa: int64(math.Round(rupees * 100))}
}

// NewMoneyFromPaisa creates Money from paisa (int64)
func NewMoneyFromPaisa(paisa int64) Money {
	return Money{Paisa: paisa}
}

// NewMoneyFromString parses Money from string (e.g., "1250.50")
func NewMoneyFromString(s string) (Money, error) {
	// Handle empty string
	s = strings.TrimSpace(s)
	if s == "" {
		return Money{}, nil
	}

	// Remove currency symbols and commas
	s = strings.ReplaceAll(s, "Rs.", "")
	s = strings.ReplaceAll(s, "Rs.", "")
	s = strings.ReplaceAll(s, "Rs", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	rupees, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Money{}, fmt.Errorf("invalid money format: %s", s)
	}

	return NewMoney(rupees), nil
}

// Rupees returns the amount in rupees as float64
func (m Money) Rupees() float64 {
	return float64(m.Paisa) / 100.0
}

// Add adds two Money values
func (m Money) Add(other Money) Money {
	return Money{Paisa: m.Paisa + other.Paisa}
}

// Subtract subtracts another Money value
func (m Money) Sub(other Money) Money {
	return Money{Paisa: m.Paisa - other.Paisa}
}

// Multiply multiplies Money by a factor
func (m Money) Multiply(factor float64) Money {
	return Money{Paisa: int64(math.Round(float64(m.Paisa) * factor))}
}

// MultiplyInt multiplies Money by an integer (for quantities)
func (m Money) MultiplyInt(factor int64) Money {
	return Money{Paisa: m.Paisa * factor}
}

// Divide divides Money by a factor
func (m Money) Divide(divisor float64) Money {
	if divisor == 0 {
		return Money{Paisa: 0}
	}
	return Money{Paisa: int64(math.Round(float64(m.Paisa) / divisor))}
}

// DivideInt divides Money by an integer
func (m Money) DivideInt(divisor int64) Money {
	if divisor == 0 {
		return Money{Paisa: 0}
	}
	return Money{Paisa: m.Paisa / divisor}
}

// Percentage calculates percentage of Money
func (m Money) Percentage(percent float64) Money {
	return m.Multiply(percent / 100.0)
}

// PercentageChange calculates percentage change from another Money value
func (m Money) PercentageChange(from Money) float64 {
	if from.Paisa == 0 {
		if m.Paisa == 0 {
			return 0
		}
		return 100 // 100% gain from zero
	}

	change := float64(m.Paisa - from.Paisa)
	return (change / float64(from.Paisa)) * 100.0
}

// IsZero returns true if the amount is zero
func (m Money) IsZero() bool {
	return m.Paisa == 0
}

// IsPositive returns true if the amount is positive
func (m Money) IsPositive() bool {
	return m.Paisa > 0
}

// IsNegative returns true if the amount is negative
func (m Money) IsNegative() bool {
	return m.Paisa < 0
}

// Abs returns the absolute value
func (m Money) Abs() Money {
	if m.Paisa < 0 {
		return Money{Paisa: -m.Paisa}
	}
	return m
}

// Compare returns -1 if m < other, 0 if equal, 1 if m > other
func (m Money) Compare(other Money) int {
	if m.Paisa < other.Paisa {
		return -1
	}
	if m.Paisa > other.Paisa {
		return 1
	}
	return 0
}

// Equal returns true if two Money values are equal
func (m Money) Equal(other Money) bool {
	return m.Paisa == other.Paisa
}

// String returns formatted string representation
func (m Money) String() string {
	return m.FormatNPR()
}

// FormatNPR formats as Nepali Rupees with proper formatting
func (m Money) FormatNPR() string {
	rupees := m.Rupees()

	// Handle negative values
	if rupees < 0 {
		return fmt.Sprintf("-Rs.%s", formatPositiveAmount(-rupees))
	}

	return fmt.Sprintf("Rs.%s", formatPositiveAmount(rupees))
}

// FormatSimple returns simple decimal format (e.g., "1250.50")
func (m Money) FormatSimple() string {
	return fmt.Sprintf("%.2f", m.Rupees())
}

// FormatWithSign returns formatted string with explicit + or - sign
func (m Money) FormatWithSign() string {
	if m.Paisa > 0 {
		return "+" + m.FormatNPR()
	}
	return m.FormatNPR()
}

// formatPositiveAmount formats a positive amount with commas
func formatPositiveAmount(amount float64) string {
	// Convert to string with 2 decimal places
	str := fmt.Sprintf("%.2f", amount)

	// Split integer and decimal parts
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := parts[1]

	// Add commas to integer part (Indian number system)
	if len(intPart) > 3 {
		// For amounts >= 1000, add commas
		result := ""
		for i, digit := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result += ","
			}
			result += string(digit)
		}
		intPart = result
	}

	// Remove trailing zeros from decimal part
	decPart = strings.TrimRight(decPart, "0")
	if decPart == "" {
		return intPart
	}

	return intPart + "." + decPart
}

// Zero returns a zero Money value
func Zero() Money {
	return Money{Paisa: 0}
}

// Sum calculates the sum of multiple Money values
func Sum(amounts ...Money) Money {
	total := Zero()
	for _, amount := range amounts {
		total = total.Add(amount)
	}
	return total
}
