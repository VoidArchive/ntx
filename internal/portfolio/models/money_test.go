package models

import (
	"testing"
)

func TestMoneyBasicOperations(t *testing.T) {
	m1 := NewMoney(1250.50)
	m2 := NewMoney(750.25)

	// Test addition
	sum := m1.Add(m2)
	expected := NewMoney(2000.75)
	if !sum.Equal(expected) {
		t.Errorf("addition failed: expected %s, got %s", expected.FormatSimple(), sum.FormatSimple())
	}

	// Test subtraction
	diff := m1.Sub(m2)
	expected = NewMoney(500.25)
	if !diff.Equal(expected) {
		t.Errorf("subtraction failed: expected %s, got %s", expected.FormatSimple(), diff.FormatSimple())
	}

	// Test multiplication
	doubled := m1.MultiplyInt(2)
	expected = NewMoney(2501.00)
	if !doubled.Equal(expected) {
		t.Errorf("multiplication failed: expected %s, got %s", expected.FormatSimple(), doubled.FormatSimple())
	}
}

func TestMoneyPrecision(t *testing.T) {
	// Test precision with paisa
	m := NewMoneyFromPaisa(125050) // Rs. 1250.50
	
	if m.Rupees() != 1250.50 {
		t.Errorf("rupees conversion failed: expected 1250.50, got %.2f", m.Rupees())
	}

	if m.Paisa != 125050 {
		t.Errorf("paisa storage failed: expected 125050, got %d", m.Paisa)
	}
}

func TestMoneyFormatting(t *testing.T) {
	tests := []struct {
		paisa    int64
		expected string
	}{
		{125050, "₹1,250.5"},
		{100000, "₹1,000"},
		{99950, "₹999.5"},
		{50, "₹0.5"},
		{0, "₹0"},
		{-125050, "-₹1,250.5"},
	}

	for _, tt := range tests {
		m := NewMoneyFromPaisa(tt.paisa)
		actual := m.FormatNPR()
		if actual != tt.expected {
			t.Errorf("formatting failed for %d paisa: expected %s, got %s", tt.paisa, tt.expected, actual)
		}
	}
}

func TestMoneyPercentageCalculations(t *testing.T) {
	current := NewMoney(1300.00)
	original := NewMoney(1000.00)

	pctChange := current.PercentageChange(original)
	expected := 30.0 // 30% increase

	if pctChange != expected {
		t.Errorf("percentage change failed: expected %.1f%%, got %.1f%%", expected, pctChange)
	}

	// Test percentage of amount
	thirty_percent := current.Percentage(30.0)
	expectedAmount := NewMoney(390.00)
	if !thirty_percent.Equal(expectedAmount) {
		t.Errorf("percentage amount failed: expected %s, got %s", 
			expectedAmount.FormatSimple(), thirty_percent.FormatSimple())
	}
}

func TestMoneyComparisons(t *testing.T) {
	m1 := NewMoney(1250.50)
	m2 := NewMoney(1250.50)
	m3 := NewMoney(1000.00)

	// Test equality
	if !m1.Equal(m2) {
		t.Error("equal amounts should be equal")
	}

	if m1.Equal(m3) {
		t.Error("different amounts should not be equal")
	}

	// Test comparisons
	if m1.Compare(m2) != 0 {
		t.Error("equal amounts should compare as 0")
	}

	if m1.Compare(m3) != 1 {
		t.Error("larger amount should compare as 1")
	}

	if m3.Compare(m1) != -1 {
		t.Error("smaller amount should compare as -1")
	}
}

func TestMoneyStringParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Money
	}{
		{"1250.50", NewMoney(1250.50)},
		{"₹1,250.50", NewMoney(1250.50)},
		{"Rs. 1250", NewMoney(1250.00)},
		{"Rs 999.99", NewMoney(999.99)},
		{"", Zero()},
	}

	for _, tt := range tests {
		actual, err := NewMoneyFromString(tt.input)
		if err != nil {
			t.Errorf("parsing '%s' failed: %v", tt.input, err)
			continue
		}

		if !actual.Equal(tt.expected) {
			t.Errorf("parsing '%s': expected %s, got %s", 
				tt.input, tt.expected.FormatSimple(), actual.FormatSimple())
		}
	}

	// Test invalid input
	_, err := NewMoneyFromString("invalid")
	if err == nil {
		t.Error("expected error for invalid input")
	}
}

func TestMoneyZeroAndNegative(t *testing.T) {
	zero := Zero()
	if !zero.IsZero() {
		t.Error("zero should be zero")
	}

	positive := NewMoney(100.00)
	if !positive.IsPositive() {
		t.Error("positive amount should be positive")
	}

	negative := NewMoney(-100.00)
	if !negative.IsNegative() {
		t.Error("negative amount should be negative")
	}

	// Test absolute value
	abs := negative.Abs()
	if !abs.Equal(positive) {
		t.Error("absolute value of negative should equal positive")
	}
}

func TestMoneySum(t *testing.T) {
	amounts := []Money{
		NewMoney(100.00),
		NewMoney(200.50),
		NewMoney(300.25),
	}

	total := Sum(amounts...)
	expected := NewMoney(600.75)

	if !total.Equal(expected) {
		t.Errorf("sum failed: expected %s, got %s", 
			expected.FormatSimple(), total.FormatSimple())
	}

	// Test empty sum
	emptySum := Sum()
	if !emptySum.IsZero() {
		t.Error("empty sum should be zero")
	}
}