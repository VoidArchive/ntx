// Package domain contains tests for the FIFO queue implementation
package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFIFOQueue_Basic(t *testing.T) {
	fq := NewFIFOQueue("ADBL")

	assert.Equal(t, "ADBL", fq.StockSymbol)
	assert.Equal(t, 0, fq.TotalShares())
	assert.Equal(t, Zero(), fq.TotalCost())
	assert.Equal(t, Zero(), fq.WeightedAverageCost())
	assert.True(t, fq.IsEmpty())
}

func TestFIFOQueue_SingleBuy(t *testing.T) {
	fq := NewFIFOQueue("ADBL")
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	err := fq.Buy(100, NewMoney(500.0), date)
	require.NoError(t, err)

	assert.Equal(t, 100, fq.TotalShares())
	assert.Equal(t, NewMoney(50000.0), fq.TotalCost())
	assert.Equal(t, NewMoney(500.0), fq.WeightedAverageCost())
	assert.False(t, fq.IsEmpty())

	lots := fq.GetLots()
	require.Len(t, lots, 1)
	assert.Equal(t, 100, lots[0].Quantity)
	assert.Equal(t, NewMoney(500.0), lots[0].Price)
	assert.Equal(t, date, lots[0].Date)
}

func TestFIFOQueue_MultipleBuys(t *testing.T) {
	fq := NewFIFOQueue("ADBL")

	// First buy: 100 shares @ Rs.500
	err := fq.Buy(100, NewMoney(500.0), time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	// Second buy: 50 shares @ Rs.600
	err = fq.Buy(50, NewMoney(600.0), time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	assert.Equal(t, 150, fq.TotalShares())
	assert.Equal(t, NewMoney(80000.0), fq.TotalCost()) // 50000 + 30000

	// WAC should be 80000/150 = 533.33
	expectedWAC := NewMoney(533.33)
	actualWAC := fq.WeightedAverageCost()
	assert.True(t, actualWAC.Subtract(expectedWAC).Abs().LessThan(NewMoney(0.01)),
		"Expected WAC ~%s, got %s", expectedWAC.String(), actualWAC.String())

	lots := fq.GetLots()
	require.Len(t, lots, 2)

	// First lot (oldest)
	assert.Equal(t, 100, lots[0].Quantity)
	assert.Equal(t, NewMoney(500.0), lots[0].Price)

	// Second lot
	assert.Equal(t, 50, lots[1].Quantity)
	assert.Equal(t, NewMoney(600.0), lots[1].Price)
}

func TestFIFOQueue_PartialSale(t *testing.T) {
	fq := NewFIFOQueue("ADBL")

	// Buy 100 shares @ Rs.500 on Jan 15
	buyDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	err := fq.Buy(100, NewMoney(500.0), buyDate)
	require.NoError(t, err)

	// Sell 60 shares @ Rs.550 on March 15 (59 days later - short term)
	saleDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	result, err := fq.Sell(60, NewMoney(550.0), saleDate)
	require.NoError(t, err)

	// Verify sale result
	assert.Equal(t, 60, result.SharesSold)
	assert.Equal(t, NewMoney(33000.0), result.TotalProceeds)  // 60 * 550
	assert.Equal(t, NewMoney(30000.0), result.TotalCostBasis) // 60 * 500
	assert.Equal(t, NewMoney(3000.0), result.TotalGainLoss)   // 33000 - 30000

	require.Len(t, result.RealizedGains, 1)
	gain := result.RealizedGains[0]
	assert.Equal(t, "ADBL", gain.StockSymbol)
	assert.Equal(t, 60, gain.Quantity)
	assert.Equal(t, NewMoney(550.0), gain.SalePrice)
	assert.Equal(t, NewMoney(500.0), gain.CostBasis)
	assert.Equal(t, NewMoney(3000.0), gain.GainLoss)
	assert.Equal(t, 59, gain.HoldingDays)
	assert.False(t, gain.IsLongTerm) // < 365 days

	// Verify remaining holdings
	assert.Equal(t, 40, fq.TotalShares())                      // 100 - 60
	assert.Equal(t, NewMoney(20000.0), fq.TotalCost())         // 40 * 500
	assert.Equal(t, NewMoney(500.0), fq.WeightedAverageCost()) // Still Rs.500
}

func TestFIFOQueue_MultipleLotSale(t *testing.T) {
	fq := NewFIFOQueue("NABIL")

	// First buy: 100 shares @ Rs.1000 on Jan 1
	err := fq.Buy(100, NewMoney(1000.0), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	// Second buy: 150 shares @ Rs.1200 on Feb 1
	err = fq.Buy(150, NewMoney(1200.0), time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	// Sell 180 shares @ Rs.1300 on June 1
	// This should consume all 100 from first lot + 80 from second lot
	saleDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	result, err := fq.Sell(180, NewMoney(1300.0), saleDate)
	require.NoError(t, err)

	// Verify sale spans multiple lots
	require.Len(t, result.RealizedGains, 2)

	// First realized gain (from first lot)
	gain1 := result.RealizedGains[0]
	assert.Equal(t, 100, gain1.Quantity)
	assert.Equal(t, NewMoney(1000.0), gain1.CostBasis)
	assert.Equal(t, NewMoney(30000.0), gain1.GainLoss) // (1300-1000)*100
	assert.Equal(t, 151, gain1.HoldingDays)            // Jan 1 to June 1
	assert.False(t, gain1.IsLongTerm)

	// Second realized gain (from second lot)
	gain2 := result.RealizedGains[1]
	assert.Equal(t, 80, gain2.Quantity)
	assert.Equal(t, NewMoney(1200.0), gain2.CostBasis)
	assert.Equal(t, NewMoney(8000.0), gain2.GainLoss) // (1300-1200)*80
	assert.Equal(t, 120, gain2.HoldingDays)           // Feb 1 to June 1
	assert.False(t, gain2.IsLongTerm)

	// Total result
	assert.Equal(t, 180, result.SharesSold)
	assert.Equal(t, NewMoney(234000.0), result.TotalProceeds)  // 180 * 1300
	assert.Equal(t, NewMoney(196000.0), result.TotalCostBasis) // 100*1000 + 80*1200
	assert.Equal(t, NewMoney(38000.0), result.TotalGainLoss)   // 30000 + 8000

	// Verify remaining holdings (70 shares from second lot)
	assert.Equal(t, 70, fq.TotalShares())
	assert.Equal(t, NewMoney(84000.0), fq.TotalCost()) // 70 * 1200
	assert.Equal(t, NewMoney(1200.0), fq.WeightedAverageCost())

	lots := fq.GetLots()
	require.Len(t, lots, 1)
	assert.Equal(t, 70, lots[0].Quantity)
	assert.Equal(t, NewMoney(1200.0), lots[0].Price)
}

func TestFIFOQueue_CompleteSellOff(t *testing.T) {
	fq := NewFIFOQueue("NIFRA")

	// Buy some shares
	err := fq.Buy(100, NewMoney(300.0), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	// Sell all shares
	result, err := fq.Sell(100, NewMoney(350.0), time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	// Verify complete selloff
	assert.Equal(t, 0, fq.TotalShares())
	assert.Equal(t, Zero(), fq.TotalCost())
	assert.Equal(t, Zero(), fq.WeightedAverageCost())
	assert.True(t, fq.IsEmpty())

	// Verify realized gain
	require.Len(t, result.RealizedGains, 1)
	gain := result.RealizedGains[0]
	assert.Equal(t, 334, gain.HoldingDays) // ~11 months
	assert.False(t, gain.IsLongTerm)       // 334 days < 365 days
}

func TestFIFOQueue_LongTermGain(t *testing.T) {
	fq := NewFIFOQueue("NIFRA")

	// Buy shares
	buyDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	err := fq.Buy(100, NewMoney(300.0), buyDate)
	require.NoError(t, err)

	// Sell after more than 1 year
	saleDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC) // ~17 months later
	result, err := fq.Sell(100, NewMoney(350.0), saleDate)
	require.NoError(t, err)

	require.Len(t, result.RealizedGains, 1)
	gain := result.RealizedGains[0]
	assert.True(t, gain.HoldingDays > 365)
	assert.True(t, gain.IsLongTerm)
}

func TestFIFOQueue_BonusShares(t *testing.T) {
	fq := NewFIFOQueue("EBL")

	// Buy 100 shares @ Rs.600
	err := fq.Buy(100, NewMoney(600.0), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	// 20% bonus (20 shares @ Rs.0)
	err = fq.Buy(20, Zero(), time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	// Verify WAC adjustment
	assert.Equal(t, 120, fq.TotalShares())
	assert.Equal(t, NewMoney(60000.0), fq.TotalCost())         // Same total cost
	assert.Equal(t, NewMoney(500.0), fq.WeightedAverageCost()) // 60000/120 = 500

	lots := fq.GetLots()
	require.Len(t, lots, 2)
	assert.Equal(t, NewMoney(600.0), lots[0].Price) // Original lot unchanged
	assert.Equal(t, Zero(), lots[1].Price)          // Bonus lot
}

func TestFIFOQueue_ErrorCases(t *testing.T) {
	fq := NewFIFOQueue("TEST")

	// Test invalid buy parameters
	err := fq.Buy(0, NewMoney(100.0), time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")

	err = fq.Buy(-10, NewMoney(100.0), time.Now())
	assert.Error(t, err)

	err = fq.Buy(10, NewMoney(-100.0), time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "price cannot be negative")

	// Test selling more than available
	err = fq.Buy(50, NewMoney(100.0), time.Now())
	require.NoError(t, err)

	_, err = fq.Sell(60, NewMoney(110.0), time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot sell 60 shares, only 50 available")

	// Test invalid sell parameters
	_, err = fq.Sell(0, NewMoney(110.0), time.Now())
	assert.Error(t, err)

	_, err = fq.Sell(10, NewMoney(-110.0), time.Now())
	assert.Error(t, err)
}
