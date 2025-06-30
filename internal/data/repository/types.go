package repository

import (
	"database/sql"
	"time"
)

// Portfolio request types
type CreatePortfolioRequest struct {
	Name        string
	Description *string
	Currency    string
}

type UpdatePortfolioRequest struct {
	ID          int64
	Name        string
	Description *string
}

// Holding request types
type CreateHoldingRequest struct {
	PortfolioID       int64
	Symbol           string
	Quantity         int64
	AverageCostPaisa int64
	LastPricePaisa   *int64
}

type UpdateHoldingRequest struct {
	ID               int64
	Quantity         int64
	AverageCostPaisa int64
	LastPricePaisa   *int64
}

// Transaction request types
type CreateTransactionRequest struct {
	PortfolioID      int64
	Symbol          string
	TransactionType string // "buy" or "sell"
	Quantity        int64
	PricePaisa      int64
	CommissionPaisa int64
	TaxPaisa        int64
	TransactionDate time.Time
	Notes           *string
}

type UpdateTransactionRequest struct {
	ID              int64
	Symbol          string
	TransactionType string
	Quantity        int64
	PricePaisa      int64
	CommissionPaisa int64
	TaxPaisa        int64
	TransactionDate time.Time
	Notes           *string
}

type ListTransactionsByDateRangeRequest struct {
	PortfolioID int64
	StartDate   time.Time
	EndDate     time.Time
}

// Corporate Action request types
type CreateCorporateActionRequest struct {
	Symbol           string
	ActionType       string // "bonus", "dividend", "split", "rights"
	AnnouncementDate time.Time
	RecordDate       time.Time
	ExecutionDate    *time.Time
	RatioFrom        *int64
	RatioTo          *int64
	AmountPaisa      *int64
	Notes            *string
}

type UpdateCorporateActionRequest struct {
	ID               int64
	Symbol           string
	ActionType       string
	AnnouncementDate time.Time
	RecordDate       time.Time
	ExecutionDate    *time.Time
	RatioFrom        *int64
	RatioTo          *int64
	AmountPaisa      *int64
	Notes            *string
}

type ListCorporateActionsByDateRangeRequest struct {
	StartDate time.Time
	EndDate   time.Time
}

// Helper functions to convert between domain types and SQL types
func nullStringFromPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func ptrFromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullInt64FromPtr(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func ptrFromNullInt64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

func nullTimeFromPtr(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func ptrFromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}