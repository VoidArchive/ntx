package models

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Money represents a monetary value in paisa (1 rupee = 100 paisa)
// Uses integer storage to avoid floating-point precision errors
type Money int64

// Percentage represents a percentage value in basis points (1% = 100 bp)
// Uses integer storage for precise calculations
type Percentage int64

// Quantity represents a quantity of shares or units
type Quantity int64

// Paisa conversion constants
const (
	PaisaPerRupee     = 100
	BasisPointsPerOne = 100 // 1% = 100 basis points
)

// Money constructor functions

// NewMoneyFromRupees creates Money from rupees (supports up to 2 decimal places)
func NewMoneyFromRupees(rupees float64) Money {
	return Money(math.Round(rupees * PaisaPerRupee))
}

// NewMoneyFromPaisa creates Money from paisa
func NewMoneyFromPaisa(paisa int64) Money {
	return Money(paisa)
}

// NewMoneyFromString parses a string representation of money
// Supports formats: "1234.56", "1,234.56", "Rs. 1,234.56"
func NewMoneyFromString(s string) (Money, error) {
	// Clean the string
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "Rs.", "")
	s = strings.ReplaceAll(s, "Rs", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	if s == "" {
		return Money(0), nil
	}

	// Parse as float
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Money(0), fmt.Errorf("invalid money format: %s", s)
	}

	return NewMoneyFromRupees(val), nil
}

// Money methods

// Paisa returns the value in paisa
func (m Money) Paisa() int64 {
	return int64(m)
}

// Rupees returns the value in rupees as float64
func (m Money) Rupees() float64 {
	return float64(m) / PaisaPerRupee
}

// String returns formatted string representation
func (m Money) String() string {
	rupees := m.Rupees()
	return fmt.Sprintf("%.2f", rupees)
}

// FormattedString returns formatted string with currency symbol
func (m Money) FormattedString() string {
	return fmt.Sprintf("Rs. %s", m.String())
}

// CommaSeparated returns comma-separated formatted string
func (m Money) CommaSeparated() string {
	rupees := m.Rupees()
	str := fmt.Sprintf("%.2f", rupees)
	
	// Split integer and decimal parts
	parts := strings.Split(str, ".")
	integer := parts[0]
	decimal := parts[1]
	
	// Add commas to integer part
	if len(integer) > 3 {
		var result strings.Builder
		for i, digit := range integer {
			if i > 0 && (len(integer)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		return fmt.Sprintf("Rs. %s.%s", result.String(), decimal)
	}
	
	return fmt.Sprintf("Rs. %s", str)
}

// Arithmetic operations

// Add adds two Money values
func (m Money) Add(other Money) Money {
	return Money(int64(m) + int64(other))
}

// Subtract subtracts another Money value
func (m Money) Subtract(other Money) Money {
	return Money(int64(m) - int64(other))
}

// Multiply multiplies Money by a factor
func (m Money) Multiply(factor float64) Money {
	return Money(math.Round(float64(m) * factor))
}

// MultiplyByQuantity multiplies Money by a Quantity
func (m Money) MultiplyByQuantity(q Quantity) Money {
	return Money(int64(m) * int64(q))
}

// Divide divides Money by a divisor
func (m Money) Divide(divisor float64) Money {
	if divisor == 0 {
		return Money(0)
	}
	return Money(math.Round(float64(m) / divisor))
}

// DivideByQuantity divides Money by a Quantity
func (m Money) DivideByQuantity(q Quantity) Money {
	if q == 0 {
		return Money(0)
	}
	return Money(int64(m) / int64(q))
}

// Comparison operations

// IsZero checks if money is zero
func (m Money) IsZero() bool {
	return m == 0
}

// IsPositive checks if money is positive
func (m Money) IsPositive() bool {
	return m > 0
}

// IsNegative checks if money is negative
func (m Money) IsNegative() bool {
	return m < 0
}

// Abs returns absolute value
func (m Money) Abs() Money {
	if m < 0 {
		return -m
	}
	return m
}

// Percentage constructor functions

// NewPercentageFromFloat creates Percentage from float (e.g., 5.25 for 5.25%)
func NewPercentageFromFloat(percent float64) Percentage {
	return Percentage(math.Round(percent * BasisPointsPerOne))
}

// NewPercentageFromBasisPoints creates Percentage from basis points
func NewPercentageFromBasisPoints(bp int64) Percentage {
	return Percentage(bp)
}

// NewPercentageFromString parses a string representation of percentage
// Supports formats: "5.25", "5.25%", "+5.25%", "-2.5%"
func NewPercentageFromString(s string) (Percentage, error) {
	// Clean the string
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "%", "")
	s = strings.TrimSpace(s)

	if s == "" {
		return Percentage(0), nil
	}

	// Parse as float
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Percentage(0), fmt.Errorf("invalid percentage format: %s", s)
	}

	return NewPercentageFromFloat(val), nil
}

// Percentage methods

// BasisPoints returns the value in basis points
func (p Percentage) BasisPoints() int64 {
	return int64(p)
}

// Float returns the percentage as float64
func (p Percentage) Float() float64 {
	return float64(p) / BasisPointsPerOne
}

// String returns formatted string representation
func (p Percentage) String() string {
	return fmt.Sprintf("%.2f%%", p.Float())
}

// SignedString returns formatted string with explicit sign
func (p Percentage) SignedString() string {
	val := p.Float()
	if val >= 0 {
		return fmt.Sprintf("+%.2f%%", val)
	}
	return fmt.Sprintf("%.2f%%", val)
}

// Percentage arithmetic operations

// Add adds two Percentage values
func (p Percentage) Add(other Percentage) Percentage {
	return Percentage(int64(p) + int64(other))
}

// Subtract subtracts another Percentage value
func (p Percentage) Subtract(other Percentage) Percentage {
	return Percentage(int64(p) - int64(other))
}

// Percentage comparison operations

// IsZero checks if percentage is zero
func (p Percentage) IsZero() bool {
	return p == 0
}

// IsPositive checks if percentage is positive
func (p Percentage) IsPositive() bool {
	return p > 0
}

// IsNegative checks if percentage is negative
func (p Percentage) IsNegative() bool {
	return p < 0
}

// Abs returns absolute value
func (p Percentage) Abs() Percentage {
	if p < 0 {
		return -p
	}
	return p
}

// Quantity constructor functions

// NewQuantity creates a new Quantity
func NewQuantity(q int64) Quantity {
	return Quantity(q)
}

// NewQuantityFromString parses a string representation of quantity
func NewQuantityFromString(s string) (Quantity, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	
	if s == "" {
		return Quantity(0), nil
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return Quantity(0), fmt.Errorf("invalid quantity format: %s", s)
	}

	return Quantity(val), nil
}

// Quantity methods

// Int64 returns the quantity as int64
func (q Quantity) Int64() int64 {
	return int64(q)
}

// String returns formatted string representation
func (q Quantity) String() string {
	return strconv.FormatInt(int64(q), 10)
}

// CommaSeparated returns comma-separated formatted string
func (q Quantity) CommaSeparated() string {
	str := q.String()
	if len(str) <= 3 {
		return str
	}
	
	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

// Quantity arithmetic operations

// Add adds two Quantity values
func (q Quantity) Add(other Quantity) Quantity {
	return Quantity(int64(q) + int64(other))
}

// Subtract subtracts another Quantity value
func (q Quantity) Subtract(other Quantity) Quantity {
	return Quantity(int64(q) - int64(other))
}

// Multiply multiplies Quantity by a factor
func (q Quantity) Multiply(factor int64) Quantity {
	return Quantity(int64(q) * factor)
}

// Quantity comparison operations

// IsZero checks if quantity is zero
func (q Quantity) IsZero() bool {
	return q == 0
}

// IsPositive checks if quantity is positive
func (q Quantity) IsPositive() bool {
	return q > 0
}

// Financial calculation helpers

// CalculatePercentageChange calculates percentage change between two Money values
func CalculatePercentageChange(oldValue, newValue Money) Percentage {
	if oldValue.IsZero() {
		return Percentage(0)
	}
	
	change := newValue.Subtract(oldValue)
	changeFloat := change.Rupees()
	oldFloat := oldValue.Rupees()
	
	percentFloat := (changeFloat / oldFloat) * 100
	return NewPercentageFromFloat(percentFloat)
}

// CalculatePortfolioValue calculates total portfolio value
func CalculatePortfolioValue(holdings []struct {
	Quantity Quantity
	Price    Money
}) Money {
	total := Money(0)
	for _, holding := range holdings {
		value := holding.Price.MultiplyByQuantity(holding.Quantity)
		total = total.Add(value)
	}
	return total
}

// ApplyPercentageToMoney applies a percentage to a money value
func ApplyPercentageToMoney(amount Money, percent Percentage) Money {
	factor := percent.Float() / 100.0
	return amount.Multiply(factor)
}