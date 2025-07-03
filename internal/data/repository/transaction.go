package repository

import (
	"context"
	"fmt"
	"ntx/internal/database"
)

type transactionRepository struct {
	queries *database.Queries
}

func NewTransactionRepository(queries *database.Queries) TransactionRepository {
	return &transactionRepository{
		queries: queries,
	}
}

func (r *transactionRepository) Create(ctx context.Context, req CreateTransactionRequest) (*database.Transactions, error) {
	params := database.CreateTransactionParams{
		PortfolioID:      req.PortfolioID,
		Symbol:          req.Symbol,
		TransactionType: req.TransactionType,
		Quantity:        req.Quantity,
		PricePaisa:      req.PricePaisa,
		CommissionPaisa: nullInt64FromPtr(&req.CommissionPaisa),
		TaxPaisa:        nullInt64FromPtr(&req.TaxPaisa),
		TransactionDate: req.TransactionDate,
		Notes:           nullStringFromPtr(req.Notes),
	}

	transaction, err := r.queries.CreateTransaction(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return &transaction, nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id int64) (*database.Transactions, error) {
	transaction, err := r.queries.GetTransaction(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

func (r *transactionRepository) ListByPortfolio(ctx context.Context, portfolioID int64) ([]database.Transactions, error) {
	transactions, err := r.queries.ListTransactionsByPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by portfolio: %w", err)
	}

	return transactions, nil
}

func (r *transactionRepository) ListBySymbol(ctx context.Context, portfolioID int64, symbol string) ([]database.Transactions, error) {
	params := database.ListTransactionsBySymbolParams{
		PortfolioID: portfolioID,
		Symbol:     symbol,
	}

	transactions, err := r.queries.ListTransactionsBySymbol(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by symbol: %w", err)
	}

	return transactions, nil
}

func (r *transactionRepository) ListByDateRange(ctx context.Context, req ListTransactionsByDateRangeRequest) ([]database.Transactions, error) {
	params := database.ListTransactionsByDateRangeParams{
		PortfolioID: req.PortfolioID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}
	
	transactions, err := r.queries.ListTransactionsByDateRange(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by date range: %w", err)
	}
	
	return transactions, nil
}

func (r *transactionRepository) Update(ctx context.Context, req UpdateTransactionRequest) (*database.Transactions, error) {
	params := database.UpdateTransactionParams{
		ID:              req.ID,
		Symbol:          req.Symbol,
		TransactionType: req.TransactionType,
		Quantity:        req.Quantity,
		PricePaisa:      req.PricePaisa,
		CommissionPaisa: nullInt64FromPtr(&req.CommissionPaisa),
		TaxPaisa:        nullInt64FromPtr(&req.TaxPaisa),
		TransactionDate: req.TransactionDate,
		Notes:           nullStringFromPtr(req.Notes),
	}

	transaction, err := r.queries.UpdateTransaction(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	return &transaction, nil
}

func (r *transactionRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeleteTransaction(ctx, id); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}

func (r *transactionRepository) GetSummary(ctx context.Context, portfolioID int64, symbol string) (*database.GetTransactionSummaryRow, error) {
	params := database.GetTransactionSummaryParams{
		PortfolioID: portfolioID,
		Symbol:     symbol,
	}

	summary, err := r.queries.GetTransactionSummary(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction summary: %w", err)
	}

	return &summary, nil
}

func (r *transactionRepository) GetPortfolioStats(ctx context.Context, portfolioID int64) (*database.GetPortfolioTransactionStatsRow, error) {
	stats, err := r.queries.GetPortfolioTransactionStats(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio transaction stats: %w", err)
	}

	return &stats, nil
}