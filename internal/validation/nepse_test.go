package validation

import (
	"testing"
	"time"

	"ntx/internal/portfolio/models"
)

func TestNEPSEValidator_ValidateSymbol(t *testing.T) {
	validator := NewNEPSEValidator()

	tests := []struct {
		name    string
		symbol  string
		wantErr bool
	}{
		{"Valid bank symbol", "NABIL", false},
		{"Valid insurance symbol", "NICA", false},
		{"Valid hydro symbol", "HIDCL", false},
		{"Empty symbol", "", true},
		{"Too short", "AB", true},
		{"Too long", "TOOLONG", true},
		{"Lowercase", "nabil", true},
		{"Numbers", "NAB1L", true},
		{"Special chars", "NAB-L", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSymbol(tt.symbol)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSymbol() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNEPSEValidator_ValidateLotSize(t *testing.T) {
	validator := NewNEPSEValidator()

	tests := []struct {
		name     string
		symbol   string
		quantity int64
		wantErr  bool
	}{
		{"Bank stock valid lot", "NABIL", 10, false},
		{"Bank stock multiple lots", "NABIL", 50, false},
		{"Bank stock invalid lot", "NABIL", 15, true},
		{"Insurance stock valid lot", "NICA", 100, false},
		{"Insurance stock invalid lot", "NICA", 150, true},
		{"Zero quantity", "NABIL", 0, true},
		{"Negative quantity", "NABIL", -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateLotSize(tt.symbol, tt.quantity)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLotSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNEPSEValidator_ValidatePriceLimit(t *testing.T) {
	validator := NewNEPSEValidator()

	lastPrice := models.NewMoney(1000) // Rs. 1000

	tests := []struct {
		name      string
		price     models.Money
		lastPrice models.Money
		wantErr   bool
	}{
		{"Within limits", models.NewMoney(1050), lastPrice, false},
		{"At upper limit", models.NewMoney(1100), lastPrice, false},
		{"At lower limit", models.NewMoney(900), lastPrice, false},
		{"Above upper limit", models.NewMoney(1101), lastPrice, true},
		{"Below lower limit", models.NewMoney(899), lastPrice, true},
		{"No last price", models.NewMoney(1200), models.Zero(), false}, // Should pass
		{"Zero price", models.Zero(), lastPrice, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePriceLimit("NABIL", tt.price, tt.lastPrice)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePriceLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNEPSEValidator_ValidateTradingHours(t *testing.T) {
	validator := NewNEPSEValidator()

	// Use a past Wednesday at 13:00 (valid trading time)
	now := time.Now()
	pastDay := getPastWeekday(now, time.Wednesday)
	validTime := time.Date(pastDay.Year(), pastDay.Month(), pastDay.Day(), 13, 0, 0, 0, pastDay.Location())

	tests := []struct {
		name    string
		time    time.Time
		wantErr bool
	}{
		{"Valid trading time", validTime, false},
		{"Before trading hours", time.Date(pastDay.Year(), pastDay.Month(), pastDay.Day(), 10, 0, 0, 0, pastDay.Location()), true},
		{"After trading hours", time.Date(pastDay.Year(), pastDay.Month(), pastDay.Day(), 16, 0, 0, 0, pastDay.Location()), true},
		{"Friday (non-trading)", getPastWeekday(now, time.Friday).Add(13 * time.Hour), true},
		{"Saturday (non-trading)", getPastWeekday(now, time.Saturday).Add(13 * time.Hour), true},
		{"Future date", time.Now().Add(24 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTradingHours(tt.time)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTradingHours() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNEPSEValidator_ValidateMinimumTransaction(t *testing.T) {
	validator := NewNEPSEValidator()

	tests := []struct {
		name     string
		quantity int64
		price    models.Money
		wantErr  bool
	}{
		{"Above minimum", 10, models.NewMoney(100), false}, // Rs. 1000 total
		{"At minimum", 5, models.NewMoney(100), false},     // Rs. 500 total
		{"Below minimum", 1, models.NewMoney(100), true},   // Rs. 100 total
		{"Zero value", 0, models.NewMoney(1000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMinimumTransaction(tt.quantity, tt.price)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMinimumTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to get the next occurrence of a specific weekday
func getNextWeekday(from time.Time, weekday time.Weekday) time.Time {
	daysUntilWeekday := int(weekday) - int(from.Weekday())
	if daysUntilWeekday <= 0 {
		daysUntilWeekday += 7
	}
	return from.AddDate(0, 0, daysUntilWeekday)
}

// Helper function to get the past occurrence of a specific weekday
func getPastWeekday(from time.Time, weekday time.Weekday) time.Time {
	daysSinceWeekday := int(from.Weekday()) - int(weekday)
	if daysSinceWeekday <= 0 {
		daysSinceWeekday += 7
	}
	return from.AddDate(0, 0, -daysSinceWeekday)
}