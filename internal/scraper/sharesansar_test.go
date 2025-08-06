package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/voidarchive/ntx/internal/domain/models"
)

func TestShareSansarScraper_GetAllQuotes(t *testing.T) {
	// Skip in short test mode since this makes real HTTP requests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	scraper := NewShareSansarScraper()
	quotes, err := scraper.GetAllQuotes()

	require.NoError(t, err, "GetAllQuotes should not return an error")
	assert.Greater(t, len(quotes), 200, "Should have at least 200+ stocks")

	// Test that we got some well-known stocks
	knownStocks := []string{"NABIL", "ADBL", "HIDCL", "API", "CHCL"}
	foundStocks := make(map[string]*models.Quote)

	for _, quote := range quotes {
		// Basic validation of each quote
		assert.NotEmpty(t, quote.Symbol, "Symbol should not be empty")
		assert.GreaterOrEqual(t, quote.LTP, 0.0, "LTP should be non-negative for %s", quote.Symbol)
		assert.GreaterOrEqual(t, quote.Open, 0.0, "Open should be non-negative for %s", quote.Symbol)
		assert.GreaterOrEqual(t, quote.High, 0.0, "High should be non-negative for %s", quote.Symbol)
		assert.GreaterOrEqual(t, quote.Low, 0.0, "Low should be non-negative for %s", quote.Symbol)
		assert.GreaterOrEqual(t, quote.Volume, 0.0, "Volume should be non-negative for %s", quote.Symbol)
		assert.GreaterOrEqual(t, quote.PrevClose, 0.0, "PrevClose should be non-negative for %s", quote.Symbol)

		// High should be >= Low
		if quote.High > 0 && quote.Low > 0 {
			assert.GreaterOrEqual(t, quote.High, quote.Low, "High should be >= Low for %s", quote.Symbol)
		}

		// Collect known stocks
		for _, known := range knownStocks {
			if quote.Symbol == known {
				foundStocks[known] = quote
			}
		}
	}

	// Verify we found major stocks
	for _, stock := range knownStocks {
		quote, found := foundStocks[stock]
		assert.True(t, found, "Should find %s in results", stock)
		if found {
			assert.Greater(t, quote.LTP, 0.0, "%s should have positive LTP", stock)
		}
	}

	t.Logf("Successfully scraped %d stocks", len(quotes))
}

func TestShareSansarScraper_GetQuote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	scraper := NewShareSansarScraper()

	// Test getting a specific stock
	quote, err := scraper.GetQuote("NABIL")
	require.NoError(t, err, "GetQuote should not return an error for NABIL")
	require.NotNil(t, quote, "Quote should not be nil")

	assert.Equal(t, "NABIL", quote.Symbol, "Symbol should match requested symbol")
	assert.Greater(t, quote.LTP, 0.0, "NABIL should have positive LTP")
	assert.Greater(t, quote.Open, 0.0, "NABIL should have positive Open")
	assert.Greater(t, quote.High, 0.0, "NABIL should have positive High")
	assert.Greater(t, quote.Low, 0.0, "NABIL should have positive Low")
	assert.GreaterOrEqual(t, quote.Volume, 0.0, "NABIL should have non-negative Volume")
	assert.Greater(t, quote.PrevClose, 0.0, "NABIL should have positive PrevClose")

	// Test market logic: High >= Low
	assert.GreaterOrEqual(t, quote.High, quote.Low, "High should be >= Low")

	t.Logf("NABIL: LTP=%.2f, Open=%.2f, High=%.2f, Low=%.2f, Volume=%.0f",
		quote.LTP, quote.Open, quote.High, quote.Low, quote.Volume)
}

func TestShareSansarScraper_GetQuote_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	scraper := NewShareSansarScraper()

	// Test with non-existent symbol
	quote, err := scraper.GetQuote("NONEXISTENT")
	assert.Error(t, err, "Should return error for non-existent symbol")
	assert.Nil(t, quote, "Quote should be nil for non-existent symbol")
	assert.Contains(t, err.Error(), "not found", "Error should mention 'not found'")
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		hasError bool
	}{
		{
			name:     "Simple number",
			input:    "123.45",
			expected: 123.45,
			hasError: false,
		},
		{
			name:     "Number with commas",
			input:    "1,234.56",
			expected: 1234.56,
			hasError: false,
		},
		{
			name:     "Number with spaces",
			input:    " 789.01 ",
			expected: 789.01,
			hasError: false,
		},
		{
			name:     "Large number with commas",
			input:    "12,345,678.90",
			expected: 12345678.90,
			hasError: false,
		},
		{
			name:     "Zero",
			input:    "0.00",
			expected: 0.00,
			hasError: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
			hasError: true,
		},
		{
			name:     "Only spaces",
			input:    "   ",
			expected: 0,
			hasError: true,
		},
		{
			name:     "Invalid text",
			input:    "abc",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFloat(tt.input)

			if tt.hasError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "Expected %f but got %f for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestNewShareSansarScraper(t *testing.T) {
	scraper := NewShareSansarScraper()

	assert.NotNil(t, scraper, "Scraper should not be nil")
	assert.NotNil(t, scraper.collector, "Collector should not be nil")
}

// Benchmark tests to measure performance
func BenchmarkGetAllQuotes(b *testing.B) {
	scraper := NewShareSansarScraper()

	for b.Loop() {
		_, err := scraper.GetAllQuotes()
		if err != nil {
			b.Fatalf("GetAllQuotes failed: %v", err)
		}
	}
}

func BenchmarkGetQuote(b *testing.B) {
	scraper := NewShareSansarScraper()

	for b.Loop() {
		_, err := scraper.GetQuote("NABIL")
		if err != nil {
			b.Fatalf("GetQuote failed: %v", err)
		}
	}
}

func BenchmarkParseFloat(b *testing.B) {
	testCases := []string{
		"123.45",
		"1,234.56",
		" 789.01 ",
		"12,345,678.90",
	}

	for b.Loop() {
		for _, tc := range testCases {
			_, _ = parseFloat(tc)
		}
	}
}
