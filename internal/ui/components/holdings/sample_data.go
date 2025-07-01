/*
NTX Portfolio Management TUI - Holdings Sample Data

Sample data generator for testing holdings display component with realistic
NEPSE stock data and calculated P/L metrics.

Used for development, testing, and demonstration purposes to ensure
component renders correctly across different scenarios.
*/

package holdings

import (
	"fmt"
	"ntx/internal/ui/themes"
)

// GenerateSampleHoldings creates realistic NEPSE portfolio data
// Includes various P/L scenarios for comprehensive component testing
func GenerateSampleHoldings() []Holding {
	return []Holding{
		{
			Symbol:        "NABIL",
			Quantity:      75,
			AvgCost:       126700,  // Rs.1,267.00 in paisa
			CurrentLTP:    132000,  // Rs.1,320.00 in paisa
			MarketValue:   9900000, // 75 * Rs.1,320.00 in paisa
			DayPL:         198000,  // +Rs.1,980 daily gain
			TotalPL:       397500,  // +Rs.3,975 total profit
			PercentChange: 4.2,     // +4.2% gain
			RSI:           67.0,    // Overbought territory
		},
		{
			Symbol:        "HIDCL",
			Quantity:      100,
			AvgCost:       42000,   // Rs.420.00 in paisa
			CurrentLTP:    44600,   // Rs.446.00 in paisa
			MarketValue:   4460000, // 100 * Rs.446.00 in paisa
			DayPL:         120000,  // +Rs.1,200 daily gain
			TotalPL:       260000,  // +Rs.2,600 total profit
			PercentChange: 6.2,     // +6.2% gain
			RSI:           45.0,    // Neutral zone
		},
		{
			Symbol:        "KTM",
			Quantity:      40,
			AvgCost:       89000,   // Rs.890.00 in paisa
			CurrentLTP:    92000,   // Rs.920.00 in paisa
			MarketValue:   3680000, // 40 * Rs.920.00 in paisa
			DayPL:         48000,   // +Rs.480 daily gain
			TotalPL:       120000,  // +Rs.1,200 total profit
			PercentChange: 3.4,     // +3.4% gain
			RSI:           52.0,    // Slightly bullish
		},
		{
			Symbol:        "EBL",
			Quantity:      30,
			AvgCost:       68000,   // Rs.680.00 in paisa
			CurrentLTP:    71000,   // Rs.710.00 in paisa
			MarketValue:   2130000, // 30 * Rs.710.00 in paisa
			DayPL:         36000,   // +Rs.360 daily gain
			TotalPL:       90000,   // +Rs.900 total profit
			PercentChange: 4.4,     // +4.4% gain
			RSI:           41.0,    // Oversold recovery
		},
		{
			Symbol:        "SANIMA",
			Quantity:      25,
			AvgCost:       55000,   // Rs.550.00 in paisa
			CurrentLTP:    52000,   // Rs.520.00 in paisa (loss position)
			MarketValue:   1300000, // 25 * Rs.520.00 in paisa
			DayPL:         -15000,  // -Rs.150 daily loss
			TotalPL:       -75000,  // -Rs.750 total loss
			PercentChange: -5.5,    // -5.5% loss
			RSI:           28.0,    // Heavily oversold
		},
		{
			Symbol:        "ADBL",
			Quantity:      50,
			AvgCost:       34500,   // Rs.345.00 in paisa
			CurrentLTP:    36200,   // Rs.362.00 in paisa
			MarketValue:   1810000, // 50 * Rs.362.00 in paisa
			DayPL:         25000,   // +Rs.250 daily gain
			TotalPL:       85000,   // +Rs.850 total profit
			PercentChange: 4.9,     // +4.9% gain
			RSI:           58.0,    // Moderately bullish
		},
		{
			Symbol:        "PCBL",
			Quantity:      80,
			AvgCost:       28000,   // Rs.280.00 in paisa
			CurrentLTP:    27200,   // Rs.272.00 in paisa (slight loss)
			MarketValue:   2176000, // 80 * Rs.272.00 in paisa
			DayPL:         -32000,  // -Rs.320 daily loss
			TotalPL:       -64000,  // -Rs.640 total loss
			PercentChange: -2.9,    // -2.9% loss
			RSI:           32.0,    // Oversold
		},
		{
			Symbol:        "GBIME",
			Quantity:      60,
			AvgCost:       31200,   // Rs.312.00 in paisa
			CurrentLTP:    33500,   // Rs.335.00 in paisa
			MarketValue:   2010000, // 60 * Rs.335.00 in paisa
			DayPL:         42000,   // +Rs.420 daily gain
			TotalPL:       138000,  // +Rs.1,380 total profit
			PercentChange: 7.4,     // +7.4% gain
			RSI:           61.0,    // Bullish
		},
		{
			Symbol:        "NLICL",
			Quantity:      35,
			AvgCost:       78000,   // Rs.780.00 in paisa
			CurrentLTP:    75500,   // Rs.755.00 in paisa (loss)
			MarketValue:   2642500, // 35 * Rs.755.00 in paisa
			DayPL:         -17500,  // -Rs.175 daily loss
			TotalPL:       -87500,  // -Rs.875 total loss
			PercentChange: -3.2,    // -3.2% loss
			RSI:           39.0,    // Bearish
		},
		{
			Symbol:        "UPPER",
			Quantity:      20,
			AvgCost:       45600,  // Rs.456.00 in paisa
			CurrentLTP:    48200,  // Rs.482.00 in paisa
			MarketValue:   964000, // 20 * Rs.482.00 in paisa
			DayPL:         10400,  // +Rs.104 daily gain
			TotalPL:       52000,  // +Rs.520 total profit
			PercentChange: 5.7,    // +5.7% gain
			RSI:           72.0,   // Overbought
		},
	}
}

// GenerateEmptyHoldings returns empty holdings array for testing empty state
func GenerateEmptyHoldings() []Holding {
	return []Holding{}
}

// GenerateLargePortfolio creates large portfolio for performance testing
// Tests component performance with 50+ holdings
func GenerateLargePortfolio() []Holding {
	baseHoldings := GenerateSampleHoldings()
	var holdings []Holding

	// Duplicate with variations to create large portfolio
	for i := 0; i < 10; i++ {
		for j, holding := range baseHoldings {
			variation := holding
			// Vary symbol to create unique entries
			variation.Symbol = fmt.Sprintf("%s%d", holding.Symbol, i+1)
			// Vary quantities and prices slightly
			variation.Quantity += int64(i * 5)
			variation.AvgCost += int64(i * 1000)
			variation.CurrentLTP += int64(i * 1200)
			// Recalculate metrics
			variation.MarketValue = variation.Quantity * variation.CurrentLTP
			variation.TotalPL = variation.MarketValue - (variation.Quantity * variation.AvgCost)
			if variation.Quantity*variation.AvgCost > 0 {
				variation.PercentChange = float64(variation.TotalPL) / float64(variation.Quantity*variation.AvgCost) * 100
			}
			variation.DayPL = variation.TotalPL / 10 // Rough daily estimate
			variation.RSI = float64(30 + (j*10)%40)  // Vary RSI values

			holdings = append(holdings, variation)
		}
	}

	return holdings
}

// GenerateMixedPortfolio creates portfolio with various P/L scenarios
// Includes strong gains, losses, and neutral positions for color testing
func GenerateMixedPortfolio() []Holding {
	return []Holding{
		{
			Symbol:        "HUGE_GAIN",
			Quantity:      100,
			AvgCost:       10000, // Rs.100.00
			CurrentLTP:    25000, // Rs.250.00
			MarketValue:   2500000,
			DayPL:         500000,  // +Rs.5,000 massive daily gain
			TotalPL:       1500000, // +Rs.15,000 total profit
			PercentChange: 150.0,   // +150% gain
			RSI:           85.0,
		},
		{
			Symbol:        "BIG_LOSS",
			Quantity:      50,
			AvgCost:       80000, // Rs.800.00
			CurrentLTP:    40000, // Rs.400.00
			MarketValue:   2000000,
			DayPL:         -100000,  // -Rs.1,000 daily loss
			TotalPL:       -2000000, // -Rs.20,000 total loss
			PercentChange: -50.0,    // -50% loss
			RSI:           15.0,
		},
		{
			Symbol:        "NEUTRAL",
			Quantity:      75,
			AvgCost:       50000, // Rs.500.00
			CurrentLTP:    50100, // Rs.501.00
			MarketValue:   3757500,
			DayPL:         3750, // +Rs.37.50 small daily gain
			TotalPL:       7500, // +Rs.75.00 minimal profit
			PercentChange: 0.2,  // +0.2% tiny gain
			RSI:           50.0,
		},
		{
			Symbol:        "VOLATILE",
			Quantity:      200,
			AvgCost:       25000, // Rs.250.00
			CurrentLTP:    24000, // Rs.240.00
			MarketValue:   4800000,
			DayPL:         -40000,  // -Rs.400 daily loss
			TotalPL:       -200000, // -Rs.2,000 total loss
			PercentChange: -4.0,    // -4.0% loss
			RSI:           35.0,
		},
	}
}

// CreateTestHoldingsDisplay creates configured display for testing
// Sets up display with sample data and proper theme integration
func CreateTestHoldingsDisplay(theme themes.Theme, width, height int) *HoldingsDisplay {
	display := NewHoldingsDisplay(theme)
	display.SetTerminalSize(width, height)
	display.UpdateHoldings(GenerateSampleHoldings())
	return display
}

// DemoHoldingsComponent demonstrates component functionality
// Shows various layouts and features for development verification
func DemoHoldingsComponent(theme themes.Theme) string {
	display := CreateTestHoldingsDisplay(theme, 120, 40)
	return display.Render()
}
