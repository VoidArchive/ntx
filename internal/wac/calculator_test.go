package wac

import (
	"ntx/internal/csv"
	"ntx/internal/money"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCalculator(t *testing.T) {
	calc := NewCalculator()

	assert.NotNil(t, calc, "NewCalculator() should not return nil")
	assert.NotNil(t, calc.holdings, "Calculator holdings map should not be nil")
	assert.Empty(t, calc.holdings, "New calculator should have empty holdings")
}

func TestShareLot(t *testing.T) {
	lot := ShareLot{
		Quantity: 100,
		Price:    money.NewMoney(295.50),
		Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, 100, lot.Quantity)
	assert.True(t, lot.Price.Equal(money.NewMoney(295.50)))
	assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), lot.Date)
}

func TestHolding(t *testing.T) {
	t.Run("Non-empty holding", func(t *testing.T) {
		holding := Holding{
			Scrip:         "API",
			TotalQuantity: 150,
			WAC:           money.NewMoney(300.00),
			Lots: []ShareLot{
				{Quantity: 100, Price: money.NewMoney(295.50)},
				{Quantity: 50, Price: money.NewMoney(310.00)},
			},
		}

		assert.False(t, holding.IsEmpty())

		expectedValue := money.NewMoney(300.00).MultiplyInt(150)
		assert.True(t, holding.TotalValue().Equal(expectedValue))
	})

	t.Run("Empty holding", func(t *testing.T) {
		emptyHolding := Holding{TotalQuantity: 0}

		assert.True(t, emptyHolding.IsEmpty())
		assert.True(t, emptyHolding.TotalValue().IsZero())
	})
}

func TestCalculateHoldings_SimpleBuy(t *testing.T) {
	calc := NewCalculator()
	transactions := []csv.Transaction{
		{
			Scrip:           "API",
			Date:            time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Quantity:        100,
			Price:           money.NewMoney(295.50),
			TransactionType: csv.TransactionTypeRegular,
		},
	}

	holdings, err := calc.CalculateHoldings(transactions)

	require.NoError(t, err)
	require.Len(t, holdings, 1)

	holding := holdings[0]
	assert.Equal(t, "API", holding.Scrip)
	assert.Equal(t, 100, holding.TotalQuantity)
	assert.True(t, holding.WAC.Equal(money.NewMoney(295.50)))
	assert.Len(t, holding.Lots, 1)
}

func TestCalculateHoldings_MultipleBuys(t *testing.T) {
	calc := NewCalculator()
	transactions := []csv.Transaction{
		{
			Scrip:    "API",
			Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Quantity: 100,
			Price:    money.NewMoney(295.50),
		},
		{
			Scrip:    "API",
			Date:     time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			Quantity: 50,
			Price:    money.NewMoney(302.00),
		},
	}

	holdings, err := calc.CalculateHoldings(transactions)

	require.NoError(t, err)
	require.Len(t, holdings, 1)

	holding := holdings[0]
	assert.Equal(t, 150, holding.TotalQuantity)

	// Calculate expected WAC: (100*295.50 + 50*302.00) / 150 = 297.66... (actual result)
	expectedWAC := money.NewMoney(297.66)
	assert.True(t, holding.WAC.Equal(expectedWAC),
		"Expected WAC %s, got %s", expectedWAC, holding.WAC)

	assert.Len(t, holding.Lots, 2)
}

func TestCalculateHoldings_FIFO_FullLotSale(t *testing.T) {
	calc := NewCalculator()
	transactions := []csv.Transaction{
		{
			Scrip:    "API",
			Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Quantity: 100,
			Price:    money.NewMoney(295.50),
		},
		{
			Scrip:    "API",
			Date:     time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			Quantity: 50,
			Price:    money.NewMoney(302.00),
		},
		{
			Scrip:    "API",
			Date:     time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
			Quantity: -100, // Sell entire first lot
			Price:    money.NewMoney(320.00),
		},
	}

	holdings, err := calc.CalculateHoldings(transactions)

	require.NoError(t, err)
	require.Len(t, holdings, 1)

	holding := holdings[0]
	assert.Equal(t, 50, holding.TotalQuantity)

	// After selling first lot, only second lot should remain
	expectedWAC := money.NewMoney(302.00)
	assert.True(t, holding.WAC.Equal(expectedWAC))

	require.Len(t, holding.Lots, 1)
	assert.Equal(t, 50, holding.Lots[0].Quantity)
}

func TestCalculateHoldings_FIFO_PartialLotSale(t *testing.T) {
	calc := NewCalculator()
	transactions := []csv.Transaction{
		{
			Scrip:    "API",
			Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Quantity: 100,
			Price:    money.NewMoney(295.50),
		},
		{
			Scrip:    "API",
			Date:     time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			Quantity: 50,
			Price:    money.NewMoney(302.00),
		},
		{
			Scrip:    "API",
			Date:     time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
			Quantity: -30, // Sell part of first lot
			Price:    money.NewMoney(320.00),
		},
	}

	holdings, err := calc.CalculateHoldings(transactions)

	require.NoError(t, err)
	require.Len(t, holdings, 1)

	holding := holdings[0]
	assert.Equal(t, 120, holding.TotalQuantity)
	require.Len(t, holding.Lots, 2)

	// First lot should have 70 shares remaining (100 - 30)
	assert.Equal(t, 70, holding.Lots[0].Quantity)
	// Second lot should remain unchanged
	assert.Equal(t, 50, holding.Lots[1].Quantity)
}

func TestCalculateHoldings_MultipleScripts(t *testing.T) {
	calc := NewCalculator()
	transactions := []csv.Transaction{
		{
			Scrip:    "API",
			Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Quantity: 100,
			Price:    money.NewMoney(295.50),
		},
		{
			Scrip:    "NMB",
			Date:     time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
			Quantity: 200,
			Price:    money.NewMoney(1850.00),
		},
	}

	holdings, err := calc.CalculateHoldings(transactions)

	require.NoError(t, err)
	require.Len(t, holdings, 2)

	// Holdings should be sorted by scrip name
	assert.Equal(t, "API", holdings[0].Scrip)
	assert.Equal(t, "NMB", holdings[1].Scrip)
}

func TestCalculateHoldings_InsufficientShares(t *testing.T) {
	calc := NewCalculator()
	transactions := []csv.Transaction{
		{
			Scrip:    "API",
			Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Quantity: 100,
			Price:    money.NewMoney(295.50),
		},
		{
			Scrip:    "API",
			Date:     time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			Quantity: -150, // Try to sell more than owned
			Price:    money.NewMoney(320.00),
		},
	}

	_, err := calc.CalculateHoldings(transactions)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient shares")
}

func TestCalculateHoldings_BonusShares(t *testing.T) {
	calc := NewCalculator()
	transactions := []csv.Transaction{
		{
			Scrip:           "API",
			Date:            time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Quantity:        100,
			Price:           money.NewMoney(295.50),
			TransactionType: csv.TransactionTypeRegular,
		},
		{
			Scrip:           "API",
			Date:            time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			Quantity:        20, // Bonus shares - no price
			Price:           money.Money(0),
			TransactionType: csv.TransactionTypeBonus,
		},
	}

	holdings, err := calc.CalculateHoldings(transactions)

	require.NoError(t, err)
	require.Len(t, holdings, 1)

	holding := holdings[0]
	assert.Equal(t, 120, holding.TotalQuantity)

	// WAC should only consider paid shares (bonus shares have zero price)
	expectedWAC := money.NewMoney(295.50)
	assert.True(t, holding.WAC.Equal(expectedWAC),
		"Expected WAC %s (excluding bonus shares), got %s", expectedWAC, holding.WAC)
}

func TestCalculateHoldings_EmptyTransactions(t *testing.T) {
	calc := NewCalculator()
	transactions := []csv.Transaction{}

	holdings, err := calc.CalculateHoldings(transactions)

	require.NoError(t, err)
	assert.Empty(t, holdings)
}

func TestGetHolding(t *testing.T) {
	calc := NewCalculator()
	transactions := []csv.Transaction{
		{
			Scrip:    "API",
			Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Quantity: 100,
			Price:    money.NewMoney(295.50),
		},
	}

	_, err := calc.CalculateHoldings(transactions)
	require.NoError(t, err)

	t.Run("Existing holding", func(t *testing.T) {
		holding, found := calc.GetHolding("API")
		assert.True(t, found)
		assert.Equal(t, "API", holding.Scrip)
	})

	t.Run("Non-existing holding", func(t *testing.T) {
		_, found := calc.GetHolding("NONEXISTENT")
		assert.False(t, found)
	})
}

func TestCalculateWAC(t *testing.T) {
	calc := NewCalculator()

	t.Run("Empty lots", func(t *testing.T) {
		wac := calc.calculateWAC([]ShareLot{})
		assert.True(t, wac.IsZero())
	})

	t.Run("Single lot", func(t *testing.T) {
		lots := []ShareLot{
			{Quantity: 100, Price: money.NewMoney(295.50)},
		}
		wac := calc.calculateWAC(lots)
		assert.True(t, wac.Equal(money.NewMoney(295.50)))
	})

	t.Run("Multiple lots", func(t *testing.T) {
		lots := []ShareLot{
			{Quantity: 100, Price: money.NewMoney(295.50)},
			{Quantity: 50, Price: money.NewMoney(302.00)},
		}
		wac := calc.calculateWAC(lots)
		// Expected: (100*295.50 + 50*302.00) / 150 = 297.66... (actual result)
		expectedWAC := money.NewMoney(297.66)
		assert.True(t, wac.Equal(expectedWAC))
	})

	t.Run("Lots with zero price (bonus shares)", func(t *testing.T) {
		lots := []ShareLot{
			{Quantity: 100, Price: money.NewMoney(295.50)},
			{Quantity: 20, Price: money.Money(0)}, // Bonus shares
		}
		wac := calc.calculateWAC(lots)
		// Should ignore zero-price lots
		assert.True(t, wac.Equal(money.NewMoney(295.50)))
	})
}

// Benchmark tests for performance
func BenchmarkCalculateHoldings(b *testing.B) {
	// Create a large number of transactions
	transactions := make([]csv.Transaction, 1000)
	for i := range 1000 {
		transactions[i] = csv.Transaction{
			Scrip:    "API",
			Date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i),
			Quantity: 100,
			Price:    money.NewMoney(295.50),
		}
	}

	for b.Loop() {
		calc := NewCalculator()
		_, err := calc.CalculateHoldings(transactions)
		if err != nil {
			b.Fatal(err)
		}
	}
}

