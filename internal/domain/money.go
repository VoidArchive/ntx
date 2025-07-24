package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Money int64

const (
	PaisaPerRupee = 100
)

func NewMoney(rupees float64) Money {
	return Money(math.Round(rupees * PaisaPerRupee))
}

func NewMoneyFromString(s string) (Money, error) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, fmt.Errorf("empty money string")
	}

	rupees, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money format: %s", s)
	}
	return NewMoney(rupees), nil
}

func NewMoneyFromPaisa(paisa int64) Money {
	return Money(paisa)
}

func Zero() Money {
	return Money(0)
}

func (m Money) Rupees() float64 {
	return float64(m) / PaisaPerRupee
}

func (m Money) Paisa() int64 {
	return int64(m)
}

func (m Money) String() string {
	rupees := m.Rupees()
	if rupees < 0 {
		return fmt.Sprintf("-Rs. %s", formatRupees(-rupees))
	}

	return fmt.Sprintf("Rs. %s", formatRupees(rupees))
}

func formatRupees(rupees float64) string {
	str := fmt.Sprintf("%.2f", rupees)
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := parts[1]

	intPart = addCommas(intPart)
	return intPart + "." + decPart
}

func addCommas(s string) string {
	if len(s) <= 3 {
		return s
	}
	var result strings.Builder
	for i, char := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(char)
	}
	return result.String()
}

func (m Money) IsZero() bool {
	return m == 0
}

func (m Money) IsPositive() bool {
	return m > 0
}

func (m Money) IsNegative() bool {
	return m < 0
}

func (m Money) Abs() Money {
	if m < 0 {
		return -m
	}
	return m
}

func (m Money) Add(other Money) Money {
	return m + other
}

func (m Money) Subtract(other Money) Money {
	return m - other
}

func (m Money) Multiply(quantity int) Money {
	return Money(int64(m) * int64(quantity))
}

func (m Money) MultiplyFloat(factor float64) Money {
	return Money(math.Round(float64(m) * factor))
}

func (m Money) Divide(quantity int) Money {
	if quantity == 0 {
		panic("division by zero")
	}
	return Money(int64(m) / int64(quantity))
}

func (m Money) DivideFloat(divisor float64) Money {
	if divisor == 0 {
		panic("division by zero")
	}
	return Money(math.Round(float64(m) / divisor))
}

func (m Money) Percentage(percent float64) Money {
	return Money(math.Round(float64(m) * percent / 100))
}

func (m Money) PercentageChange(newValue Money) float64 {
	if m == 0 {
		if newValue == 0 {
			return 0
		}
		return math.Inf(1)
	}
	return ((float64(newValue) - float64(m)) / float64(m)) * 100
}

func (m Money) Compare(other Money) int {
	if m < other {
		return -1
	}
	if m > other {
		return 1
	}
	return 0
}

func (m Money) Equal(other Money) bool {
	return m == other
}

func (m Money) GreaterThan(other Money) bool {
	return m > other
}

func (m Money) LessThan(other Money) bool {
	return m < other
}

func (m Money) Min(other Money) Money {
	if m < other {
		return m
	}
	return other
}

func (m Money) Max(other Money) Money {
	if m > other {
		return m
	}
	return other
}
