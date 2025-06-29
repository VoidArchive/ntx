package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CorporateAction represents a corporate action announcement for NEPSE stocks
type CorporateAction struct {
	ID               int64                 `json:"id" db:"id"`
	Symbol           string                `json:"symbol" db:"symbol"`
	ActionType       CorporateActionType   `json:"action_type" db:"action_type"`
	AnnouncementDate time.Time             `json:"announcement_date" db:"announcement_date"`
	RecordDate       time.Time             `json:"record_date" db:"record_date"`
	ExDate           time.Time             `json:"ex_date" db:"ex_date"`
	Ratio            string                `json:"ratio" db:"ratio"`
	DividendAmount   Money                 `json:"dividend_amount" db:"dividend_amount"`
	Processed        bool                  `json:"processed" db:"processed"`
	ProcessedDate    time.Time             `json:"processed_date" db:"processed_date"`
	Notes            string                `json:"notes" db:"notes"`
	CreatedAt        time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at" db:"updated_at"`
}

// CorporateActionType represents the type of corporate action
type CorporateActionType string

const (
	CorporateActionTypeBonus    CorporateActionType = "bonus"
	CorporateActionTypeDividend CorporateActionType = "dividend"
	CorporateActionTypeRights   CorporateActionType = "rights"
	CorporateActionTypeSplit    CorporateActionType = "split"
)

// CorporateActionRequest represents a request to create/update a corporate action
type CorporateActionRequest struct {
	Symbol           string              `json:"symbol" validate:"required,min=2,max=10"`
	ActionType       CorporateActionType `json:"action_type" validate:"required"`
	AnnouncementDate time.Time           `json:"announcement_date" validate:"required"`
	RecordDate       time.Time           `json:"record_date" validate:"required"`
	ExDate           time.Time           `json:"ex_date" validate:"required"`
	Ratio            string              `json:"ratio,omitempty" validate:"max=20"`
	DividendAmount   Money               `json:"dividend_amount,omitempty"`
	Notes            string              `json:"notes,omitempty" validate:"max=500"`
}

// CorporateActionSummary provides summarized corporate action information
type CorporateActionSummary struct {
	Symbol              string              `json:"symbol"`
	ActionType          CorporateActionType `json:"action_type"`
	ExDate              time.Time           `json:"ex_date"`
	Description         string              `json:"description"`
	Status              string              `json:"status"`
	RequiresProcessing  bool                `json:"requires_processing"`
	ImpactDescription   string              `json:"impact_description"`
	DaysUntilExDate     int                 `json:"days_until_ex_date"`
}

// BonusShareCalculation represents the calculation for bonus share distribution
type BonusShareCalculation struct {
	Symbol           string    `json:"symbol"`
	CurrentHolding   Quantity  `json:"current_holding"`
	BonusRatio       string    `json:"bonus_ratio"`
	BonusShares      Quantity  `json:"bonus_shares"`
	NewTotalShares   Quantity  `json:"new_total_shares"`
	NewAverageCost   Money     `json:"new_average_cost"`
}

// DividendCalculation represents the calculation for dividend distribution
type DividendCalculation struct {
	Symbol         string   `json:"symbol"`
	CurrentHolding Quantity `json:"current_holding"`
	DividendRate   Money    `json:"dividend_rate"`
	TotalDividend  Money    `json:"total_dividend"`
	TaxDeduction   Money    `json:"tax_deduction,omitempty"`
	NetDividend    Money    `json:"net_dividend"`
}

// RightsShareCalculation represents the calculation for rights share offering
type RightsShareCalculation struct {
	Symbol           string   `json:"symbol"`
	CurrentHolding   Quantity `json:"current_holding"`
	RightsRatio      string   `json:"rights_ratio"`
	RightsEntitled   Quantity `json:"rights_entitled"`
	RightsPrice      Money    `json:"rights_price"`
	TotalInvestment  Money    `json:"total_investment"`
	NewTotalShares   Quantity `json:"new_total_shares"`
	NewAverageCost   Money    `json:"new_average_cost"`
}

// Methods for CorporateAction

// IsValid validates the corporate action data
func (ca *CorporateAction) IsValid() bool {
	return ca.Symbol != "" &&
		len(ca.Symbol) >= 2 && len(ca.Symbol) <= 10 &&
		ca.ActionType != "" &&
		isValidCorporateActionType(ca.ActionType) &&
		!ca.AnnouncementDate.IsZero() &&
		!ca.RecordDate.IsZero() &&
		!ca.ExDate.IsZero() &&
		ca.ExDate.After(ca.AnnouncementDate) &&
		len(ca.Notes) <= 500
}

// GetDescription returns a human-readable description of the corporate action
func (ca *CorporateAction) GetDescription() string {
	switch ca.ActionType {
	case CorporateActionTypeBonus:
		if ca.Ratio != "" {
			return fmt.Sprintf("Bonus shares %s", ca.Ratio)
		}
		return "Bonus shares"
	case CorporateActionTypeDividend:
		if !ca.DividendAmount.IsZero() {
			return fmt.Sprintf("Cash dividend Rs. %s per share", ca.DividendAmount.String())
		}
		return "Cash dividend"
	case CorporateActionTypeRights:
		if ca.Ratio != "" {
			return fmt.Sprintf("Rights shares %s", ca.Ratio)
		}
		return "Rights shares"
	case CorporateActionTypeSplit:
		if ca.Ratio != "" {
			return fmt.Sprintf("Stock split %s", ca.Ratio)
		}
		return "Stock split"
	default:
		return string(ca.ActionType)
	}
}

// GetStatus returns the current status of the corporate action
func (ca *CorporateAction) GetStatus() string {
	now := time.Now()
	
	if ca.Processed {
		return "Processed"
	}
	
	if now.Before(ca.ExDate) {
		return "Pending"
	}
	
	if now.After(ca.ExDate) && !ca.Processed {
		return "Due for Processing"
	}
	
	return "Active"
}

// RequiresProcessing checks if the corporate action requires manual processing
func (ca *CorporateAction) RequiresProcessing() bool {
	now := time.Now()
	return !ca.Processed && now.After(ca.ExDate)
}

// DaysUntilExDate calculates days until ex-dividend date
func (ca *CorporateAction) DaysUntilExDate() int {
	now := time.Now()
	if now.After(ca.ExDate) {
		return 0
	}
	return int(ca.ExDate.Sub(now).Hours() / 24)
}

// CalculateBonusShares calculates bonus share distribution for a holding
func (ca *CorporateAction) CalculateBonusShares(currentHolding Quantity, currentAvgCost Money) (*BonusShareCalculation, error) {
	if ca.ActionType != CorporateActionTypeBonus {
		return nil, fmt.Errorf("not a bonus share action")
	}

	bonusShares, err := calculateBonusFromRatio(ca.Ratio, currentHolding)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate bonus shares: %w", err)
	}

	newTotalShares := currentHolding.Add(bonusShares)
	
	// New average cost = (old total cost) / (new total shares)
	totalCost := currentAvgCost.MultiplyByQuantity(currentHolding)
	newAvgCost := totalCost.DivideByQuantity(newTotalShares)

	return &BonusShareCalculation{
		Symbol:         ca.Symbol,
		CurrentHolding: currentHolding,
		BonusRatio:     ca.Ratio,
		BonusShares:    bonusShares,
		NewTotalShares: newTotalShares,
		NewAverageCost: newAvgCost,
	}, nil
}

// CalculateDividend calculates dividend amount for a holding
func (ca *CorporateAction) CalculateDividend(currentHolding Quantity) (*DividendCalculation, error) {
	if ca.ActionType != CorporateActionTypeDividend {
		return nil, fmt.Errorf("not a dividend action")
	}

	if ca.DividendAmount.IsZero() {
		return nil, fmt.Errorf("dividend amount not specified")
	}

	totalDividend := ca.DividendAmount.MultiplyByQuantity(currentHolding)
	
	// In Nepal, dividend tax is typically 5% for individuals
	taxRate := NewPercentageFromFloat(5.0)
	taxDeduction := ApplyPercentageToMoney(totalDividend, taxRate)
	netDividend := totalDividend.Subtract(taxDeduction)

	return &DividendCalculation{
		Symbol:         ca.Symbol,
		CurrentHolding: currentHolding,
		DividendRate:   ca.DividendAmount,
		TotalDividend:  totalDividend,
		TaxDeduction:   taxDeduction,
		NetDividend:    netDividend,
	}, nil
}

// CalculateRightsShares calculates rights share offering for a holding
func (ca *CorporateAction) CalculateRightsShares(currentHolding Quantity, currentAvgCost Money, rightsPrice Money) (*RightsShareCalculation, error) {
	if ca.ActionType != CorporateActionTypeRights {
		return nil, fmt.Errorf("not a rights share action")
	}

	rightsEntitled, err := calculateBonusFromRatio(ca.Ratio, currentHolding)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate rights entitlement: %w", err)
	}

	totalInvestment := rightsPrice.MultiplyByQuantity(rightsEntitled)
	newTotalShares := currentHolding.Add(rightsEntitled)
	
	// New average cost = (old total cost + rights investment) / (new total shares)
	oldTotalCost := currentAvgCost.MultiplyByQuantity(currentHolding)
	newTotalCost := oldTotalCost.Add(totalInvestment)
	newAvgCost := newTotalCost.DivideByQuantity(newTotalShares)

	return &RightsShareCalculation{
		Symbol:          ca.Symbol,
		CurrentHolding:  currentHolding,
		RightsRatio:     ca.Ratio,
		RightsEntitled:  rightsEntitled,
		RightsPrice:     rightsPrice,
		TotalInvestment: totalInvestment,
		NewTotalShares:  newTotalShares,
		NewAverageCost:  newAvgCost,
	}, nil
}

// ToSummary converts corporate action to summary format
func (ca *CorporateAction) ToSummary() CorporateActionSummary {
	return CorporateActionSummary{
		Symbol:              ca.Symbol,
		ActionType:          ca.ActionType,
		ExDate:              ca.ExDate,
		Description:         ca.GetDescription(),
		Status:              ca.GetStatus(),
		RequiresProcessing:  ca.RequiresProcessing(),
		ImpactDescription:   ca.getImpactDescription(),
		DaysUntilExDate:     ca.DaysUntilExDate(),
	}
}

// getImpactDescription returns a description of the impact on portfolio
func (ca *CorporateAction) getImpactDescription() string {
	switch ca.ActionType {
	case CorporateActionTypeBonus:
		return "Increases share quantity, reduces average cost per share"
	case CorporateActionTypeDividend:
		return "Provides cash return, no impact on share quantity"
	case CorporateActionTypeRights:
		return "Opportunity to buy additional shares at discounted price"
	case CorporateActionTypeSplit:
		return "Increases share quantity proportionally, reduces price per share"
	default:
		return "Impact varies based on action type"
	}
}

// Methods for CorporateActionRequest

// ToCorporateAction converts request to corporate action
func (car *CorporateActionRequest) ToCorporateAction() CorporateAction {
	now := time.Now()
	return CorporateAction{
		Symbol:           car.Symbol,
		ActionType:       car.ActionType,
		AnnouncementDate: car.AnnouncementDate,
		RecordDate:       car.RecordDate,
		ExDate:           car.ExDate,
		Ratio:            car.Ratio,
		DividendAmount:   car.DividendAmount,
		Notes:            car.Notes,
		Processed:        false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// IsValid validates the corporate action request
func (car *CorporateActionRequest) IsValid() bool {
	return car.Symbol != "" &&
		len(car.Symbol) >= 2 && len(car.Symbol) <= 10 &&
		car.ActionType != "" &&
		isValidCorporateActionType(car.ActionType) &&
		!car.AnnouncementDate.IsZero() &&
		!car.RecordDate.IsZero() &&
		!car.ExDate.IsZero() &&
		car.ExDate.After(car.AnnouncementDate) &&
		len(car.Ratio) <= 20 &&
		len(car.Notes) <= 500
}

// Helper functions

// isValidCorporateActionType checks if the action type is valid
func isValidCorporateActionType(actionType CorporateActionType) bool {
	switch actionType {
	case CorporateActionTypeBonus,
		CorporateActionTypeDividend,
		CorporateActionTypeRights,
		CorporateActionTypeSplit:
		return true
	default:
		return false
	}
}

// calculateBonusFromRatio calculates bonus/rights shares from ratio string
// Supports formats like "1:5" (1 bonus for every 5 held), "20%" (20% bonus)
func calculateBonusFromRatio(ratio string, currentHolding Quantity) (Quantity, error) {
	if ratio == "" {
		return Quantity(0), fmt.Errorf("ratio not specified")
	}

	ratio = strings.TrimSpace(ratio)

	// Handle percentage format (e.g., "20%")
	if strings.HasSuffix(ratio, "%") {
		percentStr := strings.TrimSuffix(ratio, "%")
		percent, err := strconv.ParseFloat(percentStr, 64)
		if err != nil {
			return Quantity(0), fmt.Errorf("invalid percentage format: %s", ratio)
		}
		
		bonusFloat := float64(currentHolding.Int64()) * (percent / 100.0)
		return Quantity(int64(bonusFloat)), nil
	}

	// Handle ratio format (e.g., "1:5")
	if strings.Contains(ratio, ":") {
		parts := strings.Split(ratio, ":")
		if len(parts) != 2 {
			return Quantity(0), fmt.Errorf("invalid ratio format: %s", ratio)
		}

		bonus, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			return Quantity(0), fmt.Errorf("invalid bonus part in ratio: %s", parts[0])
		}

		held, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return Quantity(0), fmt.Errorf("invalid held part in ratio: %s", parts[1])
		}

		if held == 0 {
			return Quantity(0), fmt.Errorf("division by zero in ratio")
		}

		bonusShares := (currentHolding.Int64() * bonus) / held
		return Quantity(bonusShares), nil
	}

	return Quantity(0), fmt.Errorf("unsupported ratio format: %s", ratio)
}

// FilterCorporateActions filters corporate actions by various criteria
func FilterCorporateActions(actions []CorporateAction, filters CorporateActionFilters) []CorporateAction {
	var filtered []CorporateAction

	for _, action := range actions {
		if shouldIncludeAction(action, filters) {
			filtered = append(filtered, action)
		}
	}

	return filtered
}

// CorporateActionFilters represents filtering criteria for corporate actions
type CorporateActionFilters struct {
	Symbol           string              `json:"symbol,omitempty"`
	ActionType       CorporateActionType `json:"action_type,omitempty"`
	ProcessedOnly    bool                `json:"processed_only,omitempty"`
	UnprocessedOnly  bool                `json:"unprocessed_only,omitempty"`
	RequiresAction   bool                `json:"requires_action,omitempty"`
	FromDate         time.Time           `json:"from_date,omitempty"`
	ToDate           time.Time           `json:"to_date,omitempty"`
}

// shouldIncludeAction checks if an action matches the filters
func shouldIncludeAction(action CorporateAction, filters CorporateActionFilters) bool {
	if filters.Symbol != "" && !strings.EqualFold(action.Symbol, filters.Symbol) {
		return false
	}

	if filters.ActionType != "" && action.ActionType != filters.ActionType {
		return false
	}

	if filters.ProcessedOnly && !action.Processed {
		return false
	}

	if filters.UnprocessedOnly && action.Processed {
		return false
	}

	if filters.RequiresAction && !action.RequiresProcessing() {
		return false
	}

	if !filters.FromDate.IsZero() && action.ExDate.Before(filters.FromDate) {
		return false
	}

	if !filters.ToDate.IsZero() && action.ExDate.After(filters.ToDate) {
		return false
	}

	return true
}