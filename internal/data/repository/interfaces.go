package repository

import (
	"context"
	"ntx/internal/database"
	"time"
)

// PortfolioRepository defines operations for portfolio management
type PortfolioRepository interface {
	Create(ctx context.Context, req CreatePortfolioRequest) (*database.Portfolios, error)
	GetByID(ctx context.Context, id int64) (*database.Portfolios, error)
	List(ctx context.Context) ([]database.Portfolios, error)
	Update(ctx context.Context, req UpdatePortfolioRequest) (*database.Portfolios, error)
	Delete(ctx context.Context, id int64) error
	GetStats(ctx context.Context, id int64) (*database.GetPortfolioStatsRow, error)
}

// HoldingRepository defines operations for holdings management
type HoldingRepository interface {
	Create(ctx context.Context, req CreateHoldingRequest) (*database.Holdings, error)
	GetByID(ctx context.Context, id int64) (*database.Holdings, error)
	GetBySymbol(ctx context.Context, portfolioID int64, symbol string) (*database.Holdings, error)
	ListByPortfolio(ctx context.Context, portfolioID int64) ([]database.Holdings, error)
	ListWithValue(ctx context.Context, portfolioID int64) ([]database.ListHoldingsWithValueRow, error)
	Update(ctx context.Context, req UpdateHoldingRequest) (*database.Holdings, error)
	UpdatePrice(ctx context.Context, portfolioID int64, symbol string, pricePaisa int64) error
	Delete(ctx context.Context, id int64) error
	GetValue(ctx context.Context, id int64) (*database.GetHoldingValueRow, error)
}

// TransactionRepository defines operations for transaction management
type TransactionRepository interface {
	Create(ctx context.Context, req CreateTransactionRequest) (*database.Transactions, error)
	GetByID(ctx context.Context, id int64) (*database.Transactions, error)
	ListByPortfolio(ctx context.Context, portfolioID int64) ([]database.Transactions, error)
	ListBySymbol(ctx context.Context, portfolioID int64, symbol string) ([]database.Transactions, error)
	ListByDateRange(ctx context.Context, req ListTransactionsByDateRangeRequest) ([]database.Transactions, error)
	Update(ctx context.Context, req UpdateTransactionRequest) (*database.Transactions, error)
	Delete(ctx context.Context, id int64) error
	GetSummary(ctx context.Context, portfolioID int64, symbol string) (*database.GetTransactionSummaryRow, error)
	GetPortfolioStats(ctx context.Context, portfolioID int64) (*database.GetPortfolioTransactionStatsRow, error)
}

// CorporateActionRepository defines operations for corporate actions management
type CorporateActionRepository interface {
	Create(ctx context.Context, req CreateCorporateActionRequest) (*database.CorporateActions, error)
	GetByID(ctx context.Context, id int64) (*database.CorporateActions, error)
	List(ctx context.Context) ([]database.CorporateActions, error)
	ListBySymbol(ctx context.Context, symbol string) ([]database.CorporateActions, error)
	ListByType(ctx context.Context, actionType string) ([]database.CorporateActions, error)
	ListByDateRange(ctx context.Context, req ListCorporateActionsByDateRangeRequest) ([]database.CorporateActions, error)
	Update(ctx context.Context, req UpdateCorporateActionRequest) (*database.CorporateActions, error)
	Delete(ctx context.Context, id int64) error
	GetPending(ctx context.Context) ([]database.CorporateActions, error)
	GetBySymbolAndDate(ctx context.Context, symbol string, date time.Time) ([]database.CorporateActions, error)
}

// Repository aggregates all repository interfaces
type Repository struct {
	Portfolio       PortfolioRepository
	Holding         HoldingRepository
	Transaction     TransactionRepository
	CorporateAction CorporateActionRepository
}

// Transactor defines transaction operations
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, repo *Repository) error) error
}