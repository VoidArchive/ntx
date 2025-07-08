package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name     string
		rupees   float64
		expected Money
	}{
		{"Zero", 0.0, Money(0)},
		{"Positive integer", 100.0, Money(10000)},
		{"Positive decimal", 295.50, Money(29550)},
		{"Small decimal", 0.01, Money(1)},
		{"Large amount", 1850.75, Money(185075)},
		{"Negative", -100.50, Money(-10050)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMoney(tt.rupees)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewMoneyFromPaisa(t *testing.T) {
	tests := []struct {
		name     string
		paisa    int64
		expected Money
	}{
		{"Zero paisa", 0, Money(0)},
		{"Positive paisa", 29550, Money(29550)},
		{"Negative paisa", -10050, Money(-10050)},
		{"Large amount", 185075, Money(185075)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMoneyFromPaisa(tt.paisa)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseMoney(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Money
		expectError bool
	}{
		{"Simple number", "295.50", NewMoney(295.50), false},
		{"Integer", "100", NewMoney(100.0), false},
		{"With NPR prefix", "NPR 295.50", NewMoney(295.50), false},
		{"With Rs. prefix", "Rs. 295.50", NewMoney(295.50), false},
		{"With commas", "1,850.75", NewMoney(1850.75), false},
		{"With NPR and commas", "NPR 1,850.75", NewMoney(1850.75), false},
		{"Empty string", "", NewMoney(0), false},
		{"Just dash", "-", NewMoney(0), false},
		{"Whitespace only", "   ", NewMoney(0), false},
		{"Zero", "0", NewMoney(0), false},
		{"Negative", "-100.50", NewMoney(-100.50), false},
		{"Invalid text", "abc", Money(0), true},
		{"Mixed valid/invalid", "100abc", Money(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMoney(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, result.Equal(tt.expected),
					"Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMoney_Rupees(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected float64
	}{
		{"Zero", Money(0), 0.0},
		{"Positive", Money(29550), 295.50},
		{"Negative", Money(-10050), -100.50},
		{"Large amount", Money(185075), 1850.75},
		{"Small amount", Money(1), 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.Rupees()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_Paisa(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected int64
	}{
		{"Zero", Money(0), 0},
		{"Positive", Money(29550), 29550},
		{"Negative", Money(-10050), -10050},
		{"Large amount", Money(185075), 185075},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.Paisa()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_String(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected string
	}{
		{"Zero", Money(0), "NPR 0.00"},
		{"Positive", Money(29550), "NPR 295.50"},
		{"Negative", Money(-10050), "NPR -100.50"},
		{"Large amount", Money(185075), "NPR 1850.75"},
		{"Small amount", Money(1), "NPR 0.01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_Add(t *testing.T) {
	tests := []struct {
		name     string
		money1   Money
		money2   Money
		expected Money
	}{
		{"Zero + Zero", Money(0), Money(0), Money(0)},
		{"Positive + Positive", Money(10000), Money(5000), Money(15000)},
		{"Positive + Negative", Money(10000), Money(-3000), Money(7000)},
		{"Negative + Positive", Money(-5000), Money(8000), Money(3000)},
		{"Negative + Negative", Money(-5000), Money(-3000), Money(-8000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money1.Add(tt.money2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_Subtract(t *testing.T) {
	tests := []struct {
		name     string
		money1   Money
		money2   Money
		expected Money
	}{
		{"Zero - Zero", Money(0), Money(0), Money(0)},
		{"Positive - Positive", Money(10000), Money(3000), Money(7000)},
		{"Positive - Negative", Money(10000), Money(-3000), Money(13000)},
		{"Negative - Positive", Money(-5000), Money(3000), Money(-8000)},
		{"Negative - Negative", Money(-5000), Money(-8000), Money(3000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money1.Subtract(tt.money2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_Multiply(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		factor   float64
		expected Money
	}{
		{"Zero * factor", Money(0), 2.5, Money(0)},
		{"Money * zero", Money(10000), 0.0, Money(0)},
		{"Money * one", Money(10000), 1.0, Money(10000)},
		{"Money * integer", Money(10000), 2.0, Money(20000)},
		{"Money * decimal", Money(10000), 1.5, Money(15000)},
		{"Money * negative", Money(10000), -2.0, Money(-20000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.Multiply(tt.factor)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_MultiplyInt(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		factor   int
		expected Money
	}{
		{"Zero * factor", Money(0), 5, Money(0)},
		{"Money * zero", Money(10000), 0, Money(0)},
		{"Money * one", Money(10000), 1, Money(10000)},
		{"Money * positive", Money(10000), 3, Money(30000)},
		{"Money * negative", Money(10000), -2, Money(-20000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.MultiplyInt(tt.factor)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_Divide(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		factor   float64
		expected Money
	}{
		{"Zero / factor", Money(0), 2.0, Money(0)},
		{"Money / one", Money(10000), 1.0, Money(10000)},
		{"Money / two", Money(10000), 2.0, Money(5000)},
		{"Money / decimal", Money(10000), 2.5, Money(4000)},
		{"Money / negative", Money(10000), -2.0, Money(-5000)},
		{"Money / zero", Money(10000), 0.0, Money(0)}, // Division by zero returns 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.Divide(tt.factor)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_DivideInt(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		factor   int
		expected Money
	}{
		{"Zero / factor", Money(0), 2, Money(0)},
		{"Money / one", Money(10000), 1, Money(10000)},
		{"Money / two", Money(10000), 2, Money(5000)},
		{"Money / three", Money(15000), 3, Money(5000)},
		{"Money / negative", Money(10000), -2, Money(-5000)},
		{"Money / zero", Money(10000), 0, Money(0)}, // Division by zero returns 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.DivideInt(tt.factor)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected bool
	}{
		{"Zero", Money(0), true},
		{"Positive", Money(1), false},
		{"Negative", Money(-1), false},
		{"Large positive", Money(10000), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.IsZero()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_IsPositive(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected bool
	}{
		{"Zero", Money(0), false},
		{"Positive", Money(1), true},
		{"Negative", Money(-1), false},
		{"Large positive", Money(10000), true},
		{"Large negative", Money(-10000), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.IsPositive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_IsNegative(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected bool
	}{
		{"Zero", Money(0), false},
		{"Positive", Money(1), false},
		{"Negative", Money(-1), true},
		{"Large positive", Money(10000), false},
		{"Large negative", Money(-10000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money.IsNegative()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_Compare(t *testing.T) {
	tests := []struct {
		name     string
		money1   Money
		money2   Money
		expected int
	}{
		{"Equal", Money(10000), Money(10000), 0},
		{"Less than", Money(5000), Money(10000), -1},
		{"Greater than", Money(10000), Money(5000), 1},
		{"Zero vs positive", Money(0), Money(1), -1},
		{"Zero vs negative", Money(0), Money(-1), 1},
		{"Negative vs negative", Money(-5000), Money(-10000), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money1.Compare(tt.money2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_Equal(t *testing.T) {
	tests := []struct {
		name     string
		money1   Money
		money2   Money
		expected bool
	}{
		{"Equal positive", Money(10000), Money(10000), true},
		{"Equal zero", Money(0), Money(0), true},
		{"Equal negative", Money(-5000), Money(-5000), true},
		{"Not equal", Money(10000), Money(5000), false},
		{"Different signs", Money(10000), Money(-10000), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money1.Equal(tt.money2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_LessThan(t *testing.T) {
	tests := []struct {
		name     string
		money1   Money
		money2   Money
		expected bool
	}{
		{"Less than", Money(5000), Money(10000), true},
		{"Equal", Money(10000), Money(10000), false},
		{"Greater than", Money(10000), Money(5000), false},
		{"Negative less than positive", Money(-1000), Money(1000), true},
		{"Negative comparison", Money(-10000), Money(-5000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money1.LessThan(tt.money2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoney_GreaterThan(t *testing.T) {
	tests := []struct {
		name     string
		money1   Money
		money2   Money
		expected bool
	}{
		{"Greater than", Money(10000), Money(5000), true},
		{"Equal", Money(10000), Money(10000), false},
		{"Less than", Money(5000), Money(10000), false},
		{"Positive greater than negative", Money(1000), Money(-1000), true},
		{"Negative comparison", Money(-5000), Money(-10000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.money1.GreaterThan(tt.money2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test precision and edge cases
func TestMoney_Precision(t *testing.T) {
	t.Run("Small amounts precision", func(t *testing.T) {
		money := NewMoney(0.01)
		assert.Equal(t, Money(1), money)
		assert.Equal(t, 0.01, money.Rupees())
	})

	t.Run("Rounding behavior", func(t *testing.T) {
		// Test that we handle floating point precision correctly with proper rounding
		money := NewMoney(295.505) // Should round to 295.51
		expected := Money(29551)   // Properly rounded result
		assert.Equal(t, expected, money)
	})

	t.Run("Large amounts", func(t *testing.T) {
		largeAmount := NewMoney(999999.99)
		assert.Equal(t, Money(99999999), largeAmount)
		assert.Equal(t, 999999.99, largeAmount.Rupees())
	})
}

// Benchmark tests for performance
func BenchmarkNewMoney(b *testing.B) {
	for b.Loop() {
		NewMoney(295.50)
	}
}

func BenchmarkParseMoney(b *testing.B) {
	for b.Loop() {
		ParseMoney("NPR 1,850.75")
	}
}

func BenchmarkMoney_String(b *testing.B) {
	money := NewMoney(295.50)

	for b.Loop() {
		_ = money.String()
	}
}

func BenchmarkMoney_Add(b *testing.B) {
	money1 := NewMoney(295.50)
	money2 := NewMoney(302.00)

	for b.Loop() {
		_ = money1.Add(money2)
	}
}

func BenchmarkMoney_Multiply(b *testing.B) {
	money := NewMoney(295.50)

	for b.Loop() {
		_ = money.MultiplyInt(100)
	}
}

